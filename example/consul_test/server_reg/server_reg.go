/*
 * @Author: calmwu
 * @Date: 2017-11-16 15:51:29
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-11-16 16:59:02
 * @Comment:
 */

package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/consul/api"
)

/*
./server_reg --ip=10.186.40.75
./server_reg --ip=10.135.138.179 --consul-ip=10.135.138.179
*/

var (
	paramName       = flag.String("name", "radium", "")
	paramIP         = flag.String("ip", "", "")
	paramPort       = flag.Int("port", 1207, "listen port")
	paramHealthPort = flag.Int("health-port", 2207, "health listen port")
	paramConsulIP   = flag.String("consul-ip", "127.0.0.1", "")
)

func healthCheck(checkIP string, checkPort int) {
	onHealthCheck := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	http.HandleFunc("/healthCheck", onHealthCheck)
	http.ListenAndServe(fmt.Sprintf("%s:%d", checkIP, checkPort), nil)
}

func RegisterServerToConsul(client *api.Client, servName string, servTags []string,
	servInstName string, servInstListenIP string, servInstListenPort int,
	servInstHealthCheckPort int) error {
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
		HTTP:     fmt.Sprintf("http://%s:%d/healthCheck", servInstListenIP, servInstHealthCheckPort),
		Interval: "3s",
		Timeout:  "3s",
		DeregisterCriticalServiceAfter: "30s", //check失败后30秒删除本服务
	}

	err := client.Agent().ServiceRegister(regInfo)
	if err != nil {
		return err
	}

	// 启动health check协程
	go healthCheck(servInstListenIP, servInstHealthCheckPort)
	return nil
}

func NewConsulClient(consulIP string) (*api.Client, error) {
	conf := api.DefaultConfig()
	if consulIP != "127.0.0.1" {
		conf.Address = fmt.Sprintf("%s:8500", consulIP)
	}
	return api.NewClient(conf)
}

func main() {
	flag.Parse()

	servTags := []string{"radium"}
	servInstName := fmt.Sprintf("%s-%s:%d", *paramName, *paramIP, *paramPort)

	client, err := NewConsulClient(*paramConsulIP)
	if err != nil {
		return
	}

	err = RegisterServerToConsul(client, *paramName, servTags, servInstName, *paramIP, *paramPort, *paramHealthPort)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
		}
	}
}
