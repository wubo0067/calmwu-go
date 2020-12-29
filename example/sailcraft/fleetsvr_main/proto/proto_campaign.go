package proto

type ProtoPassCampaignChapterRequest struct {
	CampaignID int `json:"CampaignID"`
}

type ProtoPassCampaignChapterInfo struct {
	CampaignID      int `json:"CampaignID"`
	FirstTimeToPass int `json:"FirstTimeToPass"`
}

type ProtoPassCampaignChapterResponse struct {
	CampaignInfo ProtoPassCampaignChapterInfo `json:"CampaignInfo"`
}

type ProtoMaxCampaignChapterResponse struct {
	CampaignProgressInfo ProtoPassCampaignChapterInfo `json:"CampaignInfo"`
}

type ProtoCampaignInfo struct {
	CampaignId    int                             `json:"CampaignID"`
	Events        []*ProtoCampaignEventInfo       `json:"CampaignEvents"`
	ProduceStatus *ProtoCampaignProduceStatusInfo `json:"CampaignProduceResourceInfo"`
}

type ProtoQueryCampaignResponse struct {
	CampaignInfo ProtoCampaignInfo `json:"CampaignInfo"`
}

type ProtoOnEventHappenedRequest struct {
	Events []ProtoEventInfo `json:"Events"`
}
