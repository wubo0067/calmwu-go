/*
 * @Author: calmwu
 * @Date: 2018-04-26 14:58:13
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-05-16 18:08:43
 * @Comment:
 */

package guidesvr

import (
	"sailcraft/base"
	"sailcraft/guidesvr_main/common"
	"sailcraft/guidesvr_main/proto"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
)

func (guideSvr *GuideSvrModule) ClientNavigate(c *gin.Context) {
	req, err := base.UnpackClientRequest(c)
	if err != nil {
		base.GLog.Error("UnpackClientRequest failed! reason[%s]", err.Error())
		return
	}

	var reqData proto.ProtoClientNavigateReq
	err = mapstructure.Decode(req.ReqData.Params, &reqData)
	if err != nil {
		base.GLog.Error("Uin[%d] Decode ProtoClientNavigateReq failed! reason[%s]",
			req.Uin, err.Error())
		return
	}

	clientIP := base.GetClientAddrFromGin(c)
	base.GLog.Debug("clientIP[%s] ProtoClientNavigateReq:%+v", clientIP, reqData)

	navigationUrl := common.GConfig.GetNavigationConfig().CheckVersion(reqData.ClientVersion)
	if navigationUrl != nil {
		var navigateRes proto.ProtoClientNavigateRes
		navigateRes.UrlLoginCheck = navigationUrl.URLLoginCheck
		navigateRes.UrlProxySvr = navigationUrl.URLProxySvr
		navigateRes.UrlSdkSvr = navigationUrl.URLSdkSvr

		base.GLog.Debug("Uin[%d] navigateRes:%+v", req.Uin, navigateRes)

		var res base.ProtoResponseS
		res.Version = req.Version
		res.EventId = req.EventId
		res.TimeStamp = time.Now().Unix()
		res.ReturnCode = 0
		res.ResData.InterfaceName = req.ReqData.InterfaceName
		res.ResData.Params = navigateRes

		base.SendResponseToClient(c, &res)
	}
}
