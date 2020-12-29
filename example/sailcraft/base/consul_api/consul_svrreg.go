/*
 * @Author: calmwu
 * @Date: 2017-11-24 14:52:16
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-11-24 14:56:31
 * @Comment:
 */

package consul_api

import (
	"errors"

	"github.com/hashicorp/consul/api"
)

// func defHealthCheck(checkIP string, checkPort int) {
// 	onHealthCheck := func(w http.ResponseWriter, r *http.Request) {
// 		w.WriteHeader(http.StatusOK)
// 	}
// 	http.HandleFunc("/healthCheck", onHealthCheck)
// 	http.ListenAndServe(fmt.Sprintf("%s:%d", checkIP, checkPort), nil)
// }

/*
这个配置和配置文件中提供的一样，example
{
    "service" : {
        "name" : "hello",
        "tags": ["master"],
        "address" : "10.186.40.75",
        "port" : 8990,
        "checks" : [
            {
                "http" : "http://10.186.40.75:8990/health",
                "interval": "10s"
            }
        ]
    }
}
*/

func ConsulSvrReg(client *api.Client, servName string, servTags []string, servInstName string,
	servInstListenIP string, servInstListenPort int, healthCheckUrl string) error {
	if client == nil {
		return errors.New("consul client is nil")
	}

	regInfo := new(api.AgentServiceRegistration)
	regInfo.Name = servName
	regInfo.Tags = servTags
	regInfo.Address = servInstListenIP
	regInfo.Port = servInstListenPort
	regInfo.ID = servInstName
	regInfo.Check = &api.AgentServiceCheck{
		HTTP:     healthCheckUrl,
		Interval: "30s",
		Timeout:  "30s",
		DeregisterCriticalServiceAfter: "7200s", //check失败后30秒删除本服务，有时候端口变了该项目的确不需要继续显示
	}

	err := client.Agent().ServiceRegister(regInfo)
	if err != nil {
		return err
	}

	return nil
}
