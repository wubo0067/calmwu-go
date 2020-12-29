/*
 * @Author: calmwu
 * @Date: 2017-11-16 17:04:28
 * @Last Modified by:   calmwu
 * @Last Modified time: 2017-11-16 17:04:28
 * @Comment:
 */

package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/hashicorp/consul/api"
)

// ./server_query --consul-ip=10.135.138.179
// 传入服务名，得到状态为passing的服务实例
// 771  2017-11-17 16:01:43 curl 127.0.0.1:8500/v1/agent/services
// 781  2017-11-17 16:08:26 curl 127.0.0.1:8500/v1/catalog/service/hello
// 786  2017-11-17 16:12:43 curl 127.0.0.1:8500/v1/health/checks/hello
// 792  2017-11-17 16:23:02 curl 127.0.0.1:8500/v1/health/state/passing

var (
	paramConsulIP = flag.String("consul-ip", "127.0.0.1", "")
)

func NewConsulClient(consulIP string) (*api.Client, error) {
	conf := api.DefaultConfig()
	if consulIP != "127.0.0.1" {
		conf.Address = fmt.Sprintf("%s:8500", consulIP)
	}
	return api.NewClient(conf)
}

func QueryHealthServices(client *api.Client) {
	ticker := time.NewTicker(3 * time.Second)

	for {
		select {
		case <-ticker.C:
			// 这个拿不到服务的端口和ip
			// checkServs, _, err := client.Health().State(api.HealthPassing, nil)
			// if err != nil {
			// 	fmt.Println(err.Error())
			// } else {
			// 	for index := range checkServs {
			// 		healthCheck := checkServs[index]
			// 		fmt.Printf("healthCheck:%+v\n", healthCheck)
			// 	}
			// 	fmt.Println("\n")
			// }
			// 得到所有服务列表
			servCatalog, _, err := client.Catalog().Services(nil)
			if err != nil {
				fmt.Println(err.Error)
			} else {
				for servName, _ := range servCatalog {
					// 获得健康的服务实例
					servEntrys, _, err := client.Health().Service(servName, "", true, nil)
					if err != nil {
						fmt.Println(err.Error())
					} else {
						for index := range servEntrys {
							servEntry := servEntrys[index]
							fmt.Printf("+++ServName[%s]\n\tNode:%+v\n\tService:%+v\n\tChecks:%+v\n",
								servName, servEntry.Node, servEntry.Service, servEntry.Checks)
						}
					}
				}
				fmt.Println("\n")
			}
		}
	}
}

func main() {
	flag.Parse()

	client, err := NewConsulClient(*paramConsulIP)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// 查询健康服务
	go QueryHealthServices(client)

	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
		}
	}
}
