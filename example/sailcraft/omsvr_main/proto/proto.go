/*
 * @Author: calmwu
 * @Date: 2018-05-17 20:18:41
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-05-18 19:17:21
 */

package proto

import (
	financesvr_proto "sailcraft/financesvr_main/proto"
)

type ProtoActiveInstControlS struct {
	ZoneID          int                         `json:"ZoneID"`            // 该活动归属的zone
	ActiveType      financesvr_proto.ActiveType `json:"ActiveType"`        // 和financesvr中的ActiveType对应
	ActiveID        int                         `json:"ActiveID"`          // 活动类型下的具体活动ID
	StartTimeName   string                      `json:"StartTime"`         // 活动ID的开始时间 2006-01-02 15:04:05
	DurationMinutes int                         `json:"DurationMinutes"`   // 持续的分钟
	ChannelName     string                      `xorm:"int 'ChannelName'"` // NOAREA CN US
	TimeZone        string                      `json:"TimeZone"`          // 时区，默认是Local，使用服务器时区，服务器用的是UTC
	GroupID         int                         `json:"GroupID"`           // 归属的group，group有自己的属性
}

// 导入活动实例的开启配置
type ProtoAddActiveInstCtrlsReq struct {
	Uin             uint64                    `json:"Uin"`
	ActiveInstCtrls []ProtoActiveInstControlS `json:"ActiveInstCtrls"`
}

// 控制命令，从数据库中加载Waiting状态的活动控制
type ProtoLoadWatingActiveInstCtrlsReq struct {
	Uin uint64 `json:"Uin"`
}

// 控制名利，清除所有运行+待运行的活动控制
type ProtoCleanAllActiveInstCtrlsReq struct {
	Uin uint64 `json:"Uin"`
}

// 查询现在开放的活动类型
type ProtoQueryRunningActiveTypesReq struct {
	Uin    uint64 `json:"Uin"`
	ZoneID int32  `json:"ZoneID"`
}

type ProtoQueryRunningActiveTypesRes struct {
	Uin                uint64                        `json:"Uin"`
	ZoneID             int32                         `json:"ZoneID"`
	RunningActiveTypes []financesvr_proto.ActiveType `json:"RATS" mapstructure:"RATS"` // 该用户能开到的开放中的活动类型
}
