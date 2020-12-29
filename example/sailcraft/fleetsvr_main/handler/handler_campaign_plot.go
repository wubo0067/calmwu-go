package handler

import (
	"sailcraft/base"
	"sailcraft/fleetsvr_main/custom_errors"
	"sailcraft/fleetsvr_main/errorcode"
	"sailcraft/fleetsvr_main/handlerbase"
	"sailcraft/fleetsvr_main/model"
	"sailcraft/fleetsvr_main/proto"
	"sailcraft/fleetsvr_main/table"
)

type CampaignPlotHandler struct {
	handlerbase.WebHandler
}

func (this *CampaignPlotHandler) List() (int, error) {

	campaignPlotModel := model.CampaignPlotModel{Uin: this.Request.Uin}
	plotList, err := campaignPlotModel.GetPlotsInfo()
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	var responseData proto.ProtoGetCampaignPlotListResponse
	if len(plotList) > 0 {
		for _, v := range plotList {
			plotInfo := new(proto.ProtoCampaignPlotInfo)
			retCode, err := composeProtoCampaignPlot(plotInfo, v)
			if err != nil {
				return retCode, err
			}

			responseData.CampaignPlotList = append(responseData.CampaignPlotList, plotInfo)
		}
	} else {
		responseData.CampaignPlotList = make([]*proto.ProtoCampaignPlotInfo, 0)
	}

	this.Response.ResData.Params = responseData

	return 0, nil
}

func (this *CampaignPlotHandler) Pass() (int, error) {
	var reqParams proto.ProtoPassCampaignPlotRequest
	err := this.UnpackParams(&reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if reqParams.ProtypeId <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("protype id is invalid")
	}

	campaignPlotModel := model.CampaignPlotModel{Uin: this.Request.Uin}
	plotInfo, err := campaignPlotModel.GetPlotInfo(reqParams.ProtypeId)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if plotInfo != nil {
		return errorcode.ERROR_CODE_CAMPAIGN_PLOT_ALREADY_PASSED, custom_errors.New("campaign plot has already passed")
	}

	plotInfo = new(table.TblCampaignPlot)
	plotInfo.Uin = this.Request.Uin
	plotInfo.ProtypeId = reqParams.ProtypeId
	plotInfo.PassTime = int(base.GLocalizedTime.SecTimeStamp())

	retCode, err := campaignPlotModel.AddPlotInfo(plotInfo)
	if err != nil {
		return retCode, err
	}

	var responseData proto.ProtoPassCampaignPlotResponse
	protoPlotInfo := new(proto.ProtoCampaignPlotInfo)
	composeProtoCampaignPlot(protoPlotInfo, plotInfo)
	responseData.CampaignPlotList = append(responseData.CampaignPlotList, protoPlotInfo)

	this.Response.ResData.Params = responseData

	return 0, nil
}

func composeProtoCampaignPlot(target *proto.ProtoCampaignPlotInfo, data *table.TblCampaignPlot) (int, error) {
	if target == nil || data == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.NullPoint()
	}

	target.ProtypeId = data.ProtypeId
	target.PassTime = data.PassTime

	return 0, nil
}
