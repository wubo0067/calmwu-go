package websvr

import (
	"sailcraft/base"
	"sailcraft/fleetsvr_main/handler"
	"sailcraft/fleetsvr_main/proto"
	"time"

	"github.com/gin-gonic/gin"
)

func (svr *FleetWebSvrModule) PressTestModUserInfo(c *gin.Context) {
	base.GLog.Debug("PressTestModUserInfo query from [%s]", c.Request.RemoteAddr)

	req := base.UnpackRequest(c)
	if req != nil {
		var res base.ProtoResponseS
		res.EventId = req.EventId
		res.Version = req.Version
		res.TimeStamp = time.Now().Unix()
		res.ReturnCode = -1
		res.ResData.InterfaceName = req.ReqData.InterfaceName

		userInfoChanged := proto.NewDefaultProtoUserInfoChanged()
		userInfoChanged.Gold = 100
		userInfoChanged.Gem = 100

		retCode, err := handler.UpdateUserInfoChanged(req.Uin, userInfoChanged)
		if err != nil {
			base.GLog.Error("UpdateUserInfoChanged failed! reason[%s]", err.Error())
			res.ReturnCode = base.ProtoReturnCode(retCode)
			base.SendResponse(c, &res)
			return
		}

		res.ReturnCode = 0

		base.SendResponse(c, &res)
	}
}
