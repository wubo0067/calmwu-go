package handler

import (
	"sailcraft/base"
	"sailcraft/fleetsvr_main/config"
	"sailcraft/fleetsvr_main/custom_errors"
	"sailcraft/fleetsvr_main/errorcode"
	"sailcraft/fleetsvr_main/model"
	"sailcraft/fleetsvr_main/proto"
	"sailcraft/fleetsvr_main/table"
	"sailcraft/fleetsvr_main/utils"
)

// 注意，user_info_changed 不能merge

func GetUserInfo(uin int, userInfo *table.TblUserInfo) (int, error) {
	if uin <= 0 || userInfo == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("params is not valid")
	}

	base.GLog.Debug("GetUserInfo uin %d", uin)

	userInfoModel := &model.UserInfoModel{Uin: uin}
	retCode, err := userInfoModel.GetUserInfo(userInfo)
	if err != nil {
		base.GLog.Error("GetUserInfo err uin %d msg %s", uin, err.Error())
		return retCode, err
	}

	return 0, nil
}

func UpdateUserInfo(uin int, userInfo *table.TblUserInfo) (int, error) {
	if uin <= 0 || userInfo == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("params is not valid")
	}

	userInfoModel := &model.UserInfoModel{Uin: uin}
	retCode, err := userInfoModel.UpdateUserInfo(userInfo)
	if err != nil {
		base.GLog.Error("UpdateUserInfo err uin %d msg %s", uin, err.Error())
		return retCode, err
	}

	return 0, nil
}

func UpdateUserInfoChanged(uin int, userInfoChanged *proto.ProtoUserInfo) (int, error) {
	if uin <= 0 || userInfoChanged == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("params is not valid")
	}

	attrMap, err := convertUserInfoChangedToMapInterface(userInfoChanged)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	userInfoModel := &model.UserInfoModel{Uin: uin}
	retCode, err := userInfoModel.UpdateUserInfoChanged(attrMap)
	if err != nil {
		base.GLog.Error("UpdateUserInfoChanged err uin %d msg %s", uin, err.Error())
		return retCode, err
	}

	return 0, nil
}

func CostUserGem(userInfo *table.TblUserInfo, gemCost int) bool {
	if userInfo == nil {
		return false
	}

	totalGem := userInfo.Gem + userInfo.PurchaseGem

	// 先消耗免费钻石
	if totalGem >= gemCost {
		if userInfo.Gem >= gemCost {
			userInfo.Gem = userInfo.Gem - gemCost
		} else {
			userInfo.Gem = 0
			userInfo.PurchaseGem = totalGem - gemCost
		}

		return true
	} else {
		return false
	}
}

func CostUserGold(userInfo *table.TblUserInfo, cost int) bool {
	if userInfo == nil {
		return false
	}

	if userInfo.Gold >= cost {
		userInfo.Gold = userInfo.Gold - cost
		return true
	} else {
		delta := cost - userInfo.Gold
		gemDelta, err := utils.GemToResource(delta, config.RESOURCE_ITEM_TYPE_GOLD)
		if err != nil {
			return false
		}

		userInfo.Gold = 0

		if CostUserGem(userInfo, gemDelta) {
			return true
		}

		return false
	}
}

func CostUserWood(userInfo *table.TblUserInfo, cost int) bool {
	if userInfo == nil {
		return false
	}

	if userInfo.Wood >= cost {
		userInfo.Wood = userInfo.Wood - cost
		return true
	} else {
		delta := cost - userInfo.Wood
		gemDelta, err := utils.GemToResource(delta, config.RESOURCE_ITEM_TYPE_WOOD)
		if err != nil {
			return false
		}

		userInfo.Wood = 0

		if CostUserGem(userInfo, gemDelta) {
			return true
		}

		return false
	}
}

func CostUserStone(userInfo *table.TblUserInfo, cost int) bool {
	if userInfo == nil {
		return false
	}

	if userInfo.Stone >= cost {
		userInfo.Stone = userInfo.Stone - cost
		return true
	} else {
		delta := cost - userInfo.Stone
		gemDelta, err := utils.GemToResource(delta, config.RESOURCE_ITEM_TYPE_STONE)
		if err != nil {
			return false
		}

		userInfo.Stone = 0

		if CostUserGem(userInfo, gemDelta) {
			return true
		}

		return false
	}
}

func CostUserIron(userInfo *table.TblUserInfo, cost int) bool {
	if userInfo == nil {
		return false
	}

	if userInfo.Iron >= cost {
		userInfo.Iron = userInfo.Iron - cost
		return true
	} else {
		delta := cost - userInfo.Iron
		gemDelta, err := utils.GemToResource(delta, config.RESOURCE_ITEM_TYPE_IRON)
		if err != nil {
			return false
		}

		userInfo.Iron = 0

		if CostUserGem(userInfo, gemDelta) {
			return true
		}

		return false
	}
}

func CostUserShipSoul(userInfo *table.TblUserInfo, cost int) bool {
	if userInfo == nil {
		return false
	}

	if userInfo.ShipSoul >= cost {
		userInfo.ShipSoul -= cost
		return true
	} else {
		return false
	}
}

func UserResourceCost(userInfo *table.TblUserInfo, cost config.ResourcesAttr, userInfoChanged *proto.ProtoUserInfo) bool {
	if userInfo == nil || userInfoChanged == nil {
		return false
	}

	for _, resource := range cost.ResourceItems {
		resourceCount := 0
		if resource.CountType == config.COUNT_TYPE_CONST {
			resourceCount = int(config.ConstantValue(resource.Count))
		}

		switch resource.Type {
		case config.RESOURCE_ITEM_TYPE_GOLD:
			if !CostUserGold(userInfo, resourceCount) {
				return false
			}
		case config.RESOURCE_ITEM_TYPE_WOOD:
			if !CostUserWood(userInfo, resourceCount) {
				return false
			}
		case config.RESOURCE_ITEM_TYPE_STONE:
			if !CostUserStone(userInfo, resourceCount) {
				return false
			}
		case config.RESOURCE_ITEM_TYPE_IRON:
			if !CostUserIron(userInfo, resourceCount) {
				return false
			}
		case config.RESOURCE_ITEM_TYPE_SHIP_SOUL:
			if !CostUserShipSoul(userInfo, resourceCount) {
				return false
			}
		default:
			return false
		}
	}

	userInfoChanged.Wood = userInfo.Wood
	userInfoChanged.Stone = userInfo.Stone
	userInfoChanged.Iron = userInfo.Iron
	userInfoChanged.Gold = userInfo.Gold
	userInfoChanged.Gem = userInfo.Gem
	userInfoChanged.PurchaseGem = userInfo.PurchaseGem
	userInfoChanged.ShipSoul = userInfo.ShipSoul

	return true
}

func AddUserExp(userInfo *table.TblUserInfo, userInfoChanged *proto.ProtoUserInfo, expPlus int) error {
	if userInfo == nil || userInfoChanged == nil {
		return custom_errors.NullPoint()
	}

	newExp := userInfo.Exp + expPlus
	newLevel := config.GLevelExpConfig.QueryLevelByExp(newExp)

	userInfoChanged.Exp = newExp

	if userInfo.Level != newLevel {
		userInfoChanged.Level = newLevel
	}

	userInfo.Exp = newExp
	userInfo.Level = newLevel

	return nil
}

func CalculateUinUserRealResourcesCost(uin int, costs *[]config.ResourceItem) ([]config.ResourceItem, int, error) {
	if costs == nil {
		return nil, errorcode.ERROR_CODE_DEFAULT, custom_errors.NullPoint()
	}

	if uin <= 0 {
		return nil, errorcode.ERROR_CODE_DEFAULT, custom_errors.New("uin is invalid")
	}

	// 计算资源数量
	if len(*costs) > 0 {
		var userInfo table.TblUserInfo
		userInfoModel := model.UserInfoModel{Uin: uin}
		_, err := userInfoModel.GetUserInfo(&userInfo)
		if err != nil {
			return nil, errorcode.ERROR_CODE_DEFAULT, err
		}

		return CalculateUserRealResourcesCost(&userInfo, costs)
	}

	return make([]config.ResourceItem, 0), 0, nil
}

/*
判断用户资源是否足够，并计算真实资源消耗（对于不足的资源用钻石补足）
*/
func CalculateUserRealResourcesCost(userInfo *table.TblUserInfo, costs *[]config.ResourceItem) ([]config.ResourceItem, int, error) {
	if userInfo == nil || costs == nil {
		return nil, errorcode.ERROR_CODE_DEFAULT, custom_errors.NullPoint()
	}

	if userInfo.Uin <= 0 {
		return nil, errorcode.ERROR_CODE_DEFAULT, custom_errors.New("uin is invalid")
	}

	// 计算资源数量
	if len(*costs) > 0 {
		gemCount := 0
		goldCount := 0
		ironCount := 0
		stoneCount := 0
		woodCount := 0
		shipSoulCount := 0

		for _, item := range *costs {
			if item.CountType == config.COUNT_TYPE_CONST {
				switch item.Type {
				case config.RESOURCE_ITEM_TYPE_GEM:
					gemCount += int(config.ConstantValue(item.Count))
				case config.RESOURCE_ITEM_TYPE_GOLD:
					goldCount += int(config.ConstantValue(item.Count))
				case config.RESOURCE_ITEM_TYPE_IRON:
					ironCount += int(config.ConstantValue(item.Count))
				case config.RESOURCE_ITEM_TYPE_STONE:
					stoneCount += int(config.ConstantValue(item.Count))
				case config.RESOURCE_ITEM_TYPE_WOOD:
					woodCount += int(config.ConstantValue(item.Count))
				case config.RESOURCE_ITEM_TYPE_SHIP_SOUL:
					shipSoulCount += int(config.ConstantValue(item.Count))
				default:
				}
			}
		}

		// 资源转换钻石
		if goldCount > 0 {
			if userInfo.Gold < goldCount {
				gem, err := utils.GemToResource(goldCount-userInfo.Gold, config.RESOURCE_ITEM_TYPE_GOLD)
				if err != nil {
					return nil, errorcode.ERROR_CODE_DEFAULT, err
				}
				gemCount += gem
				goldCount = userInfo.Gold
			}
		}

		if ironCount > 0 {
			if userInfo.Iron < ironCount {
				gem, err := utils.GemToResource(ironCount-userInfo.Iron, config.RESOURCE_ITEM_TYPE_IRON)
				if err != nil {
					return nil, errorcode.ERROR_CODE_DEFAULT, err
				}
				gemCount += gem
				ironCount = userInfo.Iron
			}
		}

		if woodCount > 0 {
			if userInfo.Wood < woodCount {
				gem, err := utils.GemToResource(woodCount-userInfo.Wood, config.RESOURCE_ITEM_TYPE_WOOD)
				if err != nil {
					return nil, errorcode.ERROR_CODE_DEFAULT, err
				}
				gemCount += gem
				woodCount = userInfo.Wood
			}
		}

		if stoneCount > 0 {
			if userInfo.Stone < stoneCount {
				gem, err := utils.GemToResource(stoneCount-userInfo.Stone, config.RESOURCE_ITEM_TYPE_STONE)
				if err != nil {
					return nil, errorcode.ERROR_CODE_DEFAULT, err
				}
				gemCount += gem
				stoneCount = userInfo.Stone
			}
		}

		if shipSoulCount > 0 {
			if userInfo.ShipSoul < shipSoulCount {
				return nil, errorcode.ERROR_CODE_NOT_ENOUGH_SHIP_SOUL, custom_errors.New("ship soul is not enough")
			}
		}

		if (userInfo.Gem + userInfo.PurchaseGem) < gemCount {
			return nil, errorcode.ERROR_CODE_NOT_ENOUGH_GEM, custom_errors.New("gem is not enough")
		}

		finalResources := make([]config.ResourceItem, 0)

		if goldCount > 0 {
			var goldItem config.ResourceItem
			goldItem.Type = config.RESOURCE_ITEM_TYPE_GOLD
			goldItem.CountType = config.COUNT_TYPE_CONST
			goldItem.Count = float64(goldCount)
			finalResources = append(finalResources, goldItem)
		}

		if gemCount > 0 {
			var gemItem config.ResourceItem
			gemItem.Type = config.RESOURCE_ITEM_TYPE_GEM
			gemItem.CountType = config.COUNT_TYPE_CONST
			gemItem.Count = float64(gemCount)
			finalResources = append(finalResources, gemItem)
		}

		if ironCount > 0 {
			var ironItem config.ResourceItem
			ironItem.Type = config.RESOURCE_ITEM_TYPE_IRON
			ironItem.CountType = config.COUNT_TYPE_CONST
			ironItem.Count = float64(ironCount)
			finalResources = append(finalResources, ironItem)
		}

		if stoneCount > 0 {
			var stoneItem config.ResourceItem
			stoneItem.Type = config.RESOURCE_ITEM_TYPE_STONE
			stoneItem.CountType = config.COUNT_TYPE_CONST
			stoneItem.Count = float64(stoneCount)
			finalResources = append(finalResources, stoneItem)
		}

		if woodCount > 0 {
			var woodItem config.ResourceItem
			woodItem.Type = config.RESOURCE_ITEM_TYPE_WOOD
			woodItem.CountType = config.COUNT_TYPE_CONST
			woodItem.Count = float64(woodCount)
			finalResources = append(finalResources, woodItem)
		}

		if shipSoulCount > 0 {
			var shipSoulItem config.ResourceItem
			shipSoulItem.Type = config.RESOURCE_ITEM_TYPE_SHIP_SOUL
			shipSoulItem.CountType = config.COUNT_TYPE_CONST
			shipSoulItem.Count = float64(shipSoulCount)
			finalResources = append(finalResources, shipSoulItem)
		}

		return finalResources, 0, nil
	}

	return make([]config.ResourceItem, 0), 0, nil
}

func convertUserInfoChangedToMapInterface(userInfoChanged *proto.ProtoUserInfo) (map[string]interface{}, error) {
	if userInfoChanged == nil {
		return nil, custom_errors.NullPoint()
	}

	attrMap := make(map[string]interface{})
	if userInfoChanged.Level >= 0 {
		attrMap[model.USER_INFO_TABLE_ATTR_LEVEL] = userInfoChanged.Level
	}
	if userInfoChanged.Exp >= 0 {
		attrMap[model.USER_INFO_TABLE_ATTR_EXP] = userInfoChanged.Exp
	}
	if userInfoChanged.Star >= 0 {
		attrMap[model.USER_INFO_TABLE_ATTR_STAR] = userInfoChanged.Star
	}
	if userInfoChanged.Gold >= 0 {
		attrMap[model.USER_INFO_TABLE_ATTR_GOLD] = userInfoChanged.Gold
	}
	if userInfoChanged.Wood >= 0 {
		attrMap[model.USER_INFO_TABLE_ATTR_WOOD] = userInfoChanged.Wood
	}
	if userInfoChanged.Gem >= 0 {
		attrMap[model.USER_INFO_TABLE_ATTR_GEM] = userInfoChanged.Gem
	}
	if userInfoChanged.PurchaseGem >= 0 {
		attrMap[model.USER_INFO_TABLE_ATTR_PURCHASE_GEM] = userInfoChanged.PurchaseGem
	}
	if userInfoChanged.Stone >= 0 {
		attrMap[model.USER_INFO_TABLE_ATTR_STONE] = userInfoChanged.Stone
	}
	if userInfoChanged.Iron >= 0 {
		attrMap[model.USER_INFO_TABLE_ATTR_IRON] = userInfoChanged.Iron
	}
	if userInfoChanged.GuildId == "" || ValidGuildId(userInfoChanged.GuildId) {
		attrMap[model.USER_INFO_TABLE_ATTR_GUILD_ID] = userInfoChanged.GuildId
	}

	return attrMap, nil
}

func GetMultiUserInfo(uinSlice ...int) ([]*table.TblUserInfo, error) {
	return model.GetMultiUserInfo(uinSlice...)
}

func convertUserInfoTableToProto(tblUserInfo *table.TblUserInfo, protoUserInfo *proto.ProtoUserInfo) {
	protoUserInfo.ChangeNameCount = tblUserInfo.ChangeNameCount
	protoUserInfo.Exp = tblUserInfo.Exp
	protoUserInfo.Gem = tblUserInfo.Gem
	protoUserInfo.Gold = tblUserInfo.Gold
	protoUserInfo.GuildId = tblUserInfo.GuildID
	protoUserInfo.Iron = tblUserInfo.Iron
	protoUserInfo.Level = tblUserInfo.Level
	protoUserInfo.PurchaseGem = tblUserInfo.PurchaseGem
	protoUserInfo.Star = tblUserInfo.Star
	protoUserInfo.Stone = tblUserInfo.Stone
	protoUserInfo.Wood = tblUserInfo.Wood
}

func FormatGuildId(creatorUin int, id int) string {
	return model.FormatGuildId(creatorUin, id)
}

func ValidGuildId(guildId string) bool {
	return model.ValidGuildId(guildId)
}

func ConvertGuildIdToUinAndId(guild string) (int, int, bool) {
	return model.ConvertGuildIdToUinAndId(guild)
}
