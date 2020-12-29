/*
 * @Author: calmwu
 * @Date: 2018-09-26 19:37:42
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-09-29 11:07:25
 */

package routersvr

import (
	"bytes"
	base "doyo-server-go/doyo-base-go"
	"doyo-server-go/doyo-routersvr-go/proto"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/OneOfOne/xxhash"
	"github.com/emirpasic/gods/sets/hashset"
	"github.com/pquerna/ffjson/ffjson"
)

// 查询future
type queryDestServiceTopicFutureResult struct {
	err                  error
	destServiceTopicInfo interface{} // 结果，有多种形式，对应不同的转发策略
}

type queryDestServiceTopicFuture struct {
	destServiceTopicChan chan *queryDestServiceTopicFutureResult //结果返回通道
	ToServType           string
	DispatchPolicy       proto.RouterSvrDispatchPolicy
	DispatchPolicyArg    string
	MsgID                string
}

func (qf *queryDestServiceTopicFuture) init() {
	qf.destServiceTopicChan = make(chan *queryDestServiceTopicFutureResult)
}

func (qf *queryDestServiceTopicFuture) response(destServiceTopicInfo interface{}, err error) {
	result := &queryDestServiceTopicFutureResult{
		err:                  err,
		destServiceTopicInfo: destServiceTopicInfo,
	}
	qf.destServiceTopicChan <- result
	close(qf.destServiceTopicChan)
}

func (qf *queryDestServiceTopicFuture) waitResp(timeout time.Duration) (interface{}, error) {
	select {
	case result, ok := <-qf.destServiceTopicChan:
		if ok {
			return result.destServiceTopicInfo, result.err
		}
	case <-time.After(timeout):
		return "nil", fmt.Errorf("MsgID[%s] query dest service topic timeout", qf.MsgID)
	}
	return "nil", fmt.Errorf("MsgID[%s] query dest service failed", qf.MsgID)
}

func (qf *queryDestServiceTopicFuture) String() string {
	content := fmt.Sprintf("MsgID[%s] ToServType[%s] DispatchPolicy[%s] DispatchPolicyArg[%s]",
		qf.MsgID, qf.ToServType, qf.DispatchPolicy.String(), qf.DispatchPolicyArg)
	return content
}

type serviceInstS struct {
	topicName string // topic名
	load      uint64 // 权重
	active    bool   // 是否有效
}

type dserviceTypeInfoS struct {
	typeName        string                   // 服务名
	roundRobinIndex uint64                   // roundrobin计算值
	servInstMap     map[string]*serviceInstS // 服务实例的运行参数
}

type routerSvrRoutingPolicy struct {
	healthServiceTopicTable serviceTopicTableS                // 健康的服务实例，按服务类型
	serviceStatusInfoTable  map[string]*dserviceTypeInfoS     // 运行状态信息
	updateChan              chan string                       // healthServiceTopicTable 更新通道
	queryChan               chan *queryDestServiceTopicFuture // 查询通道
	exitChan                chan struct{}
	exitWait                sync.WaitGroup
}

func newRouterSvrPolicyMgr() *routerSvrRoutingPolicy {
	routingPolicy := &routerSvrRoutingPolicy{
		healthServiceTopicTable: make(serviceTopicTableS),
		serviceStatusInfoTable:  make(map[string]*dserviceTypeInfoS),
		updateChan:              make(chan string, 8),
		queryChan:               make(chan *queryDestServiceTopicFuture, 1024),
		exitChan:                make(chan struct{}),
	}

	routingPolicy.exitWait.Add(1)
	go routingPolicy.routingPolicyRoutine()

	return routingPolicy
}

func (rp *routerSvrRoutingPolicy) String() string {
	var content bytes.Buffer
	content.WriteString(fmt.Sprintf("healthServiceTopicTable:%+v, ", rp.healthServiceTopicTable))
	content.WriteString("serviceStatusInfoTable{")
	for servType, serviceTypeInfo := range rp.serviceStatusInfoTable {
		content.WriteString(fmt.Sprintf("%s:[roundRobinIndex:%d ", servType, serviceTypeInfo.roundRobinIndex))
		for _, serviceInstInfo := range serviceTypeInfo.servInstMap {
			content.WriteString(fmt.Sprintf("%+v", *serviceInstInfo))
		}
		content.WriteString("] ")
	}
	content.WriteString("}")
	return content.String()
}

func (rp *routerSvrRoutingPolicy) stop() {
	close(rp.exitChan)
	rp.exitWait.Wait()
}

func (rp *routerSvrRoutingPolicy) updateServiceTopicTable(jsonData string) {
	rp.updateChan <- jsonData
}

func (rp *routerSvrRoutingPolicy) queryDestServiceTopics(msg *proto.RouterSvrDispatchMsg) (*queryDestServiceTopicFuture, error) {
	queryFuture := &queryDestServiceTopicFuture{
		ToServType:        msg.ToServType,
		DispatchPolicy:    msg.DispatchPolicy,
		DispatchPolicyArg: msg.DispatchPolicyArg,
		MsgID:             msg.MessageID,
	}
	queryFuture.init()
	rp.queryChan <- queryFuture
	return queryFuture, nil
}

func (rp *routerSvrRoutingPolicy) routingPolicyRoutine() {
	base.ZLog.Debug("routingPolicyRoutine running")

	defer func() {
		rp.exitWait.Done()
		if err := recover(); err != nil {
			// 回收panic，防止整个服务panic
			stackInfo := base.CallStack(1)
			base.ZLog.DPanicw("routingPolicyRoutine panic recovered! err", err, "stack", string(stackInfo))
		}
	}()

	loadTicker := time.NewTicker(60 * time.Second)

L:
	for {
		select {
		case <-rp.exitChan:
			base.ZLog.Info("routingPolicyRoutine receive exit noitfy")
			break L
		case queryFuture, ok := <-rp.queryChan:
			if ok {
				rp.processQueryFuture(queryFuture)
			}
		case healthInfo, ok := <-rp.updateChan:
			if ok {
				rp.merge(healthInfo)
			}
		case <-loadTicker.C:
			// 重置所有服务的负载
			for _, serviceTypeInfo := range rp.serviceStatusInfoTable {
				for _, serviceInstInfo := range serviceTypeInfo.servInstMap {
					serviceInstInfo.load = 0
				}
			}
		}
	}
	base.ZLog.Debug("routingPolicyRoutine exit!")
}

func (rp *routerSvrRoutingPolicy) processQueryFuture(queryFuture *queryDestServiceTopicFuture) {
	defer func() {
		if err := recover(); err != nil {
			// 回收panic，防止整个服务panic
			stackInfo := base.CallStack(1)
			base.ZLog.DPanicw("processQueryFuture panic recovered! err", err, "stack", string(stackInfo))
		}
	}()

	base.ZLog.Debugf("queryFuture: %s", queryFuture.String())

	toServType := queryFuture.ToServType
	var topicLst []string
	var exist bool = false

	// 查询目的服务类型服务
	if topicLst, exist = rp.healthServiceTopicTable[toServType]; !exist || len(topicLst) == 0 {
		base.ZLog.Errorf("ToServType[%s] is invalid!", toServType)
		queryFuture.response("nil", ErrServTypeInvalid)
		return
	}

	switch queryFuture.DispatchPolicy {
	case proto.RouterSvrDispatchPolicyBrd:
		// Broadcast
		queryFuture.response(topicLst, nil)
		rp.updateServiceLoad(toServType, topicLst)
		base.ZLog.Debugf("MsgID[%s] DestServiceTopics:%v", queryFuture.MsgID, topicLst)

	case proto.RouterSvrDispatchPolicyRandom:
		// Random
		selectPos := rand.Uint32() % uint32(len(topicLst))
		destServiceTopic := topicLst[selectPos]
		queryFuture.response(destServiceTopic, nil)
		rp.updateServiceLoad(toServType, []string{destServiceTopic})
		base.ZLog.Debugf("MsgID[%s] DestServiceTopics:%s", queryFuture.MsgID, destServiceTopic)

	case proto.RouterSvrDispatchPolicyDH:
		// Dest Hash
		//hashNum := base.HashStr2Uint32(queryFuture.DispatchPolicyArg)
		hashNum := xxhash.ChecksumString32(queryFuture.DispatchPolicyArg)
		selectPos := hashNum % uint32(len(topicLst))
		destServiceTopic := topicLst[selectPos]
		queryFuture.response(destServiceTopic, nil)
		rp.updateServiceLoad(toServType, []string{destServiceTopic})
		base.ZLog.Debugf("MsgID[%s] DestServiceTopics:%s", queryFuture.MsgID, destServiceTopic)

	case proto.RouterSvrDispatchPolicyRR:
		// RoundRobin
		rrNum := rp.serviceStatusInfoTable[toServType].roundRobinIndex
		rp.serviceStatusInfoTable[toServType].roundRobinIndex++
		selectPos := rrNum % uint64(len(topicLst))
		destServiceTopic := topicLst[selectPos]
		queryFuture.response(destServiceTopic, nil)
		rp.updateServiceLoad(toServType, []string{destServiceTopic})
		base.ZLog.Debugf("MsgID[%s] DestServiceTopics:%s", queryFuture.MsgID, destServiceTopic)

	case proto.RouterSvrDispatchPolicyLoad:
		// By Load
		destServiceTopic := rp.findLowestLoadService(toServType, topicLst)
		queryFuture.response(destServiceTopic, nil)
		rp.updateServiceLoad(toServType, []string{destServiceTopic})
		base.ZLog.Debugf("MsgID[%s] DestServiceTopics:%s", queryFuture.MsgID, destServiceTopic)

	default:
		base.ZLog.Errorf("DispatchPolicy[%s] is not support", queryFuture.DispatchPolicy.String())
		queryFuture.response("nil", ErrRoutingPolicyNotSupport)
	}
}

func (rp *routerSvrRoutingPolicy) updateServiceLoad(servType string, topicLst []string) {
	if serviceTypeInfo, typeExist := rp.serviceStatusInfoTable[servType]; typeExist {
		for index := range topicLst {
			if serviceInstInfo, instExist := serviceTypeInfo.servInstMap[topicLst[index]]; instExist {
				serviceInstInfo.load++
			}
		}
	}
}

func (rp *routerSvrRoutingPolicy) findLowestLoadService(servType string, topicLst []string) string {
	var lowestLoad uint64
	var serviceInst *serviceInstS

	if serviceTypeInfo, typeExist := rp.serviceStatusInfoTable[servType]; typeExist {
		for index := range topicLst {
			if serviceInstInfo, instExist := serviceTypeInfo.servInstMap[topicLst[index]]; instExist {
				if lowestLoad == 0 {
					lowestLoad = serviceInstInfo.load
					serviceInst = serviceInstInfo
				} else {
					if lowestLoad > serviceInstInfo.load {
						lowestLoad = serviceInstInfo.load
						serviceInst = serviceInstInfo
					}
				}
			}
		}
	}

	if serviceInst != nil {
		return serviceInst.topicName
	}
	return topicLst[0]
}

func (rp *routerSvrRoutingPolicy) merge(healthInfo string) {
	defer func() {
		if err := recover(); err != nil {
			// 回收panic，防止整个服务panic
			stackInfo := base.CallStack(1)
			base.ZLog.DPanicw("merge panic recovered! err", err, "stack", string(stackInfo))
		}
	}()

	rp.healthServiceTopicTable = make(serviceTopicTableS)
	err := ffjson.Unmarshal([]byte(healthInfo), &rp.healthServiceTopicTable)
	if err != nil {
		base.ZLog.Errorf("ffjson Unmarshal healthInfo failed! reason:%s", err.Error())
	} else {
		// 用于反查
		updateServTypeSet := hashset.New()

		for servType, servTopicLst := range rp.healthServiceTopicTable {
			updateServTypeSet.Add(servType)

			if _, exist := rp.serviceStatusInfoTable[servType]; !exist {
				// 新增一个服务类型
				servTypeInfo := &dserviceTypeInfoS{
					typeName:        servType,
					roundRobinIndex: 0,
					servInstMap:     make(map[string]*serviceInstS),
				}
				// 添加的服务实例
				for _, serviceTopic := range servTopicLst {
					servTypeInfo.servInstMap[serviceTopic] = &serviceInstS{
						topicName: serviceTopic,
						load:      0,
						active:    true,
					}
				}
				rp.serviceStatusInfoTable[servType] = servTypeInfo
			} else {
				// 更新所有的服务实例
				servTypeInfo := rp.serviceStatusInfoTable[servType]
				// 重置所有状态
				for _, serviceInst := range servTypeInfo.servInstMap {
					serviceInst.active = false
				}

				for _, serviceTopic := range servTopicLst {
					if _, exists := servTypeInfo.servInstMap[serviceTopic]; !exists {
						// 新增
						servTypeInfo.servInstMap[serviceTopic] = &serviceInstS{
							topicName: serviceTopic,
							load:      0,
							active:    true,
						}
					} else {
						servTypeInfo.servInstMap[serviceTopic].active = true
					}
				}
			}
		}

		for servType, servTypeInfo := range rp.serviceStatusInfoTable {
			if !updateServTypeSet.Contains(servType) {
				// 如果现在的服务类型不在更新信息里，说明这个服务下所有实例全部失效
				for _, serviceInst := range servTypeInfo.servInstMap {
					serviceInst.active = false
				}
			}
		}

		base.ZLog.Debugf("routingPolicy:%s", rp.String())
	}
}
