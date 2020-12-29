package handler

import (
	"sailcraft/fleetsvr_main/config"
	"sailcraft/fleetsvr_main/custom_errors"
	"sailcraft/fleetsvr_main/errorcode"
	"sailcraft/fleetsvr_main/handlerbase"
	"sailcraft/fleetsvr_main/model"
	"sailcraft/fleetsvr_main/proto"
	"sailcraft/fleetsvr_main/table"
)

type AncientRelicsHandler struct {
	handlerbase.WebHandler
}

func (this *AncientRelicsHandler) Info() (int, error) {
	relicsList, err := GetGuildAncientRelicsList(this.Request.Uin)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	propProtypes := config.GPropConfig.GetPropByType(config.PROP_TYPE_PIECES)
	piecesProtypeIds := make([]int, 0, len(propProtypes))
	for _, protype := range propProtypes {
		piecesProtypeIds = append(piecesProtypeIds, protype.Id)
	}

	pieces, err := GetMultiPropsByProtypeId(this.Request.Uin, piecesProtypeIds...)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	piecesNotPlaced := make([]int, 0, len(pieces))
	for _, piece := range pieces {
		piecesNotPlaced = append(piecesNotPlaced, piece.ProtypeId)
	}

	var responseData proto.ProtoGetAncientRelicsInfoResopnse
	responseData.Relics = make([]*proto.ProtoAncientRelicsInfo, 0, len(config.GAncientRelicsConfig.AttrArr))
	for _, relics := range relicsList {
		protype, ok := config.GAncientRelicsConfig.AttrMap[relics.ProtypeId]
		if !ok {
			continue
		}

		protoRelics := new(proto.ProtoAncientRelicsInfo)
		composeProtoAncientRelics(protoRelics, relics, protype)
		responseData.Relics = append(responseData.Relics, protoRelics)
	}

	responseData.PiecesNotPlaced = append(responseData.PiecesNotPlaced, piecesNotPlaced...)

	this.Response.ResData.Params = responseData

	return 0, nil
}

func (this *AncientRelicsHandler) PlacePiece() (int, error) {
	var reqParams proto.ProtoPlaceAncientRelicsPieceRequest
	err := this.UnpackParams(&reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	relicsProtype, ok := config.GAncientRelicsConfig.AttrMap[reqParams.ProtypeId]
	if !ok {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("protype id is invalid")
	}

	// 1. 判断碎片是否足够
	allPiecesProtype := config.GPropConfig.GetPropByType(config.PROP_TYPE_PIECES)
	allPiecesProtypeId := make([]int, 0, len(allPiecesProtype))
	for _, protype := range allPiecesProtype {
		allPiecesProtypeId = append(allPiecesProtypeId, protype.Id)
	}

	allPieces, err := GetMultiPropsByProtypeId(this.Request.Uin, allPiecesProtypeId...)
	pieceMap := make(map[int]*table.TblPropInfo)
	for _, piece := range allPieces {
		pieceMap[piece.ProtypeId] = piece
	}

	for _, pieceProtypeId := range reqParams.Pieces {
		if piece, ok := pieceMap[pieceProtypeId]; !ok || piece.PropNum <= 0 {
			return errorcode.ERROR_CODE_RELICS_NOT_ENOUGH_PIECES, custom_errors.New("piece[%d] is not enough", pieceProtypeId)
		}
	}

	// 2. 判断上古遗物状态
	relics, err := GetGuildAncientRelics(this.Request.Uin, relicsProtype.Id)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if relics.Status != model.ANCIENT_RELICS_STATUS_UNCOMPLETED {
		return errorcode.ERROR_CODE_RELICS_ALREADY_COMPLETED, custom_errors.New("ancient relics has already installed all pieces")
	}

	// 3. 判断上古遗物碎片是否已经装过了
	for _, pieceId := range reqParams.Pieces {
		for _, existingPiece := range relics.Pieces.PiecesId {
			if pieceId == existingPiece {
				return errorcode.ERROR_CODE_RELICS_PIECES_ALREADY_INSTALLED, custom_errors.New("pieces[%d] of ancient relics has already installed", pieceId)
			}
		}
	}

	relics.Pieces.PiecesId = append(relics.Pieces.PiecesId, reqParams.Pieces...)

	// 4. 判断上古遗物是否组装已经完成，由于第4步已经做了重复判断，这里只需要判断长度
	if len(relics.Pieces.PiecesId) == len(relicsProtype.NeedKeys) {
		relics.Status = model.ANCIENT_RELICS_STATUS_COMPLETED
	}

	retCode, err := UpdateAncientRelics(this.Request.Uin, relics)
	if err != nil {
		return retCode, err
	}

	var responseData proto.ProtoPlaceAncientRelicsPieceResponse
	responseData.Cost = new(proto.ProtoResourcesAttr)
	responseData.Cost.BattleShipCards = make([]*proto.ProtoBattleShipCardItem, 0)
	responseData.Cost.ResourceItems = make([]*proto.ProtoResourceItem, 0)
	responseData.Cost.PropItems = make([]*proto.ProtoPropItem, 0, len(reqParams.Pieces))
	for _, piece := range reqParams.Pieces {
		protoPropItem := new(proto.ProtoPropItem)
		protoPropItem.ProtypeId = piece
		protoPropItem.CountType = config.COUNT_TYPE_CONST
		protoPropItem.Count = 1

		responseData.Cost.PropItems = append(responseData.Cost.PropItems, protoPropItem)
	}

	protoRelics := new(proto.ProtoAncientRelicsInfo)
	composeProtoAncientRelics(protoRelics, relics, relicsProtype)
	responseData.Relics = append(responseData.Relics, protoRelics)

	needPieceMap := make(map[int]int)
	for _, pieceProtypeId := range reqParams.Pieces {
		needPieceMap[pieceProtypeId] = pieceProtypeId
	}
	responseData.PiecesNotPlaced = make([]int, 0, len(allPieces))
	for _, piece := range allPieces {
		if _, ok := needPieceMap[piece.ProtypeId]; !ok {
			responseData.PiecesNotPlaced = append(responseData.PiecesNotPlaced, piece.ProtypeId)
		}
	}

	this.Response.ResData.Params = responseData

	return 0, nil
}

func (this *AncientRelicsHandler) RecieveRelicsReward() (int, error) {
	var reqParams proto.ProtoReceiveRelicsRewardRequest
	err := this.UnpackParams(&reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	protype, ok := config.GAncientRelicsConfig.AttrMap[reqParams.ProtypeId]
	if !ok {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("can not found relics protype[%d]", reqParams.ProtypeId)
	}

	relics, err := GetGuildAncientRelics(this.Request.Uin, reqParams.ProtypeId)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	switch relics.Status {
	case model.ANCIENT_RELICS_STATUS_UNCOMPLETED:
		return errorcode.ERROR_CODE_RELICS_UNCOMPLETED, custom_errors.New("ancient relics has not completed installation")
	case model.ANCIENT_RELICS_STATUS_RECEIVED:
		return errorcode.ERROR_CODE_RELICS_ALREADY_RECEIVED, custom_errors.New("reward of ancient relics has already received")
	}

	relics.Status = model.ANCIENT_RELICS_STATUS_RECEIVED
	retCode, err := UpdateAncientRelics(this.Request.Uin, relics)
	if err != nil {
		return retCode, err
	}

	var responseData proto.ProtoReceiveRelicsRewardResponse

	protoRelics := new(proto.ProtoAncientRelicsInfo)
	composeProtoAncientRelics(protoRelics, relics, protype)
	responseData.Relics = append(responseData.Relics, protoRelics)

	responseData.Reward = new(proto.ProtoResourcesAttr)
	ResourcesConfigToProto(&protype.Reward, responseData.Reward)

	this.Response.ResData.Params = responseData

	return 0, nil
}

func GetGuildAncientRelicsList(uin int) ([]*table.TblGuildAncientRelicsInfo, error) {
	ancientRelicsModel := model.GuildAncientRelicsModel{Uin: uin}
	relicsList, err := ancientRelicsModel.GetRelicsList()
	if err != nil {
		return nil, err
	}

	relicsMap := make(map[int]*table.TblGuildAncientRelicsInfo)
	for _, relics := range relicsList {
		relicsMap[relics.ProtypeId] = relics
	}

	ret := make([]*table.TblGuildAncientRelicsInfo, 0, len(config.GAncientRelicsConfig.AttrArr))
	for _, protype := range config.GAncientRelicsConfig.AttrArr {
		if relics, ok := relicsMap[protype.Id]; ok {
			ret = append(ret, relics)
			continue
		}

		relicsInfo := new(table.TblGuildAncientRelicsInfo)
		relicsInfo.ProtypeId = protype.Id
		relicsInfo.Uin = uin
		relicsInfo.Status = model.ANCIENT_RELICS_STATUS_UNCOMPLETED
		ret = append(ret, relicsInfo)
	}

	return ret, nil
}

func GetGuildAncientRelics(uin, protypeId int) (*table.TblGuildAncientRelicsInfo, error) {
	ancientRelicsModel := model.GuildAncientRelicsModel{Uin: uin}
	relics, err := ancientRelicsModel.GetRelics(protypeId)
	if err != nil {
		return nil, err
	}

	if relics != nil {
		return relics, nil
	}

	if protype, ok := config.GAncientRelicsConfig.AttrMap[protypeId]; ok {
		relics = new(table.TblGuildAncientRelicsInfo)
		relics.ProtypeId = protype.Id
		relics.Uin = uin
		relics.Status = model.ANCIENT_RELICS_STATUS_UNCOMPLETED

		return relics, nil
	} else {
		return nil, custom_errors.New("protype id is invalid")
	}
}

func UpdateAncientRelics(uin int, record *table.TblGuildAncientRelicsInfo) (int, error) {
	if record == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.NullPoint()
	}

	ancientRelicsModel := model.GuildAncientRelicsModel{Uin: uin}
	if record.Id > 0 {
		return ancientRelicsModel.UpdateReclis(record)
	} else {
		return ancientRelicsModel.AddReclis(record)
	}
}

func composeProtoAncientRelics(target *proto.ProtoAncientRelicsInfo, data *table.TblGuildAncientRelicsInfo, protype *config.AncientRelicsProtype) (int, error) {
	if target == nil || (data == nil && protype == nil) {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.NullPoint()
	}

	if data != nil {
		target.ProtypeId = data.ProtypeId
		target.Status = data.Status
		target.Pieces = data.Pieces.PiecesId
		if len(target.Pieces) <= 0 {
			target.Pieces = make([]int, 0)
		}
	} else {
		target.ProtypeId = data.ProtypeId
		target.Pieces = make([]int, 0)
		target.Status = model.ANCIENT_RELICS_STATUS_UNCOMPLETED
	}

	return 0, nil
}
