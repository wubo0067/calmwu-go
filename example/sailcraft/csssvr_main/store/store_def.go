/*
 * @Author: calmwu
 * @Date: 2018-01-11 15:24:31
 * @Last Modified by:   calmwu
 * @Last Modified time: 2018-01-11 15:24:31
 * @Comment:
 */

package store

const (
	TBNAME_USERONLINE    = "tbl_UserOnline"
	TBNAME_USERMATCHINFO = "tbl_UserMatchInfo"
	TBNAME_USERRECHARGE  = "tbl_UserTotalRecharge"

	MAX_BATCHCDKEY_COUNT       = 500
	MAX_RECENTCDKEYBATCH_COUNT = 100
	MAX_GEOUIN_COUNT           = 50
)

type CDKeyStatusT int

const (
	ECDKEY_STATUS_ACTIVATE = 1
	ECDKEY_STATUS_RECEIVED = 2
	ECDKEY_STATUS_ABOLISH  = 3
	ECDKEY_STATUS_EXPIRE   = 4
)

type TblUserOnineS struct {
	Uin             int    `mapstructure:"Uin"`
	CreateTime      int64  `mapstructure:"CreateTime"`
	ISOCountryCode  string `mapstructure:"ISOCountryCode"`
	LoginTime       int64  `mapstructure:"LoginTime"`
	LogoutTime      int64  `mapstructure:"LogoutTime"`
	MaxOnlinetime   int64  `mapstructure:"MaxOnlinetime"`
	TotalOnlineTime int64  `mapstructure:"TotalOnlineTime"`
	VersionID       int    `mapstructure:"VersionID"`
	Platform        string `mapstructure:"Platform"`
}

type TblUserTotalRechargeS struct {
	Uin           int     `mapstructure:"Uin"`
	ChannelID     string  `mapstructure:"ChannelID"`
	TotalCost     float32 `mapstructure:"TotalCost"`
	RechargeCount int     `mapstructure:"RechargeCount"`
}

type TblUserMatchInfoS struct {
	Uin                 int         `mapstructure:"uin"`
	MatchCount          int         `mapstructure:"matchcount"`
	MatchTotalTime      int         `mapstructure:"matchtotaltime"`
	MatchMaxDuration    int         `mapstructure:"MatchMaxDuration"`
	MatchWinCount       int         `mapstructure:"matchwincount"`
	MatchLostCount      int         `mapstructure:"matchlostcount"`
	MatchTieCount       int         `mapstructure:"matchtiecount"`
	MatchSurrenderCount int         `mapstructure:"matchsurrendercount"`
	MatchEscapeCount    int         `mapstructure:"matchescapecount"`
	MatchShipStatistics map[int]int `mapstructure:"matchshipstatistics"`
	RecentlyUseShips    []int       `mapstructure:"RecentlyUseShips"`
	LineUpStatistics    map[int]int `mapstructure:"lineupstatistics"`
	VersionID           int         `mapstructure:"versionid"`
}

type UserMatchParamsS struct {
	MatchDuration int   `json:"MatchDuration"`
	MatchResult   int   `json:"MatchResult"`
	ShipIDs       []int `json:"ShipIDs"`
	LineUpId      int   `json:"LineUpId"`
}
