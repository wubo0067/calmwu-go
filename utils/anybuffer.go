/*
 * @Author: CALM.WU
 * @Date: 2023-06-06 15:05:42
 * @Last Modified by: CALM.WU
 * @Last Modified time: 2023-06-06 15:13:12
 */

package utils

import (
	"github.com/pkg/errors"
	"golang.org/x/exp/constraints"
)

type Buffer[T constraints.Unsigned] struct {
	buf []T //contents are the bytes buf[off : len(buf)]
}

var ErrTooLarge = errors.New("bytes.Buffer: too large")

const (
	maxInt            = int(^uint(0) >> 1)
	smallBufferSize   = 16
	defaultBufferSize = 64
)

func (b *Buffer[T]) Slice() []T {
	return b.buf
}

// The `empty()` function is a method of the `Buffer` struct that takes a pointer to a `Buffer`
// instance and returns a boolean value indicating whether the buffer is empty or not. It does this by
// checking the length of the buffer's underlying slice (`b.buf`) and returning `true` if the length is
// less than or equal to 0, and `false` otherwise.
func (b *Buffer[T]) empty() bool {
	return len(b.buf) <= 0
}

func (b *Buffer[T]) Len() int { return len(b.buf) - 0 }

func (b *Buffer[T]) Cap() int { return cap(b.buf) }

func (b *Buffer[T]) truncate(n int) {
	if n == 0 {
		b.Reset()
		return
	}

	if n < 0 || n > b.Len() {
		panic("bytes.Buffer: truncation out of range")
	}
	b.buf = b.buf[:0+n]
}

func (b *Buffer[T]) Reset() {
	b.buf = b.buf[:0]
}

func (b *Buffer[T]) tryGrowByReslice(n int) (int, bool) {
	if l := len(b.buf); n <= cap(b.buf)-l {
		// 剩余的容量如果满足增加n的需求，则直接resliced
		b.buf = b.buf[:l+n]
		return l, true
	}
	return 0, false
}

// n是额外增加的元素数量
func (b *Buffer[T]) grow(n int) int {
	m := b.Len()

	if i, ok := b.tryGrowByReslice(n); ok {
		return i
	}

	if b.buf == nil && n <= smallBufferSize {
		b.buf = make([]T, n, smallBufferSize)
		return 0
	}

	c := cap(b.buf)
	if n <= c/2-m {
		// We can slide things down instead of allocating a new
		// slice. We only need m+n <= c to slide, but
		// we instead let capacity get twice as large so we
		// don't spend all our time copying.
	} else if c > maxInt-c-n {
		// not enough space anywhere，如果2倍的容量加上n超过了maxInt，直接报错
		panic(ErrTooLarge)
	} else {
		// enough space at end，按2倍容量扩容
		buf := makeSlice[T](2*c + n)
		copy(buf, b.buf[0:])
		b.buf = buf
	}
	b.buf = b.buf[:m+n]
	return m
}

// 扩展capacity
func (b *Buffer[T]) Grow(n int) {
	if n < 0 {
		panic("bytes.Buffer.Grow: negative count")
	}
	m := b.grow(n)
	b.buf = b.buf[:m]
}

func (b *Buffer[T]) Extend(n int) {
	b.extend(n)
}

func (b *Buffer[T]) extend(n int) int {
	m, ok := b.tryGrowByReslice(n)
	if !ok {
		m = b.grow(n)
	}
	return m
}

func makeSlice[T constraints.Unsigned](n int) []T {
	defer func() {
		if recover() != nil {
			panic(ErrTooLarge)
		}
	}()
	return make([]T, n)
}

func NewBuffer[T constraints.Unsigned](size int) *Buffer[T] {
	if size == 0 {
		size = defaultBufferSize
	}

	return &Buffer[T]{buf: make([]T, 0, size)}
}

func NewBufferFrom[T constraints.Unsigned](buf []T) *Buffer[T] {
	return &Buffer[T]{buf: buf}
}
