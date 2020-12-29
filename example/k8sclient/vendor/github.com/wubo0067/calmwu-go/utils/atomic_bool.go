/*
 * @Author: calmwu
 * @Date: 2019-03-07 19:04:11
 * @Last Modified by:   calmwu
 * @Last Modified time: 2019-03-07 19:04:11
 */

package utils

import "sync/atomic"

// An AtomicBool is an atomic bool
type AtomicBool struct {
	v int32
}

// Set sets the value
func (a *AtomicBool) Set(value bool) {
	var n int32
	if value {
		n = 1
	}
	atomic.StoreInt32(&a.v, n)
}

// Get gets the value
func (a *AtomicBool) Get() bool {
	return atomic.LoadInt32(&a.v) != 0
}
