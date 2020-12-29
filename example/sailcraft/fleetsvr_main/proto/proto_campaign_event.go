package proto

type ProtoCampaignEventProgress struct {
	Progress int `json:"Progress"`
	Total    int `json:"Total"`
}

type ProtoCampaignEventInfo struct {
	Id          int                        `json:"Id"`
	Uin         int                        `json:"Uin"`
	CampaignId  int                        `json:"CampaignId"`
	EventId     int                        `json:"EventProtypeId"`
	StartTime   int                        `json:"StartTime"`
	ResetTime   int                        `json:"RestTime"`
	MissionId   int                        `json:"MissionId"`
	Progress    ProtoCampaignEventProgress `json:"Progress"`
	EventStatus int                        `json:"EventStatus"`
	TriggerType int                        `json:"TriggerType"`
}

type ProtoQueryCampaignEventUnifishedResponse struct {
	CampaignEventsUnfinished []*ProtoCampaignEventInfo `json:"CampaignEvents"`
}

type ProtoCampaignEventTriggered struct {
	NewCampaingEvent ProtoCampaignEventInfo `json:"TriggeredCampaignEvent"`
}

type ProtoFinishCampaignEventRequest struct {
	Id int `json:"Id"`
}

type ProtoFinishCampaignEventResponse struct {
	//ResourcesChanged ProtoResourcesAttr `json:"ResourcesChanged"`
	Cost   ProtoResourcesAttr `json:"Cost"`
	Reward ProtoResourcesAttr `json:"Reward"`
}

type ProtoEventInfo struct {
	EventName string      `json:"EventName"`
	EventData interface{} `json:"EventData"`
}
