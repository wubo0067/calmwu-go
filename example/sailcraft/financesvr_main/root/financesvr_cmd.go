package root

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"sailcraft/base"
	"sailcraft/financesvr_main/common"
	"sailcraft/sysconf"
	"syscall"

	"github.com/urfave/cli"
)

var (
	FinanceSvrFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "ip, i",
			Value: "0.0.0.0",
			Usage: "Finance Service Listen Address",
		},
		cli.IntFlag{
			Name:  "port, p",
			Value: 4000,
			Usage: "Finance Service Listen Port",
		},
		cli.StringFlag{
			Name:  "conf, c",
			Value: "./config.json",
			Usage: "Finance Service ConfigS File",
		},
		cli.StringFlag{
			Name:  "logpath, l",
			Value: "../log",
			Usage: "Finance Service Log File",
		},
		cli.IntFlag{
			Name:  "cport, t",
			Value: 4100,
			Usage: "Finance Service Control Interface Port",
		},
		cli.StringFlag{
			Name:  "consul, u",
			Value: "",
			Usage: "Finance server ip, no: not register to consul",
		},
	}

	FinanceSvrReloadFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "ips, s",
			Value: "192.168.32.129,118.89.34.64",
			Usage: "Finance Service List",
		},
		cli.StringFlag{
			Name:  "ports, p",
			Value: "4001, 4002, 4003",
			Usage: "Finance Service Control Port",
		},
	}

	FinanceSvrCmds = []cli.Command{
		{
			Name:    "finance",
			Aliases: []string{"f"},
			Usage:   "Start SailCraft Finance Version Service",
			Flags:   FinanceSvrFlags,
			Action:  FinanceSvrStart,
		},
		{
			Name:    "reload",
			Aliases: []string{"r"},
			Usage:   "Notify SailCraft Finance reload config",
			Flags:   FinanceSvrReloadFlags,
			Action:  FinanceSvrReload,
		},
	}
)

func FinanceSvrStart(c *cli.Context) error {
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
	logFileName := fmt.Sprintf("%s/financesvr_%d.log", servLogPath, webListenPort)
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
	err = common.InitRedis(sysconf.GRedisConfig.SingletonRedisAddrsss)
	if err != nil {
		return err
	}

	// 初始化数据库
	err = common.InitMysql(sysconf.GMysqlConfig.ConfigMap["user_finance"])

	base.GLog.Info("FinanceSvr[%s:%d] configFile[%s] Running", webListenIP, webListenPort, servConfigFile)

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
			ginRouter.GET("/FinanceSvr/healthCheck", onHealthCheck)
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

	RunWebServ(webListenIP, webListenPort)

	base.GLog.Info("FinanceSvr[%s:%d] Exit!", webListenIP, webListenPort)
	return nil
}

func FinanceSvrReload(c *cli.Context) error {
	reloadIps := c.String("ips")
	ports := c.String("ports")
	doReloadCmd(reloadIps, ports)
	return nil
}
