/*
 * @Author: calmwu
 * @Date: 2018-09-20 11:39:48
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-10-24 17:58:29
 */

package routersvr

import (
	base "doyo-server-go/doyo-base-go"
	"doyo-server-go/doyo-routersvr-go/doyo-kafka-go/doyokafka"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/urfave/cli"
	"go.uber.org/zap/zapcore"
)

var (
	// 命令参数
	RouterSvrFlags = []cli.Flag{
		cli.IntFlag{
			Name:  "id",
			Value: 1,
			Usage: "RouterSvr instance ID",
		},
		cli.StringFlag{
			Name:  "topic",
			Value: "topic-routersvr",
			Usage: "kafka topic",
		},
		cli.StringFlag{
			Name:  "logpath",
			Value: "../log",
			Usage: "routersvr log path",
		},
		cli.StringFlag{
			Name:  "conf",
			Value: "../conf/conf.json",
			Usage: "routersvr config file",
		},
		cli.StringFlag{
			Name:  "consulhealthcheckaddr",
			Value: "localhost:7002",
			Usage: "Consul health check address",
		},
	}
)

func Main(c *cli.Context) error {
	var err error
	// 获取参数
	routerSvrID := c.Int("id")
	kafkaTopic := c.String("topic")
	conf := c.String("conf")
	logPath := c.String("logpath")
	consulHealthCheckAddr := c.String("consulhealthcheckaddr")

	// 初始化log
	base.InitDefaultZapLog(fmt.Sprintf("%s/DoyoRouterSvr_%d.log", logPath, routerSvrID), zapcore.DebugLevel)

	err = loadConfig(conf)
	if err != nil {
		return err
	}

	base.SeedMathRand()

	consulListenAddr := confMgr.ConsulConf.ConsulListenAddr
	if nil == net.ParseIP(consulListenAddr) {
		// 如果不是ip，通过设备名获取
		consulListenAddr, err = base.GetIPByIfname(consulListenAddr)
		if err != nil {
			base.ZLog.Errorf("NewRouterSvrConsul GetIPByIfname failed: %s", err.Error())
			return err
		}
	}

	// 路由策略管理
	policyMgr := newRouterSvrPolicyMgr()

	// 服务健康状态查询
	healthCheck, err := newRouterSvrHealthCheck(kafkaTopic, routerSvrID, consulListenAddr, consulHealthCheckAddr)
	if err != nil {
		return err
	}
	// 注册到consul
	err = healthCheck.start(policyMgr)
	if err != nil {
		return err
	}

	// 等待下，完成路由表的创建
	time.Sleep(3 * time.Second)

	// 初始化kafka
	doyoKfk, err := doyokafka.InitModule(confMgr.KfaConf.Brokers, []string{kafkaTopic}, confMgr.KfaConf.GroupID, base.ZLog)
	if err != nil {
		base.ZLog.Errorf("doyokafka InitModule failed: %s", err.Error())
		return err
	}

	// 转发逻辑
	routerSvrForward, err := newRourterSvrForward(confMgr.DispatchRoutineCount, policyMgr, doyoKfk)
	if err != nil {
		return err
	}

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
				routerSvrForward.forwardData(d)
			}
		}
	}
	// 停止转发
	routerSvrForward.stop()
	// 停止健康服务状态同步
	healthCheck.stop()
	// 停止路由策略
	policyMgr.stop()
	// 停止推送数据到kafka
	doyoKfk.ShutDown()

	base.ZLog.Infof("%s-%d exit!", kafkaTopic, routerSvrID)
	return nil
}
