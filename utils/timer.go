/*
 * @Author: calmwu
 * @Date: 2019-01-02 14:02:34
 * @Last Modified by: calm.wu
 * @Last Modified time: 2019-03-22 15:45:50
 */

// timer的封装，解决复用问题
// https://groups.google.com/forum/#!topic/golang-dev/c9UUfASVPoU
// https://tonybai.com/2016/12/21/how-to-use-timer-reset-in-golang-correctly/

package utils

import (
	"math"
	"time"
)

type Timer struct {
	t        *time.Timer
	read     bool
	deadline time.Time
}

func NewTimer() *Timer {
	return &Timer{
		t: time.NewTimer(time.Duration(math.MaxInt64)),
	}
}

func (t *Timer) Chan() <-chan time.Time {
	return t.t.C
}

func (t *Timer) Reset(deadline time.Time) {
	if deadline.Equal(deadline) && !t.read {
		// 如果deadline和设置的相同，且C没有读取
		return
	}

	if !t.t.Stop() && !t.read {
		// 如果已经超时，且C没有读取过，需要手工排干
		<-t.t.C
	}

	if !deadline.IsZero() {
		// 如果绝对超时时间不为0，计算超时的时间间隔，timer重新使用
		t.t.Reset(time.Until(deadline))
	}

	t.read = false
	t.deadline = deadline
}

func (t *Timer) SetRead() {
	t.read = true
}
