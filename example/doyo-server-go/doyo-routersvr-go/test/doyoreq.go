/*
 * @Author: calmwu
 * @Date: 2018-10-03 11:30:23
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-10-12 11:05:56
 */

package main

import (
	base "doyo-server-go/doyo-base-go"
	"doyo-server-go/doyo-routersvr-go/doyo-kafka-go/doyokafka"
	"doyo-server-go/doyo-routersvr-go/doyo-routerstub"
	"doyo-server-go/doyo-routersvr-go/proto"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pquerna/ffjson/ffjson"

	"go.uber.org/zap/zapcore"
)

var (
	cmdParamSvrName    = flag.String("svrname", "DoyoReqApp", "")
	cmdParamID         = flag.Int("id", 1, "")
	cmdParamKfkBrokers = flag.String("brokers", "", "")
	cmdParamHostIP     = flag.String("ip", "127.0.0.1", "")
)

func OnReceive(doyoRsm *routerstub.RouterStubModule, msgID string, fromTopic string, data []byte) {
	base.ZLog.Debugf("msgID[%s] fromTopic[%s] data[%s]", msgID, fromTopic, string(data))
}

func sendHelloNotify(doyoRsm *routerstub.RouterStubModule, svrInstanceTopic string) {
	helloNtf := HelloNtf{
		HelloInfo: "Hello kafka message, from reqapp",
	}

	jsonNtf, _ := ffjson.Marshal(helloNtf)
	appMsg := AppMsg{
		Cmd:  CMD_NTF_HELLO,
		Body: jsonNtf,
	}

	jsonMsg, _ := ffjson.Marshal(appMsg)
	msgID, err := routerstub.Notify(doyoRsm, svrInstanceTopic, "DoyoResApp", proto.RouterSvrDispatchPolicyRR, "", jsonMsg)
	if err != nil {
		base.ZLog.Errorf("notify failed! reason:%s", err.Error())
		return
	}
	base.ZLog.Debugf("notifty msgid[%s]", msgID)
}

func testCall(doyoRsm *routerstub.RouterStubModule, svrInstanceTopic string) {
	time.Sleep(3 * time.Second)
	reverseStrs := []string{"1234567890", "abcdefg"}

	i := 0
	for i < len(reverseStrs) {
		reverseStrReq := ReverseStringReq{
			NormalString: reverseStrs[i],
		}
		jsonReq, _ := ffjson.Marshal(reverseStrReq)

		appMsg := AppMsg{
			Cmd:  CMD_REQ_REVERSESTRING,
			Body: jsonReq,
		}
		jsonMsg, _ := ffjson.Marshal(appMsg)

		base.ZLog.Debug("----------Call-----------------")

		msgID, responserTopic, resPayLoad, err := routerstub.Call(doyoRsm, svrInstanceTopic, "DoyoResApp", proto.RouterSvrDispatchPolicyRR, "",
			jsonMsg, 5*time.Second)
		if err != nil {
			base.ZLog.Errorf("+++++++++++++++++++msgID[%s] call failed! reason:%s", msgID, err.Error())
		} else {
			base.ZLog.Debugf("Call MsgID[%s]", msgID)
			var resAppMsg AppMsg
			ffjson.Unmarshal(resPayLoad, &resAppMsg)
			base.ZLog.Debugf("msgID[%s] responserTopic[%s] resAppMsg:%+v", msgID, responserTopic, resAppMsg)
			var reverseStrRes ReverseStringRes
			ffjson.Unmarshal(resAppMsg.Body, &reverseStrRes)
			base.ZLog.Debugf("reverseStrRes:%+v", reverseStrRes)
		}
		i++
	}
}

func main() {
	flag.Parse()

	base.InitDefaultZapLog("DoyoReqApp.log", zapcore.DebugLevel)

	svrInstanceTopic := fmt.Sprintf("%s-%d", *cmdParamSvrName, *cmdParamID)

	doyoKfk, err := doyokafka.InitModule(*cmdParamKfkBrokers, []string{svrInstanceTopic}, svrInstanceTopic, base.ZLog)
	if err != nil {
		base.ZLog.Errorf("doyokafka InitModule failed: %s", err.Error())
		return
	}

	doyoRsm, err := routerstub.NewRouterStubModule(*cmdParamSvrName, svrInstanceTopic, doyoKfk, OnReceive,
		*cmdParamHostIP, 5500)
	if err != nil {
		base.ZLog.Errorf("NewRouterStubModule failed, reason:%s", err.Error())
		return
	}

	// 信号处理
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)

	// i := 0
	// for i < 2 {
	// 	sendHelloNotify(doyoRsm, svrInstanceTopic)
	// 	i++
	// }

	go testCall(doyoRsm, svrInstanceTopic)

L:
	for {
		select {
		case sig := <-sigchan:
			switch sig {
			case syscall.SIGINT:
				fallthrough
			case syscall.SIGTERM:
				base.ZLog.Warnf("Receive Stop Notify sig: %s", sig.String())
				// 停止从kafka拉取数据
				doyoKfk.StopPull()
			case syscall.SIGUSR1:
				base.ZLog.Warnf("Receive Reload Notify")
			}
		case data := <-doyoKfk.PullChan():
			switch d := data.(type) {
			// 从kafka拉取数据结束
			case *doyokafka.DoyoKafkaEofData:
				base.ZLog.Info("receive end notify, The data has been read all!")
				break L
			case *doyokafka.DoyoKafkaReadData:
				doyoRsm.ReceiveDoyoKfkData(d)
			}
		}
	}

	doyoRsm.Stop()
	doyoKfk.ShutDown()
}
