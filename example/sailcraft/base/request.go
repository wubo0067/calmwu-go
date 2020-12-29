/*
 * @Author: calmwu
 * @Date: 2017-09-20 17:12:00
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-06-13 14:26:45
 * @Comment:
 */

package base

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
)

func UnpackRequest(c *gin.Context) *ProtoRequestS {
	bodyData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		GLog.Error("Read request body failed! reason[%s]", err.Error())
		return nil
	}

	GLog.Debug("Request Data:\n%s", bodyData)

	var req ProtoRequestS
	err = json.Unmarshal(bodyData, &req)
	if err != nil {
		GLog.Error("decode body failed! reason[%s]", err.Error())
		return nil
	}

	return &req
}

func SendResponse(c *gin.Context, res *ProtoResponseS) {
	response, err := json.Marshal(res)
	if err == nil {
		GLog.Debug("send respone to %s\nResponse Data:\n%s", c.Request.RemoteAddr, response)
		c.Data(http.StatusOK, "text/plain; charset=utf-8", response)
	} else {
		GLog.Error("Json Marshal ProtoResponseS failed! reason[%s]", err.Error())
	}
}

func GetClientAddrFromGin(c *gin.Context) string {
	var remoteAddr string
	remoteAddrLst, ok := c.Request.Header["X-Real-Ip"]
	if !ok {
		remoteAddr = "Unknown"
	} else {
		remoteAddr = remoteAddrLst[0]
	}
	return remoteAddr
}

func UnpackClientRequest(c *gin.Context) (*ProtoRequestS, error) {
	var req ProtoRequestS
	dcompressR, _ := zlib.NewReader(c.Request.Body)
	err := json.NewDecoder(dcompressR).Decode(&req)
	return &req, err
}

func SendResponseToClient(c *gin.Context, res *ProtoResponseS) {
	var compressBuf bytes.Buffer
	compressW := zlib.NewWriter(&compressBuf)
	json.NewEncoder(compressW).Encode(res)
	compressW.Close()
	//GLog.Debug("compress size[%d]", compressBuf.Len())
	c.Data(http.StatusOK, "text/plain; charset=utf-8", compressBuf.Bytes())
}

func PostRequest(url string, req *ProtoRequestS) (*ProtoResponseS, error) {
	serialData, err := json.Marshal(req)
	if err != nil {
		GLog.Error("PostRequest to url[%s] Marshal failed! reason[%s]",
			url, err.Error())
		return nil, err
	}

	res, err := http.Post(url, "text/plain; charset=utf-8", strings.NewReader(string(serialData)))
	if err != nil {
		GLog.Error("PostRequest to url[%s] Post failed! reason[%s]",
			url, err.Error())
		return nil, err
	}

	bodyData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		GLog.Error("Read body failed! reason[%s]", err.Error())
		return nil, err
	}

	GLog.Debug("Response Data:\n%s", bodyData)

	var protoRes ProtoResponseS
	err = json.Unmarshal(bodyData, &protoRes)
	if err != nil {
		GLog.Error("decode ProtoResponseS failed! reason[%s]", err.Error())
		return nil, err
	}
	return &protoRes, nil
}

func MapstructUnPackByJsonTag(m interface{}, rawVal interface{}) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:  "json",
		Metadata: nil,
		Result:   rawVal,
	})
	if err != nil {
		GLog.Error("mapstructure.NewDecoder failed! reason[%s]", err.Error())
		return err
	}

	err = decoder.Decode(m)
	if err != nil {
		GLog.Error("Decode %s failed! reason[%s]", reflect.TypeOf(m).String(), err.Error())
		return err
	}
	return nil
}

type WebItfResData struct {
	Param   interface{}
	RetCode int
}

type webItfResponseFunc func()

func RequestPretreatment(c *gin.Context, interfaceName string, realReqPtr interface{}) (*WebItfResData, webItfResponseFunc, error) {
	var err error
	req := UnpackRequest(c)
	if req == nil {
		err = fmt.Errorf("unpack interface[%s] request failed!", interfaceName)
		GLog.Error(err.Error())
		return nil, nil, err
	}

	err = MapstructUnPackByJsonTag(req.ReqData.Params, realReqPtr)
	if err != nil {
		err = fmt.Errorf("Uin[%d] Decode %s failed! reason[%s]",
			req.Uin, reflect.Indirect(reflect.ValueOf(realReqPtr)).Type().String(), err.Error())
		GLog.Error(err.Error())
		return nil, nil, err
	}

	webItfResData := new(WebItfResData)

	return webItfResData, func() {
		if req != nil {
			var res ProtoResponseS
			res.Version = req.Version
			res.EventId = req.EventId
			res.ReturnCode = ProtoReturnCode(webItfResData.RetCode)
			res.TimeStamp = time.Now().UTC().Unix()
			res.ResData.InterfaceName = req.ReqData.InterfaceName
			if err != nil {
				res.ResData.Params = err.Error()
			} else {
				res.ResData.Params = webItfResData.Param
			}
			SendResponse(c, &res)
		}
	}, nil
}
