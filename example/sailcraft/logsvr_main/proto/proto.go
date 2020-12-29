/*
 * @Author: calmwu
 * @Date: 2017-08-31 11:23:23
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-09-01 17:19:45
 * @Comment:
 */

package proto

import l4g "log4go"

// 日志的json对象
type ProtoLogInfoS struct {
	HostIP     string    `json:"HostIP"`
	ServerID   string    `json:"ServerID"`
	LogLevel   l4g.Level `json:"LogLevel"` // 	FINEST = 0, FINE, DEBUG, TRACE, INFO, WARNING, ERROR, CRITICAL
	FileName   string    `json:"FileName"`
	LineNo     int       `json:"LineNo"`
	LogContent string    `json:"LogContent"`
}
