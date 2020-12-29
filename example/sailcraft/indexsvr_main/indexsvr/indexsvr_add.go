/*
 * @Author: CALM.WU
 * @Date: 2017-09-26 11:21:20
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-05-19 22:00:46
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

func (indexSvr *IndexSvrModule) AddGuildIndex(c *gin.Context) {
	base.GLog.Debug("AddGuildIndex query from [%s]", c.Request.RemoteAddr)

	req := base.UnpackRequest(c)
	if req != nil {
		var reqParams proto.ProtoAddGuildNameReqParamsS
		err := mapstructure.Decode(req.ReqData.Params, &reqParams)
		if err != nil {
			base.GLog.Error("Decode params to ProtoAddGuildNameReqParamsS faield! reason[%s]", reqParams)
			return
		}

		var addDataMetaI = &proto.GuildInfoS{ID: reqParams.ID, GuildName: reqParams.GuildName, Creator: reqParams.Creator, PerformId: reqParams.PerformId}

		base.GLog.Debug("add Guild[%s] Name [%s]", reqParams.ID, reqParams.GuildName)

		data.GDataMgr.Set(reqParams.GuildName, addDataMetaI)
		data.GDataMgr.Set(reqParams.PerformId, addDataMetaI)

		// 向其它indexsvr广播，同样适用http调用
		common.IndexSvrSync(common.GConfig.IndexSvrCluster, req.ReqData.InterfaceName, req)
	} else {
		base.GLog.Error("UnpackRequest failed!")
	}
}

func (indexSvr *IndexSvrModule) AddUserIndex(c *gin.Context) {
	base.GLog.Debug("AddUserIndex query from [%s]", c.Request.RemoteAddr)

	req := base.UnpackRequest(c)
	if req != nil {
		var reqParams proto.ProtoAddUserNameReqParamsS
		err := mapstructure.Decode(req.ReqData.Params, &reqParams)
		if err != nil {
			base.GLog.Error("Decode params to ProtoAddUserNameReqParamsS faield! reason[%s]", reqParams)
			return
		}

		var addDataMetaI = &proto.UserInfoS{Uin: reqParams.Uin, UserName: reqParams.UserName}

		base.GLog.Debug("add User[%s] Name [%s]", reqParams.Uin, reqParams.UserName)

		data.GDataMgr.Set(reqParams.UserName, addDataMetaI)
		// 为uin也加上索引
		data.GDataMgr.Set(reqParams.Uin, addDataMetaI)

		// 向其它indexsvr广播，同样适用http调用
		common.IndexSvrSync(common.GConfig.IndexSvrCluster, req.ReqData.InterfaceName, req)

	} else {
		base.GLog.Error("UnpackRequest failed!")
	}
}
