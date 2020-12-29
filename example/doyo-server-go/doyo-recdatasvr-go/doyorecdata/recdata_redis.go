/*
 * @Author: calmwu
 * @Date: 2018-11-07 15:03:10
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-11-15 20:25:45
 */

package doyorecdata

import (
	"doyo-server-go/doyo-base-go/redistool"
	"fmt"
	"time"
)

type DoyoRecDataRedisKeyMgr struct {
	fmtFollowing          string
	fmtFollower           string
	fmtFriends            string
	fmtRoomInfo           string
	fmtCountryHeatRooms   string
	HeatRooms             string
	fmtUser               string
	onlineUsersRedisKey   string
	fmtOnlineCountryUsers string
}

var (
	doyoRecDataRedisKeyMgr *DoyoRecDataRedisKeyMgr
	DoyoRecDataRedisMgr    *redistool.RedisMgr
)

const (
	RedisRoomInfoRecyclingDelayDuration = 300 * time.Second
)

func init() {
	doyoRecDataRedisKeyMgr = new(DoyoRecDataRedisKeyMgr)
	doyoRecDataRedisKeyMgr.fmtFollowing = "following-set-%s"
	doyoRecDataRedisKeyMgr.fmtFollower = "follower-set-%s"
	doyoRecDataRedisKeyMgr.fmtFriends = "friends-set-%s"
	doyoRecDataRedisKeyMgr.fmtRoomInfo = "roominfo-hash-%s"                 // 房间信息
	doyoRecDataRedisKeyMgr.fmtCountryHeatRooms = "countryheatrooms-zset-%s" // 按国家的房间热度
	doyoRecDataRedisKeyMgr.HeatRooms = "heatrooms"                          // 房间热度
	doyoRecDataRedisKeyMgr.fmtUser = "user-%s"
	doyoRecDataRedisKeyMgr.onlineUsersRedisKey = "onlineuser-set"
	doyoRecDataRedisKeyMgr.fmtOnlineCountryUsers = "onlineuser-set-%s"
}

func (km *DoyoRecDataRedisKeyMgr) getUserRedisKey(userID string) string {
	return fmt.Sprintf(km.fmtUser, userID)
}

func (km *DoyoRecDataRedisKeyMgr) getOnlineCountryUsersRedisKey(countryIsoCode string) string {
	return fmt.Sprintf(km.fmtOnlineCountryUsers, countryIsoCode)
}

func (km *DoyoRecDataRedisKeyMgr) getRoomRedisKey(roomID string) string {
	return fmt.Sprintf(km.fmtRoomInfo, roomID)
}
