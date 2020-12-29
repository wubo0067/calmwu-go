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
	"sailcraft/fleetsvr_main/utils"
	"sort"
	"time"
)

type ActivityTaskHandler struct {
	handlerbase.WebHandler
}

func (this *ActivityTaskHandler) SimpleInfo() (int, error) {
	activityTaskFreshModel := model.ActivityTaskFreshModel{Uin: this.Request.Uin}
	freshInfo, err := activityTaskFreshModel.GetActivityTaskFresh()
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	timeZone := utils.QueryPlayerTimeZone(this.Request.Uin)
	base.GLocalizedTime.SetLocale(timeZone)
	defer base.GLocalizedTime.SetLocale("Local")

	refreshTime, err := base.GLocalizedTime.TodayClock(23, 59, 59)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	var responseData proto.ProtoGetActivitySimpleInfoResponse
	responseData.Uin = this.Request.Uin
	if freshInfo == nil || !base.GLocalizedTime.IsToday(int64(freshInfo.FreshTime)) {
		responseData.Vitality = 0
	} else {
		responseData.Vitality = freshInfo.Score
	}
	responseData.RestTimeToReset = int(refreshTime.Sub(base.GLocalizedTime.Now()).Seconds())

	this.Response.ResData.Params = responseData

	return 0, nil
}

// 获取每日任务相关信息（任务列表，奖励列表，活跃度）
func (this *ActivityTaskHandler) DetailInfo() (int, error) {
	// calmwu
	timeZone := utils.QueryPlayerTimeZone(this.Request.Uin)
	base.GLocalizedTime.SetLocale(timeZone)
	defer base.GLocalizedTime.SetLocale("Local")

	activityTaskFreshModel := model.ActivityTaskFreshModel{Uin: this.Request.Uin}
	freshInfo, err := activityTaskFreshModel.GetActivityTaskFresh()
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	var responseData proto.ProtoGetActivityInfoResponse
	responseData.Uin = this.Request.Uin
	refreshTime, err := base.GLocalizedTime.TodayClock(23, 59, 59)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}
	responseData.RestTimeToReset = int(refreshTime.Sub(base.GLocalizedTime.Now()).Seconds())

	protoTaskList := make([]*proto.ProtoActivityTaskInfo, 0)
	protoActivityScoreRewardList := make([]*proto.ProtoActivityScoreRewardInfo, 0)

	if freshInfo != nil && base.GLocalizedTime.IsToday(int64(freshInfo.FreshTime)) {
		// 任务列表
		activityTaskModel := model.ActivityTaskModel{Uin: this.Request.Uin}
		tblTaskList, err := activityTaskModel.GetActivityTaskList()
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		taskMap := make(map[int]*table.TblActivityTask)
		for _, tblTask := range tblTaskList {
			taskMap[tblTask.ProtypeId] = tblTask
		}

		for _, protype := range config.GActivityTaskConfig.AttrMap {
			protoTask := new(proto.ProtoActivityTaskInfo)
			if tblTask, ok := taskMap[protype.Id]; ok {
				err = composeProtoActivityTask(protoTask, tblTask, protype)
			} else {
				err = composeProtoActivityTask(protoTask, nil, protype)
			}
			if err != nil {
				return errorcode.ERROR_CODE_DEFAULT, err
			}
			protoTaskList = append(protoTaskList, protoTask)
		}

		// 活跃度奖励列表
		activityScoreRewardModel := model.ActivityScoreRewardModel{Uin: this.Request.Uin}
		tblRewardList, err := activityScoreRewardModel.GetActivityScoreRewardList()
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		rewardMap := make(map[string]*table.TblActivityScoreReward)
		for _, tblReward := range tblRewardList {
			rewardMap[tblReward.RewardId] = tblReward
		}

		for _, protype := range config.GActivityScoreRewardConfig.AttrMap {
			protoReward := new(proto.ProtoActivityScoreRewardInfo)
			if tblReward, ok := rewardMap[protype.Id]; ok {
				err = composeProtoActivityScoreReward(protoReward, tblReward, protype)
			} else {
				err = composeProtoActivityScoreReward(protoReward, nil, protype)
			}

			if err != nil {
				return errorcode.ERROR_CODE_DEFAULT, err
			}

			protoActivityScoreRewardList = append(protoActivityScoreRewardList, protoReward)
		}

		responseData.Vitality = freshInfo.Score
		responseData.ActivityTaskFreshed = 0
	} else {
		// 任务列表
		for _, protype := range config.GActivityTaskConfig.AttrMap {
			protoTask := new(proto.ProtoActivityTaskInfo)
			err = composeProtoActivityTask(protoTask, nil, protype)
			if err != nil {
				return errorcode.ERROR_CODE_DEFAULT, err
			}
			protoTaskList = append(protoTaskList, protoTask)
		}

		// 活跃度奖励列表
		for _, protype := range config.GActivityScoreRewardConfig.AttrMap {
			protoReward := new(proto.ProtoActivityScoreRewardInfo)
			err = composeProtoActivityScoreReward(protoReward, nil, protype)
			if err != nil {
				return errorcode.ERROR_CODE_DEFAULT, err
			}

			protoActivityScoreRewardList = append(protoActivityScoreRewardList, protoReward)
		}

		responseData.Vitality = 0
		responseData.ActivityTaskFreshed = 1
	}

	sort.Slice(protoTaskList, func(i, j int) bool {
		if protoTaskList[i].Status == model.ACTIVITY_TASK_STATUS_COMPLETED {
			return true
		}

		if protoTaskList[j].Status == model.ACTIVITY_TASK_STATUS_COMPLETED {
			return false
		}

		return protoTaskList[i].ProtypeId < protoTaskList[j].ProtypeId
	})

	responseData.RewardList = protoActivityScoreRewardList
	responseData.TaskList = protoTaskList

	this.Response.ResData.Params = responseData

	return 0, nil
}

func (this *ActivityTaskHandler) Status() (int, error) {
	activityTaskFreshModel := model.ActivityTaskFreshModel{Uin: this.Request.Uin}
	freshInfo, err := activityTaskFreshModel.GetActivityTaskFresh()
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	// calmwu
	timeZone := utils.QueryPlayerTimeZone(this.Request.Uin)
	base.GLocalizedTime.SetLocale(timeZone)
	defer base.GLocalizedTime.SetLocale("Local")

	taskList := make([]*table.TblActivityTask, 0)
	if freshInfo == nil && !base.GLocalizedTime.IsToday(int64(freshInfo.FreshTime)) {
		list, err := resetActivityTask(this.Request.Uin)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}
		_, err = resetActivityScoreReward(this.Request.Uin)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		if freshInfo == nil {
			freshInfo = new(table.TblActivityTaskFresh)
			freshInfo.Uin = this.Request.Uin
			freshInfo.Score = 0
		}

		freshInfo.FreshTime = int(base.GLocalizedTime.SecTimeStamp())
		_, err = activityTaskFreshModel.UpdateActivityTaskFresh(freshInfo)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		taskList = append(taskList, list...)
	} else {
		activityTaskModel := model.ActivityTaskModel{Uin: this.Request.Uin}
		list, err := activityTaskModel.GetActivityTaskList()
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		taskList = append(taskList, list...)
	}

	taskList = sortTblActivityTask(taskList)
	if len(taskList) > 0 && taskList[0].Status != model.ACTIVITY_TASK_STATUS_COMPLETED {
		var responseData proto.ProtoGetActivityTaskStatusNormalResponse
		switch taskList[0].Status {
		case model.ACTIVITY_TASK_STATUS_WAIT_FOR_RECEIVE:
			responseData.Status = 0
		case model.ACTIVITY_TASK_STATUS_UNCOMPLETED:
			responseData.Status = 1
		}

		responseData.TaskInfo = new(proto.ProtoActivityTaskInfo)
		err := composeProtoActivityTask(responseData.TaskInfo, taskList[0], nil)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}
		this.Response.ResData.Params = responseData
	} else {
		var responseData proto.ProtoGetActivityTaskStatusNothingResponse
		responseData.Status = 2
		this.Response.ResData.Params = responseData
	}

	return 0, nil
}

func (this *ActivityTaskHandler) ReceiveReward() (int, error) {
	var reqParams proto.ProtoReceiveActivityTaskRewardRequest
	err := this.UnpackParams(&reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	// 检测任务是否存在
	if _, ok := config.GActivityTaskConfig.AttrMap[reqParams.ProtypeId]; !ok {
		return errorcode.ERROR_CODE_ACTIVITY_TASK_ID_NOT_EXIST, custom_errors.New("activity task[%d] not exist", reqParams.ProtypeId)
	}

	activityTaskFreshModel := model.ActivityTaskFreshModel{Uin: this.Request.Uin}
	freshInfo, err := activityTaskFreshModel.GetActivityTaskFresh()
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	// calmwu
	timeZone := utils.QueryPlayerTimeZone(this.Request.Uin)
	base.GLocalizedTime.SetLocale(timeZone)
	defer base.GLocalizedTime.SetLocale("Local")

	// 任务刷新时间检测
	if freshInfo == nil || !base.GLocalizedTime.IsToday(int64(freshInfo.FreshTime)) {
		return errorcode.ERROR_CODE_ACTIVITY_TASK_UNFINISHED, custom_errors.New("activity task[%d] is unfinished", reqParams.ProtypeId)
	}

	activityTaskModel := model.ActivityTaskModel{Uin: this.Request.Uin}
	taskInfo, err := activityTaskModel.GetActivityTaskByProtypeId(reqParams.ProtypeId)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if taskInfo == nil {
		return errorcode.ERROR_CODE_ACTIVITY_TASK_UNFINISHED, custom_errors.New("activity task[%d] is unfinished", reqParams.ProtypeId)
	}

	// 任务状态检测
	if taskInfo.Status == model.ACTIVITY_TASK_STATUS_COMPLETED {
		return errorcode.ERROR_CODE_ACTIVITY_TASK_FINISHED, custom_errors.New("activity task[%d] is already finished.", reqParams.ProtypeId)
	}

	protype := config.GActivityTaskConfig.AttrMap[reqParams.ProtypeId]
	if taskInfo.Status == model.ACTIVITY_TASK_STATUS_UNCOMPLETED {
		var progress MissionProgress
		err = json.Unmarshal([]byte(taskInfo.ProgressData), &progress)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		if progress.Current < protype.Parameters.Int(config.MISSION_PARAMETER_TARGET, 0) {
			return errorcode.ERROR_CODE_ACTIVITY_TASK_UNFINISHED, custom_errors.New("activity task[%d] is unfinished", reqParams.ProtypeId)
		}
	}

	// 更新任务
	taskInfo.Status = model.ACTIVITY_TASK_STATUS_COMPLETED
	retCode, err := activityTaskModel.UpdateActivityTask(taskInfo)
	if err != nil {
		return retCode, err
	}

	vitalityRes := protype.Reward.GetResourceItem(config.RESOURCE_ITEM_TYPE_VITALITY)
	oldScore := freshInfo.Score
	if vitalityRes != nil {
		freshInfo.Score += base.ConvertToInt(vitalityRes.Count, 0)
		retCode, err = activityTaskFreshModel.UpdateActivityTaskFresh(freshInfo)
		if err != nil {
			return retCode, err
		}
	}

	var responseData proto.ProtoReceiveActivityTaskRewardResponse
	responseData.Vitality = freshInfo.Score
	protoTask := new(proto.ProtoActivityTaskInfo)
	composeProtoActivityTask(protoTask, taskInfo, protype)
	responseData.TaskList = append(responseData.TaskList, protoTask)
	ResourcesConfigToProtoWithOmit(&protype.Reward, &responseData.Rewards, config.RESOURCE_ITEM_TYPE_VITALITY)

	newActivityVitalityRewardProtype := config.GActivityScoreRewardConfig.GetScoreReward(oldScore, freshInfo.Score)
	if len(newActivityVitalityRewardProtype) > 0 {

		activityRewardModel := model.ActivityScoreRewardModel{Uin: this.Request.Uin}
		scoreRewardList, err := activityRewardModel.GetActivityScoreRewardList()
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		scoreRewardMap := make(map[string]*table.TblActivityScoreReward)
		for _, scoreRewardInfo := range scoreRewardList {
			scoreRewardMap[scoreRewardInfo.RewardId] = scoreRewardInfo
		}

		responseData.NewVitalityReward = 0
		for _, vitalityRewardProtype := range newActivityVitalityRewardProtype {
			// 这里激活了新的奖励列表，需要更新数据库
			scoreRewardInfo := scoreRewardMap[vitalityRewardProtype.Id]
			if scoreRewardInfo != nil && scoreRewardInfo.Status == model.ACTIVITY_SCORE_REWARD_STATUS_UNACTIVE {
				scoreRewardInfo.Status = model.ACTIVITY_SCORE_REWARD_STATUS_UNRECEIVE
				// 直接更新数据库
				_, err := activityRewardModel.UpdateActivityScoreReward(scoreRewardInfo)
				if err != nil {
					return errorcode.ERROR_CODE_DEFAULT, err
				}

				protoRewardInfo := new(proto.ProtoActivityScoreRewardInfo)
				err = composeProtoActivityScoreReward(protoRewardInfo, scoreRewardInfo, vitalityRewardProtype)
				if err != nil {
					return errorcode.ERROR_CODE_DEFAULT, err
				}
				responseData.NewVitalityRewardList = append(responseData.NewVitalityRewardList, protoRewardInfo)
			}
		}

		if len(responseData.NewVitalityRewardList) > 0 {
			responseData.NewVitalityReward = 1
		}
	} else {
		responseData.NewVitalityReward = 0
	}

	this.Response.ResData.Params = responseData

	return 0, nil
}

func (this *ActivityTaskHandler) ReceiveVitalityReward() (int, error) {
	var reqParams proto.ProtoReceiveActivityScoreRewardRequest
	err := this.UnpackParams(&reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	// 检测Score奖励是否存在
	protype, ok := config.GActivityScoreRewardConfig.AttrMap[reqParams.RewardId]
	if !ok {
		return errorcode.ERROR_CODE_ACTIVITY_SCORE_REWARD_ID_NOT_EXIST, custom_errors.New("reward id[%s] not exist", reqParams.RewardId)
	}

	// 检测分数是否达到
	activityTaskFreshModel := model.ActivityTaskFreshModel{Uin: this.Request.Uin}
	freshInfo, err := activityTaskFreshModel.GetActivityTaskFresh()
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	// calmwu
	timeZone := utils.QueryPlayerTimeZone(this.Request.Uin)
	base.GLocalizedTime.SetLocale(timeZone)
	defer base.GLocalizedTime.SetLocale("Local")

	if freshInfo == nil || !base.GLocalizedTime.IsToday(int64(freshInfo.FreshTime)) {
		return errorcode.ERROR_CODE_ACTIVITY_SCORE_REWARD_NOT_REACH, custom_errors.New("reward[%s] score is not reach ", reqParams.RewardId)
	}

	if freshInfo.Score < protype.Score {
		return errorcode.ERROR_CODE_ACTIVITY_SCORE_REWARD_NOT_REACH, custom_errors.New("reward[%s] score is not reach ", reqParams.RewardId)
	}

	activityScoreRewardModel := model.ActivityScoreRewardModel{Uin: this.Request.Uin}
	scoreRewardInfo, err := activityScoreRewardModel.GetActivityScoreRewardByRewardId(reqParams.RewardId)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if scoreRewardInfo != nil {
		if scoreRewardInfo.Status == model.ACTIVITY_SCORE_REWARD_STATUS_RECEIVED {
			return errorcode.ERROR_CODE_ACTIVITY_SCORE_REWARD_RECEIVED_REPEAT, custom_errors.New("reward[%s] is already received", reqParams.RewardId)
		} else {
			scoreRewardInfo.Status = model.ACTIVITY_SCORE_REWARD_STATUS_RECEIVED
			retCode, err := activityScoreRewardModel.UpdateActivityScoreReward(scoreRewardInfo)
			if err != nil {
				return retCode, err
			}
		}
	} else {
		scoreRewardInfo := new(table.TblActivityScoreReward)
		scoreRewardInfo.RewardId = reqParams.RewardId
		scoreRewardInfo.Status = model.ACTIVITY_SCORE_REWARD_STATUS_RECEIVED
		scoreRewardInfo.Uin = this.Request.Uin
		retCode, err := activityScoreRewardModel.AddActivityScoreReward(scoreRewardInfo)
		if err != nil {
			return retCode, err
		}
	}

	var responseData proto.ProtoReceiveActivityScoreRewardResponse
	protoTaskReward := new(proto.ProtoActivityScoreRewardInfo)
	composeProtoActivityScoreReward(protoTaskReward, scoreRewardInfo, protype)
	responseData.RewardList = append(responseData.RewardList, protoTaskReward)
	ResourcesConfigToProto(&protype.Reward, &responseData.Rewards)

	this.Response.ResData.Params = responseData

	return 0, nil
}

func HandleActivityTaskProgress(req *base.ProtoRequestS, resParams *map[string]interface{}, eventDataSet *EventsHappenedDataSet) (int, error) {
	if req == nil || resParams == nil || eventDataSet == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.NullPoint()
	}

	if req.Uin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.InvalidUin()
	}

	(*resParams)["NewActivityTasksWaitingForReceive"] = 0

	activityTaskFreshModel := model.ActivityTaskFreshModel{Uin: req.Uin}
	freshInfo, err := activityTaskFreshModel.GetActivityTaskFresh()
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	timeZone := utils.QueryPlayerTimeZone(req.Uin)
	base.GLocalizedTime.SetLocale(timeZone)
	time.LoadLocation(timeZone)
	defer base.GLocalizedTime.SetLocale("Local")
	defer time.LoadLocation("Local")

	clockStart := time.Now()
	// 刷新日常任务相关数据
	var taskList []*table.TblActivityTask
	if freshInfo == nil {
		freshInfo = new(table.TblActivityTaskFresh)
		freshInfo.Uin = req.Uin
		freshInfo.Score = 0
		freshInfo.FreshTime = int(base.GLocalizedTime.SecTimeStamp())
		retCode, err := activityTaskFreshModel.AddActivityTaskFresh(freshInfo)
		if err != nil {
			return retCode, err
		}

		taskList, err = resetActivityTask(req.Uin)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}
		_, err = resetActivityScoreReward(req.Uin)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}
	} else if !base.GLocalizedTime.IsToday(int64(freshInfo.FreshTime)) {
		freshInfo.Score = 0
		freshInfo.FreshTime = int(base.GLocalizedTime.SecTimeStamp())
		retCode, err := activityTaskFreshModel.UpdateActivityTaskFresh(freshInfo)
		if err != nil {
			return retCode, err
		}

		taskList, err = resetActivityTask(req.Uin)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}
		_, err = resetActivityScoreReward(req.Uin)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}
	} else {
		activityTaskModel := model.ActivityTaskModel{Uin: req.Uin}
		taskList, err = activityTaskModel.GetActivityTaskList()
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}
	}

	base.GLog.Debug("Init Cost: %dns", time.Since(clockStart).Nanoseconds())

	clockStart = time.Now()
	taskMap := make(map[int]*table.TblActivityTask)
	for _, tblTask := range taskList {
		taskMap[tblTask.ProtypeId] = tblTask
	}

	eventProgress := make(map[int]*MissionProgress)

	var oldProgress MissionProgress

	if eventDataSet.BattleEndData != nil {
		switch eventDataSet.BattleEndData.BattleType {
		case BATTLE_TYPE_PVP:
			if list, ok := config.GActivityTaskConfig.TypeMap[config.MISSION_TYPE_LEAGUE_PVP_TIMES]; ok {
				for _, protype := range list {
					if task, ok := taskMap[protype.Id]; ok {
						if task.Status != model.ACTIVITY_TASK_STATUS_UNCOMPLETED {
							continue
						}

						err = json.Unmarshal([]byte(task.ProgressData), &oldProgress)
						if err != nil {
							continue
						}
						progress := new(MissionProgress)
						*progress = oldProgress
						progress.Current += 1
						eventProgress[protype.Id] = progress
					} else {
						progress := new(MissionProgress)
						progress.Current = 1
						eventProgress[protype.Id] = progress
					}
				}
			}

			if eventDataSet.BattleEndData.SinkShipCount > 0 {
				if list, ok := config.GActivityTaskConfig.TypeMap[config.MISSION_TYPE_LEAGUE_PVP_SINK_BATTLE_SHIPS]; ok {
					for _, protype := range list {
						if task, ok := taskMap[protype.Id]; ok {
							if task.Status != model.ACTIVITY_TASK_STATUS_UNCOMPLETED {
								continue
							}

							err = json.Unmarshal([]byte(task.ProgressData), &oldProgress)
							if err != nil {
								continue
							}

							progress := new(MissionProgress)
							*progress = oldProgress
							progress.Current += eventDataSet.BattleEndData.SinkShipCount
							eventProgress[protype.Id] = progress
						} else {
							progress := new(MissionProgress)
							progress.Current = eventDataSet.BattleEndData.SinkShipCount
							eventProgress[protype.Id] = progress
						}

					}
				}
			}

			if list, ok := config.GActivityTaskConfig.TypeMap[config.MISSION_TYPE_LEAGUE_PVP_COMBO_TIMES]; ok {
				for _, protype := range list {
					if task, ok := taskMap[protype.Id]; ok {
						if task.Status != model.ACTIVITY_TASK_STATUS_UNCOMPLETED {
							continue
						}
					}

					if eventDataSet.BattleEndData.MaxCombos >= protype.Parameters.Int(config.MISSION_PARAMETER_COMBO, 0) {
						progress := new(MissionProgress)
						progress.Current = 1
						eventProgress[protype.Id] = progress
					}
				}
			}

			switch eventDataSet.BattleEndData.BattleResult {
			case BATTLE_RESULT_SUCCESS:
				if list, ok := config.GActivityTaskConfig.TypeMap[config.MISSION_TYPE_LEAGUE_PVP_WIN_TIMES]; ok {
					for _, protype := range list {
						if task, ok := taskMap[protype.Id]; ok {
							if task.Status != model.ACTIVITY_TASK_STATUS_UNCOMPLETED {
								continue
							}

							err = json.Unmarshal([]byte(task.ProgressData), &oldProgress)
							if err != nil {
								continue
							}

							progress := new(MissionProgress)
							*progress = oldProgress
							progress.Current += 1
							eventProgress[protype.Id] = progress
						} else {
							progress := new(MissionProgress)
							progress.Current = 1
							eventProgress[protype.Id] = progress
						}
					}
				}
			}
		case BATTLE_TYPE_PVE:
			if list, ok := config.GActivityTaskConfig.TypeMap[config.MISSION_TYPE_CAMPAIGN_BATTLE_TIMES]; ok {
				for _, protype := range list {
					if task, ok := taskMap[protype.Id]; ok {
						if task.Status != model.ACTIVITY_TASK_STATUS_UNCOMPLETED {
							continue
						}

						err = json.Unmarshal([]byte(task.ProgressData), &oldProgress)
						if err != nil {
							continue
						}

						progress := new(MissionProgress)
						*progress = oldProgress
						progress.Current += 1
						eventProgress[protype.Id] = progress
					} else {
						progress := new(MissionProgress)
						progress.Current = 1
						eventProgress[protype.Id] = progress
					}
				}
			}
		}
	}

	if eventDataSet.CampaignProduceResourcesReceivedEventData != nil {
		if list, ok := config.GActivityTaskConfig.TypeMap[config.MISSION_TYPE_CAMPAIGN_CLAIM_RESOURCES_TIMES]; ok {
			for _, protype := range list {
				if task, ok := taskMap[protype.Id]; ok {
					if task.Status != model.ACTIVITY_TASK_STATUS_UNCOMPLETED {
						continue
					}

					err = json.Unmarshal([]byte(task.ProgressData), &oldProgress)
					if err != nil {
						continue
					}

					progress := new(MissionProgress)
					*progress = oldProgress
					progress.Current += 1
					eventProgress[protype.Id] = progress
				} else {
					progress := new(MissionProgress)
					progress.Current = 1
					eventProgress[protype.Id] = progress
				}

			}
		}
	}

	if eventDataSet.ShopPurchasedEventData != nil {
		if eventDataSet.ShopPurchasedEventData.Goods != nil && eventDataSet.ShopPurchasedEventData.Goods.Props != nil {
			for _, propItem := range eventDataSet.ShopPurchasedEventData.Goods.Props {
				if propProtype, ok := config.GPropConfig.AttrMap[propItem.ProtypeId]; ok {
					if propProtype.PropType == config.PROP_TYPE_CARDPACK {
						if list, ok := config.GActivityTaskConfig.TypeMap[config.MISSION_TYPE_SHOP_BUY_CARDPACKS]; ok {
							for _, protype := range list {
								if protype.Parameters.Int(config.MISSION_PARAMETER_CARDPACK_ID, 0) != propItem.ProtypeId {
									continue
								}

								if task, ok := taskMap[protype.Id]; ok {
									if task.Status != model.ACTIVITY_TASK_STATUS_UNCOMPLETED {
										continue
									}

									err = json.Unmarshal([]byte(task.ProgressData), &oldProgress)
									if err != nil {
										continue
									}

									progress := new(MissionProgress)
									*progress = oldProgress
									progress.Current += propItem.Count
									eventProgress[protype.Id] = progress
								} else {
									progress := new(MissionProgress)
									progress.Current = propItem.Count
									eventProgress[protype.Id] = progress
								}
							}
						}
					}
				}
			}
		}
	}

	if eventDataSet.ShareToSocialEventData != nil {
		if list, ok := config.GActivityTaskConfig.TypeMap[config.MISSION_TYPE_SOCIAL_SHARE_TIMES]; ok {
			for _, protype := range list {
				if task, ok := taskMap[protype.Id]; ok {
					if task.Status != model.ACTIVITY_TASK_STATUS_UNCOMPLETED {
						continue
					}

					err = json.Unmarshal([]byte(task.ProgressData), &oldProgress)
					if err != nil {
						continue
					}
					progress := new(MissionProgress)
					*progress = oldProgress
					progress.Current += 1
					eventProgress[protype.Id] = progress
				} else {
					progress := new(MissionProgress)
					progress.Current = 1
					eventProgress[protype.Id] = progress
				}
			}
		}
	}

	if eventDataSet.ResourcesCostEventData != nil {
		for _, resItem := range eventDataSet.ResourcesCostEventData.Resources.Resources {
			if resItem.Type == config.RESOURCE_ITEM_TYPE_GEM {
				if list, ok := config.GActivityTaskConfig.TypeMap[config.MISSION_TYPE_CONSUME_GEM_COUNT]; ok {
					for _, protype := range list {
						if task, ok := taskMap[protype.Id]; ok {
							if task.Status != model.ACTIVITY_TASK_STATUS_UNCOMPLETED {
								continue
							}

							err = json.Unmarshal([]byte(task.ProgressData), &oldProgress)
							if err != nil {
								continue
							}

							progress := new(MissionProgress)
							*progress = oldProgress
							progress.Current += resItem.Count
							eventProgress[protype.Id] = progress

							base.GLog.Debug("ResCount:[%+v] progress:[%+v]", resItem.Count, progress)
						} else {
							progress := new(MissionProgress)
							progress.Current = resItem.Count
							eventProgress[protype.Id] = progress
						}
					}
				}
			}
		}
	}

	// 登录事件
	if list, ok := config.GActivityTaskConfig.TypeMap[config.MISSION_TYPE_DAILY_LOGIN]; ok {
		for _, protype := range list {
			if task, ok := taskMap[protype.Id]; ok {
				if task.Status != model.ACTIVITY_TASK_STATUS_UNCOMPLETED {
					continue
				}

				err = json.Unmarshal([]byte(task.ProgressData), &oldProgress)
				if err != nil {
					continue
				}

				progress := new(MissionProgress)
				*progress = oldProgress
				progress.Current = 1
				eventProgress[protype.Id] = progress
			} else {
				progress := new(MissionProgress)
				progress.Current = 1
				eventProgress[protype.Id] = progress
			}
		}
	}

	// 这里的newTaskWaitingForReceive全量返回，方便处理小红点
	newTaskWaitingForReceive := make([]*proto.ProtoActivityTaskInfo, 0)

	for protypeId, task := range taskMap {
		// 已经是完成的的任务
		if task.Status == model.ACTIVITY_TASK_STATUS_WAIT_FOR_RECEIVE {
			if protype, ok := config.GActivityTaskConfig.AttrMap[protypeId]; ok {
				protoTask := new(proto.ProtoActivityTaskInfo)
				composeProtoActivityTask(protoTask, task, protype)
				newTaskWaitingForReceive = append(newTaskWaitingForReceive, protoTask)
			}
		}
	}

	updateActivityTasks := make([]*table.TblActivityTask, 0)
	addActivityTasks := make([]*table.TblActivityTask, 0)
	// 更新任务进度和添加新的任务到数据库
	for proId, prog := range eventProgress {
		if protype, ok := config.GActivityTaskConfig.AttrMap[proId]; ok {
			task, isUpdate := taskMap[proId]
			if !isUpdate {
				task = new(table.TblActivityTask)
				task.ProtypeId = proId
				task.Uin = req.Uin
				task.Status = model.ACTIVITY_TASK_STATUS_UNCOMPLETED
			}

			// 进度
			prog.Total = protype.Parameters.Int(config.CAMPAIGN_MISSION_PARAMETER_TARGET, 0)
			data, err := json.Marshal(prog)
			if err != nil {
				base.GLog.Error("json marshal activity task progress[%+v] err[%s]", prog, err)
				continue
			}
			task.ProgressData = string(data)

			if prog.Current >= prog.Total {
				prog.Current = prog.Total
				task.Status = model.ACTIVITY_TASK_STATUS_WAIT_FOR_RECEIVE
			}

			if isUpdate {
				updateActivityTasks = append(updateActivityTasks, task)
			} else {
				addActivityTasks = append(addActivityTasks, task)
			}

			if task.Status == model.ACTIVITY_TASK_STATUS_WAIT_FOR_RECEIVE {
				protoTask := new(proto.ProtoActivityTaskInfo)
				composeProtoActivityTask(protoTask, task, protype)
				newTaskWaitingForReceive = append(newTaskWaitingForReceive, protoTask)
			}
		}
	}

	base.GLog.Debug("Comupte Cost: %dns", time.Since(clockStart).Nanoseconds())

	clockStart = time.Now()

	activityTaskModel := model.ActivityTaskModel{Uin: req.Uin}

	if len(updateActivityTasks) > 0 {
		retCode, err := activityTaskModel.UpdateMultiActivityTask(updateActivityTasks)
		if err != nil {
			return retCode, err
		}
	}

	if len(addActivityTasks) > 0 {
		retCode, err := activityTaskModel.AddMultiActivityTask(addActivityTasks)
		if err != nil {
			return retCode, err
		}
	}

	base.GLog.Debug("Update Cost: %dns", time.Since(clockStart).Nanoseconds())

	// 组织返回内容
	if len(newTaskWaitingForReceive) > 0 {
		(*resParams)["NewActivityTasksWaitingForReceive"] = 1
		(*resParams)["ActivityTasksWaitingForReceive"] = newTaskWaitingForReceive
	} else {
		(*resParams)["NewActivityTasksWaitingForReceive"] = 0
	}

	return 0, nil
}

func composeProtoActivityTask(target *proto.ProtoActivityTaskInfo, data *table.TblActivityTask, protype *config.ActivityTaskProtype) error {
	if target == nil || (data == nil && protype == nil) {
		return custom_errors.NullPoint()
	}

	if data != nil {
		target.ProtypeId = data.ProtypeId
		target.Status = data.Status

		var progress MissionProgress
		err := json.Unmarshal([]byte(data.ProgressData), &progress)
		if err != nil {
			target.Current = 0
			target.Total = 0
			base.GLog.Error("json unmarshal progress data[%s] err[%s]", data.ProgressData, err)
		} else {
			target.Current = progress.Current
			target.Total = progress.Total
		}
	} else if protype != nil {
		target.ProtypeId = protype.Id
		target.Current = 0
		target.Total = protype.Parameters.Int(config.MISSION_PARAMETER_TARGET, 1)
		target.Status = model.ACTIVITY_TASK_STATUS_UNCOMPLETED
	}

	return nil
}

func composeProtoActivityScoreReward(target *proto.ProtoActivityScoreRewardInfo, data *table.TblActivityScoreReward, protype *config.ActivityScoreRewardProtype) error {
	if target == nil || (data == nil && protype == nil) {
		return custom_errors.NullPoint()
	}

	if data != nil {
		target.RewardId = data.RewardId
		target.Status = data.Status
	} else if protype != nil {
		target.RewardId = protype.Id
		target.Status = model.ACTIVITY_SCORE_REWARD_STATUS_UNACTIVE
	}

	return nil
}

func getActivityTaskFreshInfo(uin int) (*table.TblActivityTaskFresh, error) {
	activityTaskFreshModel := model.ActivityTaskFreshModel{Uin: uin}
	freshInfo, err := activityTaskFreshModel.GetActivityTaskFresh()
	if err != nil {
		return nil, err
	}

	return freshInfo, nil
}

func sortTblActivityTask(taskSlice []*table.TblActivityTask) []*table.TblActivityTask {
	sort.Slice(taskSlice, func(i, j int) bool {
		// 1. 状态排序
		if taskSlice[i].Status != taskSlice[j].Status {

			switch taskSlice[i].Status {
			case model.ACTIVITY_TASK_STATUS_WAIT_FOR_RECEIVE:
				return true
			case model.ACTIVITY_TASK_STATUS_UNCOMPLETED:
				return taskSlice[j].Status == model.ACTIVITY_TASK_STATUS_COMPLETED
			case model.ACTIVITY_TASK_STATUS_COMPLETED:
				return false
			}
		}

		// 2. Id排序
		return taskSlice[i].ProtypeId < taskSlice[j].ProtypeId
	})

	return taskSlice
}

// 重置每日任务
func resetActivityTask(uin int) ([]*table.TblActivityTask, error) {
	base.GLog.Debug("Reset Activity Task")

	if uin <= 0 {
		return nil, custom_errors.InvalidUin()
	}

	activityTaskModel := model.ActivityTaskModel{Uin: uin}
	taskList, err := activityTaskModel.GetActivityTaskList()
	if err != nil {
		return nil, err
	}

	taskMap := make(map[int]*table.TblActivityTask)
	for _, tblTask := range taskList {
		taskMap[tblTask.ProtypeId] = tblTask
	}

	refreshedList := make([]*table.TblActivityTask, 0)
	var progress MissionProgress
	for _, protype := range config.GActivityTaskConfig.AttrMap {
		progress.Total = protype.Parameters.Int(config.MISSION_PARAMETER_TARGET, 0)
		data, err := json.Marshal(progress)
		if err != nil {
			base.GLog.Error("json marshal progress[%+v] error[%s]", progress, err)
			continue
		}
		progressData := string(data)

		tblTask, ok := taskMap[protype.Id]
		if ok {
			tblTask.Status = model.ACTIVITY_TASK_STATUS_UNCOMPLETED
			tblTask.ProgressData = progressData
			_, err := activityTaskModel.UpdateActivityTask(tblTask)
			if err != nil {
				return nil, err
			}
		} else {
			tblTask = new(table.TblActivityTask)
			tblTask.Uin = uin
			tblTask.ProtypeId = protype.Id
			tblTask.ProgressData = progressData
			tblTask.Status = model.ACTIVITY_TASK_STATUS_UNCOMPLETED
			_, err := activityTaskModel.AddActivityTask(tblTask)
			if err != nil {
				return nil, err
			}
		}

		refreshedList = append(refreshedList, tblTask)
	}

	return refreshedList, nil
}

// 重置每日活跃度奖励
func resetActivityScoreReward(uin int) ([]*table.TblActivityScoreReward, error) {
	base.GLog.Debug("Reset Activity Score Reward")

	if uin <= 0 {
		return nil, custom_errors.InvalidUin()
	}

	activityScoreRewardModel := model.ActivityScoreRewardModel{Uin: uin}
	rewardList, err := activityScoreRewardModel.GetActivityScoreRewardList()
	if err != nil {
		return nil, err
	}

	rewardMap := make(map[string]*table.TblActivityScoreReward)
	for _, tblReward := range rewardList {
		rewardMap[tblReward.RewardId] = tblReward
	}

	refreshedList := make([]*table.TblActivityScoreReward, 0)
	for _, protype := range config.GActivityScoreRewardConfig.AttrMap {
		var tblReward *table.TblActivityScoreReward
		if tblReward, ok := rewardMap[protype.Id]; ok {
			tblReward.Status = model.ACTIVITY_SCORE_REWARD_STATUS_UNACTIVE
			_, err := activityScoreRewardModel.UpdateActivityScoreReward(tblReward)
			if err != nil {
				return nil, err
			}
		} else {
			tblReward = new(table.TblActivityScoreReward)
			tblReward.Uin = uin
			tblReward.RewardId = protype.Id
			tblReward.Status = model.ACTIVITY_SCORE_REWARD_STATUS_UNACTIVE

			_, err := activityScoreRewardModel.AddActivityScoreReward(tblReward)
			if err != nil {
				return nil, err
			}
		}

		refreshedList = append(refreshedList, tblReward)
	}

	return refreshedList, nil
}
