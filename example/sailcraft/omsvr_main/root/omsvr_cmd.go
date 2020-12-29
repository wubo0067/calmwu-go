/*
 * @Author: calmwu
 * @Date: 2018-05-18 10:38:45
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-05-18 12:17:58
 * @Comment:
 */

package root

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"sailcraft/base"
	"sailcraft/omsvr_main/activemgr"
	"sailcraft/omsvr_main/common"
	"sailcraft/sysconf"
	"syscall"

	"github.com/urfave/cli"
)

var (
	OMSvrFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "ip, i",
			Value: "0.0.0.0",
			Usage: "OperationManger Service Listen Address",
		},
		cli.IntFlag{
			Name:  "port, p",
			Value: 2000,
			Usage: "OperationManger Service Listen Port",
		},
		cli.StringFlag{
			Name:  "conf, c",
			Value: "./config.json",
			Usage: "OperationManger Service ConfigS File",
		},
		cli.StringFlag{
			Name:  "logpath, l",
			Value: "../log",
			Usage: "OperationManger Service Log File",
		},
		cli.IntFlag{
			Name:  "cport, t",
			Value: 2100,
			Usage: "OperationManger Service Control Interface Port",
		},
		cli.StringFlag{
			Name:  "consul, u",
			Value: "",
			Usage: "OperationManger server ip, no: not register to consul",
		},
	}

	OMSvrReloadFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "ips, s",
			Value: "192.168.32.129,118.89.34.64",
			Usage: "OperationManger Service List",
		},
		cli.StringFlag{
			Name:  "ports, p",
			Value: "2100, 2101",
			Usage: "OperationManger Service Control Port",
		},
	}

	OMSvrCmds = []cli.Command{
		{
			Name:    "omsvr",
			Aliases: []string{"o"},
			Usage:   "Start SailCraft OperationManger Version Service",
			Flags:   OMSvrFlags,
			Action:  OMSvrStart,
		},
		{
			Name:    "reload",
			Aliases: []string{"r"},
			Usage:   "Notify SailCraft OperationManger reload config",
			Flags:   OMSvrReloadFlags,
			Action:  OMSvrReload,
		},
	}
)

func OMSvrStart(c *cli.Context) error {
	webListenIP := c.String("ip")
	webListenPort := c.Int("port")
	servControlPort := c.Int("cport")
	servConfigFile := c.String("conf")
	servLogPath := c.String("logpath")
	consulServerIP := c.String("consul")

	// 判断目录是否存在
	err := base.CheckDir(servLogPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return err
	}

	//初始化log
	logFileName := fmt.Sprintf("%s/omsvr_%d.log", servLogPath, webListenPort)
	base.InitLog(logFileName)
	defer base.GLog.Close()

	// 读取配置
	err = common.GConfig.Init(servConfigFile)
	if err != nil {
		return err
	}

	// 获得系统配置
	err = sysconf.Initialize(common.GConfig.GetSysConfPath())
	if err != nil {
		return err
	}

	// 初始化数据库
	err = common.InitMysql(sysconf.GMysqlConfig.ConfigMap["omsdb"])
	if err != nil {
		base.GLog.Error(err.Error())
		return err
	}
	base.GLog.Info("OMSvr[%s:%d] configFile[%s] Running", webListenIP, webListenPort, servConfigFile)

	// 启动管理端口
	err = InitSvrCtrl(webListenIP, servControlPort)
	if err != nil {
		return err
	}

	// 注册到consul
	if len(consulServerIP) != 0 && net.ParseIP(consulServerIP) != nil {
		err = common.RegisterToConsul(webListenIP, webListenPort, consulServerIP, webListenPort)
		if err != nil {
			base.GLog.Error("registerToConsul failed! reason[%s]", err.Error())
			return err
		} else {
			ginRouter.GET("/OMSvr/healthCheck", onHealthCheck)
			base.GLog.Info("registerToConsul successed!")
		}
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGUSR1)
	go func() {
		for range sc {
			base.DumpStacks()
		}
	}()

	// 启动活动管理
	activemgr.CreateActiveInstCtrlMgr()
	if activemgr.GActiveInstCtrlMgr == nil {
		base.GLog.Debug("activemgr.ActiveInstCtrlMgr == nil")
	}

	RunWebServ(webListenIP, webListenPort)

	base.GLog.Info("OMSvr[%s:%d] Exit!", webListenIP, webListenPort)
	return nil
}

func OMSvrReload(c *cli.Context) error {
	reloadIps := c.String("ips")
	ports := c.String("ports")
	doReloadCmd(reloadIps, ports)
	return nil
}
