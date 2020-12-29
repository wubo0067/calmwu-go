/*
 * @Author: calmwu
 * @Date: 2018-01-08 09:51:34
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-06-13 15:37:58
 */

package guidesvr

import (
	"sailcraft/base"
	"sailcraft/base/consul_api"
	"sailcraft/guidesvr_main/common"

	"github.com/gin-gonic/gin"
)

func (guideSvr *GuideSvrModule) TuitionStepReport(c *gin.Context) {
	clientIP := base.GetClientAddrFromGin(c)

	req, err := base.UnpackClientRequest(c)
	if err != nil {
		base.GLog.Error("UnpackClientRequest failed! reason[%s]", err.Error())
		return
	}
	req.ChannelUID = clientIP
	consul_api.PostBaseRequstByConsulDns("TuitionStepReport", req, common.ConsulClient,
		"CassandraSvr")
}

func (guideSvr *GuideSvrModule) UploadUserAction(c *gin.Context) {
	clientIP := base.GetClientAddrFromGin(c)

	req, err := base.UnpackClientRequest(c)
	if err != nil {
		base.GLog.Error("UnpackClientRequest failed! reason[%s]", err.Error())
		return
	}
	req.ChannelUID = clientIP
	consul_api.PostBaseRequstByConsulDns("UploadUserAction", req, common.ConsulClient,
		"CassandraSvr")
}
