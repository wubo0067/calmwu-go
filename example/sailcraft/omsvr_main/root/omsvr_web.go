/*
 * @Author: calmwu
 * @Date: 2018-05-18 11:03:36
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-05-18 11:05:09
 * @Comment:
 */

package root

import (
	"fmt"
	"net/http"
	"sailcraft/base"
	"sailcraft/omsvr_main/web"

	"github.com/gin-gonic/gin"
)

var (
	ginRouter *gin.Engine
)

func init() {
	gin.SetMode(gin.DebugMode)
	ginRouter = gin.Default()
}

func RunWebServ(webListenIP string, webListenPort int) error {
	// 注册接口
	err := base.GinRegisterWebModule(ginRouter, web.WebOMSModule)
	if err != nil {
		base.GLog.Error("GinRegisterWebModule failed! reason[%s]", err.Error())
		return err
	}

	servAddr := fmt.Sprintf("%s:%d", webListenIP, webListenPort)
	base.GLog.Debug("OMSvr watch[%s]", servAddr)
	ginRouter.Run(servAddr)
	return nil
}

func onHealthCheck(c *gin.Context) {
	c.Status(http.StatusOK)
}
