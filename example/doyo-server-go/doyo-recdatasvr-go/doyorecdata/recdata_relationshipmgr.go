/*
 * @Author: calmwu
 * @Date: 2018-11-13 10:35:51
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-11-13 10:37:17
 */

// 处理用户的 followerlist followinglist friendlist

package doyorecdata

import (
	base "doyo-server-go/doyo-base-go"
	"doyo-server-go/doyo-recdatasvr-go/proto"
	routerstub "doyo-server-go/doyo-routersvr-go/doyo-routerstub"
	"time"
)

func processUserAddFriends(doyoRsm *routerstub.RouterStubModule, msgID string, fromTopic string,
	cmd proto.DoyoRecDataCmdType, cmdData []byte) {

	now := time.Now()
	defer base.TimeTaken(now, "processUserAddFriends")
}

func processUserAddFollow(doyoRsm *routerstub.RouterStubModule, msgID string, fromTopic string,
	cmd proto.DoyoRecDataCmdType, cmdData []byte) {

	now := time.Now()
	defer base.TimeTaken(now, "processUserAddFollow")
}

func processUserDelFollow(doyoRsm *routerstub.RouterStubModule, msgID string, fromTopic string,
	cmd proto.DoyoRecDataCmdType, cmdData []byte) {

	now := time.Now()
	defer base.TimeTaken(now, "processUserDelFollow")
}
