/*
 * @Author: calmwu
 * @Date: 2018-05-14 16:53:27
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-05-19 16:16:14
 * @Comment:
 */

package guidesvr

import (
	"encoding/json"
	"fmt"
	"sailcraft/base"
	"sailcraft/guidesvr_main/common"
	"sailcraft/guidesvr_main/proto"
	"sailcraft/sysconf"
	"time"

	"github.com/fzzy/radix/redis"
	"github.com/gin-gonic/gin"
)

func (guideSvr *GuideSvrModule) SetMaintainInfo(c *gin.Context) {
	req := base.UnpackRequest(c)
	if req == nil {
		base.GLog.Error("SetMaintainInfo read request failed!")
		return
	}

	var err error

	var res base.ProtoResponseS
	res.Version = req.Version
	res.EventId = req.EventId
	res.TimeStamp = time.Now().UTC().Unix()
	res.ReturnCode = 0
	res.ResData.InterfaceName = req.ReqData.InterfaceName

	var maintainInfo proto.MaintainInfoS
	err = base.MapstructUnPackByJsonTag(req.ReqData.Params, &maintainInfo)
	if err != nil {
		err = fmt.Errorf("Decode MaintainInfoS failed! reason[%s]", err.Error())
		base.GLog.Error(err.Error())

		res.ReturnCode = -1
		failInfo := new(base.ProtoFailInfoS)
		failInfo.FailureReason = err.Error()
		res.ResData.Params = failInfo
	} else {
		maintainKey := common.GConfig.GetGMKey()
		base.GLog.Debug("maintainKey[%s] maintainParams:%+v", maintainKey, maintainInfo.MaintainKey)

		maintainValStr, err := json.Marshal(maintainInfo.MaintainKey)
		if err != nil {
			err = fmt.Errorf("[%v] json marshal failed! reason[%s]", maintainInfo.MaintainKey, err.Error())
			base.GLog.Error(err.Error())

			res.ReturnCode = -1
			failInfo := new(base.ProtoFailInfoS)
			failInfo.FailureReason = err.Error()
			res.ResData.Params = failInfo
		} else {
			// 得到redis cluster的ip列表
			for index := range sysconf.GRedisConfig.ClusterRedisAddressList {
				addr := &sysconf.GRedisConfig.ClusterRedisAddressList[index]
				redisClient, err := redis.DialTimeout("tcp", *addr, 5*time.Second)
				if err != nil {
					base.GLog.Error("Connect to Redis[%s] failed! error[%s]", addr, err.Error())
					continue
				}
				defer redisClient.Close()

				err = redisClient.Cmd("SET", maintainKey, maintainValStr).Err
				if err != nil {
					base.GLog.Error("SET cmd key[%s] value[%s] addr[%s] failed! reason[%s]",
						maintainKey, maintainValStr, *addr, err.Error())
					continue
				} else {
					base.GLog.Debug("SET cmd key[%s] value[%s] addr[%s] successed!", maintainKey, maintainValStr, *addr)
				}
			}
		}

	}

	// 返回给客户端
	base.SendResponse(c, &res)
}
