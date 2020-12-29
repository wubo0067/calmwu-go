/*
 * @Author: calmwu
 * @Date: 2018-12-08 16:06:26
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-12-10 11:47:54
 */

package utils

import (
	"fmt"
	"sync"
	"time"
)

// WorkerHandler 回调函数类型定义
type WorkerHandler func(interface{}) error

// WorkerPool 协程池对象
type WorkerPool struct {
	workerHandler WorkerHandler

	maxWorkersCount int

	maxIdleWorkerDuration time.Duration

	lock         sync.Mutex
	workersCount int
	mustStop     bool

	ready []*workerChan

	stopCh chan struct{}

	workerChanPool sync.Pool
}

// 外界和routine沟通的通道，传入参数或者终止符
type workerChan struct {
	lastUseTime time.Time
	ch          chan interface{}
}

// DefaultConcurrency 默认的并发数量
const DefaultConcurrency = 256 * 1024

// StartWorkerPool 启动协程池
func StartWorkerPool(workerFunc WorkerHandler, maxWorkersCount int, maxIdelWorkerDuration time.Duration) (*WorkerPool, error) {
	wp := &WorkerPool{
		workerHandler:         workerFunc,
		maxWorkersCount:       maxWorkersCount,
		maxIdleWorkerDuration: maxIdelWorkerDuration,
	}

	err := wp.start()

	return wp, err
}

func (wp *WorkerPool) start() error {
	if wp.stopCh != nil {
		err := fmt.Errorf("WorkerPool already started!")
		ZLog.Error(err.Error())
		return err
	}
	wp.stopCh = make(chan struct{})

	go func() {
		var scratch []*workerChan
		for {
			wp.clean(&scratch)
			select {
			case <-wp.stopCh:
				Info("WorkPool clean routine exit!")
				return
			default:
				time.Sleep(wp.getMaxIdleWorkerDuration())
			}
		}
	}()

	return nil
}

// Stop 停止协程池
func (wp *WorkerPool) Stop() {
	if wp.stopCh == nil {
		Error("WorkerPool wasn't started")
		return
	}

	close(wp.stopCh)
	wp.stopCh = nil

	wp.lock.Lock()
	ready := wp.ready

	// 清晰的标识出释放的对象
	Infof("WorkerPool stop %d workerChan", len(ready))
	for i, ch := range ready {
		// 给所有工作中的routine退出，close(ch.ch)
		ch.ch <- nil
		ready[i] = nil
	}
	wp.ready = ready[:0]
	wp.mustStop = true
	wp.lock.Unlock()
}

func (wp *WorkerPool) getMaxIdleWorkerDuration() time.Duration {
	if wp.maxIdleWorkerDuration <= 0 {
		return 10 * time.Second
	}
	return wp.maxIdleWorkerDuration
}

// Serve 调用
func (wp *WorkerPool) Serve(arg interface{}) bool {
	wch := wp.getWorkerChan()
	if wch == nil {
		Error("WorkerPool get workerChan failed!")
		return false
	}
	wch.ch <- arg
	return true
}

func (wp *WorkerPool) getWorkerChan() *workerChan {
	var wch *workerChan
	createWorker := false

	wp.lock.Lock()
	ready := wp.ready
	// nil对象也可以用len
	n := len(ready) - 1
	if n < 0 {
		// 没有空闲的routine
		if wp.workersCount < wp.maxWorkersCount {
			// 判断是否可以创建新的
			createWorker = true
			wp.workersCount++
		} else {
			Warnf("WorkerPool wokersCount reach limit:%d", wp.workersCount)
		}
	} else {
		// 从末尾取
		wch = ready[n]
		ready[n] = nil
		wp.ready = ready[:n]
	}
	wp.lock.Unlock()

	if wch == nil {
		if !createWorker {
			// 达到上限无法创建
			return nil
		}
		// workerChan对象从pool中创建
		vch := wp.workerChanPool.Get()
		if vch == nil {
			vch = &workerChan{
				ch: make(chan interface{}, 1),
			}
		}
		wch = vch.(*workerChan)
		// 启动routine
		go func() {
			// 传入通道
			wp.workerFunc(wch)
			// 完毕后回收
			wp.workerChanPool.Put(vch)
		}()
	}
	return wch
}

func (wp *WorkerPool) workerFunc(wch *workerChan) {
	var err error
	// 读取通道数据
	for c := range wch.ch {
		if c == nil {
			Info("workerFunc receive exit notify!")
			break
		}

		if err = wp.workerHandler(c); err != nil {
			Errorf("workerHandler error:%s", err.Error())
		}

		if !wp.release(wch) {
			break
		}
	}

	// 退出
	wp.lock.Lock()
	// 具体的工作routine数量递减
	wp.workersCount--
	wp.lock.Unlock()
}

func (wp *WorkerPool) release(wch *workerChan) bool {
	// 标记最后使用时间
	wch.lastUseTime = time.Now()
	wp.lock.Lock()
	if wp.mustStop {
		Info("WorkerPool worker must stop")
		wp.lock.Unlock()
		return false
	}
	// 加入空闲ready队列尾部
	wp.ready = append(wp.ready, wch)
	wp.lock.Unlock()
	return true
}

// 回收空闲的wokerChan
func (wp *WorkerPool) clean(scratch *[]*workerChan) {
	maxIdleWorkerDuration := wp.getMaxIdleWorkerDuration()

	currentTime := time.Now()

	wp.lock.Lock()
	// 查看空闲列表
	ready := wp.ready
	n := len(ready)
	i := 0
	// 从头查找所有超时的对象，尾部是最新的，前面是idle很久的
	for i < n && currentTime.Sub(ready[i].lastUseTime) > maxIdleWorkerDuration {
		i++
	}
	// 将清理对象插入scratch中
	*scratch = append((*scratch)[:0], ready[:i]...)
	if i > 0 {
		// 排干
		m := copy(ready, ready[i:])
		for i = m; i < n; i++ {
			ready[i] = nil
		}
		wp.ready = ready[:m]
	}
	wp.lock.Unlock()

	tmp := *scratch
	Debugf("WorkerPool release %d wokerChan", len(tmp))
	for i, ch := range tmp {
		// 通知routine结束
		ch.ch <- nil
		tmp[i] = nil
	}
}
