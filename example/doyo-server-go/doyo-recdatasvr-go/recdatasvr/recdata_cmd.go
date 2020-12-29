/*
 * @Author: calmwu
 * @Date: 2018-11-01 11:02:57
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-11-07 19:34:59
 */

package recdatasvr

import (
	base "doyo-server-go/doyo-base-go"
	"doyo-server-go/doyo-base-go/redistool"
	"doyo-server-go/doyo-recdatasvr-go/doyorecdata"
	"doyo-server-go/doyo-routersvr-go/doyo-kafka-go/doyokafka"
	routerstub "doyo-server-go/doyo-routersvr-go/doyo-routerstub"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/urfave/cli"
	"go.uber.org/zap/zapcore"
)

const (
	ServTypeName = "DoyoRecDataSvr"
)

var (
	// 命令参数
	DoyoRecDataSvrFlags = []cli.Flag{
		cli.IntFlag{
			Name:  "id",
			Value: 1,
			Usage: "DoyoRecDataSvr instance ID",
		},
		cli.StringFlag{
			Name:  "logpath",
			Value: "../log",
			Usage: "DoyoRecDataSvr log path",
		},
		cli.StringFlag{
			Name:  "conf",
			Value: "../conf/conf.json",
			Usage: "DoyoRecDataSvr config file",
		},
	}
)

func Main(c *cli.Context) error {
	var err error

	servID := c.Int("id")
	conf := c.String("conf")
	logPath := c.String("logpath")
	svrInstanceTopic := fmt.Sprintf("%s-%d", ServTypeName, servID)

	// 初始化log
	base.InitDefaultZapLog(fmt.Sprintf("%s/%s.log", logPath, svrInstanceTopic), zapcore.DebugLevel)

	base.SeedMathRand()

	err = loadConfig(conf)
	if err != nil {
		return err
	}

	// 初始化redis
	doyorecdata.DoyoRecDataRedisMgr = redistool.NewRedisMgr(strings.Split(confMgr.RedisConf.ServerAddrs, ","),
		confMgr.RedisConf.SessionCount, confMgr.RedisConf.IsCluster == 1, confMgr.RedisConf.Password)
	err = doyorecdata.DoyoRecDataRedisMgr.Start()
	if err != nil {
		base.ZLog.Errorf("redisMgr[%s] start failed! reason:%s", confMgr.RedisConf.ServerAddrs,
			err.Error())
		return err
	}

	// 初始化kafka
	doyoKfk, err := doyokafka.InitModule(confMgr.KafkaConf.Brokers, []string{svrInstanceTopic}, svrInstanceTopic, base.ZLog)
	if err != nil {
		base.ZLog.Errorf("doyokafka InitModule failed: %s", err.Error())
		return err
	}

	doyoRsm, err := routerstub.NewRouterStubModule(ServTypeName, svrInstanceTopic, doyoKfk, doyorecdata.OnReceive,
		confMgr.HealthCheckConf.ConsulHost, confMgr.HealthCheckConf.CheckHost,
		confMgr.HealthCheckConf.CheckPort)
	if err != nil {
		base.ZLog.Errorf("DoyoRecDataSvr NewRouterStubModule failed, reason:%s", err.Error())
		return err
	}

	// 初始化统计
	recDataStatistics := doyorecdata.InitDoyoRecDataStatistics()

	// 信号处理
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)

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
	doyorecdata.DoyoRecDataRedisMgr.Stop()
	recDataStatistics.Stop()

	return nil
}
