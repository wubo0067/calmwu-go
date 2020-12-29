/*
 * @Author: calmwu
 * @Date: 2018-01-10 16:34:48
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-01-10 16:35:28
 * @Comment:
 */

package common

import (
	"fmt"
	"sailcraft/base"
	"sailcraft/base/consul_api"

	"github.com/hashicorp/consul/api"
)

var (
	ConsulServInstanceID string
	ConsulClient         *api.Client
)

const (
	ConsulServName          string = "SailCraft-CassandraSvr"
	ConsulHealthCheckUrlFmt string = "http://%s:%d/CassandraSvr/healthCheck"
)

func RegisterToConsul(listenIP string, listenPort int, consulServerIP string, healthCheckPort int) error {
	var err error
	ConsulClient, err = consul_api.NewConsulClient(consulServerIP)
	if err != nil {
		base.GLog.Error("registerToConsul failed! reason[%s]", err.Error())
		return err
	}

	servTags := []string{"CssSvr"}
	ConsulServInstanceID = fmt.Sprintf("CassandraSvr-%s:%d", listenIP, listenPort)
	healthCheckUrl := fmt.Sprintf(ConsulHealthCheckUrlFmt, listenIP, healthCheckPort)
	return consul_api.ConsulSvrReg(ConsulClient, ConsulServName, servTags, ConsulServInstanceID, listenIP, listenPort, healthCheckUrl)
}
