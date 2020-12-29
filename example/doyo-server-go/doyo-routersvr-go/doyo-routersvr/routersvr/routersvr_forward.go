/*
 * @Author: calmwu
 * @Date: 2018-09-20 16:26:11
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-09-29 11:25:26
 */

package routersvr

import (
	base "doyo-server-go/doyo-base-go"
	"doyo-server-go/doyo-routersvr-go/doyo-kafka-go/doyokafka"
	"doyo-server-go/doyo-routersvr-go/proto"
	"sync"
	"time"

	"github.com/pquerna/ffjson/ffjson"
)

const (
	queryRoutingPolicyTryTimes     = 3
	queryRoutimePolicyWaitDuration = 5 * time.Second
)

// 消息转发
type routerSvrForwardMsg struct {
	isEof bool                         // 是否是结束消息
	kdata *doyokafka.DoyoKafkaReadData // kafka数据
}

type routerSvrForward struct {
	routineCount int
	exitWait     sync.WaitGroup
	msgChan      chan *routerSvrForwardMsg
}

func newRourterSvrForward(drc int, policyMgr *routerSvrRoutingPolicy, doyoKfk *doyokafka.KafkaModule) (*routerSvrForward, error) {
	forward := &routerSvrForward{
		routineCount: drc,
		msgChan:      make(chan *routerSvrForwardMsg, 1024),
	}

	for i := 0; i < drc; i++ {
		go forward.forwardMsgRoutine(policyMgr, doyoKfk)
		forward.exitWait.Add(1)
	}
	return forward, nil
}

func (rf *routerSvrForward) forwardData(data *doyokafka.DoyoKafkaReadData) {
	rf.msgChan <- &routerSvrForwardMsg{
		isEof: false,
		kdata: data,
	}
}

func (rf *routerSvrForward) stop() {
	for index := 0; index < rf.routineCount; index++ {
		rf.msgChan <- &routerSvrForwardMsg{
			isEof: true,
		}
	}
	rf.exitWait.Wait()
}

func (rf *routerSvrForward) forwardMsgRoutine(policyMgr *routerSvrRoutingPolicy, doyoKfk *doyokafka.KafkaModule) {
	base.ZLog.Debug("forwardMsgRoutine running")

	defer func() {
		rf.exitWait.Done()
		if err := recover(); err != nil {
			// 回收panic，防止整个服务panic
			stackInfo := base.CallStack(1)
			base.ZLog.DPanicw("forwardMsgRoutine panic recovered! err", err, "stack", string(stackInfo))
		}
	}()

	var dispatchMsg proto.RouterSvrDispatchMsg
L:
	for {
		select {
		case msg, ok := <-rf.msgChan:
			if ok {
				if msg.isEof {
					base.ZLog.Info("forwardMsgRoutine receive exit noitfy")
					break L
				}
				// 解包
				err := ffjson.Unmarshal(msg.kdata.Data(), &dispatchMsg)
				if err != nil {
					base.ZLog.Errorf("ffjson Unmarshal failed! %s", err.Error())
				} else {
					base.ZLog.Debugw("Forward", "msgid", dispatchMsg.MessageID, "dispatchPolicy", dispatchMsg.DispatchPolicy.String())
					rf.processMsg(&dispatchMsg, policyMgr, doyoKfk)
				}
			} else {
				base.ZLog.Errorf("forwardMsgRoutine read from msgChan failed!")
			}
		}
	}

	base.ZLog.Debug("forwardMsgRoutine exit!")
}

func (rf *routerSvrForward) processMsg(msg *proto.RouterSvrDispatchMsg, policyMgr *routerSvrRoutingPolicy, doyoKfk *doyokafka.KafkaModule) {
	defer func() {
		if err := recover(); err != nil {
			// 回收panic，防止整个服务panic
			stackInfo := base.CallStack(1)
			base.ZLog.DPanicw("forwardMsgRoutine panic recovered! err", err, "stack", string(stackInfo))
		}
	}()

	// 超时重试
	queryOk := false
	var err error
	var iDestServiceTopics interface{}

	for index := 0; index < queryRoutingPolicyTryTimes; index++ {
		queryFuture, _ := policyMgr.queryDestServiceTopics(msg)
		iDestServiceTopics, err = queryFuture.waitResp(queryRoutimePolicyWaitDuration)
		if err != nil {
			if err == ErrQueryRoutinePolicyTimeOut {
				base.ZLog.Warnf("MsgID[%s] queryDestServiceTopics timeout! reason:%s, can retry", msg.MessageID, err.Error())
			} else {
				base.ZLog.Warnf("MsgID[%s] queryDestServiceTopics failed! reason:%s", msg.MessageID, err.Error())
				break
			}
		} else {
			queryOk = true
			break
		}
	}

	if !queryOk {
		base.ZLog.Errorf("MsgID[%s] queryDestServiceTopics failed!", msg.MessageID)
		return
	}

	// 转发
	switch msg.DispatchPolicy {
	case proto.RouterSvrDispatchPolicyBrd:
		if destServiceTopics, ok := iDestServiceTopics.([]string); ok {
			jsonData, _ := ffjson.Marshal(msg)
			for _, destTopic := range destServiceTopics {
				base.ZLog.Debugf("Forward msgID[%s] to Dest[%s]", msg.MessageID, destTopic)
				doyoKfk.PushKfkData(destTopic, jsonData)
			}
		}
	case proto.RouterSvrDispatchPolicyRandom:
		fallthrough
	case proto.RouterSvrDispatchPolicyDH:
		fallthrough
	case proto.RouterSvrDispatchPolicyLoad:
		fallthrough
	case proto.RouterSvrDispatchPolicyRR:
		if destServiceTopic, ok := iDestServiceTopics.(string); ok {
			jsonData, _ := ffjson.Marshal(msg)
			base.ZLog.Debugf("Forward msgID[%s] to Dest[%s]", msg.MessageID, destServiceTopic)
			doyoKfk.PushKfkData(destServiceTopic, jsonData)
		}
	}
}
