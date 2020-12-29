package proto

type ProtoCampaignProduceResourceInfo struct {
	Uin                  int                `json:"Uin"`
	LastReceivedTime     int                `json:"LastReceivedTime"`
	CanReceive           int                `json:"CanReceive"`
	RestTimeToReceive    int                `json:"RestTimeToReceive"`
	TotalProduceResource ProtoResourcesAttr `json:"TotalProduceResource"`
}

type ProtoCampaignProduceStatusInfo struct {
	Uin               int `json:"Uin"`
	LastReceivedTime  int `json:"LastReceivedTime"`
	CanReceive        int `json:"CanReceive"`
	RestTimeToReceive int `json:"RestTimeToReceive"`
}

type ProtoQueryCampaignProduceResourceResponse struct {
	CampaginProduceResourceInfo ProtoCampaignProduceResourceInfo `json:"CampaignProduceResourceInfo"`
}

type ProtoQueryCampaignProduceStatusResponse struct {
	CampaignProduceStatusInfo ProtoCampaignProduceStatusInfo `json:"CampaignProduceResourceInfo"`
}

type ProtoReceiveCampaignProduceResourceResponse struct {
	CampaignProduceStatusInfo ProtoCampaignProduceStatusInfo `json:"CampaignProduceResourceInfo"`
}
