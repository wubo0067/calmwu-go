/*
 * @Author: calmwu
 * @Date: 2017-09-23 10:33:08
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-05-19 22:28:15
 * @Comment:
 */

package indexsvr

import (
	"sailcraft/base"
	"sailcraft/indexsvr_main/common"
	"sailcraft/indexsvr_main/data"
	"sailcraft/indexsvr_main/proto"

	"github.com/mitchellh/mapstructure"

	"github.com/gin-gonic/gin"
)

func (indexSvr *IndexSvrModule) ModifyUserName(c *gin.Context) {
	base.GLog.Debug("ModifyUserName query from [%s]", c.Request.RemoteAddr)

	req := base.UnpackRequest(c)
	if req != nil {
		var reqParams proto.ProtoModifyUserNameReqParamsS
		err := mapstructure.Decode(req.ReqData.Params, &reqParams)
		if err != nil {
			base.GLog.Error("Decode params to ProtoModifyUserNameReqParamsS faield! reason[%s]", reqParams)
			return
		}

		var oldDataMetaI = &proto.UserInfoS{Uin: reqParams.Uin, UserName: reqParams.UserName}
		var newDataMetaI = &proto.UserInfoS{Uin: reqParams.Uin, UserName: reqParams.NewUserName}

		base.GLog.Debug("change User[%s] Name [%s] ===> [%s]", reqParams.Uin, reqParams.UserName, reqParams.NewUserName)

		data.GDataMgr.Modify(reqParams.UserName, reqParams.NewUserName, oldDataMetaI, newDataMetaI)

		// 向其它indexsvr广播，同样适用http调用
		common.IndexSvrSync(common.GConfig.IndexSvrCluster, req.ReqData.InterfaceName, req)
	} else {
		base.GLog.Error("UnpackRequest failed!")
	}
}

func (indexSvr *IndexSvrModule) ModifyGuildName(c *gin.Context) {
	base.GLog.Debug("ModifyGuildName query from [%s]", c.Request.RemoteAddr)

	req := base.UnpackRequest(c)
	if req != nil {
		var reqParams proto.ProtoModifyGuildNameReqParamsS
		err := mapstructure.Decode(req.ReqData.Params, &reqParams)
		if err != nil {
			base.GLog.Error("Decode params to ProtoModifyGuildNameReqParamsS faield! reason[%s]", reqParams)
			return
		}

		var oldDataMetaI = &proto.GuildInfoS{ID: reqParams.ID, GuildName: reqParams.GuildName}
		var newDataMetaI = &proto.GuildInfoS{ID: reqParams.ID, GuildName: reqParams.NewGuildName}

		base.GLog.Debug("change Guild[%s] Name [%s] ===> [%s]", reqParams.ID, reqParams.GuildName, reqParams.NewGuildName)

		data.GDataMgr.Modify(reqParams.GuildName, reqParams.NewGuildName, oldDataMetaI, newDataMetaI)

		// 向其它indexsvr广播，同样适用http调用
		common.IndexSvrSync(common.GConfig.IndexSvrCluster, req.ReqData.InterfaceName, req)
	} else {
		base.GLog.Error("UnpackRequest failed!")
	}
}
