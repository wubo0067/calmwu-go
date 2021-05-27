/*
 * @Author: CALM.WU
 * @Date: 2021-04-21 11:10:44
 * @Last Modified by: CALM.WU
 * @Last Modified time: 2021-04-21 14:14:22
 */

package utils

// RingArrayGrowing is a growing ring array
// Not thread safe
type RingArrayGrowing struct {
	items        []interface{}
	size         int // size of array
	beginPos     int // first item
	readableSize int // number of data items available
}

// NewRingArrayGrowing construct a new RingArrayGrowing instance
func NewRingArrayGrowing(initialSize int) *RingArrayGrowing {
	return &RingArrayGrowing{
		items: make([]interface{}, initialSize),
		size:  initialSize,
	}
}

// ReadOne consumes first item from the array if it is avaliable, otherwise return false
func (r *RingArrayGrowing) ReadOne() (interface{}, bool) {
	if r.readableSize == 0 {
		return nil, false
	}

	r.readableSize--
	item := r.items[r.beginPos]
	r.items[r.beginPos] = nil // remove reference, help GC

	if r.beginPos == r.size-1 {
		r.beginPos = 0
	} else {
		r.beginPos++
	}
	return item, true
}

// WriteOne add an item to the end of the array, growing it if it is full
func (r *RingArrayGrowing) WriteOne(item interface{}) {
	if r.readableSize == r.size {
		// 扩展 * 2
		newSize := r.size * 2
		newItems := make([]interface{}, newSize)
		to := r.beginPos + r.readableSize
		if to <= r.readableSize {
			copy(newItems, r.items[r.beginPos:to])
		} else {
			copied := copy(newItems, r.items[r.beginPos:])
			copy(newItems[copied:], r.items[:(to%r.size)])
		}
		r.beginPos = 0
		r.items = newItems
		r.size = newSize
	}
	r.items[(r.readableSize+r.beginPos)%r.size] = item
	r.readableSize++
}
