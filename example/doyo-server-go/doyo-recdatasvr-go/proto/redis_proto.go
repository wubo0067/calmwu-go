/*
 * @Author: calmwu
 * @Date: 2018-10-25 15:00:30
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-11-29 19:27:25
 */

package proto

import "time"

// redis中的数据对象

// 用户，运行时信息
type DoyoRedisUserInfo struct {
	ProtoDoyoRecDataUserLogin
	FriendCount   int       `json:"FriendCount"`    // 好友数量
	FollwerCount  int       `json:"FollwerCount"`   // 粉丝数量
	FollwingCount int       `json:"FollowingCount"` // 关注数量
	IsOnline      int       `json:"IsOnline"`       // 是否在线 0：下线，1：上线
	IsAnchor      int       `json:"IsAnchor"`       // 是否是主播 1：主播，0：非主播
	IsBan         int       `json:"IsBan"`          // 是否被禁止
	LoginTime     time.Time `json:"LoginTime"`      // 上线时间
	LogoutTime    time.Time `json:"LogoutTime"`     // 下线时间
}

type DoyoRedisUserAction struct {
	ViewerEnterRoomLst    []string `json:"ViewerEnterRoomLst"`    // 用户最近进入房间列表，保存20个
	ViewerWatchContentLst []string `json:"ViewerWatchContentLst"` // 用户最近观看内容列表，保存20个
	AncherPlayContentlst  []string `json:"AncherPlayContentlst"`  // 主播最近播放内容列表，保存20个
}

// 房间信息
type DoyoRedisRoomInfo struct {
	ProtoDoyoRecDataAnchorStartRoom
	ViewerCount int       `json:"ViewerCount"` // 同时观看人数 HINCRBY
	StartTime   time.Time `json:"StartTime"`   // 开播时间
	StopTime    time.Time `json:"StopTime"`    // 关播时间
}
