package handler

import (
	"sailcraft/base"
	"sailcraft/fleetsvr_main/config"
	"sailcraft/fleetsvr_main/custom_errors"
	"sailcraft/fleetsvr_main/errorcode"
	"sailcraft/fleetsvr_main/handlerbase"
	"sailcraft/fleetsvr_main/model"
	"sailcraft/fleetsvr_main/proto"
	"sailcraft/fleetsvr_main/table"
)

type BattleShipHandler struct {
	handlerbase.WebHandler
}

func (this *BattleShipHandler) Init() (int, error) {
	var reqParams proto.ProtoInitBattleShipRequest
	err := this.UnpackParams(&reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	battleShipModel := &model.BattleShipModel{Uin: this.Request.Uin}
	_, err = battleShipModel.InitBattleShip(reqParams.ProtypeIDs)
	if err != nil {
		base.GLog.Error("InitBattleShip error %s", err.Error())
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	this.Response.ResData.Params = this.Request.ReqData.Params
	return 0, nil
}

func (this *BattleShipHandler) Compose() (int, error) {
	var reqParams proto.ProtoComposeBattleShipRequest
	err := this.UnpackParams(&reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	battleShipInfo := new(table.TblBattleShip)

	battleShipModel := &model.BattleShipModel{Uin: this.Request.Uin}
	retCode, err := battleShipModel.GetBattleShipInfoByID(reqParams.ShipID, battleShipInfo)
	if err != nil {
		return retCode, err
	}

	if retCode == 0 {
		// 记录不存在，返回错误
		return errorcode.ERROR_CODE_BATTLE_SHIP_NOT_EXSIT, custom_errors.New("battle ship is not exist")
	}

	//先拉取用户的基本信息
	userInfo := new(table.TblUserInfo)
	retCode, err = GetUserInfo(this.Request.Uin, userInfo)
	if err != nil {
		return retCode, err
	}

	var protoOldUserInfo proto.ProtoUserInfo
	convertUserInfoTableToProto(userInfo, &protoOldUserInfo)

	// 检查升级是否合法 更新数据结构
	userInfoChanged := proto.NewDefaultProtoUserInfoChanged()
	retCode, err = updateComposeShipData(battleShipInfo, userInfo, userInfoChanged)
	if err != nil {
		return retCode, err
	}

	// 更新数据库
	err = battleShipModel.UpgradeBattleShipInfoByID(battleShipInfo)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	retCode, err = UpdateUserInfoChanged(this.Request.Uin, userInfoChanged)
	if err != nil {
		return retCode, err
	}

	var protoUserInfo proto.ProtoUserInfo
	convertUserInfoTableToProto(userInfo, &protoUserInfo)

	var responseData proto.ProtoComposeBattleShipResponse
	responseData.BattleShip = packResponseBattleShipProto(battleShipInfo)
	responseData.OldUserInfo = protoOldUserInfo
	responseData.UserInfo = protoUserInfo

	this.Response.ResData.Params = responseData

	return 0, nil
}

func (this *BattleShipHandler) AddCard() (int, error) {
	var reqParams proto.ProtoAddBattleShipCardNumberRequest
	err := this.UnpackParams(&reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if len(reqParams.ShipCards) <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("request param ship cards length is wrong")
	}

	retCode, err := validBattleShipProtypeIDs(reqParams.ShipCards)
	if err != nil {
		return retCode, err
	}

	// 先查出用户所有的战舰，然后更新碎片数量，如果不存在，那么插入新的船碎片即可。
	battleShipModel := &model.BattleShipModel{Uin: this.Request.Uin}
	battleShips, err := battleShipModel.GetBattleShipList()
	if err != nil {
		base.GLog.Error("GetBattleShipList error %s", err.Error())
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	effectedBattleShips := make([]*table.TblBattleShip, 0)

	newBattleShips := make([]*table.TblBattleShip, 0)
	// 对每个战舰，检查下是否存在，如果存在，则加碎片；不存在则插入碎片，记住status=0
	for _, shipCard := range reqParams.ShipCards {
		battleShipInfo := queryBattleShipInfo(battleShips, shipCard.ProtypeID)
		if battleShipInfo != nil {
			// 直接添加碎片数量即可
			if shipCard.CardNumber != 0 {
				battleShipInfo.CardNumber = battleShipInfo.CardNumber + shipCard.CardNumber
				if battleShipInfo.CardNumber < 0 {
					return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("battle ship[%d] card number will less than 0", battleShipInfo.ProtypeID)
				}

				err = battleShipModel.UpgradeBattleShipInfoByID(battleShipInfo)
				if err != nil {
					base.GLog.Error("UpgradeBattleShipInfoByID error %s", err.Error())
					return errorcode.ERROR_CODE_DEFAULT, err
				}
			}

			effectedBattleShips = append(effectedBattleShips, battleShipInfo)
		} else {
			// 新生成的战舰碎片
			newBattleShip := table.NewDefaultBattleShip()
			newBattleShip.Status = table.BATTLE_SHIP_STATUS_CARD
			newBattleShip.Level = 0
			newBattleShip.CardNumber = shipCard.CardNumber
			newBattleShip.ProtypeID = shipCard.ProtypeID
			newBattleShip.Uin = this.Request.Uin

			if newBattleShip.CardNumber < 0 {
				return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("battle ship[%d] card number will less than 0", battleShipInfo.ProtypeID)
			}

			newBattleShips = append(newBattleShips, newBattleShip)
		}
	}

	if len(newBattleShips) > 0 {
		affected, err := battleShipModel.AddBattleShip(newBattleShips)
		base.GLog.Debug("AddBattleShip affected %d", affected)

		if err != nil {
			base.GLog.Error("AddBattleShip error %s", err.Error())
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		// 重新查下数据库的战舰列表
		battleShips, err = battleShipModel.GetBattleShipList()
		if err != nil {
			base.GLog.Error("GetBattleShipList error %s", err.Error())
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		for _, shipInfo := range newBattleShips {
			battleShipInfo := queryBattleShipInfo(battleShips, shipInfo.ProtypeID)
			if battleShipInfo != nil {
				effectedBattleShips = append(effectedBattleShips, battleShipInfo)
			}
		}
	}

	var responseData proto.ProtoAddBattleShipCardNumberResponse
	responseData.BattleShips = make([]proto.ProtoBattleShipInfo, 0)
	for _, battleShipInfo := range effectedBattleShips {
		resBattleShipInfo := packResponseBattleShipProto(battleShipInfo)
		responseData.BattleShips = append(responseData.BattleShips, resBattleShipInfo)
	}

	this.Response.ResData.Params = responseData

	return 0, nil
}

func (this *BattleShipHandler) ValidAddCard() (int, error) {
	var reqParams proto.ProtoCheckBattleShipCardNumberRequest
	err := this.UnpackParams(&reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if len(reqParams.ShipCards) <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("request param ship cards length is wrong")
	}

	retCode, err := validBattleShipProtypeIDs(reqParams.ShipCards)
	if err != nil {
		return retCode, err
	}

	battleShipModel := &model.BattleShipModel{Uin: this.Request.Uin}
	battleShips, err := battleShipModel.GetBattleShipList()
	if err != nil {
		base.GLog.Error("GetBattleShipList error %s", err.Error())
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	for _, shipCard := range reqParams.ShipCards {
		if shipCard.CardNumber < 0 {
			battleShipInfo := queryBattleShipInfo(battleShips, shipCard.ProtypeID)
			if battleShipInfo == nil || battleShipInfo.CardNumber < (-shipCard.CardNumber) {
				return errorcode.ERROR_CODE_NOT_ENOUGH_BATTLE_SHIP_CARD, custom_errors.New("battle ship[%d] has not enough cards", shipCard.ProtypeID)
			}
		}
	}

	return 0, nil
}

func (this *BattleShipHandler) ModifyCard() (int, error) {
	var reqParams proto.ProtoModifyBattleShipCardNumberRequest
	err := this.UnpackParams(&reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if len(reqParams.ShipCards) <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("request param ship cards length is wrong")
	}

	retCode, err := validBattleShipProtypeIDs(reqParams.ShipCards)
	if err != nil {
		return retCode, err
	}

	// 先查出用户所有的战舰，然后更新碎片数量，如果不存在，那么插入新的船碎片即可。
	battleShipModel := &model.BattleShipModel{Uin: this.Request.Uin}
	battleShips, err := battleShipModel.GetBattleShipList()
	if err != nil {
		base.GLog.Error("GetBattleShipList error %s", err.Error())
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	effectedBattleShips := make([]*table.TblBattleShip, 0, len(reqParams.ShipCards))
	newBattleShips := make([]*table.TblBattleShip, 0)
	for _, shipCard := range reqParams.ShipCards {
		battleShipInfo := queryBattleShipInfo(battleShips, shipCard.ProtypeID)
		if battleShipInfo != nil {
			// 直接添加碎片数量即可
			if shipCard.CardNumber >= 0 {
				battleShipInfo.CardNumber = shipCard.CardNumber
				err = battleShipModel.UpgradeBattleShipInfoByID(battleShipInfo)
				if err != nil {
					base.GLog.Error("UpgradeBattleShipInfoByID error %s", err.Error())
					return errorcode.ERROR_CODE_DEFAULT, err
				}
				effectedBattleShips = append(effectedBattleShips, battleShipInfo)
			}
		} else {
			// 新生成的战舰碎片
			if shipCard.CardNumber >= 0 {
				newBattleShip := table.NewDefaultBattleShip()
				newBattleShip.Status = table.BATTLE_SHIP_STATUS_CARD
				newBattleShip.Level = 0
				newBattleShip.CardNumber = shipCard.CardNumber
				newBattleShip.ProtypeID = shipCard.ProtypeID
				newBattleShip.Uin = this.Request.Uin

				newBattleShips = append(newBattleShips, newBattleShip)
			}
		}
	}

	if len(newBattleShips) > 0 {
		affected, err := battleShipModel.AddBattleShip(newBattleShips)
		base.GLog.Debug("AddBattleShip affected %d", affected)

		if err != nil {
			base.GLog.Error("AddBattleShip error %s", err.Error())
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		// 重新查下数据库的战舰列表
		battleShips, err = battleShipModel.GetBattleShipList()
		if err != nil {
			base.GLog.Error("GetBattleShipList error %s", err.Error())
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		for _, shipInfo := range newBattleShips {
			battleShipInfo := queryBattleShipInfo(battleShips, shipInfo.ProtypeID)
			if battleShipInfo != nil {
				effectedBattleShips = append(effectedBattleShips, battleShipInfo)
			}
		}
	}

	var responseData proto.ProtoModifyBattleShipCardNumberResponse
	responseData.BattleShips = make([]*proto.ProtoBattleShipInfo, 0)
	for _, battleShipInfo := range effectedBattleShips {
		resBattleShipInfo := packResponseBattleShipProto(battleShipInfo)
		responseData.BattleShips = append(responseData.BattleShips, &resBattleShipInfo)
	}

	this.Response.ResData.Params = responseData

	return 0, nil
}

func (this *BattleShipHandler) Reclaim() (int, error) {
	var reqParams proto.ProtoReclaimBattleShipCardsRequest
	err := this.UnpackParams(&reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if len(reqParams.ShipCards) <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("request param ship cards length is wrong")
	}

	retCode, err := validBattleShipProtypeIDs(reqParams.ShipCards)
	if err != nil {
		return retCode, err
	}

	battleShipModel := &model.BattleShipModel{Uin: this.Request.Uin}
	battleShips, err := battleShipModel.GetBattleShipList()
	if err != nil {
		base.GLog.Error("GetBattleShipList error %s", err.Error())
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	qualityCardsCost := make(map[string]int)
	var cost config.ResourcesAttr

	for _, shipCard := range reqParams.ShipCards {
		if shipCard.CardNumber > 0 {
			battleShipInfo := queryBattleShipInfo(battleShips, shipCard.ProtypeID)

			if battleShipInfo == nil || battleShipInfo.CardNumber < shipCard.CardNumber {
				return errorcode.ERROR_CODE_BATTLE_SHIP_NOT_ENOUGH_CARD, custom_errors.New("battle ship[%d] has not enough cards", shipCard.ProtypeID)
			}

			battleShipInfo.CardNumber -= shipCard.CardNumber

			var battleShipCardItem config.BattleShipCardItem
			battleShipCardItem.ProtypeId = shipCard.ProtypeID
			battleShipCardItem.CountType = config.COUNT_TYPE_CONST
			battleShipCardItem.Count = float64(shipCard.CardNumber)
			cost.AddShipCards(1, &battleShipCardItem)

			starLevelAttr, err := config.GBattleShipProtypeConfig.GetStarAttr(battleShipInfo.ProtypeID, battleShipInfo.StarLevel)
			if err != nil {
				return errorcode.ERROR_CODE_DEFAULT, err
			}
			qualityCardsCost[starLevelAttr.Rarity] += shipCard.CardNumber

		}
	}

	var reward config.ResourcesAttr
	for quality, count := range qualityCardsCost {
		qualityReward, err := config.GShipReclaimConfig.ReclaimReward(quality, count)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		reward.Add(qualityReward)
	}

	var responseData proto.ProtoReclaimBattleShipCardsResponse

	ResourcesConfigToProto(&cost, &responseData.Cost)
	ResourcesConfigToProto(&reward, &responseData.Rewards)

	this.Response.ResData.Params = responseData

	return 0, nil
}

func (this *BattleShipHandler) Disassemble() (int, error) {
	var reqParams proto.ProtoDisassembleBattleShipRequest
	err := this.UnpackParams(&reqParams)

	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if len(reqParams.ProtypeIdList) <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("request param ship id length is wrong")
	}

	battleShipModel := &model.BattleShipModel{Uin: this.Request.Uin}
	battleShips, err := battleShipModel.GetBattleShipList()
	if err != nil {
		base.GLog.Error("GetBattleShipList error %s", err.Error())
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	effectedBattleShips := make([]*table.TblBattleShip, 0, len(reqParams.ProtypeIdList))

	var cost config.ResourcesAttr
	var reward config.ResourcesAttr

	for _, protypeId := range reqParams.ProtypeIdList {
		battleShipInfo := queryBattleShipInfo(battleShips, protypeId)
		if battleShipInfo == nil || battleShipInfo.Status == table.BATTLE_SHIP_STATUS_CARD || battleShipInfo.Level <= 1 {
			return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("battle ship[%d] can not be assembled", protypeId)
		}

		battleShipStarProtype, err := config.GBattleShipProtypeConfig.GetStarAttr(battleShipInfo.ProtypeID, battleShipInfo.StarLevel)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		starUpgradeProtype, err := config.GBattleShipStrengthenConfig.GetUpgradeAttr(battleShipStarProtype.Rarity, battleShipInfo.StarLevel)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		zeroStarUpgradeProtype, err := config.GBattleShipStrengthenConfig.GetUpgradeAttr(battleShipStarProtype.Rarity, 0)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		levelUpgradeProtype, err := config.GBattleShipUpgradeConfig.GetUpgradeAttr(battleShipStarProtype.Rarity, battleShipInfo.Level)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		levelOneUpgradeProtype, err := config.GBattleShipUpgradeConfig.GetUpgradeAttr(battleShipStarProtype.Rarity, 1)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		reclaimProtype, err := config.GShipReclaimConfig.GetReclaimProtype(battleShipStarProtype.Rarity)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		cost.Add(&reclaimProtype.DowngradeCost)

		var battleShipCard config.BattleShipCardItem
		battleShipCard.ProtypeId = battleShipInfo.ProtypeID
		battleShipCard.CountType = config.COUNT_TYPE_CONST
		battleShipCard.Count = float64(starUpgradeProtype.SumCard - zeroStarUpgradeProtype.SumCard)
		reward.AddShipCards(1, &battleShipCard)

		var levelUpgradeRet config.ResourcesAttr
		levelUpgradeRet.Add(&levelUpgradeProtype.SumCost)
		levelUpgradeRet.Sub(&levelOneUpgradeProtype.SumCost)
		levelUpgradeRet.Scale(float64(reclaimProtype.ResourcesReturnRatio) / 100)
		reward.Add(&levelUpgradeRet)

		battleShipInfo.StarLevel = 0
		battleShipInfo.Level = 1
		effectedBattleShips = append(effectedBattleShips, battleShipInfo)
	}

	// 判断资源消耗是否足够
	var userInfo table.TblUserInfo
	retCode, err := GetUserInfo(this.Request.Uin, &userInfo)
	if err != nil {
		return retCode, err
	}

	_, retCode, err = CalculateUserRealResourcesCost(&userInfo, &cost.ResourceItems)
	if err != nil {
		return retCode, err
	}

	// 更新战舰
	var responseData proto.ProtoDisassembleBattleShipResponse

	for _, battleShipInfo := range effectedBattleShips {
		err = battleShipModel.UpgradeBattleShipInfoByID(battleShipInfo)
		if err != nil {
			base.GLog.Error("UpgradeBattleShipInfoByID error %s", err.Error())
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		protoBattleShipInfo := packResponseBattleShipProto(battleShipInfo)
		responseData.BattleShips = append(responseData.BattleShips, protoBattleShipInfo)
	}

	ResourcesConfigToProto(&cost, &responseData.Cost)
	ResourcesConfigToProto(&reward, &responseData.Rewards)

	this.Response.ResData.Params = responseData

	return 0, nil
}

func (this *BattleShipHandler) Upgrade() (int, error) {
	var reqParams proto.ProtoUpgradeBattleShipLevelRequest
	err := this.UnpackParams(&reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	battleShipInfo := new(table.TblBattleShip)

	battleShipModel := &model.BattleShipModel{Uin: this.Request.Uin}
	retCode, err := battleShipModel.GetBattleShipInfoByID(reqParams.ShipID, battleShipInfo)
	if err != nil {
		return retCode, err
	}

	if retCode == 0 {
		// 记录不存在，返回错误
		return errorcode.ERROR_CODE_BATTLE_SHIP_NOT_EXSIT, custom_errors.New("battle ship is not exist")
	}

	//先拉取用户的基本信息
	userInfo := new(table.TblUserInfo)
	retCode, err = GetUserInfo(this.Request.Uin, userInfo)
	if err != nil {
		return retCode, err
	}

	var protoOldUserInfo proto.ProtoUserInfo
	convertUserInfoTableToProto(userInfo, &protoOldUserInfo)

	base.GLog.Debug("GetUserInfo data is [%+v]", *userInfo)

	// 检查升级是否合法 更新数据结构
	userInfoChanged := proto.NewDefaultProtoUserInfoChanged()
	retCode, err = updateUpgradeShipLevelData(battleShipInfo, userInfo, userInfoChanged)
	if err != nil {
		return retCode, err
	}
	// 更新数据库
	err = battleShipModel.UpgradeBattleShipInfoByID(battleShipInfo)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	retCode, err = UpdateUserInfoChanged(this.Request.Uin, userInfoChanged)
	if err != nil {
		return retCode, err
	}

	var protoUserInfo proto.ProtoUserInfo
	convertUserInfoTableToProto(userInfo, &protoUserInfo)

	var responseData proto.ProtoUpgradeBattleShipLevelResponse
	responseData.BattleShip = packResponseBattleShipProto(battleShipInfo)
	responseData.OldUserInfo = protoOldUserInfo
	responseData.UserInfo = protoUserInfo

	this.Response.ResData.Params = responseData

	return 0, nil
}

func (this *BattleShipHandler) UpgradeStarLevel() (int, error) {
	var reqParams proto.ProtoUpgradeBattleShipStarLevelRequest
	err := this.UnpackParams(&reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	battleShipInfo := new(table.TblBattleShip)

	battleShipModel := &model.BattleShipModel{Uin: this.Request.Uin}
	retCode, err := battleShipModel.GetBattleShipInfoByID(reqParams.ShipID, battleShipInfo)
	if err != nil {
		return retCode, err
	}

	if retCode == 0 {
		// 记录不存在，返回错误
		return errorcode.ERROR_CODE_BATTLE_SHIP_NOT_EXSIT, custom_errors.New("battle ship is not exist")
	}

	//先拉取用户的基本信息
	userInfo := new(table.TblUserInfo)
	retCode, err = GetUserInfo(this.Request.Uin, userInfo)
	if err != nil {
		return retCode, err
	}

	var protoOldUserInfo proto.ProtoUserInfo
	convertUserInfoTableToProto(userInfo, &protoOldUserInfo)
	// 检查升级是否合法 更新数据结构
	userInfoChanged := proto.NewDefaultProtoUserInfoChanged()
	retCode, err = updateUpgradeShipStarData(battleShipInfo, userInfo, userInfoChanged)
	if err != nil {
		return retCode, err
	}

	// 更新数据库
	err = battleShipModel.UpgradeBattleShipInfoByID(battleShipInfo)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	retCode, err = UpdateUserInfoChanged(this.Request.Uin, userInfoChanged)
	if err != nil {
		return retCode, err
	}

	var protoUserInfo proto.ProtoUserInfo
	convertUserInfoTableToProto(userInfo, &protoUserInfo)

	var responseData proto.ProtoUpgradeBattleShipStarLevelResponse
	responseData.BattleShip = packResponseBattleShipProto(battleShipInfo)
	responseData.OldUserInfo = protoOldUserInfo
	responseData.UserInfo = protoUserInfo

	this.Response.ResData.Params = responseData

	return 0, nil
}

func (this *BattleShipHandler) List() (int, error) {
	battleShipModel := &model.BattleShipModel{Uin: this.Request.Uin}
	records, err := battleShipModel.GetBattleShipList()
	if err != nil {
		base.GLog.Error("GetBattleShipList error %s", err.Error())
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	var responseData proto.ProtoGetBattleShipListResponse
	responseData.BattleShips = make([]proto.ProtoBattleShipInfo, 0)
	for _, battleShipInfo := range records {
		resBattleShipInfo := packResponseBattleShipProto(battleShipInfo)
		responseData.BattleShips = append(responseData.BattleShips, resBattleShipInfo)
	}

	this.Response.ResData.Params = responseData

	return 0, nil
}

func packResponseBattleShipProto(battleShipInfo *table.TblBattleShip) proto.ProtoBattleShipInfo {
	return proto.ProtoBattleShipInfo{
		Id:         battleShipInfo.Id,
		Uin:        battleShipInfo.Uin,
		ProtypeID:  battleShipInfo.ProtypeID,
		Level:      battleShipInfo.Level,
		StarLevel:  battleShipInfo.StarLevel,
		CardNumber: battleShipInfo.CardNumber,
		Status:     battleShipInfo.Status,
	}
}

func updateUpgradeShipLevelCost(battleShipInfo *table.TblBattleShip, userInfo *table.TblUserInfo, userInfoChanged *proto.ProtoUserInfo) (int, error) {
	if battleShipInfo == nil || userInfo == nil || userInfoChanged == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.NullPoint()
	}

	if battleShipInfo.Status != table.BATTLE_SHIP_STATUS_SHIP {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("ship is not exist")
	}

	protypeid := battleShipInfo.ProtypeID
	starLevel := battleShipInfo.StarLevel

	shipStarLevelAttr, err := config.GBattleShipProtypeConfig.GetStarAttr(protypeid, starLevel)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	// 看看对应星级等级是否满足
	strengthenAttr, err := config.GBattleShipStrengthenConfig.GetUpgradeAttr(shipStarLevelAttr.Rarity, shipStarLevelAttr.StarLevel)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	newLevel := battleShipInfo.Level + 1
	if newLevel > strengthenAttr.ShipLevelMax {
		return errorcode.ERROR_CODE_BATTLE_SHIP_LEVEL_REACH_STAR_LIMIT, custom_errors.New("battle ship level reach star limit")
	}

	// 这里需要取下一级的消耗
	upgradeAttr, err := config.GBattleShipUpgradeConfig.GetUpgradeAttr(shipStarLevelAttr.Rarity, newLevel)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	// 检查cost消耗
	if !UserResourceCost(userInfo, upgradeAttr.Cost, userInfoChanged) {
		return errorcode.ERROR_CODE_NOT_ENOUGH_RESOURCE, custom_errors.New("not enough resource")
	}

	// 检查碎片数量
	if battleShipInfo.CardNumber < upgradeAttr.Card {
		return errorcode.ERROR_CODE_BATTLE_SHIP_NOT_ENOUGH_CARD, custom_errors.New("not enough card number")
	}

	battleShipInfo.CardNumber = battleShipInfo.CardNumber - upgradeAttr.Card
	battleShipInfo.Level = battleShipInfo.Level + 1

	return 0, nil
}

func updateUpgradeShipLevelData(battleShipInfo *table.TblBattleShip, userInfo *table.TblUserInfo, userInfoChanged *proto.ProtoUserInfo) (int, error) {
	if battleShipInfo == nil || userInfo == nil || userInfoChanged == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.NullPoint()
	}

	protypeid := battleShipInfo.ProtypeID
	level := battleShipInfo.Level

	// 检查等级是否ok
	_, err := config.GBattleShipProtypeConfig.GetLevelAttr(protypeid, level)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	// 看看下一级是否存在
	_, err = config.GBattleShipProtypeConfig.GetLevelAttr(protypeid, level+1)
	if err != nil {
		return errorcode.ERROR_CODE_BATTLE_SHIP_LEVEL_LIMIT, err
	}

	return updateUpgradeShipLevelCost(battleShipInfo, userInfo, userInfoChanged)
}

func updateUpgradeShipStarData(battleShipInfo *table.TblBattleShip, userInfo *table.TblUserInfo, userInfoChanged *proto.ProtoUserInfo) (int, error) {
	if battleShipInfo == nil || userInfo == nil || userInfoChanged == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.NullPoint()
	}

	if battleShipInfo == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.NullPoint()
	}

	if battleShipInfo.Status != table.BATTLE_SHIP_STATUS_SHIP {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("ship is not exist")
	}

	protypeID := battleShipInfo.ProtypeID
	starLevel := battleShipInfo.StarLevel

	_, err := config.GBattleShipProtypeConfig.GetStarAttr(protypeID, starLevel)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	newLevel := starLevel + 1

	// 看看下一级是否存在
	shipStarLevelAttr, err := config.GBattleShipProtypeConfig.GetStarAttr(protypeID, newLevel)
	if err != nil {
		return errorcode.ERROR_CODE_BATTLE_SHIP_STAR_LEVEL_LIMIT, err
	}

	// 查看是否满足等级需求
	strengthenAttr, err := config.GBattleShipStrengthenConfig.GetUpgradeAttr(shipStarLevelAttr.Rarity, newLevel)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if strengthenAttr.ShipLevelNeed > battleShipInfo.Level {
		return errorcode.ERROR_CODE_BATTLE_SHIP_STAR_REACH_LEVEL_LIMIT, custom_errors.New("battle ship star level reach level limit")
	}

	// 消耗是否满足
	if !UserResourceCost(userInfo, strengthenAttr.Cost, userInfoChanged) {
		return errorcode.ERROR_CODE_NOT_ENOUGH_RESOURCE, custom_errors.New("not enough resource")
	}

	// 检查碎片数量
	if battleShipInfo.CardNumber < strengthenAttr.Card {
		return errorcode.ERROR_CODE_BATTLE_SHIP_NOT_ENOUGH_CARD, custom_errors.New("not enough card number")
	}

	battleShipInfo.CardNumber = battleShipInfo.CardNumber - strengthenAttr.Card
	battleShipInfo.StarLevel = battleShipInfo.StarLevel + 1

	return 0, nil
}

func queryBattleShipInfo(battleShips []*table.TblBattleShip, protypeID int) *table.TblBattleShip {
	for _, info := range battleShips {
		if info.ProtypeID == protypeID {
			return info
		}
	}

	return nil
}

func queryBattleShipInfoById(battleShips []*table.TblBattleShip, shipId int) *table.TblBattleShip {
	for _, info := range battleShips {
		if info.Id == shipId {
			return info
		}
	}

	return nil
}

func validBattleShipProtypeIDs(shipCards []*proto.ProtoBattleShipCardNumberParams) (int, error) {
	for _, card := range shipCards {
		protypeid := card.ProtypeID
		_, err := config.GBattleShipProtypeConfig.GetLevelAttr(protypeid, 1)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}
	}

	return 0, nil
}

func updateComposeShipData(battleShipInfo *table.TblBattleShip, userInfo *table.TblUserInfo, userInfoChanged *proto.ProtoUserInfo) (int, error) {
	if battleShipInfo == nil || userInfo == nil || userInfoChanged == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.NullPoint()
	}

	if battleShipInfo.Status == table.BATTLE_SHIP_STATUS_SHIP {
		return errorcode.ERROR_CODE_BATTLE_SHIP_IS_NOT_CARD_TYPE, custom_errors.New("ship is not card status")
	}

	protypeID := battleShipInfo.ProtypeID
	starLevel := battleShipInfo.StarLevel

	shipStarLevelAttr, err := config.GBattleShipProtypeConfig.GetStarAttr(protypeID, starLevel)
	if err != nil {
		return errorcode.ERROR_CODE_BATTLE_SHIP_STAR_LEVEL_LIMIT, err
	}

	// 查询消耗信息
	strengthenAttr, err := config.GBattleShipStrengthenConfig.GetUpgradeAttr(shipStarLevelAttr.Rarity, starLevel)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	// 消耗是否满足
	if !UserResourceCost(userInfo, strengthenAttr.Cost, userInfoChanged) {
		return errorcode.ERROR_CODE_NOT_ENOUGH_RESOURCE, custom_errors.New("not enough resource")
	}

	if strengthenAttr.Card > battleShipInfo.CardNumber {
		return errorcode.ERROR_CODE_BATTLE_SHIP_NOT_ENOUGH_CARD, custom_errors.New("not enough card")
	}

	battleShipInfo.CardNumber = battleShipInfo.CardNumber - strengthenAttr.Card
	battleShipInfo.Level = 1
	battleShipInfo.Status = table.BATTLE_SHIP_STATUS_SHIP

	return 0, nil
}

func GetBattleShipListByUin(uin int) ([]*table.TblBattleShip, int, error) {
	battleShipModel := model.BattleShipModel{Uin: uin}

	records, err := battleShipModel.GetBattleShipList()

	if err != nil {
		return nil, errorcode.ERROR_CODE_DEFAULT, err
	}

	return records, 0, nil
}
