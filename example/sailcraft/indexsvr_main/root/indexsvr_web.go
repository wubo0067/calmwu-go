/*
 * @Author: calmwu
 * @Date: 2017-09-20 14:22:36
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-09-20 16:05:43
 * @Comment:
 */

package root

import (
	"fmt"
	"net/http"
	"sailcraft/base"
	"sailcraft/indexsvr_main/indexsvr"

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
	err := base.GinRegisterWebModule(ginRouter, indexsvr.WebIndexSvrModule)
	if err != nil {
		base.GLog.Error("GinRegisterWebModule failed! reason[%s]", err.Error())
		return err
	}

	servAddr := fmt.Sprintf("%s:%d", webListenIP, webListenPort)
	base.GLog.Debug("IndexSvr watch[%s]", servAddr)
	ginRouter.Run(servAddr)
	return nil
}

func onHealthCheck(c *gin.Context) {
	c.Status(http.StatusOK)
}
