package handlerbase

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sailcraft/base"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type WebHandler struct {
	Ctx      *gin.Context
	Request  *base.ProtoRequestS
	Response *base.ProtoResponseS
	Handler  string
	Action   string
}

type WebHandlerInterface interface {
	SetContext(*gin.Context)
	Prepare() error
	Finish()
	UnpackRequest() error
	UnpackParams(interface{}) error
	SetReturnCode(int64)
}

func (this *WebHandler) SetContext(ctx *gin.Context) {
	this.Ctx = ctx
	this.Handler = this.Ctx.Param("handler")
	this.Action = strings.TrimLeft(this.Ctx.Param("action"), "/")
}

func (this *WebHandler) Prepare() error {
	err := this.UnpackRequest()
	if err != nil {
		return fmt.Errorf("Unpack request error! [%s]", err.Error())
	}

	this.Response = new(base.ProtoResponseS)
	this.Response.EventId = this.Request.EventId
	this.Response.Version = this.Request.Version
	this.Response.TimeStamp = time.Now().Unix()
	this.Response.ReturnCode = -1
	this.Response.ResData.InterfaceName = fmt.Sprintf("%s/%s", this.Handler, this.Action)

	if this.Request.Uin <= 0 {
		return fmt.Errorf("uin is invalid")
	}

	base.GLog.Debug("query[%s/%s] from %s", this.Handler, this.Action, this.Ctx.Request.RemoteAddr)

	return nil
}

func (this *WebHandler) Finish() {
	if this.Response == nil {
		return
	}

	response, err := json.Marshal(this.Response)
	if err == nil {
		base.GLog.Debug("send respone to %s\nResponse Data:\n%s", this.Ctx.Request.RemoteAddr, response)
		this.Ctx.Data(http.StatusOK, "text/plain; charset=utf-8", response)
	} else {
		base.GLog.Error("Json Marshal ProtoResponseS failed! reason[%s]", err.Error())
	}
}

func (this *WebHandler) UnpackRequest() error {
	bodyData, err := ioutil.ReadAll(this.Ctx.Request.Body)
	if err != nil {
		return fmt.Errorf("Read request body failed! reason[%s]", err.Error())
	}
	base.GLog.Debug("Request Data:\n%s", bodyData)

	this.Request = new(base.ProtoRequestS)
	err = json.Unmarshal(bodyData, this.Request)
	if err != nil {
		return fmt.Errorf("decode body failed! reason[%s]", err.Error())
	}

	return nil
}

func (this *WebHandler) UnpackParams(i interface{}) error {
	err := base.ConvertHashToObj(this.Request.ReqData.Params, i, "json")
	if err != nil {
		return fmt.Errorf("UnpackParams failed! reason:[%s], typeof Params: %T", err.Error(), i)
	}

	return nil
}

func (this *WebHandler) SetReturnCode(retCode int64) {
	if this.Response != nil {
		this.Response.ReturnCode = base.ProtoReturnCode(retCode)
	}
}
