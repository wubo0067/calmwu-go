package handler

import (
	"sailcraft/base"
	"sailcraft/fleetsvr_main/errorcode"
	"sailcraft/fleetsvr_main/handlerbase"
	"sailcraft/fleetsvr_main/proto"
)

type RefreshHandler struct {
	handlerbase.WebHandler
}

func (this *RefreshHandler) Info() (int, error) {
	freshInfo, err := getActivityTaskFreshInfo(this.Request.Uin)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	var responseData proto.ProtoRefreshResponse
	if freshInfo == nil || base.GLocalizedTime.IsToday(int64(freshInfo.FreshTime)) {
		responseData.ActivityTaskFreshed = 0
	} else {
		responseData.ActivityTaskFreshed = 1
	}

	refreshTime, err := base.GLocalizedTime.TodayClock(23, 59, 59)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}
	responseData.ActivityTaskRestTimeToFresh = int(refreshTime.Sub(base.GLocalizedTime.Now()).Seconds())

	this.Response.ResData.Params = responseData

	return 0, nil
}
