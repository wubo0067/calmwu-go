/*
 * @Author: calmwu
 * @Date: 2017-11-24 14:52:16
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-09-25 19:55:42
 * @Comment:
 */

package consul-api

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
		HTTP:                           healthCheckUrl,
		Interval:                       "3s",
		Timeout:                        "3s",
		DeregisterCriticalServiceAfter: "600s", //check失败后多久从consul集群中删除，这样页面就看不到了
	}

	err := client.Agent().ServiceRegister(regInfo)
	if err != nil {
		return err
	}

	return nil
}
