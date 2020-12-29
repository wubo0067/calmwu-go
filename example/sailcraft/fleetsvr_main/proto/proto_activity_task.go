package proto

type ProtoActivityTaskInfo struct {
	ProtypeId int `json:"ProtypeId"`
	Current   int `json:"Current"`
	Total     int `json:"Total"`
	Status    int `json:"Status"`
}

type ProtoActivityScoreRewardInfo struct {
	RewardId string `json:"RewardId"`
	Status   int    `json:"Status"`
}

type ProtoGetActivityInfoResponse struct {
	Uin                 int                             `json:"Uin"`
	Vitality            int                             `json:"Vitality"`
	RestTimeToReset     int                             `json:"RestTimeToReset"`
	ActivityTaskFreshed int                             `json:"ActivityTaskFreshed"`
	TaskList            []*ProtoActivityTaskInfo        `json:"TaskList"`
	RewardList          []*ProtoActivityScoreRewardInfo `json:"RewardList"`
}

type ProtoGetActivitySimpleInfoResponse struct {
	Uin             int `json:"Uin"`
	Vitality        int `json:"Vitality"`
	RestTimeToReset int `json:"RestTimeToReset"`
}

type ProtoReceiveActivityTaskRewardRequest struct {
	ProtypeId int `json:"ProtypeId"`
}

type ProtoReceiveActivityTaskRewardResponse struct {
	Vitality              int                             `json:"Vitality"`
	TaskList              []*ProtoActivityTaskInfo        `json:"TaskList"`
	Rewards               ProtoResourcesAttr              `json:"Rewards"`
	NewVitalityReward     int                             `json:"NewVitalityReward"`
	NewVitalityRewardList []*ProtoActivityScoreRewardInfo `json:"NewVitalityRewardList"`
}

type ProtoReceiveActivityScoreRewardRequest struct {
	RewardId string `json:"RewardId"`
}

type ProtoReceiveActivityScoreRewardResponse struct {
	RewardList []*ProtoActivityScoreRewardInfo `json:"RewardList"`
	Rewards    ProtoResourcesAttr              `json:"Rewards"`
}

type ProtoGetActivityTaskStatusNormalResponse struct {
	Status   int                    `json:"Status"`
	TaskInfo *ProtoActivityTaskInfo `json:"TaskInfo"`
}

type ProtoGetActivityTaskStatusNothingResponse struct {
	Status int `json:"Status"`
}
