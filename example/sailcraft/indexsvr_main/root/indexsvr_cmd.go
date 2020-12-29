/*
 * @Author: calmwu
 * @Date: 2017-09-18 10:08:52
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-04-28 16:38:47
 * @Comment:
 */

package root

import (
	"fmt"
	"net"
	"os"
	"sailcraft/base"
	"sailcraft/base/word_filter"
	"sailcraft/indexsvr_main/common"
	"sailcraft/indexsvr_main/data"
	"sailcraft/indexsvr_main/proto"

	"github.com/urfave/cli"
)

var (
	indexSvrFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "ip, i",
			Value: "0.0.0.0",
			Usage: "Service Listen Address",
		},
		cli.IntFlag{
			Name:  "port, p",
			Value: 5000,
			Usage: "Service Listen Port",
		},
		cli.StringFlag{
			Name:  "conf, c",
			Value: "./config.json",
			Usage: "Service ConfigS File",
		},
		cli.StringFlag{
			Name:  "logpath, l",
			Value: "../log",
			Usage: "Service Log File",
		},
		cli.IntFlag{
			Name:  "cport, t",
			Value: 5100,
			Usage: "Service Control Interface Port",
		},
		cli.StringFlag{
			Name:  "consul, u",
			Value: "",
			Usage: "consul server ip",
		},
	}

	indexSvrControlFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "ips, s",
			Value: "192.168.32.129,118.89.34.64",
			Usage: "Index Service List",
		},
		cli.StringFlag{
			Name:  "ports, p",
			Value: "505",
			Usage: "Index Control Port",
		},
		cli.StringFlag{
			Name:  "cmd, c",
			Value: "reload-conf/reload-data",
			Usage: "Index Control cmd",
		},
	}

	IndexSvrCmds = []cli.Command{
		{
			// 索引服务
			Name:    "index",
			Aliases: []string{"i"},
			Usage:   "Start SailCraft Index Service",
			Flags:   indexSvrFlags,
			Action:  actionIndexSvrStart,
		},
		{
			Name:    "control",
			Aliases: []string{"r"},
			Usage:   "SailCraft Query Service Control",
			Flags:   indexSvrControlFlags,
			Action:  actionIndexSvrControl,
		},
	}
)

func actionIndexSvrStart(c *cli.Context) error {
	// 启动服务
	webListenIP := c.String("ip")
	webListenPort := c.Int("port")
	servControlPort := c.Int("cport")
	servConfigFile := c.String("conf")
	servLogPath := c.String("logpath")
	consulServerIP := c.String("consul")

	common.GServName = c.App.Name
	common.GServListenIP = webListenIP
	common.GServListenCtrlPort = servControlPort

	// 判断目录是否存在
	err := base.CheckDir(servLogPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return err
	}

	//初始化log
	logFileName := fmt.Sprintf("%s/indexsvr_%d.log", servLogPath, webListenPort)
	base.InitLog(logFileName)
	defer base.GLog.Close()

	err = common.LoadConfig(servConfigFile, webListenIP)
	if err != nil {
		return err
	}

	base.GLog.Info("IndexSvr[%s:%d] configFile[%s] controlPort[%d] Running", webListenIP, webListenPort, servConfigFile,
		servControlPort)

	// 启动加载数据
	dataMgr := data.CreateDataMgr("redis")
	if dataMgr == nil {
		return fmt.Errorf("Create redis mgr failed")
	}

	err = dataMgr.Load()
	if err != nil {
		return err
	}

	// 加载脏字库
	err = base.PathExist(common.GConfig.DirtyWordFile)
	if err != nil {
		base.GLog.Error(err.Error())
		return err
	}
	word_filter.LoadDicFiles([]string{common.GConfig.DirtyWordFile})

	err = InitCtrl()
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
			ginRouter.GET("/IndexSvr/healthCheck", onHealthCheck)
			base.GLog.Info("registerToConsul successed!")
		}
	}

	RunWebServ(webListenIP, webListenPort)

	return nil
}

func actionIndexSvrControl(c *cli.Context) {
	// 发送控制命令
	ipNames := c.String("ips")
	ports := c.String("ports")
	cmd := c.String("cmd")
	if cmd == "reload-data" {
		ReloadCmd(ipNames, ports, proto.CTRLCMD_RELOADDATA_REQ)
	} else if cmd == "reload-conf" {
		ReloadCmd(ipNames, ports, proto.CTRLCMD_REOLADCONF_REQ)
	}
}
