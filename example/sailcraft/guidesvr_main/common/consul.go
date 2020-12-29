/*
 * @Author: calmwu
 * @Date: 2017-12-28 10:41:48
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-12-28 10:42:40
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
	ConsulServName          string = "SailCraft-GuideSvr"
	ConsulHealthCheckUrlFmt string = "http://%s:%d/GuideSvr/healthCheck"
)

func RegisterToConsul(listenIP string, listenPort int, consulServerIP string, healthCheckPort int) error {
	var err error
	ConsulClient, err = consul_api.NewConsulClient(consulServerIP)
	if err != nil {
		base.GLog.Error("registerToConsul failed! reason[%s]", err.Error())
		return err
	}

	servTags := []string{"GuideSvr"}
	ConsulServInstanceID = fmt.Sprintf("GuideSvr-%s:%d", listenIP, listenPort)
	healthCheckUrl := fmt.Sprintf(ConsulHealthCheckUrlFmt, listenIP, healthCheckPort)
	return consul_api.ConsulSvrReg(ConsulClient, ConsulServName, servTags, ConsulServInstanceID, listenIP, listenPort, healthCheckUrl)
}
