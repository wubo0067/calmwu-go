/*
 * @Author: calm.wu
 * @Date: 2019-03-22 15:36:53
 * @Last Modified by: calm.wu
 * @Last Modified time: 2019-03-22 16:46:03
 */

package utils

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"time"
)

var ForeverTestTimeout = time.Second * 30
var NerverStop <-chan struct{} = make(chan struct{})

type Group struct {
	wg sync.WaitGroup
}

func (g *Group) Wait() {
	g.wg.Wait()
}

func (g *Group) StartWithChannel(stopCh <-chan struct{}, f func(stopCh <-chan struct{})) {
	g.Start(func() {
		f(stopCh)
	})
}

func (g *Group) StartWithContext(ctx context.Context, f func(ctx context.Context)) {
	g.Start(func() {
		f(ctx)
	})
}

func (g *Group) Start(f func()) {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		f()
	}()
}

// 按周期定时调用
func Forever(f func(), period time.Duration) {
	Until(f, period, NerverStop)
}

// 一直周期调用，直到stop chan被close
func Until(f func(), period time.Duration, stopCh <-chan struct{}) {
	JitterUntil(f, period, 0.0, true, stopCh)
}

// 在执行后开始计算超时，超时时间不包括函数执行时间
func UntilWithContext(ctx context.Context, f func(context.Context), period time.Duration) {
	JitterUntilWithContext(ctx, f, period, 0.0, true)
}

// 在执行后开始计算超时
func JitterUntilWithContext(ctx context.Context, f func(context.Context), period time.Duration, jitterFactor float64, sliding bool) {
	// 这个func wrapper和ctx.Done()真是赞，学到了，前者统一为一个方法，闭包好啊
	JitterUntil(func() { f(ctx) }, period, jitterFactor, sliding, ctx.Done())
}

// 在执行前开始计算超时
func NonSlidingUntil(f func(), period time.Duration, stopCh <-chan struct{}) {
	JitterUntil(f, period, 0.0, false, stopCh)
}

// 在执行前开始计算超时，超时时间包括了函数执行时间
func NonSlidingUntilWithContext(ctx context.Context, f func(context.Context), period time.Duration) {
	JitterUntilWithContext(ctx, f, period, 0.0, false)
}

// If sliding is true, the period is computed after f runs. If it is false then
// period includes the runtime for f.
func JitterUntil(f func(), period time.Duration, jitterFactor float64, sliding bool, stopCh <-chan struct{}) {
	var t *time.Timer
	// 超时后正常读取
	var sawTimeout bool = false

	for {
		select {
		case <-stopCh:
			// receive close notify
			return
		default:
			// nonblock
		}

		// 计算周期时间
		jitteredPeriod := period
		if jitteredPeriod > 0.0 {
			jitteredPeriod = Jitter(period, jitterFactor)
		}

		if !sliding {
			t = resetOrReuseTimer(t, jitteredPeriod, sawTimeout)
		}

		// 执行
		func() {
			//TODO: defer recover
			f()
		}()

		if sliding {
			t = resetOrReuseTimer(t, jitteredPeriod, sawTimeout)
		}

		// 在这里等待
		select {
		case <-stopCh:
			return
		case <-t.C:
			sawTimeout = true
		}
	}
}

// Jitter returns a time.Duration between duration and duration + maxFactor *
// duration.
//
// This allows clients to avoid converging on periodic behavior. If maxFactor
// is 0.0, a suggested default value will be chosen.
func Jitter(duration time.Duration, maxFactor float64) time.Duration {
	if maxFactor <= 0.0 {
		maxFactor = 1.0
	}
	wait := duration + time.Duration(rand.Float64()*maxFactor*float64(duration))
	return wait
}

func resetOrReuseTimer(t *time.Timer, d time.Duration, sawTimeout bool) *time.Timer {
	if t == nil {
		return time.NewTimer(d)
	}

	if !t.Stop() && !sawTimeout {
		// timer超时但没有读取，手工排干
		<-t.C
	}
	// reuse
	t.Reset(d)
	return t
}

var ErrWaitTimeout = errors.New("timed out waiting for the condition")

// 定义函数类型
type ConditionFunc func() (done bool, err error)

type Backoff struct {
	Duration time.Duration
	Factor   float64
	Jitter   float64
	Steps    int
}



type WaitFunc func(done <-chan struct{}) <-chan struct{}

func Poll(interval, timeout time.Duration, condition ConditionFunc) error {
	return pollInternal(poller(interval, timeout), condition)
}

func pollInternal(wait WaitFunc, condition ConditionFunc) error {
	done := make(chan struct{})
	defer close(done)
	return WaitFor(wait, condition, done)
}

func PollInfinite(interval time.Duration, condition ConditionFunc) error {
	done := make(chan struct{})
	defer close(done)
	return WaitFor(poller(interval, 0), condition, done)
}

func WaitFor(wait WaitFunc, fn ConditionFunc, done <-chan struct{}) error {
	c := wait(done)
	for {
		// open判断c是否被关闭
		_, open := <-c
		// 执行条件函数
		ok, err := fn()
		if err != nil {
			return err
		}

		if ok {
			return nil
		}

		if !open {
			break
		}
	}
	return ErrWaitTimeout
}

// 定时间隔在channel发送一个信号
func poller(interval, timeout time.Duration) WaitFunc {
	return WaitFunc(func(done <-chan struct{}) <-chan struct{} {
		ch := make(chan struct{})

		go func() {
			// 函数执行完毕后关闭，通知外部
			defer close(ch)

			// 生成定时器
			tick := time.NewTicker(interval)
			defer tick.Stop()

			var after <-chan time.Time

			if timeout != 0 {
				timer := time.NewTimer(timeout)
				after = timer.C
				defer timer.Stop()
			}

			for {
				select {
				case <-tick.C:
					select {
					case ch <- struct{}{}:
					default:
					}
				case <-after:
					// 超时退出
					return
				case <-done:
					// 外部主动关闭
					return
				}
			}
		}()

		return ch
	})
}
