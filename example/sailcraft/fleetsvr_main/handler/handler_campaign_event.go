package handler

import (
	"encoding/json"
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

type CampaignEventHandler struct {
	handlerbase.WebHandler
}

func (this *CampaignEventHandler) UnfinishedList() (int, error) {
	// 获取数据库状态为未完成事件
	campaignEventModel := model.CampaignEventModel{Uin: this.Request.Uin}
	records, err := campaignEventModel.QueryCampaignEventUnfinished()
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}
	base.GLog.Debug("len of records : %d", len(records))

	// 区分已关闭事件和未完成事件
	unfinishedRecords := make([]*table.TblCampaignEvent, 0, len(records))
	invalidOrClosedRecords := make([]*table.TblCampaignEvent, 0, len(records))

	for _, record := range records {
		if eventProtype, ok := config.GCampaignEventConfig.AttrMap[record.EventId]; ok {
			timePassed := int(base.GLocalizedTime.SecTimeStamp()) - record.StartTime
			if timePassed > eventProtype.Term {
				base.GLog.Debug("event expired[%d, %d]", record.Id, record.EventId)
				record.EventStatus = model.CAMPAIGN_EVENT_STATUS_CLOSED
				invalidOrClosedRecords = append(invalidOrClosedRecords, record)
			} else {
				if _, ok := config.GCampaignConfig.AttrMap[record.CampaignId]; !ok {
					base.GLog.Debug("campaign id not exist[%d, %d]", record.Id, record.EventId)
					invalidOrClosedRecords = append(invalidOrClosedRecords, record)
				} else {
					unfinishedRecords = append(unfinishedRecords, record)
				}
			}
		} else {
			base.GLog.Debug("event protype id not exist[%d, %d]", record.Id, record.EventId)
			record.EventStatus = model.CAMPAIGN_EVENT_STATUS_CLOSED
			invalidOrClosedRecords = append(invalidOrClosedRecords, record)
		}
	}

	// 组织返回内容
	var responseData proto.ProtoQueryCampaignEventUnifishedResponse
	responseData.CampaignEventsUnfinished = make([]*proto.ProtoCampaignEventInfo, len(unfinishedRecords))

	for index, tblEvent := range unfinishedRecords {
		responseData.CampaignEventsUnfinished[index] = new(proto.ProtoCampaignEventInfo)
		composeCampaignEventTblToProto(tblEvent, responseData.CampaignEventsUnfinished[index])
	}

	this.Response.ResData.Params = responseData

	return 0, nil
}

func (this *CampaignEventHandler) Complete() (int, error) {
	var reqParams proto.ProtoFinishCampaignEventRequest
	err := this.UnpackParams(&reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	campaignEventModel := model.CampaignEventModel{Uin: this.Request.Uin}

	eventInfo := new(table.TblCampaignEvent)
	retCode, err := campaignEventModel.QueryCampaignEventById(reqParams.Id, eventInfo)
	if err != nil {
		return retCode, err
	}

	if retCode == 0 {
		return errorcode.ERROR_CODE_CAMPAIGN_EVENT_NOT_EXSIT, custom_errors.New("id not exist")
	}

	retCode, err = ValidCampaignEventCanReceive(eventInfo)
	if err != nil {
		return retCode, err
	}

	var costs config.ResourcesAttr

	// CampaignEvent.json根据eventid找到配置
	eventProtype := config.GCampaignEventConfig.AttrMap[eventInfo.EventId]
	switch eventProtype.EventType {
	case config.CAMPAIGN_EVENT_TYPE_MISSION:
		// 如果任务事件类型是finish_mission，通过任务id在CampaignMission.json找到任务对象
		if missionProtype, ok := config.GCampaignMissionConfig.AttrMap[eventProtype.MissionId]; ok {

			var curProgress table.TblCampaignEventProgress
			err = json.Unmarshal([]byte(eventInfo.ProgressData), &curProgress)
			if err != nil {
				return errorcode.ERROR_CODE_DEFAULT, err
			}
			// 判断进度是否完成，是否可以领取
			if curProgress.Progress < missionProtype.Parameter[config.CAMPAIGN_MISSION_PARAMETER_TARGET] {
				return errorcode.ERROR_CODE_CAMPAIGN_EVENT_UNFINISHED, custom_errors.New("campaign event unfinished")
			}
			base.GLog.Debug("Uin[%d] eventInfo:%+v", this.Request.Uin, eventInfo)
		} else {
			return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("wrong mission id")
		}
	case config.CAMPAIGN_EVENT_TYPE_EXCHANGE:
		// 判断资源是否足够
		finalCost, retCode, err := CalculateUinUserRealResourcesCost(this.Request.Uin, &eventProtype.Requirement.ResourceItems)
		if err != nil {
			return retCode, err
		}

		costs.ResourceItems = finalCost
	}

	// 刷新完成状态到数据库
	eventInfo.EventStatus = model.CAMPAIGN_EVENT_STATUS_RECEIVED
	retCode, err = campaignEventModel.UpdateCampaignEvent(eventInfo)
	if err != nil {
		return retCode, err
	}

	// var finalResourceChanged config.ResourcesAttr
	// finalResourceChanged.EscapeNil()
	// finalResourceChanged.Add(&eventProtype.Reward)
	// if eventProtype.EventType == config.CAMPAIGN_EVENT_TYPE_EXCHANGE {
	// 	finalResourceChanged.Sub(&costs)
	// }

	var responseData proto.ProtoFinishCampaignEventResponse
	ResourcesConfigToProto(&costs, &responseData.Cost)
	ResourcesConfigToProto(&eventProtype.Reward, &responseData.Reward)

	this.Response.ResData.Params = responseData

	// 统计数据
	retCode, err = UpdateCampaignEventStatistics(eventInfo)
	if err != nil {
		return retCode, err
	}

	return 0, nil
}

func OnCampaignEventTriggerForTimer(req *base.ProtoRequestS, res *map[string]interface{}) (int, error) {
	base.GLog.Debug("Enter: OnCampaignEventTriggerForTimer")
	if req == nil || res == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.NullPoint()
	}

	(*res)["NewCampaignEvent"] = 0

	if req.Uin <= 0 {
		base.GLog.Error("OnCampaignEventTriggerForTimer req uin is wrong [uin:%d]", req.Uin)
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("request param uin is wrong")
	}

	campaignEventFreshModel := model.CampaignEventFreshModel{Uin: req.Uin}

	var eventFreshInfo table.TblCampaignEventFresh
	retCode, err := campaignEventFreshModel.GetCampaignEventFreshInfo(&eventFreshInfo)

	if err != nil {
		return retCode, err
	}

	// 如果数据库中没有事件刷新的时间信息，将当前时间作为事件刷新时间
	if retCode == 0 {
		campaignEventFreshModel.RefreshCampaignEventFreshTime(&eventFreshInfo)
		return 0, nil
	}

	// 时间没到
	timePassed := int(base.GLocalizedTime.SecTimeStamp()) - eventFreshInfo.LastFreshTime
	if timePassed < config.GGlobalConfig.Campaign.TimeTriggerInterval {
		return 0, nil
	}

	newEvent, retCode, err := NewCampaignEvent(req.Uin, model.CAMPAIGN_EVENT_TRIGGER_TYPE_TIME)
	if err != nil {
		return retCode, err
	}

	if newEvent == nil {
		return 0, nil
	}

	// 触发成功，添加返回内容
	protoCampaignEventInfo := new(proto.ProtoCampaignEventInfo)
	retCode, err = composeCampaignEventTblToProto(newEvent, protoCampaignEventInfo)
	if err != nil {
		return retCode, err
	}

	resData := make([]*proto.ProtoCampaignEventInfo, 1)
	resData[0] = protoCampaignEventInfo
	(*res)["CampaignEventTriggered"] = resData
	(*res)["NewCampaignEvent"] = 1

	return 0, nil
}

func OnCampaignEventTriggerForPvpWin(req *base.ProtoRequestS, res *map[string]interface{}, eventData *EventsHappenedDataSet) (int, error) {
	base.GLog.Debug("Enter: OnCampaignEventTriggerForPvpWin")
	if req == nil || res == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.NullPoint()
	}

	(*res)["NewCampaignEvent"] = 0

	if req.Uin <= 0 {
		base.GLog.Error("OnCampaignEventTriggerForTimer req uin is wrong [uin:%d]", req.Uin)
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("request param uin is wrong")
	}

	if eventData.BattleEndData == nil {
		return 0, nil
	}

	const (
		RAND_BASE = 1000
	)

	// 概率随机
	r := rand.New(rand.NewSource(base.GLocalizedTime.SecTimeStamp()))
	rRet := r.Intn(RAND_BASE)
	if rRet > int(RAND_BASE*config.GGlobalConfig.Campaign.PVPWinTriggerProbability) {
		base.GLog.Debug("rand failed, have no event triggered.")
		return 0, nil
	}

	newEvent, retCode, err := NewCampaignEvent(req.Uin, model.CAMPAIGN_EVENT_TRIGGER_TYPE_PVP_WIN)
	if err != nil {
		return retCode, err
	}

	if newEvent == nil {
		return 0, nil
	}

	// 触发成功，添加返回内容
	protoCampaignEventInfo := new(proto.ProtoCampaignEventInfo)
	retCode, err = composeCampaignEventTblToProto(newEvent, protoCampaignEventInfo)
	if err != nil {
		return retCode, err
	}

	resData := make([]*proto.ProtoCampaignEventInfo, 1)
	resData[0] = protoCampaignEventInfo
	(*res)["CampaignEventTriggered"] = resData
	(*res)["NewCampaignEvent"] = 1

	return 0, nil
}

func HandleCampaignEventProgress(req *base.ProtoRequestS, res *map[string]interface{}, eventData *EventsHappenedDataSet) (int, error) {
	if req == nil || res == nil || eventData == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.NullPoint()
	}
	base.GLog.Debug("Enter: HandleCampaignEventProgress Uin[%d]", req.Uin)

	if req.Uin <= 0 {
		base.GLog.Error("OnCampaignEventTriggerForTimer req uin is wrong [uin:%d]", req.Uin)
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("request param uin is wrong")
	}

	data := make(map[string]int)

	if eventData.BattleEndData != nil {
		base.GLog.Debug("Uin[%d] BattleEndData:%+v", req.Uin, eventData.BattleEndData)

		if eventData.BattleEndData.BattleType == BATTLE_TYPE_PVP {
			// PVP战斗次数 +1
			data[config.CAMPAIGN_MISSION_TYPE_PVP_TIMES] = 1
			if eventData.BattleEndData.BattleResult == BATTLE_RESULT_SUCCESS {
				// PVP胜利次数 +1
				data[config.CAMPAIGN_MISSION_TYPE_PVP_WIN_TIMES] = 1
			}
		}

		if eventData.BattleEndData.SinkShipCount > 0 {
			// 沉船数 +count
			data[config.CAMPAIGN_MISSION_TYPE_SINK_SHIP] = eventData.BattleEndData.SinkShipCount
		}
	}

	base.GLog.Debug("CampaignEvent data:%+v", data)

	missionIds := make([]int, 0)
	missionCountMap := make(map[int]int)
	for missionName, count := range data {
		missionProtypes := config.GCampaignMissionConfig.TypeMap[missionName]
		for _, missionProtype := range missionProtypes {
			missionIds = append(missionIds, missionProtype.Id)
			missionCountMap[missionProtype.Id] = count
		}
	}

	base.GLog.Debug("CampaignEvent missionCountMap:%+v", missionCountMap)

	campaignEventModel := model.CampaignEventModel{Uin: req.Uin}

	records, err := campaignEventModel.GetUnifinishedEventsByMissionId(missionIds)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if len(records) > 0 {
		//var progress table.TblCampaignEventProgress
		updateRecords := make([]*table.TblCampaignEvent, 0)
		for _, record := range records {
			progress := new(table.TblCampaignEventProgress)
			err = json.Unmarshal([]byte(record.ProgressData), progress)
			if err != nil {
				base.GLog.Error("unmarshal progress failed! error[%s]", err)
				continue
			}
			// dump查询出来的事件信息
			base.GLog.Debug("CampaignEvent record:%+v", record)

			if count, ok := missionCountMap[record.MissionId]; ok && count > 0 {
				if progress.Progress < progress.Total {
					// 这里改变事件对应任务的进度情况
					progress.Progress += count
					if progress.Progress >= progress.Total {
						progress.Progress = progress.Total
						base.GLog.Debug("MissionId[%d] is finished! Total[%d]", record.MissionId, progress.Total)
					}

					newProgressData, err := json.Marshal(&progress)
					if err != nil {
						base.GLog.Error("marshal progress failed! error[%s]", err)
						continue
					}

					record.ProgressData = string(newProgressData)

					updateRecords = append(updateRecords, record)
				}
			}
		}

		if len(updateRecords) > 0 {
			retCode, err := campaignEventModel.UpdateMultiCampaignEvents(updateRecords)
			if err != nil {
				return retCode, err
			}
		}
	}

	return 0, nil
}

func ValidCampaignEventCanReceive(campaignEvent *table.TblCampaignEvent) (int, error) {
	if campaignEvent.EventStatus == model.CAMPAIGN_EVENT_STATUS_CLOSED {
		return errorcode.ERROR_CODE_CAMPAIGN_EVENT_CLOSED, custom_errors.New("event has closed")
	}

	if campaignEvent.EventStatus == model.CAMPAIGN_EVENT_STATUS_RECEIVED {
		return errorcode.ERROR_CODE_CAMPAIGN_EVENT_FINISHED, custom_errors.New("event has already done")
	}

	if eventProtype, ok := config.GCampaignEventConfig.AttrMap[campaignEvent.EventId]; ok {
		if _, ok := config.GCampaignConfig.AttrMap[campaignEvent.CampaignId]; ok {
			// 检查事件是否过期
			restTime := eventProtype.Term - (int(base.GLocalizedTime.SecTimeStamp()) - campaignEvent.StartTime)
			if restTime <= 0 {
				return errorcode.ERROR_CODE_CAMPAIGN_EVENT_EXPIRED, custom_errors.New("campaign event expired")
			}

			return 0, nil
		}

		// 关卡Id不存在
		return errorcode.ERROR_CODE_CAMPAIGN_ID_NOT_EXIST, custom_errors.New("campaign id not exist")
	} else {
		return errorcode.ERROR_CODE_CAMPAIGN_EVENT_NOT_EXSIT, custom_errors.New("event protype id [%d] not found", campaignEvent.EventId)
	}
}

func composeCampaignEventTblToProto(tblEvent *table.TblCampaignEvent, protoEvent *proto.ProtoCampaignEventInfo) (int, error) {

	if tblEvent == nil || protoEvent == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.NullPoint()
	}

	if eventProtype, ok := config.GCampaignEventConfig.AttrMap[tblEvent.EventId]; ok {
		protoEvent.CampaignId = tblEvent.CampaignId
		protoEvent.EventId = tblEvent.EventId
		protoEvent.Id = tblEvent.Id
		protoEvent.MissionId = tblEvent.MissionId
		protoEvent.StartTime = tblEvent.StartTime
		protoEvent.TriggerType = tblEvent.TriggerType
		protoEvent.Uin = tblEvent.Uin
		// 解包了进度
		json.Unmarshal([]byte(tblEvent.ProgressData), &protoEvent.Progress)
		restTime := eventProtype.Term - (int(base.GLocalizedTime.SecTimeStamp()) - protoEvent.StartTime)

		if restTime > 0 {
			protoEvent.EventStatus = model.CAMPAIGN_EVENT_STATUS_UNFINISHED
			protoEvent.ResetTime = restTime
		} else {
			protoEvent.EventStatus = model.CAMPAIGN_EVENT_STATUS_CLOSED
			protoEvent.ResetTime = 0
		}

		base.GLog.Debug("Uin[%d] CampaignId[%d] EventId[%d] Id[%d] MissionId[%d] EventStatus[%d] ResetTime[%d]",
			protoEvent.Uin, protoEvent.CampaignId, protoEvent.EventId, protoEvent.Id, protoEvent.MissionId,
			protoEvent.EventStatus, protoEvent.ResetTime)
		return 0, nil
	} else {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("event protype id [%d] not found", tblEvent.EventId)
	}
}

func UpdateCampaignEventStatistics(eventInfo *table.TblCampaignEvent) (int, error) {
	if eventInfo.Uin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("campaign event's uin is invalid")
	}

	campaignEventStatisticsModel := model.CampaignEventStatisticsModel{Uin: eventInfo.Uin}
	var statisticData table.TblCampaignEventStatistics
	switch eventInfo.TriggerType {
	case model.CAMPAIGN_EVENT_TRIGGER_TYPE_PVP_WIN:
		retCode, err := campaignEventStatisticsModel.AddPVPWinTriggerCountToday(1, &statisticData)
		if err != nil {
			return retCode, err
		}
	case model.CAMPAIGN_EVENT_TRIGGER_TYPE_TIME:
		retCode, err := campaignEventStatisticsModel.AddTimeTriggerCountToday(1, &statisticData)
		if err != nil {
			return retCode, err
		}
	}

	return 0, nil
}

func NewCampaignEvent(uin int, triggerType int) (*table.TblCampaignEvent, int, error) {
	// 获取所有事件
	campaignEventModel := model.CampaignEventModel{Uin: uin}
	allEvents, err := campaignEventModel.GetAllEvents()
	if err != nil {
		return nil, errorcode.ERROR_CODE_DEFAULT, err
	}

	// 事件分类
	unfinishedEvents := make([]*table.TblCampaignEvent, 0)
	todayEventsForTriggerType := make([]*table.TblCampaignEvent, 0)
	invalidEvents := make([]*table.TblCampaignEvent, 0)

	for _, event := range allEvents {
		year, month, day := base.GLocalizedTime.NowDate()
		startYear, startMonth, startDay := base.GLocalizedTime.UnixDate(int64(event.StartTime), 0)
		if event.EventStatus == model.CAMPAIGN_EVENT_STATUS_UNFINISHED {
			if eventProtype, ok := config.GCampaignEventConfig.AttrMap[event.EventId]; ok {
				eventTimePassed := int(base.GLocalizedTime.SecTimeStamp()) - event.StartTime

				if eventTimePassed < eventProtype.Term {
					if _, ok := config.GCampaignConfig.AttrMap[event.CampaignId]; ok {
						unfinishedEvents = append(unfinishedEvents, event)
					} else {
						invalidEvents = append(invalidEvents, event)
					}
				} else {
					if year < startYear || month < startMonth || day < startDay {
						invalidEvents = append(invalidEvents, event)
						continue
					}
				}

				if year == startYear && month == startMonth && day == startDay && event.TriggerType == triggerType {
					todayEventsForTriggerType = append(todayEventsForTriggerType, event)
				}
			} else {
				// 配置表中没有该事件
				invalidEvents = append(invalidEvents, event)
				continue
			}
		} else {
			if year == startYear && month == startMonth && day == startDay {
				if event.TriggerType == model.CAMPAIGN_EVENT_TRIGGER_TYPE_TIME {
					todayEventsForTriggerType = append(todayEventsForTriggerType, event)
				}
			} else {
				invalidEvents = append(invalidEvents, event)
			}
		}
	}

	// 今天时间触发的事件已达上限
	switch triggerType {
	case model.CAMPAIGN_EVENT_TRIGGER_TYPE_PVP_WIN:
		if len(todayEventsForTriggerType) >= config.GGlobalConfig.Campaign.PVPWinTriggerLimit {
			return nil, 0, nil
		}
	case model.CAMPAIGN_EVENT_TRIGGER_TYPE_TIME:
		if len(todayEventsForTriggerType) >= config.GGlobalConfig.Campaign.TimeTriggerLimit {
			return nil, 0, nil
		}
	}

	// 获取已通关关卡
	var chapterInfo table.TblCampaignPassChapter
	campaignPassChapterModel := model.CampaignPassChapterModel{Uin: uin}
	retCode, err := campaignPassChapterModel.QueryMaxChapterInfoByUin(&chapterInfo)
	if err != nil {
		return nil, retCode, err
	}

	// 获取已通关章节数量
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

	passChapterCount := 0
	for _, area := range config.GCampaignAreaConfig.AttrMap {
		if area.Id <= passChapter {
			passChapterCount += 1
		}
	}

	// 未完成事件限制
	if len(unfinishedEvents) >= config.GGlobalConfig.Campaign.MaxUnfinishedEventLimit[passChapterCount] {
		return nil, 0, nil
	}

	// 从已通关关卡中获取没有未完成事件且能够触发事件的所有关卡
	protypeList := make([]*config.CampaignProtype, 0)
	for attrId, attr := range config.GCampaignConfig.AttrMap {
		if attrId <= chapterInfo.CampaignId && len(attr.EventIds) > 0 {
			hasUnfinished := false
			for _, event := range unfinishedEvents {
				if event.CampaignId == attrId {
					hasUnfinished = true
					break
				}
			}

			if !hasUnfinished {
				protypeList = append(protypeList, attr)
			}
		}
	}

	pLen := len(protypeList)
	if pLen == 0 {
		return nil, 0, nil
	}

	// 随机关卡
	r := rand.New(rand.NewSource(base.GLocalizedTime.SecTimeStamp()))
	targetCampaignProtype := protypeList[r.Intn(pLen)]

	// 随机事件
	pos := r.Intn(targetCampaignProtype.TotalChance)
	var targetEventIndex = -1
	for i, eventChance := range targetCampaignProtype.EventChances {
		if pos < eventChance {
			targetEventIndex = i
			break
		}

		pos -= eventChance
	}

	targetEventProtype, ok := config.GCampaignEventConfig.AttrMap[targetCampaignProtype.EventIds[targetEventIndex]]
	if !ok {
		return nil, errorcode.ERROR_CODE_DEFAULT, custom_errors.New("can not found event id[%d]", targetCampaignProtype.EventIds[targetEventIndex])
	}

	var newEvent table.TblCampaignEvent
	newEvent.Uin = uin
	newEvent.CampaignId = targetCampaignProtype.Id
	newEvent.EventId = targetEventProtype.Id
	newEvent.EventStatus = model.CAMPAIGN_EVENT_STATUS_UNFINISHED
	newEvent.MissionId = targetEventProtype.MissionId
	newEvent.StartTime = int(base.GLocalizedTime.SecTimeStamp())
	newEvent.TriggerType = triggerType

	if targetEventProtype.MissionId > 0 {
		if protype, ok := config.GCampaignMissionConfig.AttrMap[targetEventProtype.MissionId]; ok {
			var progress table.TblCampaignEventProgress
			progress.Progress = 0
			progress.Total = protype.Parameter[config.CAMPAIGN_MISSION_PARAMETER_TARGET]

			base.GLog.Debug("Uin[%d] CampaignId[%d] EventId[%d] EventStatus[%d] MissionId[%d] Progress[0] Total[%d]",
				newEvent.Uin, newEvent.CampaignId, newEvent.EventId, newEvent.EventStatus, newEvent.MissionId, progress.Total)
			progressData, err := json.Marshal(&progress)
			if err != nil {
				return nil, errorcode.ERROR_CODE_DEFAULT, err
			}

			newEvent.ProgressData = string(progressData)
		} else {
			return nil, errorcode.ERROR_CODE_DEFAULT, custom_errors.New("can not found mission id [%d]", targetEventProtype.MissionId)
		}

	}

	// 插入db中
	retCode, err = campaignEventModel.AddCampaignEvent(&newEvent)
	if err != nil {
		return nil, retCode, err
	}
	base.GLog.Debug("Uin[%d] newCampaignEvent:%+v", newEvent)
	campaignEventFreshModel := model.CampaignEventFreshModel{Uin: uin}
	var eventFreshInfo table.TblCampaignEventFresh
	retCode, err = campaignEventFreshModel.RefreshCampaignEventFreshTime(&eventFreshInfo)
	if err != nil {
		return nil, retCode, err
	}

	// 每当触发一个新事件时，删除无效事件（不是今天触发的，并且处于已完成或者已过期状态）
	if _, err = campaignEventModel.DeleteCampaignEvents(invalidEvents); err != nil {
		base.GLog.Error(err)
	}

	return &newEvent, 0, nil
}
