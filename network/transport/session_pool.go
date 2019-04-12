/*
 * @Author: calmwu
 * @Date: 2017-12-08 15:58:24
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-12-08 16:06:03
 * @Comment:
 */

package transport

import "sync"

type SessionCmd int

const (
	E_SESSIONCMD_STARTCONN      SessionCmd = iota // 握手通过，通知业务新连接到来
	E_SESSIONCMD_STOPCONN                         // 客户端断开连接
	E_SESSIONCMD_ACTIVESTOPCONN                   // 服务器主动断开连接
	E_SESSIONCMD_TRANSFER                         // 数据透传
)

type NetSessionData struct {
	Cmd       SessionCmd
	SessionID uint32
	MsgId     uint32 // 消息序号
	Data      []byte
}

var (
	netSesesionDataPool *sync.Pool
)

func init() {
	netSesesionDataPool = new(sync.Pool)
	netSesesionDataPool.New = func() interface{} {
		return new(NetSessionData)
	}
}

func PoolGetSessionData() *NetSessionData {
	return netSesesionDataPool.Get().(*NetSessionData)
}

func PoolPutSessionData(data *NetSessionData) {
	if data != nil {
		netSesesionDataPool.Put(data)
	}
}
