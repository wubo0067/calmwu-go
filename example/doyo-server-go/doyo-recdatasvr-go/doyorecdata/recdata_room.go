/*
 * @Author: calmwu
 * @Date: 2018-10-26 15:09:08
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-11-15 20:27:08
 */

// 房间hash，房间热度，在线房间列表

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

func processAnchorStartRoom(doyoRsm *routerstub.RouterStubModule, msgID string, fromTopic string,
	cmd proto.DoyoRecDataCmdType, cmdData []byte) {
	now := time.Now()
	defer base.TimeTaken(now, "processAnchorStartRoom")

	var err error
	var anchorStartRoom proto.ProtoDoyoRecDataAnchorStartRoom

	err = ffjson.Unmarshal(cmdData, &anchorStartRoom)
	if err != nil {
		base.ZLog.Errorf("fromTopic[%s] msgID[%s] cmd[%s] Decode failed! reason[%s]",
			fromTopic, msgID, cmd.String(), err.Error())
	} else {
		roomRedisKey := doyoRecDataRedisKeyMgr.getRoomRedisKey(anchorStartRoom.RoomID)
		userRedisKey := doyoRecDataRedisKeyMgr.getUserRedisKey(anchorStartRoom.AnchorID)

		// hmget获取房间信息，判断是否是重复开播
		roomGetRedisInfo, err := DoyoRecDataRedisMgr.HashGetAll(roomRedisKey)
		if err != nil {
			base.ZLog.Error("Room[%s] HGETALL execute faield! reason:%s", roomRedisKey, err.Error())
			return
		}

		doyoRedisRoomInfo := new(proto.DoyoRedisRoomInfo)

		var isReOpen bool = false
		if len(roomGetRedisInfo) == 0 {
			// 新开播的
			base.ZLog.Infof("Room:%+v is new launch", anchorStartRoom)
		} else {
			// 重新开播
			isReOpen = true
			base.ZLog.Warnf("Room[%s]：%+v repeat the launch!!!!", anchorStartRoom.RoomID, roomGetRedisInfo)
		}

		doyoRedisRoomInfo.ProtoDoyoRecDataAnchorStartRoom = anchorStartRoom
		doyoRedisRoomInfo.ViewerCount = 0
		doyoRedisRoomInfo.StartTime = now

		roomSetRedisInfo, err := redistool.ConvertObjToInterfaceMap(doyoRedisRoomInfo)
		if err != nil {
			base.ZLog.Errorf("Room[%s] doyoRedisRoomInfo redistool.ConvertObjToInterfaceMap failed!", roomRedisKey)
			return
		}

		cmds, err := DoyoRecDataRedisMgr.Pipelined(func(pipe redis.Pipeliner) error {
			// 设置房间信息
			pipe.HMSet(roomRedisKey, roomSetRedisInfo)
			if isReOpen {
				// 如果重开要删除ttl
				pipe.Persist(roomRedisKey)
			}
			// 设置用户状态为主播
			pipe.HMSet(userRedisKey, map[string]interface{}{
				"IsAnchor": 1,
			})
			return nil
		})

		if err != nil {
			base.ZLog.Errorf("Room[%s] processAnchorStartRoom Pipelined failed! error:%s", roomRedisKey, err.Error())
		} else {
			base.ZLog.Debugf("Room[%s] processAnchorStartRoom Pipelined successed! result:%v", roomRedisKey, cmds)
		}
	}
}

func processAnchorStopRoom(doyoRsm *routerstub.RouterStubModule, msgID string, fromTopic string,
	cmd proto.DoyoRecDataCmdType, cmdData []byte) {
	now := time.Now()
	defer base.TimeTaken(now, "processAnchorStopRoom")

	var anchorStopRoom proto.ProtoDoyoRecDataAnchorStopRoom

	err := ffjson.Unmarshal(cmdData, &anchorStopRoom)
	if err != nil {
		base.ZLog.Errorf("fromTopic[%s] msgID[%s] cmd[%s] Decode failed! reason[%s]",
			fromTopic, msgID, cmd.String(), err.Error())
	} else {
		roomRedisKey := doyoRecDataRedisKeyMgr.getRoomRedisKey(anchorStopRoom.RoomID)
		userRedisKey := doyoRecDataRedisKeyMgr.getUserRedisKey(anchorStopRoom.AnchorID)
		// 判断房间是否存在，然后删除该房间
		exists, err := DoyoRecDataRedisMgr.Exists(roomRedisKey)
		if err != nil {
			base.ZLog.Error("Room[%s] Exists cmd execute faield! reason:%s", roomRedisKey, err.Error())
			return
		}

		if exists == 1 {
			cmds, err := DoyoRecDataRedisMgr.Pipelined(func(pipe redis.Pipeliner) error {
				// 修改房间结束时间
				pipe.HMSet(roomRedisKey, map[string]interface{}{
					"StopTime": base.TimeName(now),
				})
				// 设置延期清除ttl
				pipe.Expire(roomRedisKey, RedisRoomInfoRecyclingDelayDuration)
				pipe.HMSet(userRedisKey, map[string]interface{}{
					"IsAnchor": 0,
				})
				return nil
			})

			if err != nil {
				base.ZLog.Errorf("Room[%s] processAnchorStopRoom Pipelined failed! error:%s", roomRedisKey, err.Error())
			} else {
				base.ZLog.Debugf("Room[%s] processAnchorStopRoom Pipelined successed! result:%v", roomRedisKey, cmds)
			}
		} else {
			base.ZLog.Errorf("Room[%s] does no exist", roomRedisKey)
		}
	}
}
