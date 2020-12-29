/*
 * @Author: calmwu
 * @Date: 2017-12-26 15:08:41
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-07-06 12:48:35
 * @Comment:
 */

package guidesvr

import (
	"compress/zlib"
	"encoding/json"
	"sailcraft/base"
	"sailcraft/guidesvr_main/common"
	"sailcraft/guidesvr_main/proto"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
)

func getMaintainInfo() *proto.MaintainParams {
	key := common.GConfig.GetGMKey()
	redisNode, err := common.GRedisCluster.GetRedisNodeByKey(key)
	if err != nil {
		return nil
	}

	val, err := redisNode.StringGet(key)
	if err != nil {
		base.GLog.Error("Get key[%s] failed! reason[%s]", key, err.Error())
		return nil
	}

	if maintainBytes, ok := val.([]byte); ok {
		var gmToolMaintainInfo proto.MaintainParams
		err = json.Unmarshal(maintainBytes, &gmToolMaintainInfo)
		if err != nil {
			base.GLog.Error("Decode[%s] failed! error[%s]", string(maintainBytes), err.Error())
		} else {
			return &gmToolMaintainInfo
		}
	} else {
		base.GLog.Error("value type is not []byte")
	}
	return nil
}

func (guideSvr *GuideSvrModule) LoginCheck(c *gin.Context) {

	maintainInfo := getMaintainInfo()

	// 客户端ip
	clientIP := base.GetClientAddrFromGin(c)
	base.GLog.Debug("clientIP[%s] maintainInfo:%+v", clientIP, maintainInfo)

	var req base.ProtoRequestS
	if c != nil && c.Request != nil && c.Request.Body != nil && c.Request.ContentLength > 0 {
		base.GLog.Debug("ContentLength[%d]", c.Request.ContentLength)

		dcompressR, _ := zlib.NewReader(c.Request.Body)
		err := json.NewDecoder(dcompressR).Decode(&req)
		if err == nil {
			base.GLog.Debug("ClientIP[%s] LoginCheck Req[%+v]", clientIP, req)

			var res base.ProtoResponseS
			res.Version = req.Version
			res.EventId = req.EventId
			res.TimeStamp = time.Now().Unix()
			res.ReturnCode = 0
			res.ResData.InterfaceName = req.ReqData.InterfaceName

			// 接口名字检测
			if req.ReqData.InterfaceName != proto.APINAME_LoginCheck {
				base.GLog.Error("ReqData.InterFaceName[%s] is invalid!", req.ReqData.InterfaceName)

				var failureInfo base.ProtoFailInfoS
				failureInfo.FailureReason = "Interface Name is invalid!"
				res.ResData.Params = failureInfo
				res.ReturnCode = proto.ProtoRetCodeError
				base.SendResponseToClient(c, &res)
				return
			}

			// map ==》 ProtoSandMonkLoginCheckReqS
			var reqLoginCheck proto.ProtoGuideSvrLoginCheckReqS
			err := mapstructure.Decode(req.ReqData.Params, &reqLoginCheck)
			if err != nil {
				base.GLog.Error("ChannelUID[%s] Decode ProtoGuideSvrLoginCheckReqS failed! error[%s]",
					req.ChannelUID, err.Error())

				var failureInfo base.ProtoFailInfoS
				failureInfo.FailureReason = "Decode failed!"
				res.ResData.Params = failureInfo
				res.ReturnCode = proto.ProtoRetCodeError
				base.SendResponseToClient(c, &res)
				return
			}

			// token检测
			if req.CsrfToken != common.GConfig.GetToken() {
				base.GLog.Error("ChannelUID[%s] Token is invalid!", req.ChannelUID)

				var failureInfo base.ProtoFailInfoS
				failureInfo.FailureReason = "Token is invalid!"
				res.ResData.Params = failureInfo
				res.ReturnCode = proto.ProtoRetCodeError
				base.SendResponseToClient(c, &res)
				return
			}

			// 根据渠道获取配置的版本信息
			versionInfo := common.GConfig.GetVersionInfo(reqLoginCheck.ChannelName)
			if versionInfo == nil {
				base.GLog.Error("ChannelName[%s] is invalid!", reqLoginCheck.ChannelName)
				var failureInfo base.ProtoFailInfoS
				failureInfo.FailureReason = "Channel is invalid!"
				res.ResData.Params = failureInfo
				res.ReturnCode = proto.ProtoRetCodeError
				base.SendResponseToClient(c, &res)
				return
			} else {
				// 判断是否该升级
				if !versionInfo.VersionSet.Contains(reqLoginCheck.ClientVersion) {
					// 需要升级
					var versionUpdate proto.ProtoGuideSvrVersionUpdateS
					versionUpdate.ChannelName = reqLoginCheck.ChannelName
					versionUpdate.NewVersion = versionInfo.CurrVersions[len(versionInfo.CurrVersions)-1]
					versionUpdate.UpdateUrl = versionInfo.UpdateUrl

					res.ReturnCode = proto.ProtoRetCodeNeedUpdate
					res.ResData.Params = versionUpdate
					base.GLog.Warn("ChannelUID[%s] channenName[%s] clientVersion[%s] currVersion[%s] need Upgrade!!!",
						req.ChannelUID, reqLoginCheck.ChannelName, reqLoginCheck.ClientVersion, versionUpdate.NewVersion)
					base.SendResponseToClient(c, &res)
					return
				}
			}

			canLogin := false
			if maintainInfo == nil {
				// redis服务器出问题，就放开访问
				canLogin = true
			} else {
				if maintainInfo.MaintainFlag == proto.E_SERVER_STATUS_RUNING {
					// 服务运行中
					canLogin = true
				} else if maintainInfo.MaintainFlag == proto.E_SERVER_STATUS_MAINTAIN {
					// 服务维护中
					base.GLog.Warn("ChannelUID[%s] Server status is maintain", req.ChannelUID)
					canLogin = false
				} else if maintainInfo.MaintainFlag == proto.E_SERVER_STATUS_TESTING {
					// 服务灰度中
					var isWhiteUser bool = false
					if maintainInfo.WhiteFlag == 1 {
						// 白名单开启，判断用户是否在白名单中
						for index := range maintainInfo.WhiteList {
							if req.ChannelUID == maintainInfo.WhiteList[index] ||
								clientIP == maintainInfo.WhiteList[index] {
								isWhiteUser = true
							}
						}
					}

					if isWhiteUser {
						canLogin = true
						base.GLog.Warn("ChannelUID[%s] Server status is testing, You are a whilte list user, pass!", req.ChannelUID)
					} else {
						canLogin = false
						base.GLog.Warn("ChannelUID[%s] Server status is testing, You are not a whilte list user, can't pass!", req.ChannelUID)
					}
				} else {
					base.GLog.Error("ChannelUID[%s] maintain.Flag[%d] is unknown!", req.ChannelUID, maintainInfo.MaintainFlag)
				}
			}

			if canLogin {
				var loginInfo proto.ProtoGuideSvrLoginInfoS
				loginInfo.ServerIPs = make([]string, 0)
				loginInfo.ServerIPs = append(loginInfo.ServerIPs, "1.1.1.1")
				loginInfo.ServerIPs = append(loginInfo.ServerIPs, "3.3.3.3")
				loginInfo.ServerIPs = append(loginInfo.ServerIPs, "2.2.2.2")
				loginInfo.Port = 7045
				loginInfo.ClientInternetIP = clientIP

				res.ReturnCode = proto.ProtoRetCodeLoginOK
				res.ResData.Params = loginInfo
				base.GLog.Debug("Service running, ChannelUID[%s] loginInfo[%v]LoginCheck successed!", req.ChannelUID, loginInfo)
			} else {
				var maintenance proto.ProtoGuideSvrMaintenanceS
				maintenance.Bulletin = "server is in maintenance"
				maintenance.RemainingSeconds = int(maintainInfo.MainDeadLine - time.Now().UTC().Unix())
				res.ResData.Params = maintenance
				res.ReturnCode = proto.ProtoRetCodeMaintenance
				base.GLog.Debug("Service in maintenance, RemainingSeconds[%d] ChannelUID[%s] LoginCheck failed!",
					maintenance.RemainingSeconds, req.ChannelUID)
			}
			base.SendResponseToClient(c, &res)

		} else {
			base.GLog.Error("LoginCheck Body decode failed! error[%s]", err.Error())
		}
	}
	return
}
