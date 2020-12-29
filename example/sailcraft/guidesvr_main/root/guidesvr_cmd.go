package root

import (
	"fmt"
	"net"
	"os"
	"sailcraft/base"
	"sailcraft/guidesvr_main/common"
	"sailcraft/sysconf"

	"github.com/urfave/cli"
)

var (
	GuideSvrFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "ip, i",
			Value: "0.0.0.0",
			Usage: "Guide Service Listen Address",
		},
		cli.IntFlag{
			Name:  "port, p",
			Value: 8000,
			Usage: "Guide Service Listen Port",
		},
		cli.StringFlag{
			Name:  "conf, c",
			Value: "./config.json",
			Usage: "Guide Service ConfigS File",
		},
		cli.StringFlag{
			Name:  "logpath, l",
			Value: "../log",
			Usage: "Guide Service Log File",
		},
		cli.IntFlag{
			Name:  "cport, t",
			Value: 8100,
			Usage: "Guide Service Control Interface Port",
		},
		cli.StringFlag{
			Name:  "consul, u",
			Value: "",
			Usage: "Guide server ip",
		},
	}

	GuideSvrReloadFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "ips, s",
			Value: "192.168.32.129,118.89.34.64",
			Usage: "Guide Service List",
		},
		cli.StringFlag{
			Name:  "ports, p",
			Value: "10001, 10002, 10003",
			Usage: "Guide Service Control Port",
		},
	}

	GuideSvrCmds = []cli.Command{
		{
			Name:    "guide",
			Aliases: []string{"o"},
			Usage:   "Start SailCraft Guide Version Service",
			Flags:   GuideSvrFlags,
			Action:  GuideSvrStart,
		},
		{
			Name:    "reload",
			Aliases: []string{"r"},
			Usage:   "Notify SailCraft Guide reload config",
			Flags:   GuideSvrReloadFlags,
			Action:  GuideSvrReload,
		},
	}
)

func GuideSvrStart(c *cli.Context) error {
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
	logFileName := fmt.Sprintf("%s/guidesvr_%d.log", servLogPath, webListenPort)
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

	// 初始化redis
	err = common.InitRedisCluster(sysconf.GRedisConfig.ClusterRedisAddressList)
	if err != nil {
		return err
	}

	base.GLog.Info("GuideSvr[%s:%d] configFile[%s] Running", webListenIP, webListenPort, servConfigFile)

	// 启动管理端口
	err = InitSvrCtrl(webListenIP, servControlPort)
	if err != nil {
		return err
	}

	// 注册consul
	if len(consulServerIP) != 0 && net.ParseIP(consulServerIP) != nil {
		err = common.RegisterToConsul(webListenIP, webListenPort, consulServerIP, webListenPort)
		if err != nil {
			base.GLog.Error("registerToConsul failed! reason[%s]", err.Error())
			return err
		} else {
			ginRouter.GET("/GuideSvr/healthCheck", onHealthCheck)
			base.GLog.Info("registerToConsul successed!")
		}
	}

	RunWebServ(webListenIP, webListenPort)

	base.GLog.Info("GuideSvr[%s:%d] Exit!", webListenIP, webListenPort)
	return nil
}

func GuideSvrReload(c *cli.Context) error {
	reloadIps := c.String("ips")
	ports := c.String("ports")
	doReloadCmd(reloadIps, ports)
	return nil
}
