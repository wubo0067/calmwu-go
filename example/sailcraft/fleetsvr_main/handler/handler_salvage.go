package handler

import (
	"math/rand"
	"sailcraft/base"
	"sailcraft/fleetsvr_main/config"
	"sailcraft/fleetsvr_main/custom_errors"
	"sailcraft/fleetsvr_main/errorcode"
	"sailcraft/fleetsvr_main/handlerbase"
	"sailcraft/fleetsvr_main/model"
	"sailcraft/fleetsvr_main/proto"
	"sailcraft/fleetsvr_main/table"
)

type SalvagePointInfo struct {
	ProtypeId      int   `json:"protype_id"`
	SalvagedPieces []int `json:"salvaged_pieces"`
}

type SalvageHandler struct {
	handlerbase.WebHandler
}

func (this *SalvageHandler) Info() (int, error) {
	// 1. 获取打捞信息
	salvageInfo, err := GetSalvageInfo(this.Request.Uin)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	// 2. 获取打捞网以及碎片道具信息
	propProtypes := config.GPropConfig.GetPropByType(config.PROP_TYPE_NET, config.PROP_TYPE_PIECES)
	propProtypeIds := make([]int, 0, len(propProtypes))
	for _, protype := range propProtypes {
		propProtypeIds = append(propProtypeIds, protype.Id)
	}

	props, err := GetMultiPropsByProtypeId(this.Request.Uin, propProtypeIds...)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	nets := make([]*table.TblPropInfo, 0)
	pieces := make([]int, 0)
	for _, prop := range props {
		if propProtype, ok := config.GPropConfig.AttrMap[prop.ProtypeId]; ok {
			switch propProtype.PropType {
			case config.PROP_TYPE_NET:
				nets = append(nets, prop)
			case config.PROP_TYPE_PIECES:
				pieces = append(pieces, prop.ProtypeId)
			}
		}
	}

	// 3. 获取上古遗物碎片信息
	relicsList, err := GetGuildAncientRelicsList(this.Request.Uin)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}
	for _, relics := range relicsList {
		pieces = append(pieces, relics.Pieces.PiecesId...)
	}

	var responseData proto.ProtoGetSalvageInfoResponse
	retCode, err := composeProtoSalvageInfo(&responseData.SalvageInfo, salvageInfo, pieces...)
	if err != nil {
		return retCode, err
	}

	responseData.Nets = make([]*proto.ProtoPropItem, 0, len(nets))
	for _, net := range nets {
		protoProtoItem := new(proto.ProtoPropItem)
		protoProtoItem.ProtypeId = net.ProtypeId
		protoProtoItem.CountType = config.COUNT_TYPE_CONST
		protoProtoItem.Count = net.PropNum
		responseData.Nets = append(responseData.Nets, protoProtoItem)
	}

	this.Response.ResData.Params = responseData

	return 0, nil
}

func (this *SalvageHandler) Salvage() (int, error) {
	var reqParams proto.ProtoSalvageRequest
	err := this.UnpackParams(&reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if reqParams.NetProtypeId < 0 || reqParams.ProtypeId < 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("protype id is invalid")
	}

	netProtype, ok := config.GPropConfig.AttrMap[reqParams.NetProtypeId]
	if !ok {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("net protype not found")
	}

	salvageProtype, ok := config.GGuildSalvageConfig.AttrMap[reqParams.ProtypeId]
	if !ok {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("salvage point protype not found")
	}

	// 1. 获取打捞信息，判断剩余打捞次数
	salvageInfo, err := GetSalvageInfo(this.Request.Uin)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if salvageInfo.RestTimes <= 0 {
		return errorcode.ERROR_CODE_SALVAGE_NOT_ENOUGH_TIMES, custom_errors.New("guild salvage times is not enough")
	}

	// 2. 获取打捞网以及碎片信息
	propProtypes := config.GPropConfig.GetPropByType(config.PROP_TYPE_NET, config.PROP_TYPE_PIECES)
	propProtypeIds := make([]int, 0, len(propProtypes))
	for _, propProtype := range propProtypes {
		propProtypeIds = append(propProtypeIds, propProtype.Id)
	}

	props, err := GetMultiPropsByProtypeId(this.Request.Uin, propProtypeIds...)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	// 3. 打捞网和碎片分类
	var selectedNet *table.TblPropInfo
	nets := make([]*table.TblPropInfo, 0)
	piecesProtypeIds := make([]int, 0)
	for _, prop := range props {
		if propProtype, ok := config.GPropConfig.AttrMap[prop.ProtypeId]; ok {
			switch propProtype.PropType {
			case config.PROP_TYPE_PIECES:
				piecesProtypeIds = append(piecesProtypeIds, prop.ProtypeId)
			case config.PROP_TYPE_NET:
				if prop.ProtypeId == reqParams.NetProtypeId {
					selectedNet = prop
				}

				nets = append(nets, prop)
			}
		}
	}

	if selectedNet == nil || selectedNet.PropNum <= 0 {
		return errorcode.ERROR_CODE_SALVAGE_NOT_ENOUGH_NETS, custom_errors.New("net[%d] is not enough", reqParams.NetProtypeId)
	}

	selectedNet.PropNum--

	// 4. 获取上古遗物碎片信息
	relicsList, err := GetGuildAncientRelicsList(this.Request.Uin)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}
	for _, relics := range relicsList {
		piecesProtypeIds = append(piecesProtypeIds, relics.Pieces.PiecesId...)
	}

	// 5. 打捞网效果
	var netEffect config.PropNetEffect
	err = config.GPropConfig.DecodeEffect(netProtype, &netEffect)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("mapstructure decode net effect[%d] failed", netProtype.Id)
	}

	// 6. 打捞
	obtained, err := randomForSalvage(salvageProtype, &netEffect, piecesProtypeIds...)

	for _, prop := range obtained.PropItems {
		propProtype, ok := config.GPropConfig.AttrMap[prop.ProtypeId]
		if !ok {
			continue
		}

		if propProtype.PropType == config.PROP_TYPE_PIECES {
			piecesProtypeIds = append(piecesProtypeIds, propProtype.Id)
		}

		if propProtype.PropType == config.PROP_TYPE_NET {
			var existsNet *table.TblPropInfo
			for _, net := range nets {
				if net.ProtypeId == prop.ProtypeId {
					existsNet = net
				}
			}

			if existsNet != nil {
				existsNet.PropNum += prop.Count
			} else {
				obtainedNet := new(table.TblPropInfo)
				obtainedNet.PropNum = prop.Count
				obtainedNet.ProtypeId = prop.ProtypeId
				obtainedNet.Uin = this.Request.Uin
				nets = append(nets, obtainedNet)
			}
		}
	}

	salvageInfo.RestTimes--
	if salvageInfo.LastNotEnoughTime < 0 {
		salvageInfo.LastNotEnoughTime = int(base.GLocalizedTime.SecTimeStamp())
	}
	retCode, err := UpdateSalvageInfo(this.Request.Uin, salvageInfo)
	if err != nil {
		return retCode, err
	}

	var responseData proto.ProtoSalvageResponse
	// 道具消耗
	responseData.Cost.BattleShipCards = make([]*proto.ProtoBattleShipCardItem, 0)
	responseData.Cost.ResourceItems = make([]*proto.ProtoResourceItem, 0)
	protoPropItem := new(proto.ProtoPropItem)
	protoPropItem.ProtypeId = reqParams.NetProtypeId
	protoPropItem.Count = 1
	protoPropItem.CountType = config.COUNT_TYPE_CONST
	responseData.Cost.PropItems = append(responseData.Cost.PropItems, protoPropItem)
	// 道具获得
	responseData.Reward = obtained

	// 打捞网信息
	responseData.Nets = make([]*proto.ProtoPropItem, 0, len(nets))
	for _, net := range nets {
		propNet := new(proto.ProtoPropItem)
		composeProtoPropInfo(propNet, net)
		responseData.Nets = append(responseData.Nets, propNet)
	}

	// 打捞信息
	retCode, err = composeProtoSalvageInfo(&responseData.SalvageInfo, salvageInfo, piecesProtypeIds...)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	// 打捞会增加活跃度
	eventDataSet := new(EventsHappenedDataSet)
	eventDataSet.SalvageData = new(OnSalvage)
	eventResParams := make(map[string]interface{})
	retCode, err = HandleGuildMission(this.Request, &eventResParams, eventDataSet)
	if err != nil {
		return retCode, err
	}

	responseData.GuildTask = eventResParams
	this.Response.ResData.Params = responseData

	return 0, nil
}

func GetSalvageInfo(uin int) (*table.TblGuildSalvage, error) {
	guildSalvageModel := model.GuildSalvageModel{Uin: uin}
	salvageInfo, err := guildSalvageModel.GetSalvage()
	if err != nil {
		return nil, err
	}

	if salvageInfo == nil {
		salvageInfo = new(table.TblGuildSalvage)
		salvageInfo.Uin = uin
		salvageInfo.RestTimes = config.GGlobalConfig.Guild.LimitSalvageTimes
		salvageInfo.LastNotEnoughTime = -1000
	} else {
		if salvageInfo.RestTimes < config.GGlobalConfig.Guild.LimitSalvageTimes {
			diff := int(base.GLocalizedTime.SecTimeStamp()) - salvageInfo.LastNotEnoughTime
			restTimes := salvageInfo.RestTimes + diff/config.GGlobalConfig.Guild.SalvageRecoverTime
			salvageInfo.LastNotEnoughTime += (restTimes - salvageInfo.RestTimes) * config.GGlobalConfig.Guild.SalvageRecoverTime

			if restTimes > config.GGlobalConfig.Guild.LimitSalvageTimes {
				salvageInfo.RestTimes = config.GGlobalConfig.Guild.LimitSalvageTimes
			} else {
				salvageInfo.RestTimes = restTimes
			}
		}

		if salvageInfo.RestTimes >= config.GGlobalConfig.Guild.LimitSalvageTimes {
			salvageInfo.LastNotEnoughTime = -1000
		}
	}

	return salvageInfo, nil
}

func UpdateSalvageInfo(uin int, salvageInfo *table.TblGuildSalvage) (int, error) {
	if salvageInfo == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.NullPoint()
	}

	salvageInfo.Uin = uin
	guildSalvageModel := model.GuildSalvageModel{Uin: uin}
	if salvageInfo.Id <= 0 {
		return guildSalvageModel.AddSalvage(salvageInfo)
	} else {
		return guildSalvageModel.UpdateSalvage(salvageInfo)
	}
}

func composeProtoSalvageInfo(target *proto.ProtoGuildSalvageInfo, data *table.TblGuildSalvage, pieces ...int) (int, error) {
	if target == nil || data == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.NullPoint()
	}

	target.Uin = data.Uin
	target.SalvageTimes = data.RestTimes
	if data.LastNotEnoughTime > 0 {
		target.RestTime = config.GGlobalConfig.Guild.SalvageRecoverTime - (int(base.GLocalizedTime.SecTimeStamp()) - data.LastNotEnoughTime)
	} else {
		target.RestTime = -1000
	}

	pieceMap := make(map[int]int)
	for _, piece := range pieces {
		pieceMap[piece] = piece
	}

	for _, salvageProtype := range config.GGuildSalvageConfig.AttrArr {
		if salvagePoolProtype, ok := config.GGuildSalvagePoolConfig.AttrMap[salvageProtype.PiecesPool]; ok {
			protoSalvagePoint := new(proto.ProtoGuildSalvagePointInfo)
			protoSalvagePoint.ProtypeId = salvageProtype.Id

			protoSalvagePoint.SalvagedPieces = make([]int, 0)
			for _, pieceProtypeId := range salvagePoolProtype.Content {
				if piece, ok := pieceMap[pieceProtypeId]; ok {
					protoSalvagePoint.SalvagedPieces = append(protoSalvagePoint.SalvagedPieces, piece)
				}
			}

			target.SalvagePoints = append(target.SalvagePoints, protoSalvagePoint)
		}
	}

	return 0, nil
}

func randomForSalvage(protype *config.GuildSalvageProtype, netEffect *config.PropNetEffect, salvagedPieces ...int) (*proto.ProtoResourcesAttr, error) {
	protoResources := new(proto.ProtoResourcesAttr)
	protoResources.BattleShipCards = make([]*proto.ProtoBattleShipCardItem, 0)
	protoResources.PropItems = make([]*proto.ProtoPropItem, 0)
	protoResources.ResourceItems = make([]*proto.ProtoResourceItem, 0)

	pieceChance := protype.PiecesChance + netEffect.PiecesChance
	r := rand.New(rand.NewSource(base.GLocalizedTime.SecTimeStamp()))
	n := r.Intn(10000)
	if n < pieceChance {
		salvagePoolProtype, ok := config.GGuildSalvagePoolConfig.AttrMap[protype.PiecesPool]
		if !ok {
			return nil, custom_errors.New("salvage pool protype not found")
		}

		protoResources = AppendProtoResources(protoResources, randomForSalvagePool(r, salvagePoolProtype, salvagedPieces...))
	}

	minLen := len(protype.Pools)
	if minLen < len(protype.TimesLower) {
		minLen = len(protype.TimesLower)
	}
	if minLen < len(protype.TimesUpper) {
		minLen = len(protype.TimesUpper)
	}

	for i := 0; i < minLen; i++ {
		poolId := protype.Pools[i]
		timesLower := protype.TimesLower[i] + netEffect.TimesLower
		timesUpper := protype.TimesUpper[i] + netEffect.TimesUpper
		times := timesLower
		if timesUpper > timesLower {
			n = r.Intn(timesUpper - timesLower + 1)
			times = timesLower + n
		}

		salvagePoolProtype, ok := config.GGuildSalvagePoolConfig.AttrMap[poolId]
		if !ok {
			return nil, custom_errors.New("salvage pool protype not found")
		}

		for j := 0; j < times; j++ {
			protoResources = AppendProtoResources(protoResources, randomForSalvagePool(r, salvagePoolProtype))
		}

	}

	return protoResources, nil
}

func randomForSalvagePool(r *rand.Rand, protype *config.GuildSalvagePoolProtype, ommitContent ...int) *proto.ProtoResourcesAttr {
	protoResources := new(proto.ProtoResourcesAttr)
	protoResources.BattleShipCards = make([]*proto.ProtoBattleShipCardItem, 0)
	protoResources.PropItems = make([]*proto.ProtoPropItem, 0)
	protoResources.ResourceItems = make([]*proto.ProtoResourceItem, 0)

	omitMap := make(map[int]int)
	for _, c := range ommitContent {
		omitMap[c] = c
	}

	minLen := len(protype.Content)
	if minLen > len(protype.Count) {
		minLen = len(protype.Count)
	}
	if minLen > len(protype.Weight) {
		minLen = len(protype.Weight)
	}

	protypeIds := make([]int, 0, minLen)
	counts := make([]int, 0, minLen)
	weightSection := make([]int, 0, minLen)
	lastSectionEnd := 0
	for i := 0; i < minLen; i++ {
		if _, ok := omitMap[protype.Content[i]]; ok {
			continue
		}

		protypeIds = append(protypeIds, protype.Content[i])
		counts = append(counts, protype.Count[i])
		if i > 0 {

		}
		lastSectionEnd += protype.Weight[i]
		weightSection = append(weightSection, lastSectionEnd)
	}

	if len(protypeIds) > 0 {
		n := r.Intn(weightSection[len(weightSection)-1])

		for i := 0; i < len(weightSection); i++ {
			if n < weightSection[i] {
				switch protype.PoolType {
				case config.POOL_TYPE_PROPS:
					protoPropItem := new(proto.ProtoPropItem)
					protoPropItem.ProtypeId = protypeIds[i]
					protoPropItem.Count = counts[i]
					protoPropItem.CountType = config.COUNT_TYPE_CONST
					protoResources.PropItems = append(protoResources.PropItems, protoPropItem)
				case config.POOL_TYPE_CARDS:
					protoCardItem := new(proto.ProtoBattleShipCardItem)
					protoCardItem.ProtypeId = protypeIds[i]
					protoCardItem.Count = counts[i]
					protoCardItem.CountType = config.COUNT_TYPE_CONST
					protoResources.BattleShipCards = append(protoResources.BattleShipCards, protoCardItem)
				}

				break
			}
		}
	}

	return protoResources
}
