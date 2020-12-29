/*
 * @Author: ksingeryu
 * @Date: 2017-10-18 10:08:52
 * @Last Modified by: ksingeryu
 * @Last Modified time: 2017-10-20 15:58:43
 * @Comment:
 */

package fleetsvr

import (
	"fmt"
	"os"
	"sailcraft/base"
	"sailcraft/dataaccess/mysql"
	"sailcraft/dataaccess/redistool"
	"sailcraft/fleetsvr_main/config"
	"sailcraft/fleetsvr_main/utils"
	"sailcraft/sysconf"

	"github.com/urfave/cli"
)

var (
	fleetSvrFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "ip, i",
			Value: "0.0.0.0",
			Usage: "Service Listen Address",
		},
		cli.IntFlag{
			Name:  "port, p",
			Value: 2002,
			Usage: "Service Listen Port",
		},
		cli.StringFlag{
			Name:  "conf, c",
			Value: "./config.json",
			Usage: "Service ConfigS File",
		},
		cli.StringFlag{
			Name:  "sysconf, sc",
			Value: "./config.json",
			Usage: "Service SysConfigS File",
		},
		cli.StringFlag{
			Name:  "logpath, l",
			Value: "../log",
			Usage: "Service log File",
		},
		cli.IntFlag{
			Name:  "cport, t",
			Value: 202,
			Usage: "Service Control Interface Port",
		},
		cli.StringFlag{
			Name:  "consul, u",
			Value: "",
			Usage: "Finance server ip",
		},
	}

	fleetSvrControlFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "ips, s",
			Value: "192.168.32.129,118.89.34.64",
			Usage: "sailcraft Service List",
		},
		cli.StringFlag{
			Name:  "ports, p",
			Value: "202",
			Usage: "sailcraft Control Port",
		},
		cli.StringFlag{
			Name:  "cmd, c",
			Value: "reload-conf/reload-data",
			Usage: "sailcraft Control cmd",
		},
	}

	FleetSvrCmds = []cli.Command{
		{
			// 游戏逻辑服务
			Name:    "fleetsvr",
			Aliases: []string{"s"},
			Usage:   "Start SailCraft Service",
			Flags:   fleetSvrFlags,
			Action:  actionFleetSvrStart,
		},
		{
			Name:    "control",
			Aliases: []string{"r"},
			Usage:   "SailCraft Query Service Control",
			Flags:   fleetSvrControlFlags,
			Action:  actionFleetSvrControl,
		},
	}
)

func actionFleetSvrStart(c *cli.Context) error {
	// 启动服务
	webListenIP := c.String("ip")
	webListenPort := c.Int("port")
	//servControlPort := c.Int("cport")
	servConfigFile := c.String("conf")
	servSysConfigFile := c.String("sysconf")
	servLogPath := c.String("logpath")
	consulServerIP := c.String("consul")

	// 判断目录是否存在
	err := base.CheckDir(servLogPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return err
	}

	//初始化log
	logFileName := fmt.Sprintf("%s/sailcraft_%d.log", servLogPath, webListenPort)
	base.InitLog(logFileName)
	defer base.GLog.Close()

	err = sysconf.Initialize(servSysConfigFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return err
	}

	err = config.Initialize(servConfigFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return err
	}

	err = mysql.GMysqlManager.Initialize()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return err
	}

	redistool.GRedisManager.Initialize()

	fmt.Println(consulServerIP)

	if consulServerIP != "" {
		err = utils.RegisterToConsul(webListenIP, webListenPort, consulServerIP, webListenPort)
		if err != nil {
			return err
		}

		GServiceMgr.ginRouter.GET(utils.CONSUL_RELATIVE_PATH, onConsulCheck)
	}

	waitForSigUsr1()

	err = GServiceMgr.RunServ(webListenIP, webListenPort)

	if err != nil {
		return err
	}

	return nil
}

func actionFleetSvrControl(c *cli.Context) error {
	return nil
}
