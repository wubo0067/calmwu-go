/*
 * @Author: calmwu
 * @Date: 2018-01-11 10:48:59
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-06-13 15:33:38
 */

package proto

import "sailcraft/base"

//------------------------------------------------------------------------------------------------------------
type CassandraProcResultS struct {
	Ok     bool
	Result interface{}
}

type CassandraProcDataS struct {
	RemoteIP string
	ReqData  *base.ProtoRequestS
	// 处理结果通道，等待返回
	ResultChan chan<- *CassandraProcResultS
}

//------------------------------------------------------------------------------------------------------------

const (
	APINAMECssSvrUserLogin         = "CssSvrUserLogin"
	APINAMECssSvrUserLogout        = "CssSvrUserLogout"
	APINAMECssSvrUserRecharge      = "CssSvrUserRecharge"
	APINAMECssSvrTuitionStepReport = "TuitionStepReport"

	// 战斗录像
	APINAMECssSvrUploadBattleVideo = "UploadBattleVideo"
	APINAMECssSvrDeleteBattleVideo = "DeleteBattleVideo"
	APINAMECssSvrGetBattleVideo    = "GetBattleVideo"
	// 查询用户国家
	APINAMECssSvrQueryPlayerGeo = "SvrQueryPlayerGeo"
	// 老用户补偿等级
	APINAMECssSvrOldUserReceiveCompensation = "OldUserReceiveCompensation"
	// CDN资源下载
	APINAMECssSvrClientCDNResourceDownloadReport = "ClientCDNResourceDownloadReport"
	//
	APINAMECssSvrUploadUserAction = "UploadUserAction"
	//
	APINAMECssSvrQueryUserRechargeInfo = "QueryUserRechargeInfo"
)

type ProtoCssSvrUserLoginNtf struct {
	Uin              uint64 `json:"Uin" mapstructure:"Uin"`
	ClientInternetIP string `json:"ClientInternetIP"` // 玩家外网IP
	Platform         string `json:"Platform"`         // ios/android
}

type ProtoCssSvrUserLogoutNtf struct {
}

// 用户充值上报
type ProtoCssSvrUserRechargeNtf struct {
	Uin            uint64  `json:"Uin" mapstructure:"Uin"`
	RechargeAmount float32 `json:"RechargeAmount"` // 充值金额
	ChannelID      string  `json:"ChannelID"`      // CN US
	PlatForm       string  `json:"PlatForm"`       // IOS ANDROID
}

type ProtoGetBattleVideoParamsResS struct {
	BattleVideoID string `json:"BattleVideoID"`
	VideoContent  string `json:"VideoContent"`
}

type ProtoSvrQueryISOCountryCodesByUinsParamsReqS struct {
	Count int   `json:"Count"`
	Uins  []int `json:"Uins"`
}

type ProtoPlayerGeoS struct {
	Uin            int    `mapstructure:"uin"`
	IsoCountryCode string `mapstructure:"registerregion"`
}

type ProtoSvrQueryISOCountryCodesByUinsParamsResS struct {
	Count           int                `json:"Count"`
	ProtoPlayerGeos []*ProtoPlayerGeoS `json:"ProtoPlayerGeos"`
}

type ProtoQueryCountryISOByIpReq struct {
	Uin      int    `mapstructure:"uin"`
	ClientIP string `mapstructure:"ClientIP"`
}

type ProtoQueryCountryISOByIpRes struct {
	Uin            int    `mapstructure:"uin"`
	ClientIP       string `mapstructure:"ClientIP"`
	CountryISOCode string `mapstructure:"CountryISOCode"`
	CountryName    string `mapstructure:"CountryName"`
}

type ProtoTuitionStepReportParamsS struct {
	ClientVersion string `json:"ClientVersion"`
	StepId        int    `json:"StepId"`
	PlatformName  string `json:"PlatformName"`
	ChannelName   string `json:"ChannelName"`
}

type ProtoOldUserReceiveCompensationReq struct {
	DeviceID string `json:"DeviceID"`
}

type ProtoOldUserReceiveCompensationRes struct {
	DeviceID string `json:"DeviceID"`
	Result   int    `json:"Result"` // 0: 领取成功，-1：领取失败，已经领取过
	Level    int    `json:"Level"`  // 1、2、3
}

type ProtoClientCDNDownloadReportNtf struct {
	ClientVersion string `json:"ClientVersion"`
	ResourceName  string `json:"ResourceName"`
	ResourceID    int    `json:"ResourceID"`
	ElapseTime    int    `json:"ElapseTime"`
	AttemptCount  int    `json:"AttemptCount"`
	PlatformName  string `json:"PlatformName"`
	ChannelName   string `json:"ChannelName"`
}

type ProtoUserActionReportNtf struct {
	Uin              int    `json:"Uin"`
	ActionName       string `json:"ActionName"`
	DiamondCostCount int    `json:"DiamondCostCount"` // 如果是钻石消耗事件，就带上消耗的数量
}

type ProtoQueryUserRechargeInfoReq struct {
	Uin int `json:"Uin"`
}

type ProtoQueryUserRechargeInfoRes struct {
	Uin                 int     `json:"Uin"`
	MaxRechargeAmount   float32 `json:"MaxRechargeAmount"`
	TotalRechargeAmount float32 `json:"TotalRechargeAmount"`
	TotalRechargeCount  int     `json:"TotalRechargeCount"`
}

type ProtoVerifySignReq struct {
	Serial     string `json:"Serial"`   // 订单号
	SignString string `json:SignString` // key=value的键值对用&连接起来
	AuthString string `json:AuthString` // 对比数据
}

type ProtoVerifySignRes struct {
	Serial       string `json:"Serial"`     // 订单号
	VerifyResult int    `json:VerifyResult` // 鉴权结果 -1失败，0成功
}
