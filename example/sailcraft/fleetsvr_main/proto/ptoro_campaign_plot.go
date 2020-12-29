package proto

type ProtoCampaignPlotInfo struct {
	ProtypeId int `json:"ProtypeId"`
	PassTime  int `json:"PassTime"`
}

type ProtoGetCampaignPlotListResponse struct {
	CampaignPlotList []*ProtoCampaignPlotInfo `json:"CampaignPlotList"`
}

type ProtoPassCampaignPlotRequest struct {
	ProtypeId int `json:"ProtypeId"`
}

type ProtoPassCampaignPlotResponse struct {
	CampaignPlotList []*ProtoCampaignPlotInfo `json:"CampaignPlotList"`
}
