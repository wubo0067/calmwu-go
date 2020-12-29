/*
 * @Author: calmwu
 * @Date: 2018-09-18 17:05:21
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-09-26 20:07:45
 */

package main

import (
	base "doyo-server-go/doyo-base-go"
	"doyo-server-go/doyo-base-go/consul_api"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var (
	ParamFlag      = flag.String("flag", "master/inst", "")
	ParamInstTypes = flag.String("types", "DoyoTa:DoyoTb:DoyoTc", "register types")
	ParamInstCount = flag.Int("count", 3, "type instance count")
	ParamInstType  = flag.String("type", "", "")
	ParamInstId    = flag.Int("id", 0, "")
)

func main() {
	flag.Parse()
	var instanceName string
	cmd := fmt.Sprintf("./%s", os.Args[0])
	env := os.Environ()
	procAttr := &os.ProcAttr{
		Env: env,
		Files: []*os.File{
			os.Stdin,
			os.Stdout,
			os.Stderr,
		},
	}

	log := base.NewSimpleLog(nil)

	if *ParamFlag == "master" {
		instanceName = "master"
		// 负责创建子进程
		types := strings.Split(*ParamInstTypes, ":")
		for index, typeName := range types {
			//fmt.Println(typeName)
			count := *ParamInstCount
			for i := 0; i < count; i++ {
				pid, err := os.StartProcess(cmd, []string{os.Args[0], "--type", typeName, "--id", strconv.Itoa(index*count + i)}, procAttr)
				if err != nil {
					log.Printf("StartProcess failed: %s\n", err.Error())
				} else {
					log.Printf("StartProcess successed, pid: %d\n", pid.Pid)
				}
			}
		}

		go func() {
			ticker := time.NewTicker(time.Second)
			ConsulClient, _ := consul_api.NewConsulClient("10.1.41.150")

			for {
				<-ticker.C
				// 定时去consul获取所有注册服务的状态
				services, meta, err := ConsulClient.Catalog().Services(nil)
				if err != nil {
					log.Printf("Consul catalog services failed: %s", err.Error())
				} else {
					if meta.LastIndex == 0 {
						log.Printf("Consul catlog services bad meta %v", meta)
						continue
					}

					if len(services) > 0 {
						log.Printf("Services %+v\n", services)

						for svrName, _ := range services {
							if svrName == "consul" {
								continue
							}

							checks, meta, err := ConsulClient.Health().Service(svrName, "", false, nil)
							if err != nil {
								log.Printf("Consul Health Service failed: %s", err.Error())
							} else {
								if meta.LastIndex == 0 {
									log.Printf("Consul Health Service bad meta %v", meta)
									continue
								}

								if len(checks) > 0 {
									for _, check := range checks {
										checkData := check.Checks[1]
										log.Printf("check %s %s %s\n", checkData.ServiceName,
											checkData.ServiceID, checkData.Status)
									}

								}
							}
						}
					} else {
						log.Printf("Services is empty\n")
					}
				}
			}
		}()
	} else {
		// 负责注册
		ConsulServName := fmt.Sprintf("%s-Svr", *ParamInstType)
		instanceName = fmt.Sprintf("%s-%d", ConsulServName, *ParamInstId)
		port := 8000 + *ParamInstId
		checkUrl := fmt.Sprintf("/%s/health/", instanceName)
		healthCheckUrl := fmt.Sprintf("http://%s:%d%s", "127.0.0.1", port, checkUrl)

		// http health check
		go func() {
			http.HandleFunc(checkUrl, func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), nil)
		}()

		// 注册
		ConsulClient, err := consul_api.NewConsulClient("10.1.41.150")
		if err != nil {
			log.Printf("NewConsulClient failed! reason[%s]", err.Error())
			return
		}

		err = consul_api.ConsulSvrReg(ConsulClient, ConsulServName, []string{ConsulServName}, instanceName, "127.0.0.1", port, healthCheckUrl)
		if err != nil {
			log.Printf("ConsulSvrReg failed: %s", err.Error())
		}
		log.Println(healthCheckUrl)
	}

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	<-sigchan
	if *ParamFlag == "master" {
		// 对进程组发送信号
		syscall.Kill(0, syscall.SIGTERM)
	}

	log.Printf("%s exit\n", instanceName)
}
