package proto

type ProtoGrowupTaskInfo struct {
	ProtypeId string `json:"ProtypeId"`
	Current   int    `json:"Current"`
	Total     int    `json:"Total"`
	Status    int    `json:"Status"`
}

type ProtoGetGrowupTaskListResponse struct {
	GrowupTaskList []*ProtoGrowupTaskInfo `json:"GrowupTaskList"`
}

type ProtoReceiveGrowupTaskRewardRequest struct {
	ProtypeId string `json:"ProtypeId"`
}

type ProtoReceiveGrowupTaskRewardResponse struct {
	TaskList []*ProtoGrowupTaskInfo `json:"TaskList"`
	Rewards  ProtoResourcesAttr     `json:"Rewards"`
}
