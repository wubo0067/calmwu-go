package websvr

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sailcraft/base"
	"sailcraft/fleetsvr_main/handler"
	"sailcraft/fleetsvr_main/handlerbase"
	"strings"

	"github.com/gin-gonic/gin"
)

type FleetRouter struct {
	HandlerType reflect.Type
	ActionMap   map[string]reflect.Method
}

type FleetWebModuleMetas map[string]*FleetRouter

type FleetWebSvrModule struct {
	ModuleMetas     base.WebModuleMetas
	FleetHandlerMap FleetWebModuleMetas
}

const (
	API_PREFIX = "/sailcraft/api/v1/FleetModule/"
)

func (svr *FleetWebSvrModule) InitModule() {
	svr.ModuleMetas = make(base.WebModuleMetas)

	svr.FleetHandlerMap = make(FleetWebModuleMetas)

	svr.RegisterHandler("Test", &handlerbase.WebHandlerTest{})

	svr.RegisterHandler("Achievement", &handler.AchievemenHandler{})
	svr.RegisterHandler("ActivityTask", &handler.ActivityTaskHandler{})
	svr.RegisterHandler("BattleShip", &handler.BattleShipHandler{})
	svr.RegisterHandler("PVE", &handler.CampaignPassChapterHandler{})
	svr.RegisterHandler("PVEEvent", &handler.CampaignEventHandler{})
	svr.RegisterHandler("PVEResources", &handler.CampaignProduceResourcesHandler{})
	svr.RegisterHandler("Event", &handler.EventHandler{})
	svr.RegisterHandler("GrowupTask", &handler.GrowupTaskHandler{})
	svr.RegisterHandler("Guild", &handler.GuildHandler{})
	svr.RegisterHandler("Timer", &handler.TimerHandler{})
	svr.RegisterHandler("Refresh", &handler.RefreshHandler{})
	svr.RegisterHandler("PVEPlot", &handler.CampaignPlotHandler{})
	svr.RegisterHandler("GuildWar", &handler.GuildWarHandler{})
	svr.RegisterHandler("Message", &handler.MessageHandler{})
	svr.RegisterHandler("Salvage", &handler.SalvageHandler{})
	svr.RegisterHandler("Relics", &handler.AncientRelicsHandler{})

	base.RegisterModuleInterface(API_PREFIX+":handler/*action", base.HTTP_METHOD_POST, svr.HandleFunc, svr.ModuleMetas)

	// 测试---修改用户信息
	//base.RegisterModuleInterface(API_PREFIX+"PressTestModUserInfo", base.HTTP_METHOD_POST, svr.PressTestModUserInfo, svr.ModuleMetas)
}

func (svr *FleetWebSvrModule) HandleFunc(c *gin.Context) {
	base.GLog.Debug("FleetWebSvrModule Handle: \n%s\n Access Url: %s\n", c.Request.RemoteAddr, c.Request.URL)

	handlerName := strings.ToLower(c.Param("handler"))
	if handlerRouter, ok := svr.FleetHandlerMap[handlerName]; ok {
		handlerV := reflect.New(handlerRouter.HandlerType)

		handler, isHandler := handlerV.Interface().(handlerbase.WebHandlerInterface)
		if isHandler {
			handler.SetContext(c)
			err := handler.Prepare()
			if err == nil {
				action := strings.TrimLeft(strings.ToLower(c.Param("action")), "/")
				if method, ok := handlerRouter.ActionMap[action]; ok {
					// 这里通过名字来调用指定的函数
					results := handlerV.MethodByName(method.Name).Call([]reflect.Value{})
					if len(results) != 2 {
						handler.SetReturnCode(-1)
						base.GLog.Error("Handle %s failed! reason[result length wrong]", action)
					} else {
						retCode := results[0].Int()
						err := results[1].Interface()

						handler.SetReturnCode(retCode)
						if err != nil {
							base.GLog.Error("Handle %s failed! reason[%s]", action, err)
						}
					}
				} else {
					goto NotFound
				}
			}

			handler.Finish()
		} else {
			goto NotFound
		}
	} else {
		goto NotFound
	}

	return

NotFound:
	base.GLog.Debug("404 page not found")
	c.Data(http.StatusNotFound, "text/plain; charset=utf-8", []byte("404 page not found"))
}

func (svr *FleetWebSvrModule) RegisterHandler(handlerName string, handler handlerbase.WebHandlerInterface) {
	v := reflect.ValueOf(handler)
	t := v.Type()

	fmt.Println(t)
	router := new(FleetRouter)
	router.ActionMap = make(map[string]reflect.Method)
	router.HandlerType = reflect.Indirect(v).Type()

	methodNum := t.NumMethod()
	for i := 0; i < methodNum; i++ {
		method := t.Method(i)
		// 判断参数为空
		if method.Type.NumIn() != 1 {
			continue
		}

		// 判断返回值为2个
		if method.Type.NumOut() != 2 {
			continue
		}

		// 判断返回值为int, error
		if !reflect.TypeOf(0).AssignableTo(method.Type.Out(0)) {
			continue
		}

		if !reflect.TypeOf(errors.New("")).AssignableTo(method.Type.Out(1)) {
			continue
		}

		actionName := strings.ToLower(method.Name)
		router.ActionMap[actionName] = method
	}

	if len(router.ActionMap) > 0 {
		svr.FleetHandlerMap[strings.ToLower(handlerName)] = router
	}
}
