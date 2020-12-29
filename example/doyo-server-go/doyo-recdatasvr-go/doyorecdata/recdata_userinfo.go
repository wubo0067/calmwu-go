/*
 * @Author: calmwu
 * @Date: 2018-10-26 15:07:10
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-11-15 16:30:50
 */

// 用户信息，用户行为 HASH，在线用户id列表

package doyorecdata

import (
	base "doyo-server-go/doyo-base-go"
	"doyo-server-go/doyo-base-go/redistool"
	"doyo-server-go/doyo-recdatasvr-go/proto"
	routerstub "doyo-server-go/doyo-routersvr-go/doyo-routerstub"
	"time"

	"github.com/go-redis/redis"
	"github.com/pquerna/ffjson/ffjson"
)

func processUserLogin(doyoRsm *routerstub.RouterStubModule, msgID string, fromTopic string,
	cmd proto.DoyoRecDataCmdType, cmdData []byte) {
	now := time.Now()
	defer base.TimeTaken(now, "processUserLogin")

	var err error
	var userLogin proto.ProtoDoyoRecDataUserLogin

	err = ffjson.Unmarshal(cmdData, &userLogin)
	if err != nil {
		base.ZLog.Errorf("fromTopic[%s] msgID[%s] cmd[%s] Decode failed! reason[%s]",
			fromTopic, msgID, cmd.String(), err.Error())
	} else {
		userRedisKey := doyoRecDataRedisKeyMgr.getUserRedisKey(userLogin.UserID)
		onlineCountryUsersRedisKey := doyoRecDataRedisKeyMgr.getOnlineCountryUsersRedisKey(userLogin.UserCountry)

		// hmget获取用户信息
		userGetRedisInfo, err := DoyoRecDataRedisMgr.HashGetAll(userRedisKey)
		if err != nil {
			base.ZLog.Error("hgetall user[%s] from redis faield! reason:%s", userRedisKey, err.Error())
			return
		}

		base.ZLog.Debugf("userGetRedisInfo:%+v", userGetRedisInfo)
		doyoRedisUserInfo := new(proto.DoyoRedisUserInfo)

		if len(userGetRedisInfo) == 0 {
			// 这是个新用户 hashsetall
			base.ZLog.Infof("UserInfo:%+v is new player", userLogin)
			doyoRedisUserInfo.UserID = userLogin.UserID
			doyoRedisUserInfo.UserLanguage = userLogin.UserLanguage
			doyoRedisUserInfo.UserCountry = userLogin.UserCountry
			doyoRedisUserInfo.UserGender = userLogin.UserGender
			doyoRedisUserInfo.IsOnline = 1
			doyoRedisUserInfo.LoginTime = now
			doyoRedisUserInfo.LogoutTime = now
		} else {
			err = redistool.ConvertStringMapToObj(userGetRedisInfo, doyoRedisUserInfo)
			if err != nil {
				base.ZLog.Error("User[%s] ConvertStringMapToObj failed! error:%s", userLogin.UserID, err.Error())
				return
			}

			if doyoRedisUserInfo.IsOnline == 1 {
				base.ZLog.Warnf("User[%s] is already loggined in", doyoRedisUserInfo.UserID)
				return
			}
			doyoRedisUserInfo.LoginTime = now
			doyoRedisUserInfo.IsOnline = 1
		}

		userSetRedisInfo, err := redistool.ConvertObjToInterfaceMap(doyoRedisUserInfo)
		if err != nil {
			base.ZLog.Errorf("User[%s] doyoRedisUserInfo redistool.ConvertObjToInterfaceMap failed!", userRedisKey)
			return
		}

		// 执行pipeline
		cmds, err := DoyoRecDataRedisMgr.Pipelined(func(pipe redis.Pipeliner) error {
			// 加入在线列表
			pipe.SAdd(doyoRecDataRedisKeyMgr.onlineUsersRedisKey, userLogin.UserID)
			// 加入国家在线列表
			pipe.SAdd(onlineCountryUsersRedisKey, userLogin.UserID)
			// 修改用户信息
			pipe.HMSet(userRedisKey, userSetRedisInfo)
			return nil
		})

		if err != nil {
			base.ZLog.Errorf("processUserLogin Pipelined failed! error:%s", err.Error())
		} else {
			base.ZLog.Debugf("processUserLogin Pipelined successed! result:%v", cmds)
		}
	}
}

func processUserLogout(doyoRsm *routerstub.RouterStubModule, msgID string, fromTopic string,
	cmd proto.DoyoRecDataCmdType, cmdData []byte) {

	now := time.Now()
	defer base.TimeTaken(now, "processUserLogin")

	var userLogout proto.ProtoDoyoRecDataUserLogout
	err := ffjson.Unmarshal(cmdData, &userLogout)
	if err != nil {
		base.ZLog.Errorf("fromTopic[%s] msgID[%s] cmd[%s] Decode failed! reason[%s]",
			fromTopic, msgID, cmd.String(), err.Error())
	} else {
		userRedisKey := doyoRecDataRedisKeyMgr.getUserRedisKey(userLogout.UserID)

		// hmget获取用户信息
		userGetRedisInfo, err := DoyoRecDataRedisMgr.HashGetAll(userRedisKey)
		if err != nil {
			base.ZLog.Error("hgetall user[%s] from redis faield! reason:%s", userRedisKey, err.Error())
			return
		}

		onlineCountryUsersRedisKey := doyoRecDataRedisKeyMgr.getOnlineCountryUsersRedisKey(userGetRedisInfo["UserCountry"])

		base.ZLog.Debugf("User[%s] logout, userRedisKey[%s]", userLogout.UserID, userRedisKey)
		cmds, err := DoyoRecDataRedisMgr.Pipelined(func(pipe redis.Pipeliner) error {
			// remove from 在线列表
			pipe.SRem(doyoRecDataRedisKeyMgr.onlineUsersRedisKey, userLogout.UserID)
			// remove from country user online list
			pipe.SRem(onlineCountryUsersRedisKey, userLogout.UserID)
			// 修改在线标识，修改离线时间
			pipe.HMSet(userRedisKey, map[string]interface{}{
				"IsOnline":   0,
				"LogoutTime": base.TimeName(now),
			})
			return nil
		})

		if err != nil {
			base.ZLog.Errorf("processUserLogout Pipelined failed! error:%s", err.Error())
		} else {
			base.ZLog.Debugf("processUserLogout Pipelined result:%v", cmds)
		}
	}
}

/*
用户进入房间
1：递增房间人数
2：获取房间信息，记录用户行为，观看的什么内容
*/
func processUserEnterRoom(doyoRsm *routerstub.RouterStubModule, msgID string, fromTopic string,
	cmd proto.DoyoRecDataCmdType, cmdData []byte) {

	now := time.Now()
	defer base.TimeTaken(now, "processUserEnterRoom")

	var enterRoom proto.ProtoDoyoRecDataUserEnterRoom
	err := ffjson.Unmarshal(cmdData, &enterRoom)
	if err != nil {
		base.ZLog.Errorf("fromTopic[%s] msgID[%s] cmd[%s] Decode failed! reason[%s]",
			fromTopic, msgID, cmd.String(), err.Error())
	} else {
		roomRedisKey := doyoRecDataRedisKeyMgr.getRoomRedisKey(enterRoom.RoomID)

		base.ZLog.Debugf("User[%s] enter room[%s] roomRedisKey[%s]", enterRoom.UserID, enterRoom.RoomID, roomRedisKey)

		cmds, err := DoyoRecDataRedisMgr.Pipelined(func(pipe redis.Pipeliner) error {
			//
			pipe.HIncrBy(roomRedisKey, "ViewerCount", 1)
			// 将来用于机器分析
			pipe.HMGet(roomRedisKey, "AnchorLanguage", "AnchorCountry", "GameName")
			return nil
		})

		if err != nil {
			base.ZLog.Errorf("processUserLogout Pipelined failed! error:%s", err.Error())
		} else {
			base.ZLog.Debugf("processUserLogout Pipelined result:%v", cmds)
		}
	}

}

func processUserLeaveRoom(doyoRsm *routerstub.RouterStubModule, msgID string, fromTopic string,
	cmd proto.DoyoRecDataCmdType, cmdData []byte) {

	now := time.Now()
	defer base.TimeTaken(now, "processUserLeaveRoom")
}
