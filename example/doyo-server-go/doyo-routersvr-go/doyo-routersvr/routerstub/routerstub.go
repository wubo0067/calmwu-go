/*
 * @Author: calmwu
 * @Date: 2018-09-30 10:58:37
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-10-03 15:13:14
 */

package routerstub

import (
	base "doyo-server-go/doyo-base-go"
	"doyo-server-go/doyo-base-go/consul_api"
	"doyo-server-go/doyo-routersvr-go/doyo-kafka-go/doyokafka"
	"doyo-server-go/doyo-routersvr-go/proto"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pquerna/ffjson/ffjson"
)

const (
	callFutureChanSize   = 1024
	stubReceiveChanSize  = 1024
	timeOutMsgIDChanSize = 128
)

type RouterStubModule struct {
	doyoKfk              *doyokafka.KafkaModule
	callFutureChan       chan *RSCallerFuture              // 请求chan
	callFutureChanClosed bool                              // 管道关闭标志
	callFutureMap        map[string]*RSCallerFuture        // 回调使用
	timeOutMsgIDChan     chan string                       // callfuture超时管道
	stubReceiveChan      chan *doyokafka.DoyoKafkaReadData // call回应通道
	receiveChanClosed    bool                              // 管道关闭标志
	stopFlag             int32                             // 停止flag
	exitWait             sync.WaitGroup
	onReceive            OnReceive
}

func NewRouterStubModule(svrName string, svrInstName string, doyoKfk *doyokafka.KafkaModule, onReceive OnReceive,
	hostIP string, healthPort int) (*RouterStubModule, error) {
	rs := &RouterStubModule{
		doyoKfk:              doyoKfk,
		callFutureChan:       make(chan *RSCallerFuture, callFutureChanSize),
		callFutureChanClosed: false,
		callFutureMap:        make(map[string]*RSCallerFuture),
		timeOutMsgIDChan:     make(chan string, timeOutMsgIDChanSize),
		stubReceiveChan:      make(chan *doyokafka.DoyoKafkaReadData, stubReceiveChanSize),
		receiveChanClosed:    false,
		stopFlag:             0,
		onReceive:            onReceive,
	}

	checkUrl := fmt.Sprintf("/%s/health/", svrInstName)
	httpCheckUrl := fmt.Sprintf("http://%s:%d%s", hostIP, healthPort, checkUrl)

	go func() {
		http.HandleFunc(checkUrl, func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		http.ListenAndServe(fmt.Sprintf("%s:%d", hostIP, healthPort), nil)
	}()

	// 注册
	ConsulClient, err := consul_api.NewConsulClient(hostIP)
	if err != nil {
		base.ZLog.Errorf("NewConsulClient failed! reason[%s]", err.Error())
		return nil, err
	}

	err = consul_api.ConsulSvrReg(ConsulClient, svrName, []string{svrName}, svrInstName, hostIP,
		healthPort, httpCheckUrl)
	if err != nil {
		base.ZLog.Errorf("ConsulSvrReg failed: %s", err.Error())
		return nil, err
	}

	base.ZLog.Infof("httpCheckUrl:%s", httpCheckUrl)

	go rs.routerStubRoutine()
	rs.exitWait.Add(1)

	return rs, nil
}

func (rsm *RouterStubModule) Stop() {
	if n := atomic.LoadInt32(&rsm.stopFlag); n == 0 {
		atomic.StoreInt32(&rsm.stopFlag, 1)
		close(rsm.callFutureChan)
		close(rsm.stubReceiveChan)
		base.ZLog.Infof("RouterStubModule Stop!")
		rsm.exitWait.Wait()
	}
}

func (rsm *RouterStubModule) Notify(destSvrTopic string, routerMsg []byte) error {
	if n := atomic.LoadInt32(&rsm.stopFlag); n == 0 {
		// 直接发送
		rsm.doyoKfk.PushKfkData(destSvrTopic, routerMsg)
		return nil
	}
	return ErrRouterStubStop
}

func (rsm *RouterStubModule) Call(msgID string, routerMsg []byte, timeout time.Duration) (*RSCallerFuture, error) {
	if n := atomic.LoadInt32(&rsm.stopFlag); n == 0 {
		callFuture := &RSCallerFuture{
			MsgID:     msgID,
			timeout:   timeout,
			routerMsg: routerMsg,
		}
		callFuture.init()

		rsm.callFutureChan <- callFuture
		return callFuture, nil
	}
	return nil, ErrRouterStubStop
}

func (rsm *RouterStubModule) ReceiveDoyoKfkData(kfkData *doyokafka.DoyoKafkaReadData) {
	if n := atomic.LoadInt32(&rsm.stopFlag); n == 0 {
		rsm.stubReceiveChan <- kfkData
	}
}

func (rsm *RouterStubModule) routerStubRoutine() {
	base.ZLog.Debug("routerStubRoutine running")

	defer func() {
		rsm.exitWait.Done()
		if err := recover(); err != nil {
			// 回收panic，防止整个服务panic
			stackInfo := base.CallStack(1)
			base.ZLog.DPanicw("routerStubRoutine panic recovered! err", err, "stack", string(stackInfo))
		}
	}()

	//timeOutCtx, timeOutCancel := context.WithCancel(context.Background())
	callFutureTicker := time.NewTicker(1 * time.Second)
L:
	for {
		select {
		case callFuture, ok := <-rsm.callFutureChan:
			if ok {
				// 发送出去
				rsm.doyoKfk.PushKfkData(TopicRouterSvr, callFuture.routerMsg)
				// 注册
				rsm.callFutureMap[callFuture.MsgID] = callFuture
				// 启动定时器
				go func() {
					select {
					case <-time.After(callFuture.timeout):
						rsm.timeOutMsgIDChan <- callFuture.MsgID
					}
				}()
			} else {
				// 被关闭
				rsm.callFutureChanClosed = true
				//base.ZLog.Info("callFutureChan receive close event")
			}
		case timeOutMsgID, ok := <-rsm.timeOutMsgIDChan:
			if ok {
				if _, exist := rsm.callFutureMap[timeOutMsgID]; exist {
					// 超时，response没有返回
					rsm.callFutureMap[timeOutMsgID].response(nil, ErrRouterStubCallTimeOut)
					delete(rsm.callFutureMap, timeOutMsgID)
					base.ZLog.Warnf("MsgID[%s] TimeOut! callFuture response to App", timeOutMsgID)
				}
			} else {
				base.ZLog.Debugf("timeOutMsgIDChan closed")
			}
		case <-callFutureTicker.C:
			// 没有任何数据时退出
			if n := atomic.LoadInt32(&rsm.stopFlag); n == 1 {
				base.ZLog.Debugf("Now will shutdown, check callFuture count[%d]", len(rsm.callFutureMap))
				if len(rsm.callFutureMap) == 0 &&
					rsm.receiveChanClosed &&
					rsm.callFutureChanClosed {
					base.ZLog.Warnf("routerStubRoutine will exit!")
					break L
				}
			}
		case kfkData, ok := <-rsm.stubReceiveChan:
			if ok {
				var msg proto.RouterSvrDispatchMsg
				err := ffjson.Unmarshal(kfkData.Data(), &msg)
				if err != nil {
					base.ZLog.Errorf("ffjson Unmarshal failed! reason: %s", err.Error())
				} else {
					// find respone callfuture
					if _, exist := rsm.callFutureMap[msg.MessageID]; exist {
						rsm.callFutureMap[msg.MessageID].response(&msg, nil)
						delete(rsm.callFutureMap, msg.MessageID)
					} else {
						if rsm.onReceive != nil {
							// callback
							rsm.onReceive(rsm, msg.MessageID, msg.FromTopic, msg.PayLoad)
						}
					}
				}
			} else {
				// 被关闭
				rsm.receiveChanClosed = true
				//base.ZLog.Info("stubReceiveChan receive close event")
			}
		}
	}
	base.ZLog.Debug("routerStubRoutine exit!")
}
