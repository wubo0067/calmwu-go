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

type CampaignPassChapterHandler struct {
	handlerbase.WebHandler
}

func (this *CampaignPassChapterHandler) PassChapter() (int, error) {

	var reqParams proto.ProtoPassCampaignChapterRequest
	err := this.UnpackParams(&reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	base.GLog.Debug("PassCampaignChapter enter params [%v]", reqParams)

	campaignID := reqParams.CampaignID

	// 先检测配置表是否包含此关卡
	if _, ok := config.GCampaignConfig.AttrMap[campaignID]; !ok {
		return errorcode.ERROR_CODE_CAMPAIGN_ID_NOT_EXIST, custom_errors.New("campaign id not exist")
	}

	// 先检查关卡ID是否存在
	model := model.CampaignPassChapterModel{Uin: this.Request.Uin}

	chapterInfo := new(table.TblCampaignPassChapter)
	retCode, err := model.QueryChapterInfoByCampaignId(campaignID, chapterInfo)
	if err != nil {
		return retCode, err
	}

	passCampaignChapterInfo := new(proto.ProtoPassCampaignChapterInfo)
	passCampaignChapterInfo.CampaignID = campaignID
	passCampaignChapterInfo.FirstTimeToPass = 1

	if retCode == 1 {
		// 数据存在
		passCampaignChapterInfo.FirstTimeToPass = 0
	} else {
		// 全新记录，需要插入到数据库
		chapterInfo.Uin = this.Request.Uin
		chapterInfo.CampaignId = campaignID

		records := make([]*table.TblCampaignPassChapter, 0)
		records = append(records, chapterInfo)

		_, err := model.AddCampaignPassChapter(records)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}
	}

	var responseData proto.ProtoPassCampaignChapterResponse
	responseData.CampaignInfo = *passCampaignChapterInfo

	this.Response.ResData.Params = responseData

	return 0, nil
}

func (this *CampaignPassChapterHandler) MaxPassedChapter() (int, error) {
	campaignModel := model.CampaignPassChapterModel{Uin: this.Request.Uin}

	chapterInfo := new(table.TblCampaignPassChapter)
	retCode, err := campaignModel.QueryMaxChapterInfoByUin(chapterInfo)

	if err != nil {
		return retCode, err
	}

	// 默认没有通关任何关卡
	maxCampaignChapterInfo := new(proto.ProtoPassCampaignChapterInfo)
	maxCampaignChapterInfo.CampaignID = 0
	maxCampaignChapterInfo.FirstTimeToPass = 0

	if retCode == 1 {
		maxCampaignChapterInfo.CampaignID = chapterInfo.CampaignId
	}

	var responseData proto.ProtoMaxCampaignChapterResponse
	responseData.CampaignProgressInfo = *maxCampaignChapterInfo

	this.Response.ResData.Params = responseData

	return 0, nil
}

func (this *CampaignPassChapterHandler) Info() (int, error) {

	// 最大通关关卡信息
	campaignModel := model.CampaignPassChapterModel{Uin: this.Request.Uin}

	chapterInfo := new(table.TblCampaignPassChapter)
	retCode, err := campaignModel.QueryMaxChapterInfoByUin(chapterInfo)

	if err != nil {
		return retCode, err
	}

	maxCampaignChapterInfo := new(proto.ProtoPassCampaignChapterInfo)
	maxCampaignChapterInfo.CampaignID = 0
	maxCampaignChapterInfo.FirstTimeToPass = 0

	if retCode == 1 {
		maxCampaignChapterInfo.CampaignID = chapterInfo.CampaignId
	}

	// 获取未完成事件
	campaignEventModel := model.CampaignEventModel{Uin: this.Request.Uin}
	records, err := campaignEventModel.QueryCampaignEventUnfinished()
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	unfinishedRecords := make([]*table.TblCampaignEvent, 0, len(records))
	invalidOrClosedRecords := make([]*table.TblCampaignEvent, 0, len(records))

	for _, record := range records {
		if eventProtype, ok := config.GCampaignEventConfig.AttrMap[record.EventId]; ok {
			timePassed := int(base.GLocalizedTime.SecTimeStamp()) - record.StartTime
			if timePassed > eventProtype.Term {
				record.EventStatus = model.CAMPAIGN_EVENT_STATUS_CLOSED
				invalidOrClosedRecords = append(invalidOrClosedRecords, record)
			} else {
				if _, ok := config.GCampaignConfig.AttrMap[record.CampaignId]; !ok {
					invalidOrClosedRecords = append(invalidOrClosedRecords, record)
				} else {
					unfinishedRecords = append(unfinishedRecords, record)
				}
			}
		} else {
			record.EventStatus = model.CAMPAIGN_EVENT_STATUS_CLOSED
			invalidOrClosedRecords = append(invalidOrClosedRecords, record)
		}
	}

	// 这里获取为完成的进度
	campaignEvents := make([]*proto.ProtoCampaignEventInfo, len(unfinishedRecords))
	for index, tblEvent := range unfinishedRecords {
		campaignEvents[index] = new(proto.ProtoCampaignEventInfo)
		composeCampaignEventTblToProto(tblEvent, campaignEvents[index])
	}

	// 生产资源信息
	produceResourceModel := model.CampaignProduceResourceModel{Uin: this.Request.Uin}
	produceInfo := new(table.TblCampaignProduceResource)
	retCode, err = produceResourceModel.QueryCampaignProductResourceInfo(produceInfo)
	if err != nil {
		return retCode, err
	}

	campaignProduceStatusInfo := new(proto.ProtoCampaignProduceStatusInfo)
	campaignProduceStatusInfo.Uin = this.Request.Uin

	if retCode == 0 {

		retCode, err = produceResourceModel.RefreshResourceReceivedTime(int(base.GLocalizedTime.SecTimeStamp()), produceInfo, true)
		if err != nil {
			return retCode, err
		}

		composeProtoCampaignProduceStatusInfo(campaignProduceStatusInfo, nil)
	} else {
		composeProtoCampaignProduceStatusInfo(campaignProduceStatusInfo, produceInfo)
	}

	var responseData proto.ProtoQueryCampaignResponse
	responseData.CampaignInfo.CampaignId = maxCampaignChapterInfo.CampaignID
	responseData.CampaignInfo.Events = campaignEvents
	responseData.CampaignInfo.ProduceStatus = campaignProduceStatusInfo

	this.Response.ResData.Params = responseData

	return 0, nil
}

func GetMaxCampaignChapterInfo(uin int, chapterInfo *table.TblCampaignPassChapter) (int, error) {
	if chapterInfo == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.NullPoint()
	}

	campaignModel := model.CampaignPassChapterModel{Uin: uin}
	retCode, err := campaignModel.QueryMaxChapterInfoByUin(chapterInfo)
	if err != nil {
		return retCode, err
	}

	return 0, nil
}
