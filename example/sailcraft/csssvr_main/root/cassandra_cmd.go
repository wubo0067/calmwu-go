/*
 * @Author: calmwu
 * @Date: 2018-01-10 15:45:31
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-05-14 17:45:51
 * @Comment:
 */

package root

import (
	"fmt"
	"net"
	"os"
	"sailcraft/base"
	"sailcraft/csssvr_main/common"
	"sailcraft/csssvr_main/store"

	"github.com/urfave/cli"
)

var (
	CassandraSvrFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "ip, i",
			Value: "0.0.0.0",
			Usage: "Cassandra Service Listen Address",
		},
		cli.IntFlag{
			Name:  "port, p",
			Value: 9000,
			Usage: "Cassandra Service Listen Port",
		},
		cli.StringFlag{
			Name:  "conf, c",
			Value: "./config.json",
			Usage: "Cassandra Service ConfigS File",
		},
		cli.StringFlag{
			Name:  "logpath, l",
			Value: "../log",
			Usage: "Cassandra Service Log File",
		},
		cli.IntFlag{
			Name:  "cport, t",
			Value: 9100,
			Usage: "Cassandra Service Control Interface Port",
		},
		cli.StringFlag{
			Name:  "consul, u",
			Value: "",
			Usage: "Cassandra server ip",
		},
	}

	CassandraSvrReloadFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "ips, s",
			Value: "192.168.32.129,118.89.34.64",
			Usage: "Cassandra Service List",
		},
		cli.StringFlag{
			Name:  "ports, p",
			Value: "10001, 10002, 10003",
			Usage: "Cassandra Service Control Port",
		},
	}

	CassandraSvrCmds = []cli.Command{
		{
			Name:    "store",
			Aliases: []string{"s"},
			Usage:   "Start SailCraft CassandraSvr Version Service",
			Flags:   CassandraSvrFlags,
			Action:  CassandraSvrStart,
		},
		{
			Name:    "reload",
			Aliases: []string{"r"},
			Usage:   "Notify SailCraft CassandraSvr reload config",
			Flags:   CassandraSvrReloadFlags,
			Action:  CassandraSvrReload,
		},
	}
)

func CassandraSvrStart(c *cli.Context) error {
	webListenIP := c.String("ip")
	webListenPort := c.Int("port")
	//servControlPort := c.Int("cport")
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
	logFileName := fmt.Sprintf("%s/cassandrasvr_%d.log", servLogPath, webListenPort)
	base.InitLog(logFileName)
	defer base.GLog.Close()

	// 读取配置
	err = common.GConfig.Init(servConfigFile)
	if err != nil {
		return err
	}

	// 获得系统配置
	// err = sysconf.Initialize(common.GConfig.GetSysConfPath())
	// if err != nil {
	// 	return err
	// }

	base.GLog.Info("CassandraSvr[%s:%d] configFile[%s] Running", webListenIP, webListenPort, servConfigFile)

	// 启动管理端口
	// err = InitSvrCtrl(webListenIP, servControlPort)
	// if err != nil {
	// 	return err
	// }

	// 注册consul
	if len(consulServerIP) != 0 && net.ParseIP(consulServerIP) != nil {
		err = common.RegisterToConsul(webListenIP, webListenPort, consulServerIP, webListenPort)
		if err != nil {
			base.GLog.Error("registerToConsul failed! reason[%s]", err.Error())
			return err
		} else {
			ginRouter.GET("/CassandraSvr/healthCheck", onHealthCheck)
			base.GLog.Info("registerToConsul successed!")
		}
	}

	// 初始化Cassandra store
	err = store.CasMgr.InitCassandraSessions(&common.GConfig.ConfigData.CassandraConf)
	if err != nil {
		return err
	}
	// 	defer cassandra.CaMgr.FiniCassandraSessions()
	// }
	RunWebServ(webListenIP, webListenPort)

	base.GLog.Info("CassandraSvr[%s:%d] Exit!", webListenIP, webListenPort)
	return nil
}

func CassandraSvrReload(c *cli.Context) error {
	return nil
}
