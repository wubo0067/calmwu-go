/*
 * @Author: calmwu
 * @Date: 2017-08-31 14:11:49
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-09-01 17:19:11
 * @Comment:
 */

package logsvr

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"sailcraft/base"
	"sailcraft/base/consul_api"

	"github.com/urfave/cli"
)

var (
	LogSvrFlag = []cli.Flag{
		cli.StringFlag{
			Name:  "ip, i",
			Value: "0.0.0.0",
			Usage: "Service Listen Address",
		},
		cli.IntFlag{
			Name:  "port, p",
			Value: 5005,
			Usage: "Service Listen Port",
		},
		cli.StringFlag{
			Name:  "storagepath, s",
			Value: "./logstorage",
			Usage: "Service log file storage path",
		},
		cli.StringFlag{
			Name:  "logpath, l",
			Value: "./log",
			Usage: "Local log file path",
		},
		cli.StringFlag{
			Name:  "consul, u",
			Value: "",
			Usage: "consul server ip",
		},
		cli.IntFlag{
			Name:  "cport, c",
			Value: 0,
			Usage: "Service Health Check Port",
		},
	}
)

func LogSvrAction(c *cli.Context) error {
	listenIP := c.String("ip")
	listenPort := c.Int("port")
	logStoragePath := c.String("storagepath")
	logPath := c.String("logpath")

	consulServerIP := c.String("consul")
	healthCheckPort := c.Int("cport")

	err := base.CheckDir(logStoragePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return err
	}

	err = base.CheckDir(logPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return err
	}

	// 初始化log
	logFileName := fmt.Sprintf("%s/logsvr.log", logPath)
	base.InitLog(logFileName)
	defer base.GLog.Close()

	// 注册consul
	if len(consulServerIP) != 0 && net.ParseIP(consulServerIP) != nil && healthCheckPort != 0 {
		err = registerToConsul(listenIP, listenPort, consulServerIP, healthCheckPort)
		if err != nil {
			return err
		} else {
			base.GLog.Info("registerToConsul successed!")
		}
	}

	//
	GLogSvrMgr.Run(listenIP, listenPort, logStoragePath)

	base.GLog.Info("LogSvr Listen[%s:%d] LogStoragePath[%s] LogPath[%s] Running\n", listenIP, listenPort, logStoragePath, logPath)
	return nil
}

func registerToConsul(listenIP string, listenPort int, consulServerIP string, healthCheckPort int) error {
	consulClient, err := consul_api.NewConsulClient(consulServerIP)
	if err != nil {
		base.GLog.Error("registerToConsul failed! reason[%s]", err.Error())
		return err
	}

	servName := "SailCraft-LogSvr"
	servTags := []string{"LogSvr"}
	servInstName := fmt.Sprintf("LogSvr-%s:%d", listenIP, listenPort)
	healthCheckUrl := fmt.Sprintf("http://%s:%d/LogSvr/healthCheck", listenIP, healthCheckPort)

	go func() {
		onHealthCheck := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}
		http.HandleFunc("/LogSvr/healthCheck", onHealthCheck)
		http.ListenAndServe(fmt.Sprintf("%s:%d", listenIP, healthCheckPort), nil)
	}()

	return consul_api.ConsulSvrReg(consulClient, servName, servTags, servInstName, listenIP, listenPort, healthCheckUrl)
}
