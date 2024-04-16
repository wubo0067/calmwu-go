/*
 * @Author: CALM.WU
 * @Date: 2024-04-16 14:34:20
 * @Last Modified by: CALM.WU
 * @Last Modified time: 2024-04-16 14:46:46
 */

package utils

import (
	"errors"
	"time"
)

var ErrSemTimeout = errors.New("semaphore: acquire timeout")

type Semaphore chan struct{}

func NewSemaphore(value int) Semaphore {
	return make(chan struct{}, value)
}

// Acquire acquires a semaphore with an optional timeout.
// If a timeout is specified and the semaphore cannot be acquired within the given duration,
// it returns an error of type `ErrSemTimeout`.
func (s Semaphore) Acquire(timeout time.Duration) error {
	if timeout > time.Duration(0) {
		timer := time.NewTimer(timeout)

		select {
		case s <- struct{}{}:
			timer.Stop()
		case <-timer.C:
			return ErrSemTimeout
		}
	} else {
		s <- struct{}{}
	}
	return nil
}

func (s Semaphore) Release() {
	<-s
}
