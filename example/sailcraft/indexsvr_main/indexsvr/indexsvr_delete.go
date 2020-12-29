/*
 * @Author: calmwu
 * @Date: 2017-09-25 09:59:27
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-09-25 10:17:36
 */

package indexsvr

import (
	"sailcraft/base"
	"sailcraft/indexsvr_main/common"
	"sailcraft/indexsvr_main/data"
	"sailcraft/indexsvr_main/proto"

	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
)

func (indexSvr *IndexSvrModule) DeleteGuildIndex(c *gin.Context) {
	base.GLog.Debug("DeleteGuildIndex query from [%s]", c.Request.RemoteAddr)

	req := base.UnpackRequest(c)
	if req != nil {
		var reqParams proto.ProtoDeleteGuildNameReqParamsS
		err := mapstructure.Decode(req.ReqData.Params, &reqParams)
		if err != nil {
			base.GLog.Error("Decode params to ProtoDeleteGuildNameReqParamsS faield! reason[%s]", reqParams)
			return
		}

		var delDataMetaI = &proto.GuildInfoS{ID: reqParams.ID, GuildName: reqParams.GuildName, Creator: reqParams.Creator, PerformId: reqParams.PerformId}

		base.GLog.Debug("delete Guild[%s] Name [%s]", reqParams.ID, reqParams.GuildName)

		data.GDataMgr.Delete(reqParams.GuildName, delDataMetaI)
		data.GDataMgr.Delete(reqParams.PerformId, delDataMetaI)

		// 向其它indexsvr广播，同样适用http调用
		common.IndexSvrSync(common.GConfig.IndexSvrCluster, req.ReqData.InterfaceName, req)
	} else {
		base.GLog.Error("UnpackRequest failed!")
	}
}
