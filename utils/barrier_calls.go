/*
 * @Author: CALM.WU
 * @Date: 2021-02-05 21:10:37
 * @Last Modified by: CALM.WU
 * @Last Modified time: 2021-02-05 22:03:14
 */

// 场景：对缓存操作时，同一个key的调用，进行排队，防止缓存被击穿

package utils

import "sync"

type (
	// BarrierCalls 并行访问的接口
	BarrierCalls interface {
		Do(key string, args interface{}, fn func(args interface{}) (interface{}, error)) (interface{}, bool, error)
	}

	// BCall 每次请求生成的调用对象
	BCall struct {
		ch  chan struct{} // 执行同步对象
		val interface{}   // 执行的结果
		err error         // 执行的错误信息
	}

	barrierCallsGroup struct {
		calls map[string]*BCall
		mutex *sync.Mutex
	}
)

func NewBarrierCalls() BarrierCalls {
	return &barrierCallsGroup{
		calls: make(map[string]*BCall),
		mutex: &sync.Mutex{},
	}
}

// Do 执行调用，对相同的key，func串行调用
func (bcg *barrierCallsGroup) Do(key string, args interface{}, fn func(args interface{}) (interface{}, error)) (interface{}, bool, error) {
	bc, fresh := bcg.newBCall(key)
	if fresh {
		return bc.val, false, bc.err
	}

	bcg.doBCall(bc, key, args, fn)
	return bc.val, true, bc.err
}

func (bcg *barrierCallsGroup) newBCall(key string) (*BCall, bool) {
	bcg.mutex.Lock()
	if bc, exist := bcg.calls[key]; exist {
		bcg.mutex.Unlock()
		// 有相同的key操作，等待释放，直接使用缓存后的数据
		<-bc.ch
		return bc, true
	}

	bc := new(BCall)
	bc.ch = make(chan struct{})
	bcg.calls[key] = bc
	bcg.mutex.Unlock()
	return bc, false
}

func (bcg *barrierCallsGroup) doBCall(bc *BCall, key string, args interface{}, fn func(args interface{}) (interface{}, error)) {
	defer func() {
		bcg.mutex.Lock()
		delete(bcg.calls, key)
		// 广播通知所有等待的bcall
		close(bc.ch)
		bcg.mutex.Unlock()
	}()

	bc.val, bc.err = fn(args)
}
