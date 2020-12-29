/*
 * @Author: calmwu
 * @Date: 2018-05-16 13:50:46
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-05-18 17:34:08
 * @Comment:
 */

package db

import (
	financesvr_proto "sailcraft/financesvr_main/proto"
)

const TBNAME_ACTIVEINSTCTRL = "tbl_ActiveInstControl" // 活动实例控制表

type ActivePerformState int

const (
	E_ACTIVEPERFORMSTATE_WAITING ActivePerformState = iota
	E_ACTIVEPERFORMSTATE_RUNNING
	E_ACTVIEPERFROMSTATE_COMPLETED
)

// 活动配置表对象定义
type TblActiveInstControlS struct {
	Id              int64                       // 自增字段
	ZoneID          int                         `xorm:"int index 'ZoneID'"`       // 该活动归属的zone
	ActiveType      financesvr_proto.ActiveType `xorm:"int index 'ActiveType'"`   // 和financesvr中的ActiveType对应
	ActiveID        int                         `xorm:"int index 'ActiveID'"`     // 活动类型下的具体活动ID
	StartTimeName   string                      `xorm:"string 'StartTime'"`       // 活动ID的开始时间 2006-01-02 15:04:05
	DurationMinutes int                         `xorm:"int 'DurationMinutes'"`    // 持续的分钟
	ChannelName     string                      `xorm:"int 'ChannelName'"`        // NOAREA CN US
	TimeZone        string                      `xorm:"string 'TimeZone'"`        // 时区，默认是Local，使用服务器时区，服务器用的是UTC
	PerformState    ActivePerformState          `xorm:"int index 'PerformState'"` // 执行状态
	GroupID         int                         `xorm:"int 'GroupID'"`            // 归属的group，group有自己的属性
}
