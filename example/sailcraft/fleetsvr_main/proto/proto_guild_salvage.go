package proto

type ProtoGuildSalvageInfo struct {
	Uin           int                           `json:"Uin"`
	SalvageTimes  int                           `json:"SalvageTimes"`  // 剩余打捞次数
	RestTime      int                           `json:"RestTime"`      // 打捞次数恢复时间
	SalvagePoints []*ProtoGuildSalvagePointInfo `json:"SalvagePoints"` // 打捞点信息
}

type ProtoGuildSalvagePointInfo struct {
	ProtypeId      int   `json:"ProtypeId"`      // 打捞点ProtypeId
	SalvagedPieces []int `json:"SalvagedPieces"` // 打捞点打捞到的碎片ProtypeId数组
}

type ProtoGetSalvageInfoResponse struct {
	SalvageInfo ProtoGuildSalvageInfo `json:"SalvageInfo"` // 打捞信息
	Nets        []*ProtoPropItem      `json:"Nets"`        // 打捞网信息
}

type ProtoSalvageRequest struct {
	ProtypeId    int `json:"ProtypeId"`
	NetProtypeId int `json:"NetProtypeId"`
}

type ProtoSalvageResponse struct {
	SalvageInfo ProtoGuildSalvageInfo  `json:"SalvageInfo"` // 打捞信息
	Nets        []*ProtoPropItem       `json:"Nets"`        // 打捞网信息
	Cost        ProtoResourcesAttr     `json:"Cost"`        // 消耗
	Reward      *ProtoResourcesAttr    `json:"Reward"`      // 获得
	GuildTask   map[string]interface{} `json:"GuildTask"`   // 公会任务更新
}
