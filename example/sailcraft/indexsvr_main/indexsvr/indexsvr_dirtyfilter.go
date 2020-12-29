/*
 * @Author: calmwu
 * @Date: 2018-04-28 15:25:02
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-04-28 15:33:24
 * @Comment:
 */

package indexsvr

import (
	"sailcraft/base"
	"sailcraft/base/word_filter"
	"sailcraft/indexsvr_main/common"
	"sailcraft/indexsvr_main/proto"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
)

func (indexSvr *IndexSvrModule) DirtyWordFilter(c *gin.Context) {
	req := base.UnpackRequest(c)
	if req == nil {
		base.GLog.Error("DirtyWordFilter read request failed!")
		return
	}

	var reqData proto.ProtoDirtyWordFilterReq
	err := mapstructure.Decode(req.ReqData.Params, &reqData)
	if err != nil {
		base.GLog.Error("Uin[%d] Decode ProtoDirtyWordFilterReq failed! reason[%s]",
			req.Uin, err.Error())
		return
	}

	rawContent := []rune(reqData.Content)
	word_filter.FilterText(common.GConfig.DirtyWordFile, rawContent, []rune{}, '*')

	var resData proto.ProtoDirtyWordFilterRes
	resData.Uin = reqData.Uin
	resData.FilterContent = string(rawContent)
	if strings.Compare(reqData.Content, resData.FilterContent) == 0 {
		resData.HaveDirtyWords = 0
	} else {
		resData.HaveDirtyWords = 1
	}

	var res base.ProtoResponseS
	res.Version = req.Version
	res.EventId = req.EventId
	res.TimeStamp = time.Now().UTC().Unix()
	res.ReturnCode = 0
	res.ResData.InterfaceName = req.ReqData.InterfaceName
	res.ResData.Params = resData
	// 返回给客户端
	base.SendResponse(c, &res)
}
