package proto

type ProtoGuildTaskInfo struct {
	Uin        int                         `json:"Uin"`
	Vitality   int                         `json:"Vitality"`
	RestTime   int                         `json:"RestTime"`
	RewardList []*ProtoGuildTaskRewardInfo `json:"RewardList"`
}

type ProtoGuildTaskRewardInfo struct {
	ProtypeId int `json:"ProtypeId"`
	Score     int `json:"Score"`
	Status    int `json:"Status"`
}

type ProtoGuildTaskInfoResponse struct {
	ProtoGuildTaskInfo `json:",squash"`
}

type ProtoReceiveVitalityRewardRequest struct {
	ProtypeId int `json:"ProtypeId"`
}

type ProtoReceiveVitalityRewardResponse struct {
	RewardList []*ProtoGuildTaskRewardInfo `json:"RewardList"`
	Rewards    ProtoResourcesAttr          `json:"Rewards"`
}
