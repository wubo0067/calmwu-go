/*
 * @Author: calmwu
 * @Date: 2017-12-28 10:41:48
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-02-02 15:16:31
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
	ConsulServName          string = "SailCraft-FinanceSvr"
	ConsulHealthCheckUrlFmt string = "http://%s:%d/FinanceSvr/healthCheck"
)

func RegisterToConsul(listenIP string, listenPort int, consulServerIP string, healthCheckPort int) error {
	var err error
	ConsulClient, err = consul_api.NewConsulClient(consulServerIP)
	if err != nil {
		base.GLog.Error("registerToConsul failed! reason[%s]", err.Error())
		return err
	}

	servTags := []string{"FinanceSvr"}
	ConsulServInstanceID = fmt.Sprintf("FinanceSvr-%s:%d", listenIP, listenPort)
	healthCheckUrl := fmt.Sprintf(ConsulHealthCheckUrlFmt, listenIP, healthCheckPort)
	return consul_api.ConsulSvrReg(ConsulClient, ConsulServName, servTags, ConsulServInstanceID, listenIP, listenPort, healthCheckUrl)
}
