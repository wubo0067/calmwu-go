package handler

import (
	"sailcraft/base"
	"sailcraft/fleetsvr_main/config"
	"sailcraft/fleetsvr_main/custom_errors"
	"sailcraft/fleetsvr_main/errorcode"
	"sailcraft/fleetsvr_main/model"
	"sailcraft/fleetsvr_main/proto"
	"sailcraft/fleetsvr_main/table"
)

const (
	GUILD_DAILY_VITALITY_REWARD_STATUS_UNCOMPLETED = 0
	GUILD_DAILY_VITALITY_REWARD_STATUS_COMPLETED   = 1
	GUILD_DAILY_VITALITY_REWARD_STATUS_RECEIVED    = 2
)

func HandleGuildMission(req *base.ProtoRequestS, resParams *map[string]interface{}, eventDataSet *EventsHappenedDataSet) (int, error) {
	base.GLog.Debug("Enter HandleGuildMision")

	(*resParams)["NewGuildTaskRewards"] = 0

	var userInfo table.TblUserInfo
	retCode, err := GetUserInfo(req.Uin, &userInfo)
	if err != nil {
		return retCode, err
	}

	creator, gId, ok := ConvertGuildIdToUinAndId(userInfo.GuildID)
	if !ok {
		return 0, nil
	}

	vitalityIncr := 0

	if eventDataSet.BattleEndData != nil {
		base.GLog.Debug("Uin[%d] GuildID[%s] BattleEndData:%+v", userInfo.Uin, userInfo.GuildID, eventDataSet.BattleEndData)
		if eventDataSet.BattleEndData.BattleType == BATTLE_TYPE_PVP {
			if eventDataSet.BattleEndData.BattleResult == BATTLE_RESULT_SUCCESS {
				if protypeList, ok := config.GGuildTaskConfig.AttrMap[config.MISSION_TYPE_LEAGUE_PVP_WIN_TIMES]; ok {
					for _, protype := range protypeList {
						vitalityIncr += protype.Score
						base.GLog.Debug("Uin[%d] GuildID[%s] vitalityIncr[%d]", userInfo.Uin, userInfo.GuildID, vitalityIncr)
					}
				}
			}
		}
	}

	if eventDataSet.ShopPurchasedEventData != nil {
		base.GLog.Debug("Uin[%d] GuildID[%s] ShopPurchasedEventData:%+v", userInfo.Uin, userInfo.GuildID, eventDataSet.ShopPurchasedEventData)
		if eventDataSet.ShopPurchasedEventData.Goods != nil && eventDataSet.ShopPurchasedEventData.Goods.Props != nil {
			base.GLog.Debug("Uin[%d] GuildID[%s] ShopPurchasedEventData.Goods:%+v", userInfo.Uin, userInfo.GuildID, *eventDataSet.ShopPurchasedEventData.Goods)
			for _, propItem := range eventDataSet.ShopPurchasedEventData.Goods.Props {
				if propProtype, ok := config.GPropConfig.AttrMap[propItem.ProtypeId]; ok {
					base.GLog.Debug("Uin[%d] GuildID[%s] propProtype:%+v", userInfo.Uin, userInfo.GuildID, propProtype)
					if propProtype.PropType == config.PROP_TYPE_CARDPACK {
						if protypeList, ok := config.GGuildTaskConfig.AttrMap[config.MISSION_TYPE_SHOP_BUY_CARDPACKS]; ok {
							base.GLog.Debug("Uin[%d] GuildID[%s] protypeList:%+v", userInfo.Uin, userInfo.GuildID, protypeList)
							for _, protype := range protypeList {
								ret := protype.Parameters.Int(config.MISSION_PARAMETER_CARDPACK_ID, 0)
								base.GLog.Debug("Uin[%d] GuildID[%s] propItem.Count[%d] ret[%d] protype:%+v", userInfo.Uin, userInfo.GuildID, propItem.Count, ret, *protype)
								if ret == propProtype.Id {
									vitalityIncr += (protype.Score * propItem.Count)
									base.GLog.Debug("Uin[%d] GuildID[%s] vitalityIncr[%d]", userInfo.Uin, userInfo.GuildID, vitalityIncr)
								}
							}
						}
					}
				}
			}
		}
	}

	if eventDataSet.SalvageData != nil {
		base.GLog.Debug("Uin[%d] GuildID[%s] SalvageData:%+v", userInfo.Uin, userInfo.GuildID, eventDataSet.SalvageData)
		if protypeList, ok := config.GGuildTaskConfig.AttrMap[config.MISSION_TYPE_SALVAGE]; ok {
			for _, protype := range protypeList {
				vitalityIncr += protype.Score
				base.GLog.Debug("Uin[%d] GuildID[%s] vitalityIncr[%d]", userInfo.Uin, userInfo.GuildID, vitalityIncr)
			}
		}
	}

	if vitalityIncr <= 0 {
		base.GLog.Error("Uin[%d] GuildID[%s] vitalityIncr[%d] is invalid!", userInfo.Uin, userInfo.GuildID, vitalityIncr)
		return 0, nil
	}

	// 更新日活跃度
	dailyVitality, err := GetGuildDailyVitality(req.Uin)
	if err != nil {
		base.GLog.Error("Uin[%d] GuildID[%s] GetGuildDailyVitality failed!", userInfo.Uin, userInfo.GuildID)
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	oldVitaliy := 0
	rstV := config.GGlobalConfig.Guild.LimitDailyVitality - dailyVitality.Vitality
	base.GLog.Debug("Uin[%d] GuildID[%s] LimitDailyVitality[%d] dailyVitality[%d] vitalityIncr[%d] rstV[%d]", userInfo.Uin, userInfo.GuildID,
		config.GGlobalConfig.Guild.LimitDailyVitality, dailyVitality.Vitality, vitalityIncr, rstV)
	if vitalityIncr > rstV {
		vitalityIncr = rstV

		if vitalityIncr <= 0 {
			return 0, nil
		}
	}

	oldVitaliy = dailyVitality.Vitality

	dailyVitality.Vitality += vitalityIncr
	retCode, err = UdpateGuildDailyVitality(req.Uin, dailyVitality)
	if err != nil {
		base.GLog.Debug("Uin[%d] GuildID[%s] dailyVitality[%d] Update failed!", userInfo.Uin, userInfo.GuildID, dailyVitality.Vitality)
		return retCode, err
	}

	// 更新周活跃度
	weeklyVitality, err := GetGuildWeeklyVitality(creator, gId, req.Uin)
	if err != nil {
		base.GLog.Debug("Uin[%d] GuildID[%s] GetGuildWeeklyVitality failed!", userInfo.Uin, userInfo.GuildID)
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	weeklyVitality, err = SetGuildWeeklyVitalityOfMember(creator, gId, req.Uin, weeklyVitality+vitalityIncr)
	if err != nil {
		base.GLog.Debug("Uin[%d] GuildID[%s] SetGuildWeeklyVitalityOfMember failed!", userInfo.Uin, userInfo.GuildID)
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	// 更新公会活跃度奖励
	newRewards := config.GGuildTaskRewardConfig.RewardBetween(oldVitaliy, dailyVitality.Vitality)
	base.GLog.Debug("GGuildTaskRewardConfig.RewardBetween newRewards oldVitaliy %d dailyVitality.Vitality %d len(newRewards) [%d]", oldVitaliy, dailyVitality.Vitality, len(newRewards))
	if len(newRewards) > 0 {
		_, dailyVitalityRewardSlice, err := GetGuildDailyVitaliyAndRewardList(req.Uin)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		newProtoRewards := make([]*proto.ProtoGuildTaskRewardInfo, 0)

		protype2Data := make(map[int]*table.TblGuildDailyVitalityReward)
		for _, data := range dailyVitalityRewardSlice {
			protype2Data[data.ProtypeId] = data

			if data.Status == GUILD_DAILY_VITALITY_REWARD_STATUS_COMPLETED {
				protoReward := new(proto.ProtoGuildTaskRewardInfo)
				protype := config.GGuildTaskRewardConfig.GetTaskConfig(data.ProtypeId)
				composeProtoGuildTaskReward(protoReward, protype, data)
				newProtoRewards = append(newProtoRewards, protoReward)
			}
		}

		base.GLog.Debug("after merge newRewards [%v] newProtoRewards[%v] ", newRewards, newProtoRewards)

		insertRecords := make([]*table.TblGuildDailyVitalityReward, 0)
		updateRecords := make([]*table.TblGuildDailyVitalityReward, 0)
		for _, protype := range newRewards {
			record, ok := protype2Data[protype.Id]
			if !ok {
				record = new(table.TblGuildDailyVitalityReward)
				record.ProtypeId = protype.Id
				record.Status = GUILD_DAILY_VITALITY_REWARD_STATUS_COMPLETED
				record.Uin = req.Uin
				insertRecords = append(insertRecords, record)
				protype2Data[protype.Id] = record
			} else {
				record.Status = GUILD_DAILY_VITALITY_REWARD_STATUS_COMPLETED
				updateRecords = append(updateRecords, record)
			}

			protoReward := new(proto.ProtoGuildTaskRewardInfo)
			composeProtoGuildTaskReward(protoReward, protype, record)
			newProtoRewards = append(newProtoRewards, protoReward)
		}

		base.GLog.Debug("after merge insertRecords [%v] updateRecords [%v] ", insertRecords, updateRecords)

		if len(updateRecords) > 0 {
			retCode, err := UpdateMultiGuildDailyVitalityReward(req.Uin, updateRecords)
			if err != nil {
				return retCode, err
			}
		}

		if len(insertRecords) > 0 {
			retCode, err := AddMultiGuildDailyVitalityReward(req.Uin, insertRecords)
			if err != nil {
				return retCode, err
			}
		}

		(*resParams)["NewGuildTaskRewards"] = 1
		(*resParams)["GuildTaskRewards"] = newProtoRewards
	}

	// 更新公会活跃度
	leagueSeasonInfo, err := GetLeagueSeasonInfo()
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if leagueSeasonInfo.AccountFlag != model.LEAGUE_SEASON_FLAG_SETTLEMENT {
		_, err = IncreaseGuildVitality(userInfo.GuildID, vitalityIncr)
		if err != nil {
			base.GLog.Debug("Uin[%d] GuildID[%s] IncreaseGuildVitality failed!", userInfo.Uin, userInfo.GuildID)
			return errorcode.ERROR_CODE_DEFAULT, err
		}
	} else {
		base.GLog.Error("Uin[%d] GuildID[%s] leagueSeasonInfo.AccountFlag[%d] != LEAGUE_SEASON_FLAG_SETTLEMENT", userInfo.Uin, userInfo.GuildID)
	}

	return 0, nil
}

func IncreaseGuildVitality(guildId string, vitalityIncr int) (int, error) {
	if !ValidGuildId(guildId) {
		return errorcode.ERROR_CODE_GUILD_ID_FORMAT_WRONG, custom_errors.New("guild id format error")
	}

	guildVitalityModel := model.GuildVitalityModel{}
	return guildVitalityModel.IncreaseGuildVitality(guildId, vitalityIncr)
}

func GetGuildVitality(guildId string) (int, error) {
	if !ValidGuildId(guildId) {
		return errorcode.ERROR_CODE_GUILD_ID_FORMAT_WRONG, custom_errors.New("guild id format error")
	}

	guildVitalityModel := model.GuildVitalityModel{}
	return guildVitalityModel.GetGuildVitality(guildId)
}

func GetGuildRank(guildId string) (int, error) {
	guildVitalityModel := model.GuildVitalityModel{}
	return guildVitalityModel.GetRank(guildId)
}

func composeProtoGuildTaskReward(target *proto.ProtoGuildTaskRewardInfo, protype *config.GuildTaskRewardProtype, data *table.TblGuildDailyVitalityReward) (int, error) {
	if target == nil || protype == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.NullPoint()
	}

	target.ProtypeId = protype.Id
	target.Score = protype.Score

	if data == nil {
		target.Status = GUILD_DAILY_VITALITY_REWARD_STATUS_UNCOMPLETED
	} else {
		target.Status = data.Status
	}

	return 0, nil
}

func DeleteGuildVitality(guildId string) (int, error) {
	guildVitalityModel := model.GuildVitalityModel{}
	return guildVitalityModel.DeleteGuildVitality(guildId)
}
