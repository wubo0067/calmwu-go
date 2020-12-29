package handler

import (
	"encoding/json"
	"sailcraft/base"
	"sailcraft/fleetsvr_main/config"
	"sailcraft/fleetsvr_main/custom_errors"
	"sailcraft/fleetsvr_main/errorcode"
	"sailcraft/fleetsvr_main/handlerbase"
	"sailcraft/fleetsvr_main/model"
	"sailcraft/fleetsvr_main/proto"
	"sailcraft/fleetsvr_main/table"
	"time"

	"github.com/mitchellh/mapstructure"
)

type MissionProgress struct {
	Current       int `json:"current"`
	Total         int `json:"total"`
	DayContinuous int `json:"day_continuous"`
	LastDay       int `json:"last_day"`
}

type AchievemenHandler struct {
	handlerbase.WebHandler
}

func (this *AchievemenHandler) List() (int, error) {
	achievementModel := model.AchievementModel{Uin: this.Request.Uin}
	tblAchievementList, err := achievementModel.GetAchievementList()
	if err != nil {
		base.GLog.Error(err)
		return errorcode.ERROR_CODE_DEFAULT, err
	}
	achievementMap := make(map[int]*table.TblAchievement)
	for _, tblAchievement := range tblAchievementList {
		achievementMap[tblAchievement.ProtypeId] = tblAchievement
	}

	achievementList := make([]*proto.ProtoAchievementInfo, 0)

	for _, protype := range config.GAchievementConfig.AttrMap {
		protoAchievement := new(proto.ProtoAchievementInfo)
		ComposeProtoAchievement(protoAchievement, achievementMap[protype.Id], protype)

		achievementList = append(achievementList, protoAchievement)
	}

	var responseData proto.ProtoGetAchievementListResponse
	responseData.Achievements = achievementList
	this.Response.ResData.Params = responseData

	return 0, nil
}

func HandleAchievementProgress(req *base.ProtoRequestS, resParams *map[string]interface{}, eventDataSet *EventsHappenedDataSet) (int, error) {
	if req == nil || resParams == nil || eventDataSet == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.NullPoint()
	}

	(*resParams)["NewCompletedAchievement"] = 0

	if req.Uin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.InvalidUin()
	}

	//base.GLog.Debug("Event Data Set:\n %v", eventDataSet)

	achievementModel := model.AchievementModel{Uin: req.Uin}
	// 从数据库从查询获得的成就
	achievementList, err := achievementModel.GetAchievementList()
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	achievementMap := make(map[int]*table.TblAchievement)
	for _, achievement := range achievementList {
		achievementMap[achievement.ProtypeId] = achievement
	}

	eventProgress := make(map[int]*MissionProgress)

	// 战斗结束
	if eventDataSet.BattleEndData != nil {
		base.GLog.Debug("Uin[%d] BattleEndData:%+v", req.Uin, eventDataSet.BattleEndData)

		switch eventDataSet.BattleEndData.BattleType {
		case BATTLE_TYPE_PVP:
			// PVP历史战斗场次
			if list, ok := config.GAchievementConfig.TypeMap[config.MISSION_TYPE_LEAGUE_PVP_TIMES_HISTORY]; ok {
				for _, protype := range list {
					if achievement, ok := achievementMap[protype.Id]; ok {
						if achievement.Status == model.ACHIEVEMENT_STATUS_COMPLETED {
							continue
						}
					}

					missionProgress := new(MissionProgress)
					missionProgress.Current = eventDataSet.BattleEndData.TotalPvpTimes
					eventProgress[protype.Id] = missionProgress
				}
			}

			// PVP历史最大连击次数
			if list, ok := config.GAchievementConfig.TypeMap[config.MISSION_TYPE_LEAGUE_PVP_COMBO_REACH_HISTORY]; ok {
				for _, protype := range list {
					if achievement, ok := achievementMap[protype.Id]; ok {
						if achievement.Status == model.ACHIEVEMENT_STATUS_COMPLETED {
							continue
						}
						missionProgress := new(MissionProgress)
						err := json.Unmarshal([]byte(achievement.ProgressData), &missionProgress)
						if err != nil {
							base.GLog.Error(err)
							continue
						}

						if missionProgress.Current >= eventDataSet.BattleEndData.MaxCombosHistory {
							continue
						}
					}

					progress := new(MissionProgress)
					progress.Current = eventDataSet.BattleEndData.MaxCombosHistory
					eventProgress[protype.Id] = progress
				}
			}

			// PVP平局
			if eventDataSet.BattleEndData.BattleResult == BATTLE_RESULT_DRAW {
				if list, ok := config.GAchievementConfig.TypeMap[config.MISSION_TYPE_LEAGUE_PVP_DRAW]; ok {
					for _, protype := range list {

						if achievement, ok := achievementMap[protype.Id]; ok {
							if achievement.Status == model.ACHIEVEMENT_STATUS_COMPLETED {
								continue
							}

							missionProgress := new(MissionProgress)
							err := json.Unmarshal([]byte(achievement.ProgressData), missionProgress)
							if err != nil {
								continue
							}

							missionProgress.Current += 1
							eventProgress[protype.Id] = missionProgress
						} else {
							missionProgress := new(MissionProgress)
							missionProgress.Current = 1
							eventProgress[protype.Id] = missionProgress
						}

					}
				}
			}

			// 最终以护卫舰武器攻击获胜
			if eventDataSet.BattleEndData.FrigateWeaponKORival != 0 {
				if list, ok := config.GAchievementConfig.TypeMap[config.MISSION_TYPE_FRIGATE_WEPON_KO_RIVAL]; ok {
					for _, protype := range list {
						if achievement, ok := achievementMap[protype.Id]; ok {
							if achievement.Status == model.ACHIEVEMENT_STATUS_COMPLETED {
								continue
							}
						}

						missionProgress := new(MissionProgress)
						missionProgress.Current = 1
						eventProgress[protype.Id] = missionProgress
					}
				}
			}

			// 以弱胜强
			// 对手评分比自己评分大于50才判断是否符合以弱胜强
			scoreDiff := int(eventDataSet.BattleEndData.RivalScore - eventDataSet.BattleEndData.SrcScore)
			base.GLog.Debug("BattleEndData:%+v", eventDataSet.BattleEndData)
			// BattleResult: 1：失败，2：平手，3：胜利
			if scoreDiff >= config.GGlobalConfig.Fleet.ScoreDelta && eventDataSet.BattleEndData.BattleResult == 3 {
				if list, ok := config.GAchievementConfig.TypeMap[config.MISSION_TYPE_LEAGUE_PVP_WEAK_DEFEAT_STRONG]; ok {
					for _, protype := range list {
						if achievement, ok := achievementMap[protype.Id]; ok {
							if achievement.Status == model.ACHIEVEMENT_STATUS_COMPLETED {
								continue
							}
						}

						missionProgress := new(MissionProgress)
						missionProgress.Current = 1
						eventProgress[protype.Id] = missionProgress
					}
				}
			}

			// 连胜
			if list, ok := config.GAchievementConfig.TypeMap[config.MISSION_TYPE_LEAGUE_PVP_WIN_STREAK]; ok {
				for _, protype := range list {
					missionProgress := new(MissionProgress)
					missionProgress.Current = 0
					if achievement, ok := achievementMap[protype.Id]; ok {
						if achievement.Status == model.ACHIEVEMENT_STATUS_COMPLETED {
							continue
						}
						err = json.Unmarshal([]byte(achievement.ProgressData), missionProgress)
						if err != nil {
							base.GLog.Error(err)
							continue
						}
					}
					if eventDataSet.BattleEndData.BattleResult == BATTLE_RESULT_SUCCESS {
						missionProgress.Current += 1
					} else if eventDataSet.BattleEndData.BattleResult == BATTLE_RESULT_FAILED {
						missionProgress.Current = 0
					}
					eventProgress[protype.Id] = missionProgress
				}
			}

			// 只用1队获得连胜
			if list, ok := config.GAchievementConfig.TypeMap[config.MISSION_TYPE_LEAGUE_PVP_WIN_STREAK_WITH_MAIN_SHIP]; ok {
				for _, protype := range list {
					missionProgress := new(MissionProgress)
					if achievement, ok := achievementMap[protype.Id]; ok {
						if achievement.Status == model.ACHIEVEMENT_STATUS_COMPLETED {
							continue
						}

						err = json.Unmarshal([]byte(achievement.ProgressData), &missionProgress)
						if err != nil {
							base.GLog.Error(err)
							continue
						}
					}
					if eventDataSet.BattleEndData.BattleResult == BATTLE_RESULT_SUCCESS && eventDataSet.BattleEndData.MainFormationAlive == 1 {
						missionProgress.Current += 1
					} else {
						missionProgress.Current = 0
					}
					eventProgress[protype.Id] = missionProgress
				}
			}
		}
	}

	// 联赛升级
	if eventDataSet.LeagueLevelUpEventData != nil {
		base.GLog.Debug("Uin[%d] LeagueLevelUpEventData:%+v", req.Uin, eventDataSet.LeagueLevelUpEventData)

		if list, ok := config.GAchievementConfig.TypeMap[config.MISSION_TYPE_LEAGUE_LEVEL_REACH]; ok {
			for _, protype := range list {
				if achievement, ok := achievementMap[protype.Id]; ok {
					if achievement.Status == model.ACHIEVEMENT_STATUS_COMPLETED {
						continue
					}
				}

				if eventDataSet.LeagueLevelUpEventData.LeagueLevel >= protype.Parameter.Int(config.MISSION_PARAMETER_LEAGUE_LEVEL, 0) {
					missionProgress := new(MissionProgress)
					missionProgress.Current = 1
					eventProgress[protype.Id] = missionProgress
				}
			}
		}
	}

	// 公会职位
	if eventDataSet.GuildPostChangedEventData != nil {
		base.GLog.Debug("Uin[%d] GuildPostChangedEventData:%+v", req.Uin, eventDataSet.GuildPostChangedEventData)

		if eventDataSet.GuildPostChangedEventData.Post == GUILD_POST_CHAIRMAN {
			if list, ok := config.GAchievementConfig.TypeMap[config.MISSION_TYPE_GUILD_CHAIRMAIN]; ok {
				for _, protype := range list {
					if achievement, ok := achievementMap[protype.Id]; ok {
						if achievement.Status == model.ACHIEVEMENT_STATUS_COMPLETED {
							continue
						}
					}
					missionProgress := new(MissionProgress)
					missionProgress.Current = 1
					eventProgress[protype.Id] = missionProgress
				}
			}
		}
	}

	// 日常活跃分
	if eventDataSet.DailyVitalityChangedEventData != nil {
		base.GLog.Debug("Uin[%d] DailyVitalityChangedEventData:%+v", req.Uin, eventDataSet.DailyVitalityChangedEventData)

		if list, ok := config.GAchievementConfig.TypeMap[config.MISSION_TYPE_ACTIVITY_SCORE_OVERFLOW_CONTINUOUS]; ok {
			for _, protype := range list {
				if eventDataSet.DailyVitalityChangedEventData.Vitality < protype.Parameter.Int(config.MISSION_PARAMETER_ACTIVITY_SCORE, 0) {
					continue
				}

				missionProgress := new(MissionProgress)
				if achievement, ok := achievementMap[protype.Id]; ok {
					if achievement.Status == model.ACHIEVEMENT_STATUS_COMPLETED {
						continue
					}

					err = json.Unmarshal([]byte(achievement.ProgressData), missionProgress)
					if err != nil {
						base.GLog.Error(err)
						continue
					}

					if !base.GLocalizedTime.IsYesterday(int64(missionProgress.LastDay)) {
						missionProgress.DayContinuous = 0
					}
				}

				missionProgress.DayContinuous += 1
				missionProgress.LastDay = int(base.GLocalizedTime.SecTimeStamp())
				if missionProgress.DayContinuous >= protype.Parameter.Int(config.MISSION_PARAMETER_DAYS, 0) {
					missionProgress.Current = 1
				} else {
					missionProgress.Current = 0
				}
				eventProgress[protype.Id] = missionProgress
			}
		}
	}

	// 开卡包
	if eventDataSet.CardBagOpenEventData != nil {
		base.GLog.Debug("Uin[%d] CardBagOpenEventData:%+v", req.Uin, eventDataSet.CardBagOpenEventData)

		haveOrangeCards := false
		for _, shipCard := range eventDataSet.CardBagOpenEventData.ShipCards {
			base.GLog.Debug(shipCard)
			if protype, ok := config.GBattleShipProtypeConfig.AttrMap[shipCard.ProtypeId]; ok && len(protype.StarList) > 0 && protype.StarList[0].Rarity == config.BATTLE_SHIP_QUALITY_LEGENDARY {
				haveOrangeCards = true
				break
			}
		}

		if haveOrangeCards {
			if list, ok := config.GAchievementConfig.TypeMap[config.MISSION_TYPE_GAIN_LEGEND_CARD]; ok {
				for _, protype := range list {
					if achievement, ok := achievementMap[protype.Id]; ok {
						if achievement.Status == model.ACHIEVEMENT_STATUS_COMPLETED {
							continue
						}
					}

					if protype.Parameter.Int(config.MISSION_PARAMETER_CARDPACK_ID, 0) != eventDataSet.CardBagOpenEventData.ProtypeId {
						continue
					}

					missionProgress := new(MissionProgress)
					missionProgress.Current = 1
					eventProgress[protype.Id] = missionProgress
				}
			}
		}
	}

	// PVE通关
	if eventDataSet.CampaignPassEventData != nil {
		base.GLog.Debug("Uin[%d] CampaignPassEventData:%+v", req.Uin, eventDataSet.CampaignPassEventData)

		var chapterInfo table.TblCampaignPassChapter
		retCode, err := GetMaxCampaignChapterInfo(req.Uin, &chapterInfo)
		if err != nil {
			return retCode, err
		}

		// 找到最大已通关章节
		passChapter := 0
		campaignProtype := config.GCampaignConfig.AttrMap[chapterInfo.CampaignId]
		if campaignProtype != nil {
			passChapter = campaignProtype.InArea
		}

		if list, ok := config.GCampaignConfig.ChapterMap[passChapter]; ok {
			for _, protype := range list {
				if protype.Id > chapterInfo.CampaignId {
					passChapter -= 1
					break
				}
			}
		}

		// 更新成就
		if list, ok := config.GAchievementConfig.TypeMap[config.MISSION_TYPE_CAMPAIGN_PASS_CHAPTER_HISTORY]; ok {
			for _, protype := range list {
				chapterCondValue := protype.Parameter.Int(config.MISSION_PARAMETER_CHAPTER, 0)
				if chapterCondValue > passChapter {
					continue
				}

				if achievement, ok := achievementMap[protype.Id]; ok {
					if achievement.Status == model.ACHIEVEMENT_STATUS_COMPLETED {
						continue
					}
				}
				missionProgress := new(MissionProgress)
				missionProgress.Current = 1
				eventProgress[protype.Id] = missionProgress
			}
		}

	}

	// 完成PVE事件
	if eventDataSet.CampaignEventAwardReceivedEventData != nil {
		if list, ok := config.GAchievementConfig.TypeMap[config.MISSION_TYPE_CAMPAIGN_COMPLETE_EVENT]; ok {
			for _, protype := range list {

				missionProgress := new(MissionProgress)
				if achievement, ok := achievementMap[protype.Id]; ok {
					if achievement.Status == model.ACHIEVEMENT_STATUS_COMPLETED {
						continue
					}

					err = json.Unmarshal([]byte(achievement.ProgressData), missionProgress)
					if err != nil {
						base.GLog.Error(err)
						continue
					}

				} else {
					missionProgress.Current = 0
				}
				missionProgress.Current += 1
				eventProgress[protype.Id] = missionProgress
			}
		}
	}

	// 新手教学
	if eventDataSet.NewbieTechPassEventData != nil {
		if eventDataSet.NewbieTechPassEventData.Step >= eventDataSet.NewbieTechPassEventData.MaxStep {
			if list, ok := config.GAchievementConfig.TypeMap[config.MISSION_TYPE_TEACHING_CHAPTER_COMPLETE_HISTORY]; ok {
				for _, protype := range list {
					if achievement, ok := achievementMap[protype.Id]; ok {
						if achievement.Status == model.ACHIEVEMENT_STATUS_COMPLETED {
							continue
						}
					}
					missionProgress := new(MissionProgress)
					missionProgress.Current = 1
					eventProgress[protype.Id] = missionProgress
				}
			}
		}
	}

	// 连续登录
	if eventDataSet.LoginEventData != nil {
		if list, ok := config.GAchievementConfig.TypeMap[config.MISSION_TYPE_LOGIN_CONTINUOUS]; ok {
			for _, protype := range list {
				missionProgress := new(MissionProgress)
				if achievement, ok := achievementMap[protype.Id]; ok {
					if achievement.Status == model.ACHIEVEMENT_STATUS_COMPLETED {
						continue
					}

					err = json.Unmarshal([]byte(achievement.ProgressData), missionProgress)
					if err != nil {
						base.GLog.Error(err)
						continue
					}

					if base.GLocalizedTime.IsToday(int64(missionProgress.LastDay)) {
						continue
					}

					if !base.GLocalizedTime.IsYesterday(int64(missionProgress.LastDay)) {
						missionProgress.DayContinuous = 0
					}
				}

				missionProgress.DayContinuous += 1
				missionProgress.LastDay = int(base.GLocalizedTime.SecTimeStamp())
				if missionProgress.DayContinuous >= protype.Parameter.Int(config.MISSION_PARAMETER_DAYS, 0) {
					missionProgress.Current = 1
				} else {
					missionProgress.Current = 0
				}
				eventProgress[protype.Id] = missionProgress
			}
		}
	}

	// 突破重围
	if eventDataSet.BreakOutDoneEventData != nil {
		if eventDataSet.BreakOutDoneEventData.IsAllDone == 1 {
			if list, ok := config.GAchievementConfig.TypeMap[config.MISSION_TYPE_BREAK_OUT_ALL]; ok {
				for _, protype := range list {
					if achievement, ok := achievementMap[protype.Id]; ok {
						if achievement.Status == model.ACHIEVEMENT_STATUS_COMPLETED {
							continue
						}
					}
					missionProgress := new(MissionProgress)
					missionProgress.Current = 1
					eventProgress[protype.Id] = missionProgress
				}
			}
		}
	}

	// 新手引导
	if eventDataSet.NewbieGuidPassEventData != nil {
		if eventDataSet.NewbieGuidPassEventData.IsAllDone == 1 {
			if list, ok := config.GAchievementConfig.TypeMap[config.MISSION_TYPE_TUTORIAL_COMPLETE_HISTORY]; ok {
				for _, protype := range list {
					if achievement, ok := achievementMap[protype.Id]; ok {
						if achievement.Status == model.ACHIEVEMENT_STATUS_COMPLETED {
							continue
						}
					}
					missionProgress := new(MissionProgress)
					missionProgress.Current = 1
					eventProgress[protype.Id] = missionProgress
				}
			}
		}
	}

	// 战舰合成
	if eventDataSet.BattleShipMergeEventData != nil {
		// 获取战舰列表

		battleShipList, retCode, err := GetBattleShipListByUin(req.Uin)
		if err != nil {
			return retCode, err
		}

		validBattleShipList := make([]*table.TblBattleShip, 0, len(battleShipList))
		for _, battleShip := range battleShipList {
			if battleShip.Status == table.BATTLE_SHIP_STATUS_SHIP {
				validBattleShipList = append(validBattleShipList, battleShip)
			}
		}

		if len(validBattleShipList) >= len(config.GBattleShipProtypeConfig.AttrMap) {
			if list, ok := config.GAchievementConfig.TypeMap[config.MISSION_TYPE_GAIN_ALL_BATTLE_SHIPS]; ok {
				for _, protype := range list {
					if achievement, ok := achievementMap[protype.Id]; ok {
						if achievement.Status == model.ACHIEVEMENT_STATUS_COMPLETED {
							continue
						}
					}
					missionProgress := new(MissionProgress)
					missionProgress.Current = 1
					eventProgress[protype.Id] = missionProgress
				}
			}
		}
	}

	// 充值
	if eventDataSet.ChargeEventData != nil {
		if eventDataSet.ChargeEventData.VipLevel > 0 {
			if list, ok := config.GAchievementConfig.TypeMap[config.MISSION_TYPE_FIRST_PURCHASE]; ok {
				for _, protype := range list {

					if achievement, ok := achievementMap[protype.Id]; ok {
						if achievement.Status == model.ACHIEVEMENT_STATUS_COMPLETED {
							continue
						}
					}
					missionProgress := new(MissionProgress)
					missionProgress.Current = 1
					eventProgress[protype.Id] = missionProgress
				}
			}
		}
	}

	// 获得所有成就
	hasGainAllAchievement := true
	for _, protype := range config.GAchievementConfig.AttrMap {
		if protype.AchievementType == config.MISSION_TYPE_GAIN_ALL_ACHIEVEMENTS {
			continue
		}

		if achievement, ok := achievementMap[protype.Id]; ok {
			if achievement.Status == model.ACHIEVEMENT_STATUS_UNCOMPLETED {
				hasGainAllAchievement = false
				break
			}
		} else {
			hasGainAllAchievement = false
			break
		}
	}

	if hasGainAllAchievement {
		if list, ok := config.GAchievementConfig.TypeMap[config.MISSION_TYPE_GAIN_ALL_ACHIEVEMENTS]; ok {
			for _, protype := range list {
				if achievement, ok := achievementMap[protype.Id]; ok {
					if achievement.Status == model.ACHIEVEMENT_STATUS_COMPLETED {
						continue
					}
				}
				missionProgress := new(MissionProgress)
				missionProgress.Current = 1
				eventProgress[protype.Id] = missionProgress
			}
		}
	}

	base.GLog.Debug("EventsProgress: [%+v]", eventProgress)
	newCompletedAchievement := make([]*proto.ProtoAchievementInfo, 0)
	// 更新成就进度和添加新的成就到数据库
	updateAchievements := make([]*table.TblAchievement, 0)
	addAchievements := make([]*table.TblAchievement, 0)
	for proId, prog := range eventProgress {
		if protype, ok := config.GAchievementConfig.AttrMap[proId]; ok {
			achievement, isUpdate := achievementMap[proId]
			if !isUpdate {
				achievement = new(table.TblAchievement)
				achievement.ProtypeId = proId
				achievement.Uin = req.Uin
				achievement.Status = model.ACHIEVEMENT_STATUS_UNCOMPLETED
			}

			// 进度
			prog.Total = protype.Parameter.Int(config.CAMPAIGN_MISSION_PARAMETER_TARGET, 0)
			data, err := json.Marshal(prog)
			if err != nil {
				base.GLog.Error("json marshal achievement progress[%+v] err[%s]", prog, err)
				continue
			}
			achievement.ProgressData = string(data)

			if prog.Current >= prog.Total {
				achievement.Status = model.ACHIEVEMENT_STATUS_COMPLETED
				achievement.CompleteTime = int(base.GLocalizedTime.SecTimeStamp())
			}

			if isUpdate {
				updateAchievements = append(updateAchievements, achievement)
			} else {
				addAchievements = append(addAchievements, achievement)
			}

			if achievement.Status == model.ACHIEVEMENT_STATUS_COMPLETED {
				protoAchievement := new(proto.ProtoAchievementInfo)
				ComposeProtoAchievement(protoAchievement, achievement, protype)
				newCompletedAchievement = append(newCompletedAchievement, protoAchievement)
			}
		}
	}

	if len(addAchievements) > 0 {
		retCode, err := achievementModel.AddMultiAchievements(addAchievements)
		if err != nil {
			return retCode, err
		}
	}

	if len(updateAchievements) > 0 {
		retCode, err := achievementModel.UpdateMultiAchievements(updateAchievements)
		if err != nil {
			return retCode, err
		}
	}

	// 组织返回内容
	if len(newCompletedAchievement) > 0 {
		(*resParams)["NewCompletedAchievement"] = 1
		(*resParams)["CompletedAchievement"] = newCompletedAchievement
	} else {
		(*resParams)["NewCompletedAchievement"] = 0
	}

	return 0, nil
}

/*
	测试接口-------完成成就
*/
func CompleteAchievement(req *base.ProtoRequestS, res *base.ProtoResponseS) (int, error) {
	if req == nil || res == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.NullPoint()
	}

	type ProtoCompleteAchievementRequest struct {
		ProtypeId    int `json:"ProtypeId"`
		CompleteTime int `json:"CompleteTime"`
	}

	var reqParams ProtoCompleteAchievementRequest
	err := mapstructure.Decode(req.ReqData.Params, &reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	protype, ok := config.GAchievementConfig.AttrMap[reqParams.ProtypeId]
	if !ok {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("protype id not exist")
	}

	achivementModel := model.AchievementModel{Uin: req.Uin}
	record, err := achivementModel.GetAchievementByProtypeId(reqParams.ProtypeId)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	isUpdate := true
	if record == nil {
		record.Uin = req.Uin
		record.ProtypeId = protype.Id
		isUpdate = false
	} else {
		if record.Status == model.ACHIEVEMENT_STATUS_COMPLETED {
			responseData := make(map[string]interface{})
			responseData["NewCompletedAchievement"] = 0
			res.ResData.Params = responseData

			return 0, nil
		}
	}

	progress := new(MissionProgress)
	progress.Current = protype.Parameter.Int(config.MISSION_PARAMETER_TARGET, 0)
	progress.Total = progress.Current

	data, err := json.Marshal(progress)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	record.ProgressData = string(data)
	record.Status = model.ACHIEVEMENT_STATUS_COMPLETED
	record.CompleteTime = reqParams.CompleteTime

	if isUpdate {
		_, err = achivementModel.UpdateAchievement(record)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}
	} else {
		_, err = achivementModel.AddAchievement(record)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}
	}

	responseData := make(map[string]interface{})

	protoAchivement := new(proto.ProtoAchievementInfo)
	ComposeProtoAchievement(protoAchivement, record, nil)

	protoCompletedAchievements := make([]*proto.ProtoAchievementInfo, 0)
	protoCompletedAchievements = append(protoCompletedAchievements, protoAchivement)

	responseData["NewCompletedAchievement"] = 1
	responseData["CompletedAchievement"] = protoCompletedAchievements

	res.ResData.Params = responseData

	return 0, nil
}

/*
	测试接口-------添加成就进度
*/
func AddAchievementProgress(req *base.ProtoRequestS, res *base.ProtoResponseS) (int, error) {
	if req == nil || res == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.NullPoint()
	}

	type ProtoAddAchievementProgressRequest struct {
		ProtypeId int `json:"ProtypeId"`
		Progress  int `json:"Progress"`
	}

	var reqParams ProtoAddAchievementProgressRequest
	err := mapstructure.Decode(req.ReqData.Params, &reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	protype, ok := config.GAchievementConfig.AttrMap[reqParams.ProtypeId]
	if !ok {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("protype id not exist")
	}

	achivementModel := model.AchievementModel{Uin: req.Uin}
	record, err := achivementModel.GetAchievementByProtypeId(reqParams.ProtypeId)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	progress := new(MissionProgress)
	isUpdate := true
	if record == nil {
		record.Uin = req.Uin
		record.ProtypeId = protype.Id
		isUpdate = false
	} else {
		err = json.Unmarshal([]byte(record.ProgressData), progress)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}
	}

	progress.Current += reqParams.Progress
	progress.Total = protype.Parameter.Int(config.MISSION_PARAMETER_TARGET, 0)

	completedAchievements := make([]*table.TblAchievement, 0)
	if progress.Current >= progress.Total {
		progress.Current = progress.Total
		if record.Status != model.ACHIEVEMENT_STATUS_COMPLETED {
			record.Status = model.ACHIEVEMENT_STATUS_COMPLETED
			record.CompleteTime = int(time.Now().Unix())
			completedAchievements = append(completedAchievements, record)
		}
	} else {
		record.Status = model.ACHIEVEMENT_STATUS_UNCOMPLETED
	}

	data, err := json.Marshal(progress)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	record.ProgressData = string(data)

	if isUpdate {
		_, err = achivementModel.UpdateAchievement(record)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}
	} else {
		_, err = achivementModel.AddAchievement(record)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}
	}

	responseData := make(map[string]interface{})
	if len(completedAchievements) > 0 {
		protoCompletedAchievements := make([]*proto.ProtoAchievementInfo, 0)
		for _, tblAchievement := range completedAchievements {
			protoAchivement := new(proto.ProtoAchievementInfo)
			ComposeProtoAchievement(protoAchivement, tblAchievement, nil)
			protoCompletedAchievements = append(protoCompletedAchievements, protoAchivement)
		}
		responseData["NewCompletedAchievement"] = 1
		responseData["CompletedAchievement"] = protoCompletedAchievements
	} else {
		responseData["NewCompletedAchievement"] = 0
	}
	res.ResData.Params = responseData

	return 0, nil
}

/*
	测试接口-------设置成就进度
*/
func SetAchievementProgress(req *base.ProtoRequestS, res *base.ProtoResponseS) (int, error) {
	if req == nil || res == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.NullPoint()
	}

	type ProtoSetAchievementProgressRequest struct {
		ProtypeId int `json:"ProtypeId"`
		Progress  int `json:"Progress"`
	}

	var reqParams ProtoSetAchievementProgressRequest
	err := mapstructure.Decode(req.ReqData.Params, &reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	protype, ok := config.GAchievementConfig.AttrMap[reqParams.ProtypeId]
	if !ok {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("protype id not exist")
	}

	achivementModel := model.AchievementModel{Uin: req.Uin}
	record, err := achivementModel.GetAchievementByProtypeId(reqParams.ProtypeId)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	isUpdate := true
	if record == nil {
		record.Uin = req.Uin
		record.ProtypeId = protype.Id
		isUpdate = false
	}

	progress := new(MissionProgress)
	progress.Current = reqParams.Progress
	progress.Total = protype.Parameter.Int(config.MISSION_PARAMETER_TARGET, 0)

	completedAchievements := make([]*table.TblAchievement, 0)
	if progress.Current >= progress.Total {
		progress.Current = progress.Total
		if record.Status != model.ACHIEVEMENT_STATUS_COMPLETED {
			record.Status = model.ACHIEVEMENT_STATUS_COMPLETED
			record.CompleteTime = int(time.Now().Unix())
			completedAchievements = append(completedAchievements, record)
		}
	} else {
		record.Status = model.ACHIEVEMENT_STATUS_UNCOMPLETED
		record.CompleteTime = 0
	}

	data, err := json.Marshal(progress)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	record.ProgressData = string(data)

	if isUpdate {
		_, err = achivementModel.UpdateAchievement(record)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}
	} else {
		_, err = achivementModel.AddAchievement(record)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}
	}

	responseData := make(map[string]interface{})
	if len(completedAchievements) > 0 {
		protoCompletedAchievements := make([]*proto.ProtoAchievementInfo, 0)
		for _, tblAchievement := range completedAchievements {
			protoAchivement := new(proto.ProtoAchievementInfo)
			ComposeProtoAchievement(protoAchivement, tblAchievement, nil)
			protoCompletedAchievements = append(protoCompletedAchievements, protoAchivement)
		}
		responseData["NewCompletedAchievement"] = 1
		responseData["CompletedAchievement"] = protoCompletedAchievements
	} else {
		responseData["NewCompletedAchievement"] = 0
	}
	res.ResData.Params = responseData

	return 0, nil
}

func ComposeProtoAchievement(target *proto.ProtoAchievementInfo, data *table.TblAchievement, protype *config.AchievementProtype) {
	if target == nil || (data == nil && protype == nil) {
		base.GLog.Error("null point")
		return // do nothing
	}

	if data != nil {
		target.ProtypeId = data.ProtypeId
		target.Status = data.Status
		target.CompleteTime = data.CompleteTime

		var progress MissionProgress
		err := json.Unmarshal([]byte(data.ProgressData), &progress)
		if err != nil {
			base.GLog.Error("json unmarshal achievement[%+v] progress error[%s]", data, err)
			target.CurrentProgress = 0
			target.TotalProgress = 0
			return
		}

		target.CurrentProgress = progress.Current
		target.TotalProgress = progress.Total
		return
	}

	if protype != nil {
		target.ProtypeId = protype.Id
		target.Status = model.ACHIEVEMENT_STATUS_UNCOMPLETED
		target.CurrentProgress = 0
		target.TotalProgress = protype.Parameter.Int(config.MISSION_PARAMETER_TARGET, 0)
		target.CompleteTime = 0
	}
}
