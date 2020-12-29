package fleetsvr

import (
	"fmt"
	"net/http"
	"sailcraft/base"
	"sailcraft/fleetsvr_main/websvr"

	"github.com/gin-gonic/gin"
)

type ServiceMgr struct {
	ginRouter *gin.Engine
}

var (
	GServiceMgr *ServiceMgr = nil
)

func init() {
	gin.SetMode(gin.DebugMode)

	GServiceMgr = new(ServiceMgr)
	GServiceMgr.ginRouter = gin.Default()
}

func (servMgr *ServiceMgr) RunServ(webListenIP string, webListenPort int) error {
	fleetWebModule := new(websvr.FleetWebSvrModule)
	fleetWebModule.InitModule()

	err := base.GinRegisterWebModule(servMgr.ginRouter, fleetWebModule)
	if err != nil {
		return err
	}

	servAddr := fmt.Sprintf("%s:%d", webListenIP, webListenPort)
	base.GLog.Debug("Sailcraft watch[%s]", servAddr)
	servMgr.ginRouter.Run(servAddr)

	return nil
}

func onConsulCheck(c *gin.Context) {
	c.Status(http.StatusOK)
}
