/*
 * @Author: calmwu
 * @Date: 2018-01-30 14:44:39
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-01-30 15:28:05
 * @Comment:
 */

package utils

import (
	"reflect"
	"sync"
	"sync/atomic"
)

/*
* Increment the counter at the time of Acquiring the object
* Decrement the counter once the routine is done using it
* Increment the counter before passing the object inside a go-routine or channel
* Decrement the counter once the routine is done with processing the object
* An object is only put back into the pool when the reference count is zero.

 */

type ReferenceCountable interface {
	SetInstance(i interface{})
	IncrementReferenceCount()
	DecrementReferenceCount()
}

// 计数对象类型
type ReferenceCounter struct {
	count        *uint32                 `sql:"-" json:"-" yaml:"-"` // 引用计数器
	destinantion *sync.Pool              `sql:"-" json:"-" yaml:"-"` // 对象池，用来回收用
	released     *uint32                 `sql:"-" json:"-" yaml:"-"` // 统计累计释放对象次数
	Instance     interface{}             `sql:"-" json:"-" yaml:"-"`
	reset        func(interface{}) error `sql:"-" json:"-" yaml:"-"` // 用来清理Instance所有成员
	id           uint32                  `sql:"-" json:"-" yaml:"-"`
}

func (rc ReferenceCounter) IncrementReferenceCount() {
	atomic.AddUint32(rc.count, 1)
}

func (rc ReferenceCounter) DecrementReferenceCount() {
	if atomic.LoadUint32(rc.count) == 0 {
		panic("this should not happen ===>" + reflect.TypeOf(rc.Instance).String())
	}

	if atomic.AddUint32(rc.count, ^uint32(0)) == 0 {
		atomic.AddUint32(rc.released, 1)
		if err := rc.reset(rc.Instance); err != nil {
			panic("error while resetting an instance ===>" + err.Error())
		}
		rc.destinantion.Put(rc.Instance)
		rc.Instance = nil
	}
}

func (rc *ReferenceCounter) SetInstance(i interface{}) {
	rc.Instance = i
}

type FactoryReferenceCountable func(ReferenceCounter) ReferenceCountable
type ResetReferenceCountable func(interface{}) error

// 计数对象池
type ReferenceCountedPool struct {
	pool       *sync.Pool
	allocated  uint32 // 统计pool分配的对象数量，只有pool中没有空闲时才会调用new
	returned   uint32 // 返回pool的对象数量
	referenced uint32 // 统计pool返回对象的数量，这个肯定要大于allocated
}

// 生成一个计数对象池
func NewReferenceCountedPool(factory FactoryReferenceCountable, reset ResetReferenceCountable) *ReferenceCountedPool {
	rcPool := new(ReferenceCountedPool)
	rcPool.pool = new(sync.Pool)
	rcPool.pool.New = func() interface{} {
		// 创建一个计数对象
		atomic.AddUint32(&rcPool.allocated, 1)
		// 调用厂方法创建对象
		rc := factory(ReferenceCounter{
			count:        new(uint32),
			destinantion: rcPool.pool,
			released:     &rcPool.returned,
			reset:        reset,
			id:           rcPool.allocated, // 计数器作为id
		})
		return rc
	}
	return rcPool
}

func (rcp *ReferenceCountedPool) Get() ReferenceCountable {
	rc := rcp.pool.Get().(ReferenceCountable)
	rc.SetInstance(rc)
	atomic.AddUint32(&rcp.referenced, 1)
	// 增加计数
	rc.IncrementReferenceCount()
	return rc
}

// 输出pool的统计
func (rcp *ReferenceCountedPool) Stats() map[string]interface{} {
	return map[string]interface{}{"allocated": rcp.allocated, "referenced": rcp.referenced, "returned": rcp.returned}
}
