/*
 * @Author: CALM.WU
 * @Date: 2021-04-21 14:13:56
 * @Last Modified by: CALM.WU
 * @Last Modified time: 2021-04-21 16:59:48
 */

package utils

import (
	"errors"
	"sync"
)

var (
	ErrRingBufEmpty       = errors.New("ringbuffer is empty")
	ErrRingBufFull        = errors.New("ringbuffer is full")
	ErrRingSpaceNotEnough = errors.New("ringbuffer space is not enough")
)

// RingBuffer is a circular buffer
type RingBuffer struct {
	buf    []byte
	size   int // size of buffer
	rPos   int // read position
	wPos   int // write position
	isFull bool
	mu     sync.Mutex
}

// NewRingBuffer construct a new RingBuffer instance
func NewRingBuffer(initialSize int) *RingBuffer {
	return &RingBuffer{
		buf:  make([]byte, initialSize),
		size: initialSize,
	}
}

// Read reads up to len(p) bytes to p
func (r *RingBuffer) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	n, err = r.read(p)
	return
}

// ReadByte read a next byte from the buf or return ErrRingBufEmpty
func (r *RingBuffer) ReadByte() (b byte, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.wPos == r.rPos && !r.isFull {
		return 0, ErrRingBufEmpty
	}

	b = r.buf[r.rPos]
	r.rPos++
	if r.rPos == r.size {
		r.rPos = 0
	}

	r.isFull = false
	return b, nil
}

func (r *RingBuffer) read(p []byte) (n int, err error) {
	if r.wPos == r.rPos && !r.isFull {
		return 0, ErrRingBufEmpty
	}

	if r.wPos > r.rPos {
		// 没有调头，计算可读字节长度
		n = r.wPos - r.rPos
		if n > len(p) {
			// 可读数据大于要读取的长度
			n = len(p)
		}
		// 拷贝数据到缓冲区
		copy(p, r.buf[r.rPos:r.rPos+n])
		r.rPos = (r.rPos + n) % r.size
		return
	}

	// 计算要拷贝的数据长度
	n = r.size - r.rPos + r.wPos
	if n > len(p) {
		n = len(p)
	}

	if r.rPos+n < r.size {
		// 向后空间数据满足
		copy(p, r.buf[r.rPos:r.rPos+n])
	} else {
		s1 := r.size - r.rPos
		copy(p, r.buf[r.rPos:])
		s2 := n - s1
		copy(p[s1:], r.buf[:s2])
	}

	r.rPos = (r.rPos + n) % r.size
	r.isFull = false

	return n, err
}

// Write write len p bytes to underlying buf
func (r *RingBuffer) Write(p []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	n, err = r.write(p)
	return n, err
}

func (r *RingBuffer) WriteByte(b byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.wPos == r.rPos && r.isFull {
		return ErrRingBufFull
	}

	r.buf[r.wPos] = b
	r.wPos++

	if r.wPos == r.size {
		r.wPos = 0
	}

	if r.wPos == r.rPos {
		r.isFull = true
	}

	return nil
}

// WriteString writes the contents of the string s to buffer, which accepts a slice of bytes.
func (r *RingBuffer) WriteString(s string) (n int, err error) {
	bs := String2Bytes(s)

	r.mu.Lock()
	defer r.mu.Unlock()
	n, err = r.write(bs)
	return n, err
}

func (r *RingBuffer) write(p []byte) (n int, err error) {
	if r.isFull {
		return 0, ErrRingBufFull
	}

	// 计算可写的空间大小
	var canWriteSize int
	if r.wPos >= r.rPos {
		canWriteSize = r.size - r.wPos + r.rPos
	} else {
		canWriteSize = r.rPos - r.wPos
	}

	if len(p) > canWriteSize {
		return 0, ErrRingSpaceNotEnough
	}

	// 写入数据的长度
	n = len(p)

	if r.wPos >= r.rPos {
		// 向后，没有调头
		s1 := r.size - r.wPos
		if s1 >= n {
			copy(r.buf[r.wPos:], p)
			r.wPos += n
		} else {
			// 分段拷贝
			copy(r.buf[r.wPos:], p[:s1])
			s2 := n - s1
			copy(r.buf[0:], p[s1:])
			r.wPos = s2
		}
	} else {
		// 调头
		copy(r.buf[r.wPos:], p)
		r.wPos += n
	}

	if r.wPos == r.size {
		r.wPos = 0
	}

	if r.wPos == r.rPos {
		r.isFull = true
	}

	return n, err
}

// Length return the length of available read bytes
func (r *RingBuffer) Length() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.wPos == r.rPos {
		if r.isFull {
			return r.size
		}
		return 0
	}

	if r.wPos > r.rPos {
		return r.wPos - r.rPos
	}

	return r.size - r.rPos + r.wPos
}

// Capacity return size of buffer
func (r *RingBuffer) Capacity() int {
	return r.size
}

// Free return length of available bytes to write
func (r *RingBuffer) Free() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.wPos == r.rPos {
		if r.isFull {
			return 0
		}
		return r.size
	}

	if r.wPos < r.rPos {
		return r.rPos - r.wPos
	}

	return r.size - r.wPos + r.rPos
}

// IsFull return this ringbuffer is full
func (r *RingBuffer) IsFull() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.isFull
}

// IsEmpty return this ringbuffer is empty
func (r *RingBuffer) IsEmpty() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	return !r.isFull && r.wPos == r.rPos
}

// Reset set read & write pos to zero
func (r *RingBuffer) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.rPos = 0
	r.wPos = 0
	r.isFull = false
}
