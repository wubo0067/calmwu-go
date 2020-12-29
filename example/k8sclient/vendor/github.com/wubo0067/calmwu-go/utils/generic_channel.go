/*
 * @Author: calmwu
 * @Date: 2019-12-02 16:04:57
 * @Last Modified by: calmwu
 * @Last Modified time: 2019-12-02 16:30:21
 */

package utils

// NOTE: this is how easy it is to define a generic type
import (
	"sync"

	"github.com/cheekybits/genny/generic"
)

// ChannelCustomType channel的类型
type ChannelCustomType generic.Type

// ChannelCustomName  channel的名字
type ChannelCustomName generic.Type

// ChannelCustomNameChannel channel的封装对象
type ChannelCustomNameChannel struct {
	C          chan ChannelCustomType
	mutex      sync.Mutex
	closedFlag bool
}

// NewChannelCustomNameChannel 创建函数
func NewChannelCustomNameChannel(size int) *ChannelCustomNameChannel {
	customChannel := new(ChannelCustomNameChannel)
	if size > 0 {
		customChannel.C = make(chan ChannelCustomType, size)
	} else {
		customChannel.C = make(chan ChannelCustomType)
	}
	customChannel.closedFlag = false
	return customChannel
}

// IsClosed 判断是否被关闭
func (cc *ChannelCustomNameChannel) IsClosed() bool {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()
	return cc.closedFlag
}

// SafeClose 安全的关闭channel
func (cc *ChannelCustomNameChannel) SafeClose() {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()
	if !cc.closedFlag {
		close(cc.C)
		cc.closedFlag = true
	}
}

// SafeSend 安全的发送数据
func (cc *ChannelCustomNameChannel) SafeSend(value ChannelCustomType, block bool) (ok, closed bool) {
	defer func() {
		if recover() != nil {
			closed = true
			ok = false
		}
	}()

	if block {
		cc.C <- value
	} else {
		select {
		case cc.C <- value:
			ok = true
		default:
			ok = false
		}
	}
	closed = false
	return
}

// Read 读取
func (cc *ChannelCustomNameChannel) Read(block bool) (val ChannelCustomType, ok bool) {
	if block {
		val, ok = <-cc.C
	} else {
		select {
		case val, ok = <-cc.C:
			if !ok && !cc.closedFlag {
				cc.mutex.Lock()
				defer cc.mutex.Unlock()
				cc.closedFlag = true
			}
		default:
			ok = false
		}
	}
	return
}
