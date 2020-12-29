/*
 * @Author: calmwu
 * @Date: 2018-10-25 14:14:01
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-11-15 15:51:12
 */

package proto

type DoyoRecDataCmdType int

const (
	DoyoRecDataCmdUserLogin DoyoRecDataCmdType = iota
	DoyoRecDataCmdUserLogout
	DoyoRecDataCmdAnchorStartRoom
	DoyoRecDataCmdAnchorStopRoom
	DoyoRecDataCmdEnterRoom
	DoyoRecDataCmdLeaveRoom
	DoyoRecDataCmdAddFriends
	DoyoRecDataCmdAddFollow
	DoyoRecDataCmdDelFollow
)

type ProtoDoyoRecDataUserLogin struct {
	UserID       string `json:"UserID"`
	UserLanguage string `json:"UserLanguage"` // http://www.lingoes.cn/zh/translator/langcode.htm
	UserCountry  string `json:"UserCountry"`  // https://zh.wikipedia.org/wiki/ISO_3166-1
	UserGender   int    `json:"UserGender"`   // 性别
}

type ProtoDoyoRecDataUserLogout struct {
	UserID string `json:"UserID"`
}

// 开播
type ProtoDoyoRecDataAnchorStartRoom struct {
	AnchorID        string `json:"AnchorID"`
	RoomID          string `json:"RoomID"`
	AnchorLanguage  string `json:"AnchorLanguage"` // 主播语言
	AnchorCountry   string `json:"AnchorCountry"`  // 主播国家
	RoomTitle       string `json:"RoomTitle"`
	DefinitionLabel string `json:"DefinitionLabel"`
	GameName        string `json:"GameName"`
}

// 关播
type ProtoDoyoRecDataAnchorStopRoom struct {
	AnchorID string `json:"AnchorID"`
	RoomID   string `json:"RoomID"`
}

// 进入房间，计算房间热度
type ProtoDoyoRecDataUserEnterRoom struct {
	UserID string `json:"UserID"`
	RoomID string `json:"RoomID"`
}

// 离开房间
type ProtoDoyoRecDataUserLeaveRoom ProtoDoyoRecDataUserEnterRoom

// 添加好友
type ProtoDoyoRecDataAddFriends struct {
	UserID    string   `json:"UserID"`
	FriendLst []string `json:"FriendLst"`
}

// 用户添加关注
type ProtoDoyoRecDataAddFollow struct {
	UserID   string `json:"UserID"`
	FollowID string `json:"FollowID"`
}

// 取消关注
type ProtoDoyoRecDataDelFollow struct {
	UserID   string `json:"UserID"`
	FollowID string `json:"FollowID"`
}
