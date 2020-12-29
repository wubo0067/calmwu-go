/*
 * @Author: calmwu
 * @Date: 2017-12-26 14:49:39
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-05-19 15:46:14
 * @Comment:
 */

package proto

const (
	CHANNEL_NAME_GOOGLE = "GooglePlay"
	CHANNEL_NAME_APPLE  = "AppStore"

	PLATFORM_NAME_GOOGLE = "Android"
	PLATFORM_NAME_APPLE  = "Ios"
)

const (
	E_SERVER_STATUS_RUNING   = 0
	E_SERVER_STATUS_MAINTAIN = 1
	E_SERVER_STATUS_TESTING  = 2
)

const (
	ProtoRetCodeError            = -1
	ProtoRetCodeLoginOK          = 0 // 可以登陆
	ProtoRetCodeProhibitionLogin = 1 // 禁止登陆
	ProtoRetCodeNeedUpdate       = 2 // 版本需要更新
	ProtoRetCodeMaintenance      = 3 // 服务器维护中
)

const (
	APINAME_LoginCheck = "LoginCheck"
)

// 登录检查请求
type ProtoGuideSvrLoginCheckReqS struct {
	ClientVersion string `json:"ClientVersion"`
	PlatformName  string `json:"PlatformName"`
	ChannelName   string `json:"ChannelName"`
}

// 版本更新回应
type ProtoGuideSvrVersionUpdateS struct {
	NewVersion  string `json:"NewVersion"`
	ChannelName string `json:"ChannelName"`
	UpdateUrl   string `json:"UpdateUrl"`
}

// 登录信息
type ProtoGuideSvrLoginInfoS struct {
	ServerIPs        []string `json:"ServerIPs"` // 根据负载排序的接入IP列表
	Port             int      `json:Port`
	ClientInternetIP string   `json:"ClientInternetIP"`
}

// 服务器维护回应
type ProtoGuideSvrMaintenanceS struct {
	Bulletin         string `json:"Bulletin"`         // 维护公告
	RemainingSeconds int    `json:"RemainingSeconds"` // 开服预计剩余秒数
}

type MaintainParams struct {
	WhiteList    []string `json:"white_list"`
	MaintainFlag int      `json:"game_maintain_flag"`
	WhiteFlag    int      `json:"white_flag"`
	MainDeadLine int64    `json:"main_dead_line"`
}

type MaintainInfoS struct {
	MaintainKey MaintainParams `json:"gm_tool_game_maintain_key"`
}

//--------------------------------------------------------------------------------------------------------
// 根据客户端版本下发客户端的登录信息
type ProtoClientNavigateReq struct {
	ClientVersion string `json:"ClientVersion"`
	PlatformName  string `json:"PlatformName"`
	ChannelName   string `json:"ChannelName"`
}

type ProtoClientNavigateRes struct {
	UrlLoginCheck string `json:"ULC"` // chksvrsc2.uqsoft.com:800 chksvrsc2audit.uqsoft.com:800
	UrlProxySvr   string `json:"UPS"` // proxysc2.uqsoft.com:6483 proxysc2audit.uqsoft.com:6483
	UrlSdkSvr     string `json:"USS"` // sdksvrsc2.uqsoft.com:80 sdksvrsc2audit.uqsoft.com:80
}
