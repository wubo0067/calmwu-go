package proto

type ProtoAncientRelicsInfo struct {
	ProtypeId int   `json:"ProtypeId"`
	Pieces    []int `json:"Pieces"`
	Status    int   `json:"Status"`
}

type ProtoGetAncientRelicsInfoResopnse struct {
	Relics          []*ProtoAncientRelicsInfo `json:"AncientRelics"`
	PiecesNotPlaced []int                     `json:"PiecesNotPlaced"`
}

type ProtoPlaceAncientRelicsPieceRequest struct {
	ProtypeId int   `json:"ProtypeId"`
	Pieces    []int `json:"Pieces"`
}

type ProtoPlaceAncientRelicsPieceResponse struct {
	Relics          []*ProtoAncientRelicsInfo `json:"AncientRelics"`
	PiecesNotPlaced []int                     `json:"PiecesNotPlaced"`
	Cost            *ProtoResourcesAttr       `json:"Cost"`
}

type ProtoReceiveRelicsRewardRequest struct {
	ProtypeId int `json:"ProtypeId"`
}

type ProtoReceiveRelicsRewardResponse struct {
	Relics []*ProtoAncientRelicsInfo `json:"AncientRelics"`
	Reward *ProtoResourcesAttr       `json:"Reward"`
}
