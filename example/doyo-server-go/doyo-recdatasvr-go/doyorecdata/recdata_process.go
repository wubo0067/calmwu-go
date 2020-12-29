/*
 * @Author: calmwu
 * @Date: 2018-11-05 10:58:22
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-11-15 15:44:16
 */

package doyorecdata

import (
	base "doyo-server-go/doyo-base-go"
	"doyo-server-go/doyo-recdatasvr-go/proto"
	routerstub "doyo-server-go/doyo-routersvr-go/doyo-routerstub"
	routerproto "doyo-server-go/doyo-routersvr-go/proto"
	"sync"

	"github.com/pquerna/ffjson/ffjson"
)

type doyoRecDataCmdProcess func(doyoRsm *routerstub.RouterStubModule, msgID string, fromTopic string,
	cmd proto.DoyoRecDataCmdType, cmdData []byte)

var (
	doyoRecDataProcessMap = map[proto.DoyoRecDataCmdType]doyoRecDataCmdProcess{
		proto.DoyoRecDataCmdUserLogin:       processUserLogin,
		proto.DoyoRecDataCmdUserLogout:      processUserLogout,
		proto.DoyoRecDataCmdAnchorStartRoom: processAnchorStartRoom,
		proto.DoyoRecDataCmdAnchorStopRoom:  processAnchorStopRoom,
		proto.DoyoRecDataCmdEnterRoom:       processUserEnterRoom,
		proto.DoyoRecDataCmdLeaveRoom:       processUserLeaveRoom,
		proto.DoyoRecDataCmdAddFriends:      processUserAddFriends,
		proto.DoyoRecDataCmdAddFollow:       processUserAddFollow,
		proto.DoyoRecDataCmdDelFollow:       processUserDelFollow,
	}

	appServMsgPool = sync.Pool{
		New: func() interface{} {
			return new(routerproto.AppServMsg)
		},
	}
)

func OnReceive(doyoRsm *routerstub.RouterStubModule, msgID string, fromTopic string, data []byte) {
	base.ZLog.Debugf("msgID[%s] fromTopic[%s]", msgID, fromTopic)

	appServMsg := appServMsgPool.Get().(*routerproto.AppServMsg)
	err := ffjson.Unmarshal(data, appServMsg)
	if err != nil {
		base.ZLog.Errorf("ffjson Unmarshal msgID[%s] fromTopic[%s] failed! reason:%s",
			msgID, fromTopic, err.Error())
	} else {
		// TODO：用goroutine pool
		go processMsg(doyoRsm, msgID, fromTopic, appServMsg)
	}
}

func processMsg(doyoRsm *routerstub.RouterStubModule, msgID string, fromTopic string, appServMsg *routerproto.AppServMsg) {
	defer func() {
		appServMsgPool.Put(appServMsg)
		recDataStatistics.decRunningCmdCount()

		if err := recover(); err != nil {
			// 回收panic，防止整个服务panic
			stackInfo := base.CallStack(1)
			base.ZLog.DPanicw("cmdProcessor panic recovered! err", err, "stack", string(stackInfo))
		}
	}()

	recDataStatistics.incRunningCmdCount()

	cmdID := proto.DoyoRecDataCmdType(appServMsg.AppServCmdID)
	if processor, ok := doyoRecDataProcessMap[cmdID]; ok {
		if processor != nil {
			processor(doyoRsm, msgID, fromTopic, cmdID, appServMsg.AppServCmdData)
		}
	} else {
		base.ZLog.Errorf("cmdID[%s] processor is not register", cmdID)
	}
}
