/*
 * @Author: calmwu
 * @Date: 2018-05-18 20:40:30
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-05-18 20:48:25
 */

package common

import (
	"fmt"
	"sailcraft/base"
	"sailcraft/base/consul_api"
	"time"
	//financesvr_proto "sailcraft/financesvr_main/proto"
)

func SendReqToFinanceSvr(interfaceName string, realReq interface{}) (*base.ProtoResponseS, error) {
	// 通过consul的dns获取健康的服务实例
	servInsts, err := consul_api.ConsulServDns(ConsulClient, "SailCraft-FinanceSvr")
	if err != nil {
		base.GLog.Error("Query SailCraft-FinanceSvr insts from consul failed! reason[%s]", err.Error())
		return nil, err
	}

	if len(servInsts) == 0 {
		err := fmt.Errorf("Query SailCraft-FinanceSvr insts count is empty!")
		base.GLog.Error(err.Error())
		return nil, err
	}

	url := fmt.Sprintf("http://%s:%d/sailcraft/api/v1/FinanceSvr/%s",
		servInsts[0].IP, servInsts[0].Port, interfaceName)
	base.GLog.Debug("Dispatch url[%s]", url)

	req := base.ProtoRequestS{
		ProtoRequestHeadS: base.ProtoRequestHeadS{
			Version:    1,
			EventId:    998,
			TimeStamp:  time.Now().Unix(),
			ChannelUID: "OMSvr",
			Uin:        10000000,
			CsrfToken:  "02cf14994a3be74301657dbcd9c0a189",
		},
		ReqData: base.ProtoData{
			InterfaceName: interfaceName,
			Params:        realReq,
		},
	}
	return base.PostRequest(url, &req)
}
