/*
 * @Author: calmwu
 * @Date: 2017-09-01 11:33:44
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-09-02 22:32:25
 * @Comment:
 */

package logsvr

import (
	"encoding/json"
	"fmt"
	l4g "log4go"
	"net"
	"sailcraft/base"
	"sailcraft/logsvr_main/proto"
	"strings"
	"time"
)

type LogSvrMgr struct {
	LogHandleMap        map[string]l4g.Logger
	logInfoDispatchChan chan *proto.ProtoLogInfoS
}

var (
	// 全局对象
	GLogSvrMgr *LogSvrMgr
)

func init() {
	GLogSvrMgr = new(LogSvrMgr)
	GLogSvrMgr.LogHandleMap = make(map[string]l4g.Logger)
	GLogSvrMgr.logInfoDispatchChan = make(chan *proto.ProtoLogInfoS, 10240)
}

// 运行
func (logsvrMgr *LogSvrMgr) Run(listenIP string, listenPort int, logStoragePath string) error {

	// 启动dipatch goroutine
	go logsvrDispatchLogInfoRoutine(logsvrMgr, logStoragePath)

	// 启动网络
	listener, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.ParseIP(listenIP),
		Port: listenPort})

	if err != nil {
		base.GLog.Error("Listen[%s:%d] failed! reason[%s]", listenIP, listenPort, err.Error())
		return err
	}
	base.GLog.Debug("LogSvr Listen[%s:%d]", listenIP, listenPort)

	// 读取网络数据
	logInfoBuf := make([]byte, 65535)
	for {
		n, remoteAddr, err := listener.ReadFromUDP(logInfoBuf)

		if err != nil {
			base.GLog.Error("ReadFromUDP[%s] failed! reason[%s]", listener.RemoteAddr().String(), err.Error())
		} else {
			// 处理命令
			realDataBuf := logInfoBuf[:n]
			// 解包
			logInfo := new(proto.ProtoLogInfoS)
			err := json.Unmarshal(realDataBuf, logInfo)
			if err != nil {
				base.GLog.Error("Unmarshal remote[%s] protoLogInfo failed! reason[%s]",
					remoteAddr, err.Error())
			} else {
				//
				logInfo.HostIP = strings.Split(remoteAddr.String(), ":")[0]
				logsvrMgr.logInfoDispatchChan <- logInfo
			}
		}
	}
	return nil
}

func logsvrDispatchLogInfoRoutine(logsvrMgr *LogSvrMgr, logStoragePath string) {
	base.GLog.Info("logsvrDispatchLogInfoRoutine running!")

	for {
		select {
		case logInfo, ok := <-logsvrMgr.logInfoDispatchChan:
			if ok {
				remoteServerTag := fmt.Sprintf("%s:%s", logInfo.HostIP, logInfo.ServerID)
				logW, ok := logsvrMgr.LogHandleMap[remoteServerTag]
				if !ok {
					// 生成log目录
					serverLogStoragePath := fmt.Sprintf("%s/%s/%s", logStoragePath, logInfo.HostIP, logInfo.ServerID)
					if err := base.MkDir(serverLogStoragePath); err != nil {
						base.GLog.Error("Create path[%s] failed! reason[%s]", serverLogStoragePath, err.Error())
						continue
					} else {
						// 生成log句柄
						serverLogFile := fmt.Sprintf("%s/%s.log", serverLogStoragePath, logInfo.ServerID)
						l4g.LogBufferLength = 1024
						logFilter := l4g.NewFileLogWriter(serverLogFile, false)
						if logFilter == nil {
							base.GLog.Error("Create logFile[%s] failed", serverLogFile)
							continue
						} else {
							logFilter.SetRotateDaily(true)
							logFilter.SetRotateMaxBackup(100)
							logFilter.SetFormat("%M")
							logW = make(l4g.Logger)
							logW.AddFilter("normal", l4g.FINE, logFilter)
							// 插入
							logsvrMgr.LogHandleMap[remoteServerTag] = logW
						}
					}
				}
				// 更加logLevel来记录信息
				// 自定义format，[time] [level] [file:line] content
				content := fmt.Sprintf("[%s] [%s] [%s:%d] %s", time.Now().UTC().String(), logInfo.LogLevel.String(),
					logInfo.FileName, logInfo.LineNo, logInfo.LogContent)
				logW.Logf(logInfo.LogLevel, content)
			}
		}
	}
}
