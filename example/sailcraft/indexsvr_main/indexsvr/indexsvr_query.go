/*
 * @Author: calmwu
 * @Date: 2017-09-20 15:15:29
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-10-11 10:35:25
 * @Comment:
 */

package indexsvr

import (
	"sailcraft/base"
	"sailcraft/indexsvr_main/common"
	"sailcraft/indexsvr_main/data"
	"sailcraft/indexsvr_main/proto"
	"time"

	"github.com/emirpasic/gods/lists/singlylinkedlist"
	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
)

func (indexSvr *IndexSvrModule) FindGuidsByName(c *gin.Context) {
	base.GLog.Debug("FindGuidsByName query from [%s]", c.Request.RemoteAddr)

	req := base.UnpackRequest(c)
	if req != nil {
		var res base.ProtoResponseS
		res.EventId = req.EventId
		res.Version = req.Version
		res.TimeStamp = time.Now().Unix()
		res.ReturnCode = -1
		res.ResData.InterfaceName = req.ReqData.InterfaceName

		var reqParams proto.ProtoFindGuildsByNameRequestParamsS
		err := mapstructure.Decode(req.ReqData.Params, &reqParams)
		if err != nil {
			base.GLog.Error("Decode params to ProtoFindGuildsByNameRequestParamsS failed! reason[%s]", err.Error())
			base.SendResponse(c, &res)
			return
		}

		var protoGuildInfos proto.ProtoFindGuildsByNameResponseParamsS
		// 查询
		var likeRes *singlylinkedlist.List = nil
		if reqParams.QueryType == common.QUERYTYPE_LIKE {
			likeRes = data.GDataMgr.Like(reqParams.GuildName, proto.E_DATATYPE_GUILDINFO, reqParams.QueryCount)
		} else if reqParams.QueryType == common.QUERYTYPE_MATCH {
			likeRes = data.GDataMgr.Match(reqParams.GuildName, proto.E_DATATYPE_GUILDINFO, reqParams.QueryCount)
		} else {
			base.GLog.Error("Request queryType[%s] is invalid!", reqParams.QueryType)
			base.SendResponse(c, &res)
			return
		}

		protoGuildInfos.GuildCount = likeRes.Size()
		protoGuildInfos.GuildInfos = make([]*proto.GuildInfoS, likeRes.Size())

		res.ReturnCode = 0
		res.ResData.Params = protoGuildInfos

		likeRes.Each(func(index int, value interface{}) {
			if guildInfo, ok := value.(*proto.GuildInfoS); ok {
				protoGuildInfos.GuildInfos[index] = guildInfo
			} else {
				base.GLog.Error("index[%d] value:%v is not *proto.GuildInfoS")
			}
		})
		base.SendResponse(c, &res)
	}
}

func (indexSvr *IndexSvrModule) FindUsersByName(c *gin.Context) {
	base.GLog.Debug("FindUsersByName query from [%s]", c.Request.RemoteAddr)

	req := base.UnpackRequest(c)
	if req != nil {
		var res base.ProtoResponseS
		res.EventId = req.EventId
		res.Version = req.Version
		res.TimeStamp = time.Now().Unix()
		res.ReturnCode = -1
		res.ResData.InterfaceName = req.ReqData.InterfaceName

		var reqParams proto.ProtoFindUsersByNameRequestParamsS
		err := mapstructure.Decode(req.ReqData.Params, &reqParams)
		if err != nil {
			base.GLog.Error("Decode request to ProtoFindUsersByNameRequestParamsS failed! reason[%s]", err.Error())
			base.SendResponse(c, &res)
			return
		}

		var protoUserInfos proto.ProtoFindUsersByNameResponseParamsS
		// 查询
		var likeRes *singlylinkedlist.List = nil
		if reqParams.QueryType == common.QUERYTYPE_LIKE {
			likeRes = data.GDataMgr.Like(reqParams.UserName, proto.E_DATATYPE_USERINFO, reqParams.QueryCount)
		} else if reqParams.QueryType == common.QUERYTYPE_MATCH {
			likeRes = data.GDataMgr.Match(reqParams.UserName, proto.E_DATATYPE_USERINFO, reqParams.QueryCount)
		} else {
			base.GLog.Error("Request queryType[%s] is invalid!", reqParams.QueryType)
			base.SendResponse(c, &res)
			return
		}

		protoUserInfos.UserCount = likeRes.Size()
		protoUserInfos.UserInfos = make([]*proto.UserInfoS, likeRes.Size())

		res.ReturnCode = 0
		res.ResData.Params = protoUserInfos

		likeRes.Each(func(index int, value interface{}) {
			if userInfo, ok := value.(*proto.UserInfoS); ok {
				protoUserInfos.UserInfos[index] = userInfo
			} else {
				base.GLog.Error("index[%d] value:%v is not *proto.UserInfoS")
			}
		})
		base.SendResponse(c, &res)
	}
}
