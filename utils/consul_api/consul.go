/*
 * @Author: calmwu
 * @Date: 2017-11-21 14:57:39
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-09-18 19:36:15
 * @Comment:
 */

package consulapi

import (
	"fmt"

	"github.com/hashicorp/consul/api"
)

func NewConsulClient(consulIP string) (*api.Client, error) {
	conf := api.DefaultConfig()
	if consulIP != "127.0.0.1" {
		conf.Address = fmt.Sprintf("%s:8500", consulIP)
	}
	return api.NewClient(conf)
}
