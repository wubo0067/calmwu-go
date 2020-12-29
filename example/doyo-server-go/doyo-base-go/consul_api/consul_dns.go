/*
 * @Author: calmwu
 * @Date: 2017-11-21 14:51:29
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-11-30 17:25:29
 * @Comment:
 */

package consul_api

import (
	base "doyo-server-go/doyo-base-go"
	"fmt"
	"time"

	"github.com/hashicorp/consul/api"
)

type ConsulServInstS struct {
	ID   string
	IP   string
	Port int
}

// 通过服务名，获取健康的服务实例列表
func ConsulServDns(client *api.Client, servName string) ([]*ConsulServInstS, error) {
	//servCatalog, _, err := client.Catalog().Services(nil)
	servEntrys, _, err := client.Health().Service(servName, "", true, nil)
	if err != nil {
		base.ZLog.Errorf(err.Error())
		return nil, err
	} else {
		consulServInstSlice := make([]*ConsulServInstS, len(servEntrys))
		for index := range servEntrys {
			consulServInstSlice[index] = &ConsulServInstS{
				ID:   servEntrys[index].Service.ID,
				IP:   servEntrys[index].Service.Address,
				Port: servEntrys[index].Service.Port,
			}
		}
		return consulServInstSlice, nil
	}
	//return nil, fmt.Errorf("Query Server[%s] dns info failed! Unknown Error!", servName)
}

func PostRequstByConsulDns(uin uint64, interfaceName string, realReq interface{}, client *api.Client, svrName string) (*base.ProtoResponseS, error) {
	if client == nil {
		err := fmt.Errorf("Consul Client is Nil!")
		base.ZLog.Errorf(err.Error())
		return nil, err
	}
	consulSvrName := fmt.Sprintf("SailCraft-%s", svrName)
	servInsts, err := ConsulServDns(client, consulSvrName)
	if err != nil {
		base.ZLog.Errorf("Query %s insts from consul failed! reason[%s]", consulSvrName, err.Error())
		return nil, err
	}

	if len(servInsts) == 0 {
		err := fmt.Errorf("Query %s insts count is empty!", consulSvrName)
		base.ZLog.Error(err.Error())
		return nil, err
	}

	url := fmt.Sprintf("http://%s:%d/sailcraft/api/v1/%s/%s",
		servInsts[0].IP, servInsts[0].Port, svrName, interfaceName)
	base.ZLog.Debugf("Dispatch url[%s]", url)

	req := base.ProtoRequestS{
		ProtoRequestHeadS: base.ProtoRequestHeadS{
			Version:    1,
			EventId:    998,
			TimeStamp:  time.Now().Unix(),
			ChannelUID: "ConsulDns",
			Uin:        int(uin),
			CsrfToken:  "02cf14994a3be74301657dbcd9c0a189",
		},
		ReqData: base.ProtoData{
			InterfaceName: interfaceName,
			Params:        realReq,
		},
	}
	return base.PostRequest(url, &req)
}

func PostBaseRequstByConsulDns(interfaceName string, req *base.ProtoRequestS, client *api.Client, svrName string) (*base.ProtoResponseS, error) {
	if client == nil {
		err := fmt.Errorf("Consul Client is Nil!")
		base.ZLog.Errorf(err.Error())
		return nil, err
	}
	consulSvrName := fmt.Sprintf("SailCraft-%s", svrName)
	servInsts, err := ConsulServDns(client, consulSvrName)
	if err != nil {
		base.ZLog.Errorf("Query %s insts from consul failed! reason[%s]", consulSvrName, err.Error())
		return nil, err
	}

	if len(servInsts) == 0 {
		err := fmt.Errorf("Query %s insts count is empty!", consulSvrName)
		base.ZLog.Errorf(err.Error())
		return nil, err
	}

	url := fmt.Sprintf("http://%s:%d/sailcraft/api/v1/%s/%s",
		servInsts[0].IP, servInsts[0].Port, svrName, interfaceName)
	base.ZLog.Debugf("Dispatch url[%s]", url)

	return base.PostRequest(url, req)
}
