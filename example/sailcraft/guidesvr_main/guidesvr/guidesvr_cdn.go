/*
 * @Author: calmwu
 * @Date: 2018-05-16 10:23:17
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-06-13 15:02:54
 * @Comment:
 */

package guidesvr

import (
	"sailcraft/base"
	"sailcraft/base/consul_api"
	"sailcraft/guidesvr_main/common"

	"github.com/gin-gonic/gin"
)

func (guideSvr *GuideSvrModule) ClientCDNResourceDownloadReport(c *gin.Context) {
	clientIP := base.GetClientAddrFromGin(c)

	req, err := base.UnpackClientRequest(c)
	if err != nil {
		base.GLog.Error("UnpackClientRequest failed! reason[%s]", err.Error())
		return
	}
	req.ChannelUID = clientIP
	consul_api.PostBaseRequstByConsulDns("ClientCDNResourceDownloadReport", req, common.ConsulClient,
		"CassandraSvr")
}
