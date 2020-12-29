package proto

type ProtoAchievementInfo struct {
	ProtypeId       int `json:"ProtypeId"`
	CurrentProgress int `json:"Current"`
	TotalProgress   int `json:"Total"`
	Status          int `json:"Status"`
	CompleteTime    int `json:"CompleteTime"`
}

type ProtoGetAchievementListResponse struct {
	Achievements []*ProtoAchievementInfo `json:"Achievements"`
}
