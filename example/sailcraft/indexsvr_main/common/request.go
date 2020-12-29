/*
 * @Author: calmwu
 * @Date: 2017-09-20 17:12:00
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-05-19 22:37:21
 * @Comment:
 */

package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sailcraft/base"
	"strings"

	"github.com/gin-gonic/gin"
)

func UnpackRequest(c *gin.Context) *base.ProtoRequestS {
	bodyData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		base.GLog.Error("Read request body failed! reason[%s]", err.Error())
		return nil
	}

	var req base.ProtoRequestS
	err = json.Unmarshal(bodyData, &req)
	if err != nil {
		base.GLog.Error("decode body failed! reason[%s]", err.Error())
		return nil
	}

	return &req
}

func SendResponse(c *gin.Context, res *base.ProtoResponseS) {
	response, err := json.Marshal(res)
	if err == nil {
		base.GLog.Debug("send respone to %s", c.Request.RemoteAddr)
		c.Data(http.StatusOK, "text/plain; charset=utf-8", response)
	} else {
		base.GLog.Error("Json Marshal ProtoResponseS failed! reason[%s]", err.Error())
	}
}

func IndexSvrSync(indexSvrIPs []string, interfaceName string, req *base.ProtoRequestS) {
	if req.Version == -1 && req.EventId == -1 && req.TimeStamp == -1 {
		base.GLog.Debug("don't dispatch continue!!!")
		return
	}

	req.Version = -1
	req.EventId = -1
	req.TimeStamp = -1

	base.GLog.Debug("Now dispatch to other IndexSvrs:%+v", indexSvrIPs)

	for _, ip := range indexSvrIPs {
		url := fmt.Sprintf("http://%s:5000/sailcraft/api/v1/IndexSvr/%s", ip, interfaceName)

		marshalData, err := json.Marshal(req)
		if err != nil {
			base.GLog.Error("Json marsh req failed! reason[%s]", err.Error())
			return
		}

		_, err = http.Post(url, "text/plain; charset=utf-8", strings.NewReader(string(marshalData)))
		if err != nil {
			base.GLog.Error("Post invoke url[%s] failed! reason[%s]", url, err.Error())
			return
		}
		base.GLog.Debug("Post invoke url[%s] successed!", url)
	}
}
