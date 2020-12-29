/*
 * @Author: calmwu
 * @Date: 2018-09-20 16:26:56
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-09-27 19:25:38
 */

package routersvr

import (
	base "doyo-server-go/doyo-base-go"
	"doyo-server-go/doyo-base-go/consul_api"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/consul/api"
	"github.com/pquerna/ffjson/ffjson"
)

type serviceTopicTableS map[string][]string

type routerSvrHealthSync struct {
	routerSvrTopicName string // 所有的routersvr使用同一个topic，所以topic和servType等同
	routerSvrInstName  string

	consulHealthCheckIP   string
	consulHealthCheckPort int
	consulHealthCheckPath string
	consulHealthCheckURL  string
	consulClient          *api.Client
	exitChan              chan struct{}
	exitWait              sync.WaitGroup
}

func newRouterSvrHealthCheck(routerSvrTopicName string, routerSvrID int, consulListenIP string, consulHealthCheckAddr string) (*routerSvrHealthSync, error) {
	var err error
	rc := new(routerSvrHealthSync)

	rc.routerSvrTopicName = routerSvrTopicName
	rc.routerSvrInstName = fmt.Sprintf("%s-%d", rc.routerSvrTopicName, routerSvrID)

	rc.consulHealthCheckIP = strings.Split(consulHealthCheckAddr, ":")[0]
	rc.consulHealthCheckPort, _ = strconv.Atoi(strings.Split(consulHealthCheckAddr, ":")[1])

	rc.consulHealthCheckPath = fmt.Sprintf("/%s/health/", rc.routerSvrInstName)
	rc.consulHealthCheckURL = fmt.Sprintf("http://%s%s", consulHealthCheckAddr, rc.consulHealthCheckPath)

	rc.consulClient, err = consul_api.NewConsulClient(consulListenIP)
	if err != nil {
		base.ZLog.Errorf("New consul client failed: %s", err.Error())
		return nil, err
	}

	base.ZLog.Debugw("Consul Info.", "consulIP", consulListenIP, "consulHealthCheckURL", rc.consulHealthCheckURL)

	rc.exitChan = make(chan struct{})

	return rc, nil
}

func (rc *routerSvrHealthSync) start(policyMgr *routerSvrRoutingPolicy) error {

	go func() {
		http.HandleFunc(rc.consulHealthCheckPath, func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		http.ListenAndServe(fmt.Sprintf("%s:%d", rc.consulHealthCheckIP, rc.consulHealthCheckPort), nil)
	}()

	// 注册到consul
	err := consul_api.ConsulSvrReg(rc.consulClient, rc.routerSvrTopicName, []string{rc.routerSvrTopicName}, rc.routerSvrInstName,
		rc.consulHealthCheckIP, rc.consulHealthCheckPort, rc.consulHealthCheckURL)
	if err != nil {
		base.ZLog.Errorf("ConsulSvrReg failed: %s", err.Error())
		return err
	}

	base.ZLog.Infof("ConsulSvrReg successed! consulHealthCheckURL: %s", rc.consulHealthCheckURL)

	// 启动health同步goroutine
	rc.exitWait.Add(1)
	go rc.healthSyncRoutine(policyMgr)

	return nil
}

func (rc *routerSvrHealthSync) stop() {
	close(rc.exitChan)
	rc.exitWait.Wait()
}

func (rc *routerSvrHealthSync) healthSyncRoutine(policyMgr *routerSvrRoutingPolicy) {
	base.ZLog.Debug("healthSyncRoutine running")

	defer func() {
		rc.exitWait.Done()
		if err := recover(); err != nil {
			// 回收panic，防止整个服务panic
			stackInfo := base.CallStack(1)
			base.ZLog.DPanicw("healthSyncRoutine panic recovered! err", err, "stack", string(stackInfo))
		}
	}()

	ticker := time.NewTicker(time.Second)
	prevServiceTopicTable := make(serviceTopicTableS) // 服务实例信息，服务类型---->实例
L:
	for {
		select {
		case <-rc.exitChan:
			base.ZLog.Info("healthSyncRoutine receive exit noitfy")
			break L
		case <-ticker.C:
			serviceTopicTable := make(serviceTopicTableS)

			servTypes, meta, err := rc.consulClient.Catalog().Services(nil)
			if err != nil {
				base.ZLog.Errorf("Consul catalog services failed: %s", err.Error())
			} else {
				if meta.LastIndex == 0 {
					base.ZLog.Errorf("Consul catalog services bad meta %v", meta)
					continue
				}

				if len(servTypes) > 0 {
					for servType := range servTypes {
						// 只检查Doyo开头的服务
						if strings.HasPrefix(servType, "Doyo") {

							// 忽略routerSvr自己
							if servType == "DoyoRouterSvr" {
								continue
							}

							// 只获取状态为passing的服务实例
							serviceEntrys, meta, err := rc.consulClient.Health().Service(servType, "", true, nil)
							if err != nil {
								base.ZLog.Errorf("Consul Health Services failed: %s", err.Error())
							} else {
								if meta.LastIndex == 0 {
									base.ZLog.Errorf("Consul Health Services bad meta %v", meta)
									continue
								}

								if len(serviceEntrys) > 0 {
									for _, serviceEntry := range serviceEntrys {
										serviceTopicTable[servType] = append(serviceTopicTable[servType], serviceEntry.Service.ID)
									}
								}
							}
						}
					}
				} else {
					base.ZLog.Errorf("servTypes is empty\n")
				}
			}

			if !cmp.Equal(prevServiceTopicTable, serviceTopicTable) {
				prevServiceTopicTable = serviceTopicTable
				base.ZLog.Info("serviceTopicTable is updated")
				// 这里通知路由策略模块
				jsonData, err := ffjson.Marshal(prevServiceTopicTable)
				if err != nil {
					base.ZLog.Errorf("ffjson Marshal prevServiceTopicTable failed! reason:%s", err.Error())
				} else {
					policyMgr.updateServiceTopicTable(string(jsonData))
				}

			}
		}
	}
	base.ZLog.Debug("healthSyncRoutine exit!")
}
