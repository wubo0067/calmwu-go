/*
 * @Author: calmwu
 * @Date: 2017-12-28 10:41:48
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-02-02 15:16:31
 */

package utils

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
	CONSUL_URL_FORMAT    = "http://%s:%d%s"
	CONSUL_RELATIVE_PATH = "/FleetSvr/Index"
	CONSUL_SERVER_NAME   = "SailCraft-FleetSvr"
)

func RegisterToConsul(listenIP string, listenPort int, consulServerIP string, healthCheckPort int) error {
	var err error
	ConsulClient, err = consul_api.NewConsulClient(consulServerIP)
	if err != nil {
		base.GLog.Error("registerToConsul failed! reason[%s]", err.Error())
		return err
	}

	servTags := []string{"FleetSvr"}
	ConsulServInstanceID = fmt.Sprintf("FleetSvr-%s:%d", listenIP, listenPort)
	healthCheckUrl := fmt.Sprintf(CONSUL_URL_FORMAT, listenIP, healthCheckPort, CONSUL_RELATIVE_PATH)

	return consul_api.ConsulSvrReg(ConsulClient, CONSUL_SERVER_NAME, servTags, ConsulServInstanceID, listenIP, listenPort, healthCheckUrl)
}
