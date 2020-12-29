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

type CampaignProduceResourcesHandler struct {
	handlerbase.WebHandler
}

func (this *CampaignProduceResourcesHandler) DetailInfo() (int, error) {
	produceResourceModel := model.CampaignProduceResourceModel{Uin: this.Request.Uin}

	produceInfo := new(table.TblCampaignProduceResource)

	// 读取数据库
	retCode, err := produceResourceModel.QueryCampaignProductResourceInfo(produceInfo)

	if err != nil {
		return retCode, err
	}

	campaignProduceResourceInfo := new(proto.ProtoCampaignProduceResourceInfo)
	campaignProduceResourceInfo.Uin = this.Request.Uin

	if retCode == 0 {
		retCode, err = composeProtoCampaignProduceResourceInfo(campaignProduceResourceInfo, nil)
		if err != nil {
			return retCode, err
		}
	} else {
		retCode, err = composeProtoCampaignProduceResourceInfo(campaignProduceResourceInfo, produceInfo)
		if err != nil {
			return retCode, err
		}
	}

	var responseData proto.ProtoQueryCampaignProduceResourceResponse
	responseData.CampaginProduceResourceInfo = *campaignProduceResourceInfo
	this.Response.ResData.Params = responseData

	return 0, nil
}

func (this *CampaignProduceResourcesHandler) SimpleInfo() (int, error) {
	produceResourceModel := model.CampaignProduceResourceModel{Uin: this.Request.Uin}

	produceInfo := new(table.TblCampaignProduceResource)

	// 读取数据库
	retCode, err := produceResourceModel.QueryCampaignProductResourceInfo(produceInfo)

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

	var responseData proto.ProtoQueryCampaignProduceStatusResponse
	responseData.CampaignProduceStatusInfo = *campaignProduceStatusInfo
	this.Response.ResData.Params = responseData

	return 0, nil
}

// 领取生产资源
func (this *CampaignProduceResourcesHandler) Receive() (int, error) {
	produceResourceModel := model.CampaignProduceResourceModel{Uin: this.Request.Uin}

	produceInfo := new(table.TblCampaignProduceResource)

	// 读取数据库
	retCode, err := produceResourceModel.QueryCampaignProductResourceInfo(produceInfo)

	if err != nil {
		return retCode, err
	}

	// 计算剩余时间，判断是否可以领取
	if retCode == 0 {
		retCode, err = produceResourceModel.RefreshResourceReceivedTime(int(base.GLocalizedTime.SecTimeStamp()), produceInfo, true)
		if err != nil {
			return retCode, err
		}

		return errorcode.ERROR_CODE_CAMPAIGN_PRODUCE_RESOURCE_TIME_NOT_COMMING, custom_errors.New("it's not time to receive resource")
	} else {
		restTime, err := restTimeToReceive(produceInfo.LastReceiveTime)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		if restTime > 0 {
			return errorcode.ERROR_CODE_CAMPAIGN_PRODUCE_RESOURCE_TIME_NOT_COMMING, custom_errors.New("it's not time to receive resource")
		}
		retCode, err = produceResourceModel.RefreshResourceReceivedTime(int(base.GLocalizedTime.SecTimeStamp()), produceInfo, true)
		if err != nil {
			return retCode, err
		}
	}

	campaignProduceStatusInfo := new(proto.ProtoCampaignProduceStatusInfo)
	campaignProduceStatusInfo.Uin = this.Request.Uin
	composeProtoCampaignProduceStatusInfo(campaignProduceStatusInfo, produceInfo)

	var responseData proto.ProtoReceiveCampaignProduceResourceResponse
	responseData.CampaignProduceStatusInfo = *campaignProduceStatusInfo
	this.Response.ResData.Params = responseData

	return 0, nil
}

func (this CampaignProduceResourcesHandler) Init() (int, error) {
	produceResourceModel := model.CampaignProduceResourceModel{Uin: this.Request.Uin}

	produceInfo := new(table.TblCampaignProduceResource)
	retCode, err := produceResourceModel.RefreshResourceReceivedTime(int(base.GLocalizedTime.SecTimeStamp()), produceInfo, false)
	if err != nil {
		return retCode, err
	}

	return 0, nil
}

func composeProtoCampaignProduceResourceInfo(protoProduceResourceInfo *proto.ProtoCampaignProduceResourceInfo, produceInfo *table.TblCampaignProduceResource) (int, error) {
	if protoProduceResourceInfo == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.NullPoint()
	}

	// 计算剩余时间，判断是否可以领取
	if produceInfo == nil {
		protoProduceResourceInfo.LastReceivedTime = 0
		protoProduceResourceInfo.CanReceive = 0

		restTime, err := restTimeToReceive(int(base.GLocalizedTime.SecTimeStamp()))
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		protoProduceResourceInfo.RestTimeToReceive = restTime
	} else {
		protoProduceResourceInfo.LastReceivedTime = produceInfo.LastReceiveTime
		restTime, err := restTimeToReceive(produceInfo.LastReceiveTime)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		if restTime > 0 {
			protoProduceResourceInfo.CanReceive = 0
			protoProduceResourceInfo.RestTimeToReceive = restTime
		} else {
			protoProduceResourceInfo.CanReceive = 1
			protoProduceResourceInfo.RestTimeToReceive = 0
		}
	}

	// 计算领取资源数

	chapterInfo := new(table.TblCampaignPassChapter)

	passCampaignChapterModel := model.CampaignPassChapterModel{Uin: protoProduceResourceInfo.Uin}
	retCode, err := passCampaignChapterModel.QueryMaxChapterInfoByUin(chapterInfo)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	var totalResources config.ResourcesAttr
	totalResources.EscapeNil()
	totalResources.Add(&config.GGlobalConfig.Campaign.ProduceResource)

	if retCode == 1 {
		for _, campaignProtype := range config.GCampaignConfig.AttrMap {
			if campaignProtype.Id <= chapterInfo.CampaignId {
				totalResources.Add(&campaignProtype.Tribute)
			}
		}
	}

	ResourcesConfigToProto(&totalResources, &protoProduceResourceInfo.TotalProduceResource)

	return 0, nil
}

func composeProtoCampaignProduceStatusInfo(protoProduceStatusInfo *proto.ProtoCampaignProduceStatusInfo, produceInfo *table.TblCampaignProduceResource) {
	if protoProduceStatusInfo != nil {
		// 计算剩余时间，判断是否可以领取
		if produceInfo == nil {
			protoProduceStatusInfo.LastReceivedTime = 0
			protoProduceStatusInfo.CanReceive = 0
			protoProduceStatusInfo.RestTimeToReceive = 0
		} else {
			protoProduceStatusInfo.LastReceivedTime = produceInfo.LastReceiveTime
			restTime, _ := restTimeToReceive(produceInfo.LastReceiveTime)
			if restTime > 0 {
				protoProduceStatusInfo.CanReceive = 0
				protoProduceStatusInfo.RestTimeToReceive = restTime
			} else {
				protoProduceStatusInfo.CanReceive = 1
				protoProduceStatusInfo.RestTimeToReceive = 0
			}
		}
	}
}

// 获取领取资源剩余时间
func restTimeToReceive(lastReceiveTime int) (int, error) {
	const (
		ReceiveInterval = 12 * 60 * 60
	)

	nowTimeStamp := base.GLocalizedTime.SecTimeStamp()
	timePassed := int(nowTimeStamp) - lastReceiveTime

	if timePassed > ReceiveInterval {
		return 0, nil
	}

	sixOclock, err := base.GLocalizedTime.TodayClock(6, 0, 0)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("time format error")
	}

	sixOclockTS := sixOclock.Unix()

	if lastReceiveTime < int(sixOclockTS) {
		if nowTimeStamp > sixOclockTS {
			return 0, nil
		}

		return int(sixOclockTS - nowTimeStamp), nil
	} else {
		eighteenOclockTS := sixOclockTS + ReceiveInterval

		if lastReceiveTime < int(eighteenOclockTS) {
			if nowTimeStamp > eighteenOclockTS {
				return 0, nil
			}

			return int(eighteenOclockTS - nowTimeStamp), nil
		}

		sixOclockTSTomorrow := eighteenOclockTS + ReceiveInterval

		return int(sixOclockTSTomorrow - nowTimeStamp), nil
	}
}
