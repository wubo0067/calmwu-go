/*
 * @Author: calmwu
 * @Date: 2018-11-13 15:38:37
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-11-19 16:49:45
 */

package main

import (
	base "doyo-server-go/doyo-base-go"
	"doyo-server-go/doyo-routersvr-go/doyo-kafka-go/doyokafka"
	"doyo-server-go/doyo-routersvr-go/doyo-routerstub"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap/zapcore"
)

var (
	cmdParamSvrName    = flag.String("svrname", "DoyoReqApp", "")
	cmdParamID         = flag.Int("id", 1, "")
	cmdParamKfkBrokers = flag.String("brokers", "192.168.68.230:9094,192.168.68.230:9092,192.168.68.230:9093", "")
	cmdParamHostIP     = flag.String("ip", "192.168.68.228", "")
)

func OnReceive(doyoRsm *routerstub.RouterStubModule, msgID string, fromTopic string, data []byte) {
	base.ZLog.Debugf("msgID[%s] fromTopic[%s] data[%s]", msgID, fromTopic, string(data))
}

// func notifyUserLogin(doyoRsm *routerstub.RouterStubModule, svrInstanceTopic string) {
// 	userLogin := recdata_proto.ProtoDoyoRecDataUserLogin{
// 		UserID:       "doyo123987654",
// 		UserLanguage: "VN",
// 		UserCountry:  "VN",
// 		UserGender:   1,
// 	}

// 	jsonLogin, _ := ffjson.Marshal(userLogin)
// 	appMsg := routersvr_proto.AppServMsg{
// 		AppServCmdID:   int(recdata_proto.DoyoRecDataCmdUserLogin),
// 		AppServCmdData: jsonLogin,
// 	}

// 	jsonMsg, _ := ffjson.Marshal(appMsg)
// 	msgID, err := routerstub.Notify(doyoRsm, svrInstanceTopic, "DoyoRecDataSvr", routersvr_proto.RouterSvrDispatchPolicyRR, "", jsonMsg)
// 	if err != nil {
// 		base.ZLog.Errorf("notify failed! reason:%s", err.Error())
// 		return
// 	}
// 	base.ZLog.Debugf("notifty msgid[%s]", msgID)
// }

// func notifyUserLogout(doyoRsm *routerstub.RouterStubModule, svrInstanceTopic string) {
// 	userLogout := recdata_proto.ProtoDoyoRecDataUserLogout{
// 		UserID: "doyo123987654",
// 	}

// 	jsonLogout, _ := ffjson.Marshal(userLogout)
// 	appMsg := routersvr_proto.AppServMsg{
// 		AppServCmdID:   int(recdata_proto.DoyoRecDataCmdUserLogout),
// 		AppServCmdData: jsonLogout,
// 	}

// 	jsonMsg, _ := ffjson.Marshal(appMsg)
// 	msgID, err := routerstub.Notify(doyoRsm, svrInstanceTopic, "DoyoRecDataSvr", routersvr_proto.RouterSvrDispatchPolicyRR, "", jsonMsg)
// 	if err != nil {
// 		base.ZLog.Errorf("notify failed! reason:%s", err.Error())
// 		return
// 	}
// 	base.ZLog.Debugf("notifty msgid[%s]", msgID)
// }

func main() {
	flag.Parse()

	base.InitDefaultZapLog("test_user.log", zapcore.DebugLevel)

	svrInstanceTopic := fmt.Sprintf("%s-%d", *cmdParamSvrName, *cmdParamID)

	doyoKfk, err := doyokafka.InitModule(*cmdParamKfkBrokers, []string{svrInstanceTopic}, svrInstanceTopic, base.ZLog)
	if err != nil {
		base.ZLog.Errorf("doyokafka InitModule failed: %s", err.Error())
		return
	}

	doyoRsm, err := routerstub.NewRouterStubModule(*cmdParamSvrName, svrInstanceTopic, doyoKfk, OnReceive,
		"192.168.68.229", *cmdParamHostIP, 5500)
	if err != nil {
		base.ZLog.Errorf("NewRouterStubModule failed, reason:%s", err.Error())
		return
	}

	// 信号处理
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)

	notifyUserLogin(doyoRsm, svrInstanceTopic, "doyo123987654", "VN", "VN")

	time.Sleep(10 * time.Second)

	//notifyUserLogout(doyoRsm, svrInstanceTopic)

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
