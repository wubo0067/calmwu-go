/*
 * @Author: calmwu
 * @Date: 2017-09-18 10:38:09
 * @Last Modified by: mikey.zhaopeng
 * @Last Modified time: 2018-08-17 10:54:49
 * @Comment:
 */

package base

import (
	"io"
	"log"
	l4g "log4go"
	"os"
)

var (
	GLog l4g.Logger = nil
)

func init() {
	if GLog == nil {
		GLog = make(l4g.Logger)
	}
}

func InitLog(logFilefullPath string) {
	l4g.LogBufferLength = 1024
	log_writer := l4g.NewFileLogWriter(logFilefullPath, false)
	log_writer.SetRotate(true)
	log_writer.SetRotateDaily(true)
	log_writer.SetRotateMaxBackup(7)
	GLog.AddFilter("normal", l4g.FINE, log_writer)
	return
}

func NewSimpleLog(out io.Writer) *log.Logger {
	logOutput := out
	if out == nil {
		logOutput = os.Stderr
	}
	return log.New(logOutput, "", log.Ldate|log.Lmicroseconds|log.Lshortfile)
}
