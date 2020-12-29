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
)

type GrowupTaskHandler struct {
	handlerbase.WebHandler
}

func (this *GrowupTaskHandler) List() (int, error) {
	growupTaskModel := model.GrowupTaskModel{Uin: this.Request.Uin}
	tblTaskList, err := growupTaskModel.GetGrowupTaskList()
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	var userInfo table.TblUserInfo
	retCode, err := GetUserInfo(this.Request.Uin, &userInfo)
	if err != nil {
		return retCode, err
	}

	taskMap := make(map[string]*table.TblGrowupTask)
	for _, tblTask := range tblTaskList {
		taskMap[tblTask.ProtypeId] = tblTask
	}

	taskList := make([]*proto.ProtoGrowupTaskInfo, 0)

	for _, protype := range config.GGrowupTaskConfig.AttrMap {
		// 等级未到
		if protype.LevelLimit > userInfo.Level {
			continue
		}

		// 前置任务未完成
		if _, ok := config.GGrowupTaskConfig.AttrMap[protype.PreTaskId]; ok {
			if tblTask, ok := taskMap[protype.PreTaskId]; !ok || tblTask.Status != model.GROWUP_TASK_STATUS_COMPLETED {
				continue
			}
		}

		tblTask, ok := taskMap[protype.Id]
		if ok {
			if tblTask.Status == model.GROWUP_TASK_STATUS_COMPLETED {
				// 当前任务已完成，有后置任务，且后置任务达到显示等级
				hasNextTask := false
				for _, nextProtypeId := range protype.NextTaskIds {
					if nextProtype, ok := config.GGrowupTaskConfig.AttrMap[nextProtypeId]; ok && nextProtype.LevelLimit <= userInfo.Level {
						hasNextTask = true
						break
					}
				}

				if hasNextTask {
					continue
				}
			}

			task := new(proto.ProtoGrowupTaskInfo)
			composeProtoGrowupTask(task, tblTask, protype)
			taskList = append(taskList, task)
		} else {
			task := new(proto.ProtoGrowupTaskInfo)
			composeProtoGrowupTask(task, nil, protype)
			taskList = append(taskList, task)
		}
	}

	var responseData proto.ProtoGetGrowupTaskListResponse
	responseData.GrowupTaskList = taskList

	this.Response.ResData.Params = responseData

	return 0, nil
}

func (this *GrowupTaskHandler) ReceiveReward() (int, error) {
	var reqParams proto.ProtoReceiveGrowupTaskRewardRequest
	err := this.UnpackParams(&reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	// 1. 检测该任务是否存在
	protype, ok := config.GGrowupTaskConfig.AttrMap[reqParams.ProtypeId]
	if !ok {
		return errorcode.ERROR_CODE_GROWUP_TASK_ID_NOT_EXIST, custom_errors.New("growup task[%s] not exist", reqParams.ProtypeId)
	}

	// 2. 检测任务是否达到等级限制
	var userInfo table.TblUserInfo
	retCode, err := GetUserInfo(this.Request.Uin, &userInfo)
	if err != nil {
		return retCode, err
	}

	// 3. 检测任务是否可领取
	growupTaskModel := model.GrowupTaskModel{Uin: this.Request.Uin}

	// 3.1 检测前置任务是否已完成并已领取
	if protype.PreTaskId != "" {
		tblPreTask, err := growupTaskModel.GetGrowupTaskByProtypeId(protype.PreTaskId)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		if tblPreTask == nil || tblPreTask.Status != model.GROWUP_TASK_STATUS_COMPLETED {
			return errorcode.ERROR_CODE_GROWUP_TASK_NOT_REACH, custom_errors.New("growup task[%s] can not receive", reqParams.ProtypeId)
		}
	}

	// 3.2 检测当前任务是否已完成
	tblTask, err := growupTaskModel.GetGrowupTaskByProtypeId(reqParams.ProtypeId)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	// 未完成
	if tblTask == nil || tblTask.Status == model.GROWUP_TASK_STATUS_UNCOMPLETED {
		return errorcode.ERROR_CODE_GROWUP_TASK_NOT_REACH, custom_errors.New("growup task[%s] uncompleted", reqParams.ProtypeId)
	}

	// 已领取
	if tblTask.Status == model.GROWUP_TASK_STATUS_COMPLETED {
		return errorcode.ERROR_CODE_GROWUP_TASK_REWARD_RECEIVED_REPEAT, custom_errors.New("growup task[%s] has already received", reqParams.ProtypeId)
	}

	// 更新任务状态
	tblTask.Status = model.GROWUP_TASK_STATUS_COMPLETED
	retCode, err = growupTaskModel.UpdateGrowupTask(tblTask)
	if err != nil {
		return retCode, err
	}

	var responseData proto.ProtoReceiveGrowupTaskRewardResponse

	// 任务新状态
	updatedTask := new(proto.ProtoGrowupTaskInfo)
	composeProtoGrowupTask(updatedTask, tblTask, protype)
	responseData.TaskList = append(responseData.TaskList, updatedTask)

	// 任务奖励
	ResourcesConfigToProto(&protype.Reward, &responseData.Rewards)

	// 判断后置任务
	if len(protype.NextTaskIds) > 0 {
		for _, nextProtypeId := range protype.NextTaskIds {
			if nextProtype, ok := config.GGrowupTaskConfig.AttrMap[nextProtypeId]; ok {
				if nextProtype.LevelLimit <= userInfo.Level {
					nextTblTask, err := growupTaskModel.GetGrowupTaskByProtypeId(nextProtypeId)
					if err != nil {
						return errorcode.ERROR_CODE_DEFAULT, err
					}

					task := new(proto.ProtoGrowupTaskInfo)
					composeProtoGrowupTask(task, nextTblTask, nextProtype)
					responseData.TaskList = append(responseData.TaskList, task)
				}
			}
		}
	}

	this.Response.ResData.Params = responseData

	return 0, nil
}

func HandleGrowupTaskProgress(req *base.ProtoRequestS, resParams *map[string]interface{}, eventDataSet *EventsHappenedDataSet) (int, error) {
	base.GLog.Debug("HandlerGrowupTaskProgress")
	if req == nil || resParams == nil || eventDataSet == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.NullPoint()
	}

	if req.Uin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.InvalidUin()
	}

	(*resParams)["NewGrowupTasksTriggered"] = 0
	(*resParams)["NewGrowupTasksWaitingForReceive"] = 0

	growupTaskModel := model.GrowupTaskModel{Uin: req.Uin}
	tblTaskList, err := growupTaskModel.GetGrowupTaskList()
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	taskMap := make(map[string]*table.TblGrowupTask)
	for _, tblTask := range tblTaskList {
		taskMap[tblTask.ProtypeId] = tblTask
	}

	eventProgress := make(map[string]*MissionProgress)
	if eventDataSet.PlayerLevelUpEventData != nil {
		// 等级上升任务
		if list, ok := config.GGrowupTaskConfig.TypeMap[config.MISSION_TYPE_PLAYER_LEVEL_UP]; ok {
			for _, protype := range list {
				if task, ok := taskMap[protype.Id]; ok {
					if task.Status != model.GROWUP_TASK_STATUS_UNCOMPLETED {
						continue
					}
				}

				progress := new(MissionProgress)
				progress.Current = eventDataSet.PlayerLevelUpEventData.Level
				progress.Total = protype.Parameter.Int(config.MISSION_PARAMETER_TARGET, 0)
				eventProgress[protype.Id] = progress
			}
		}

		// 等级上升触发新任务
		newTriggeredTaskList := make([]*proto.ProtoGrowupTaskInfo, 0)
		for _, protype := range config.GGrowupTaskConfig.AttrMap {
			// 等级达到限制
			if protype.LevelLimit > eventDataSet.PlayerLevelUpEventData.Level {
				continue
			}

			// 由于等级上升导致条件满足
			if protype.LevelLimit <= eventDataSet.PlayerLevelUpEventData.Level && protype.LevelLimit > eventDataSet.PlayerLevelUpEventData.OldLevel {
				// 前置任务完成
				if preTask, ok := taskMap[protype.PreTaskId]; !ok || preTask.Status != model.GROWUP_TASK_STATUS_COMPLETED {
					continue
				}

				// 当前任务非已完成状态
				if task, ok := taskMap[protype.Id]; ok {
					if task.Status == model.GROWUP_TASK_STATUS_COMPLETED {
						continue
					}
				}

				base.GLog.Debug("Triggered new growup task[%+v]", protype)
				tblTask := taskMap[protype.Id]
				protoTask := new(proto.ProtoGrowupTaskInfo)
				composeProtoGrowupTask(protoTask, tblTask, protype) // tblTask为nil时，该函数会构造一个进度为0的任务结构
				newTriggeredTaskList = append(newTriggeredTaskList, protoTask)
			}
		}

		if len(newTriggeredTaskList) > 0 {
			(*resParams)["HaveNewGrowupTasks"] = 1
			(*resParams)["NewGrowupTasks"] = newTriggeredTaskList
		}
	}

	if eventDataSet.BattleShipLevelUpEventData != nil || eventDataSet.BattleShipStarLevelUpEventData != nil || eventDataSet.BattleShipMergeEventData != nil {
		// 获取战舰列表
		battleShipList, retCode, err := GetBattleShipListByUin(req.Uin)
		if err != nil {
			return retCode, err
		}

		// 战舰升级
		if eventDataSet.BattleShipLevelUpEventData != nil {
			if list, ok := config.GGrowupTaskConfig.TypeMap[config.MISSION_TYPE_BATTLE_SHIP_LEVEL_UP]; ok {
				for _, protype := range list {
					if task, ok := taskMap[protype.Id]; ok {
						if task.Status != model.GROWUP_TASK_STATUS_UNCOMPLETED {
							continue
						}
					}

					levelCond := protype.Parameter.Int(config.MISSION_PARAMETER_LEVEL, 0)

					progress := new(MissionProgress)
					progress.Current = 0
					countLimit := protype.Parameter.Int(config.MISSION_PARAMETER_TARGET, 0)

					for _, battleShip := range battleShipList {
						if battleShip.Level >= levelCond {
							progress.Current += 1
						}
						if progress.Current >= countLimit {
							break
						}
					}

					eventProgress[protype.Id] = progress
				}
			}
		}

		// 战舰升星
		if eventDataSet.BattleShipStarLevelUpEventData != nil {
			if list, ok := config.GGrowupTaskConfig.TypeMap[config.MISSION_TYPE_BATTLE_SHIP_STRENGTHEN]; ok {
				for _, protype := range list {
					if task, ok := taskMap[protype.Id]; ok {
						if task.Status != model.GROWUP_TASK_STATUS_UNCOMPLETED {
							continue
						}
					}

					starLevelCond := protype.Parameter.Int(config.MISSION_PARAMETER_STAR_LEVEL, 0)

					progress := new(MissionProgress)
					progress.Current = 0
					countLimit := protype.Parameter.Int(config.MISSION_PARAMETER_TARGET, 0)

					for _, battleShip := range battleShipList {
						if battleShip.StarLevel >= starLevelCond {
							progress.Current += 1
						}
						if progress.Current >= countLimit {
							break
						}
					}

					eventProgress[protype.Id] = progress
				}
			}
		}

		// 战舰合成
		if eventDataSet.BattleShipMergeEventData != nil {
			if list, ok := config.GGrowupTaskConfig.TypeMap[config.MISSION_TYPE_COLLECT_BATTLE_SHIP]; ok {
				for _, protype := range list {
					if task, ok := taskMap[protype.Id]; ok {
						if task.Status != model.GROWUP_TASK_STATUS_UNCOMPLETED {
							continue
						}
					}

					qualityCond := protype.Parameter.String(config.MISSION_PARAMETER_QUALITY, "")

					progress := new(MissionProgress)
					progress.Current = 0
					countLimit := protype.Parameter.Int(config.MISSION_PARAMETER_TARGET, 0)

					for _, battleShip := range battleShipList {
						shipStarLevelProtype, err := config.GBattleShipProtypeConfig.GetStarAttr(battleShip.ProtypeID, battleShip.StarLevel)
						if err != nil {
							continue
						}
						if shipStarLevelProtype.Rarity == qualityCond {
							progress.Current += 1
						}
						if progress.Current >= countLimit {
							break
						}
					}

					eventProgress[protype.Id] = progress
				}
			}
		}
	}

	// 护卫舰升级
	if eventDataSet.FrigateLevelUpEventData != nil {
		if list, ok := config.GGrowupTaskConfig.TypeMap[config.MISSION_TYPE_FRIGATE_LEVEL_UP]; ok {
			for _, protype := range list {
				if task, ok := taskMap[protype.Id]; ok {
					if task.Status != model.GROWUP_TASK_STATUS_UNCOMPLETED {
						continue
					}
				}

				progress := new(MissionProgress)
				progress.Current = eventDataSet.FrigateLevelUpEventData.Level
				eventProgress[protype.Id] = progress
			}
		}
	}

	// 护卫舰武器升级
	if eventDataSet.FirgateWeaponLevelUpEventData != nil {
		if list, ok := config.GGrowupTaskConfig.TypeMap[config.MISSION_TYPE_FRIGATE_WEAPON_LEVEL_UP]; ok {
			for _, protype := range list {
				if protype.Parameter.Int(config.MISSION_PARAMETER_WEAPON_ID, 0) != eventDataSet.FirgateWeaponLevelUpEventData.ProtypeId {
					continue
				}

				if task, ok := taskMap[protype.Id]; ok {
					if task.Status != model.GROWUP_TASK_STATUS_UNCOMPLETED {
						continue
					}
				}

				progress := new(MissionProgress)
				progress.Current = eventDataSet.FirgateWeaponLevelUpEventData.Level
				eventProgress[protype.Id] = progress
			}
		}
	}

	// 护卫舰技能升级
	if eventDataSet.FrigateSkillLevelUpEventData != nil {
		if list, ok := config.GGrowupTaskConfig.TypeMap[config.MISSION_TYPE_FRIGATE_SPELL_LEVEL_UP]; ok {
			for _, protype := range list {
				if protype.Parameter.Int(config.MISSION_PARAMETER_SPELL_ID, 0) != eventDataSet.FrigateSkillLevelUpEventData.ProtypeId {
					continue
				}

				if task, ok := taskMap[protype.Id]; ok {
					if task.Status != model.GROWUP_TASK_STATUS_UNCOMPLETED {
						continue
					}
				}

				progress := new(MissionProgress)
				progress.Current = eventDataSet.FrigateSkillLevelUpEventData.Level
				eventProgress[protype.Id] = progress
			}
		}
	}

	// PVE通关
	if eventDataSet.CampaignPassEventData != nil {
		var chapterInfo table.TblCampaignPassChapter
		campaignPassChapterModel := model.CampaignPassChapterModel{Uin: req.Uin}
		retCode, err := campaignPassChapterModel.QueryMaxChapterInfoByUin(&chapterInfo)
		if err != nil {
			return retCode, err
		}

		if list, ok := config.GGrowupTaskConfig.TypeMap[config.MISSION_TYPE_CAMPAIGN_PASS_LEVEL]; ok {
			for _, protype := range list {
				if protype.Parameter.Int(config.MISSION_PARAMETER_LEVEL, 0) > chapterInfo.CampaignId {
					continue
				}

				if task, ok := taskMap[protype.Id]; ok {
					if task.Status != model.GROWUP_TASK_STATUS_UNCOMPLETED {
						continue
					}
				}

				progress := new(MissionProgress)
				progress.Current = 1
				eventProgress[protype.Id] = progress
			}
		}
	}

	// 联赛升级
	if eventDataSet.LeagueLevelUpEventData != nil {
		if list, ok := config.GGrowupTaskConfig.TypeMap[config.MISSION_TYPE_LEAGUE_LEVEL_UP]; ok {
			for _, protype := range list {
				if task, ok := taskMap[protype.Id]; ok {
					if task.Status != model.GROWUP_TASK_STATUS_UNCOMPLETED {
						continue
					}
				}

				progress := new(MissionProgress)
				progress.Current = eventDataSet.LeagueLevelUpEventData.MaxLeagueLevel
				eventProgress[protype.Id] = progress
			}
		}
	}

	// PVE生产资源领取
	if eventDataSet.CampaignProduceResourcesReceivedEventData != nil {
		if list, ok := config.GGrowupTaskConfig.TypeMap[config.MISSION_TYPE_CAMPAIGN_COLLECT_RESOURCES]; ok {
			for _, protype := range list {
				progress := new(MissionProgress)
				if task, ok := taskMap[protype.Id]; ok {
					if task.Status != model.GROWUP_TASK_STATUS_UNCOMPLETED {
						continue
					}

					err := json.Unmarshal([]byte(task.ProgressData), progress)
					if err != nil {
						continue
					}
				} else {
					progress.Current = 0
				}

				count := 0
				resTypeCond := protype.Parameter.String(config.MISSION_PARAMETER_RESOURCE_TYPE, "")
				for _, resItem := range eventDataSet.CampaignProduceResourcesReceivedEventData.Rewards.Resources {
					if resItem.Type == resTypeCond {
						count += resItem.Count
					}
				}

				if count > 0 || progress.Current > 0 {
					progress.Current += count
					eventProgress[protype.Id] = progress
				}
			}
		}
	}

	// 更新或者添加到数据库
	completedGrowupTasks := make([]*table.TblGrowupTask, 0)
	updateGrowupTasks := make([]*table.TblGrowupTask, 0)
	addGrowupTasks := make([]*table.TblGrowupTask, 0)
	for proId, prog := range eventProgress {
		if protype, ok := config.GGrowupTaskConfig.AttrMap[proId]; ok {
			tblTask, isUpdate := taskMap[proId]
			if !isUpdate {
				tblTask = new(table.TblGrowupTask)
				tblTask.ProtypeId = proId
				tblTask.Uin = req.Uin
			}

			// 进度
			prog.Total = protype.Parameter.Int(config.CAMPAIGN_MISSION_PARAMETER_TARGET, 0)
			if prog.Current > prog.Total {
				prog.Current = prog.Total
			}
			data, err := json.Marshal(prog)
			if err != nil {
				base.GLog.Error("json marshal activity task progress[%+v] err[%s]", prog, err)
				continue
			}
			tblTask.ProgressData = string(data)

			if prog.Current >= prog.Total {
				tblTask.Status = model.GROWUP_TASK_STATUS_WAIT_FOR_RECEIVE
			}

			if isUpdate {
				updateGrowupTasks = append(updateGrowupTasks, tblTask)
			} else {
				addGrowupTasks = append(addGrowupTasks, tblTask)
			}

			if tblTask.Status == model.GROWUP_TASK_STATUS_WAIT_FOR_RECEIVE {
				if preTaskProtype, ok := config.GGrowupTaskConfig.AttrMap[protype.PreTaskId]; ok {
					if preTask, ok := taskMap[preTaskProtype.Id]; !ok || preTask.Status != model.GROWUP_TASK_STATUS_COMPLETED {
						continue
					}
				}

				completedGrowupTasks = append(completedGrowupTasks, tblTask)
			}
		}
	}

	newTaskWaitingForReceive := make([]*proto.ProtoGrowupTaskInfo, 0)
	if len(completedGrowupTasks) > 0 {
		var userInfo table.TblUserInfo
		_, err := GetUserInfo(req.Uin, &userInfo)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		for _, tblTask := range completedGrowupTasks {
			if protype, ok := config.GGrowupTaskConfig.AttrMap[tblTask.ProtypeId]; ok {
				if userInfo.Level >= protype.LevelLimit {
					protoTask := new(proto.ProtoGrowupTaskInfo)
					composeProtoGrowupTask(protoTask, tblTask, protype)
					newTaskWaitingForReceive = append(newTaskWaitingForReceive, protoTask)
				}
			}
		}
	}

	// 更新数据库
	if len(updateGrowupTasks) > 0 {
		retCode, err := growupTaskModel.UpdateMultiGrowupTask(updateGrowupTasks)
		if err != nil {
			return retCode, err
		}
	}

	if len(addGrowupTasks) > 0 {
		retCode, err := growupTaskModel.AddMultiGrowupTask(addGrowupTasks)
		if err != nil {
			return retCode, err
		}
	}

	// 返回内容
	if len(newTaskWaitingForReceive) > 0 {
		(*resParams)["HaveNewGrowupTasksCanReceive"] = 1
		(*resParams)["NewGrowupTasksCanReceive"] = newTaskWaitingForReceive
	}

	return 0, nil
}

func composeProtoGrowupTask(target *proto.ProtoGrowupTaskInfo, data *table.TblGrowupTask, protype *config.GrowUpTaskProtype) {
	if target == nil || (data == nil && protype == nil) {
		base.GLog.Error("null point")
		return
	}

	if data != nil {
		var progress MissionProgress
		err := json.Unmarshal([]byte(data.ProgressData), &progress)
		if err != nil {
			base.GLog.Error("json unmarshal progress[%+v] error[%s]", data.ProgressData, err)
			return
		}

		target.Current = progress.Current
		target.Total = progress.Total
		target.ProtypeId = data.ProtypeId
		target.Status = data.Status
	} else if protype != nil {
		target.Current = 0
		target.Total = protype.Parameter.Int(config.MISSION_PARAMETER_TARGET, 0)
		target.ProtypeId = protype.Id
		target.Status = model.GROWUP_TASK_STATUS_UNCOMPLETED
	}
}
