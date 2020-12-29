/*
 * @Author: calmwu
 * @Date: 2018-07-17 10:20:51
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-07-17 10:32:12
 * @Comment:
 */

package web

import (
	"sailcraft/base"
	"sailcraft/csssvr_main/proto"
	"sailcraft/csssvr_main/store"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	INTERFACE_OPTYPE_PUT int = iota
	INTERFACE_OPTYPE_GET
)

func webItfProcess(c *gin.Context, interfaceName string, itfOpType int) {
	req := base.UnpackRequest(c)
	if req == nil {
		base.GLog.Error("unpack interface[%s] request failed!", interfaceName)
	}

	var res base.ProtoResponseS
	res.Version = req.Version
	res.TimeStamp = time.Now().Unix()
	res.EventId = req.EventId
	res.ReturnCode = -1
	res.ResData.InterfaceName = req.ReqData.InterfaceName

	if itfOpType == INTERFACE_OPTYPE_PUT {

		var cpd *proto.CassandraProcDataS = new(proto.CassandraProcDataS)
		cpd.RemoteIP = c.Request.RemoteAddr
		cpd.ReqData = req
		base.GLog.Debug("%s client[%s] channeluid[%s]", req.ReqData.InterfaceName, cpd.RemoteIP, req.ChannelUID)
		store.CasMgr.SubmitRequest(cpd)
		res.ReturnCode = 0

	} else if itfOpType == INTERFACE_OPTYPE_GET {

		queryResultI := store.CasMgr.QueryResult(req, c.Request.RemoteAddr)
		if queryResultI != nil {
			// 设置返回数据
			base.GLog.Debug("%s query result", req.ReqData.InterfaceName)
			res.ReturnCode = 0
			res.ResData.Params = queryResultI
		} else {
			base.GLog.Error("%s query result is nil!!", req.ReqData.InterfaceName)
		}
	}

	base.SendResponse(c, &res)
}
