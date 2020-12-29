/*
 * @Author: calmwu
 * @Date: 2018-09-20 17:45:46
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-11-05 11:10:03
 */

package proto

type RouterSvrDispatchPolicy int

const (
	RouterSvrDispatchPolicyRandom    RouterSvrDispatchPolicy = iota // 随机转发
	RouterSvrDispatchPolicyRR                                       // RR，
	RouterSvrDispatchPolicyDH                                       // Destination Hashing，hash计算DispatchPolicyArg，对服务实例取模
	RouterSvrDispatchPolicyBrd                                      // 广播，对所有服务实例转发
	RouterSvrDispatchPolicyBrdMS                                    // 广播，对所有服务实例转发选择Master和slave
	RouterSvrDispatchPolicyBrdMaster                                // 广播，接收的为master，处理请求并响应
	RouterSvrDispatchPolicyBrdSlave                                 // 广播，接收的为slave，处理请求
	RouterSvrDispatchPolicyLoad                                     // 选择负载最小服务实例
)

type ServTypeMatchPolicy int

const (
	ServTypeMatchPolicyStrict ServTypeMatchPolicy = iota // 严格匹配
	ServtypeMatchPolicyPrefix                            // 前缀匹配
)

type RouterSvrDispatchMsg struct {
	MessageID         string                  `json:"MessageID"`         // 消息id，用uuid
	FromTopic         string                  `json:"FromTopic"`         // 消息发送方topic，每个producer都有自己的topic
	ToServType        string                  `json:"ToServType"`        // 目的业务业务类型
	DispatchPolicy    RouterSvrDispatchPolicy `json:"DispatchPolicy"`    // 转发策略
	DispatchPolicyArg string                  `json:"DispatchPolicyArg"` // 转发策略运算参数
	PayLoad           []byte                  `json:"PayLoad"`           // 业务转发的具体消息
}

type AppServType int

const (
	DoyoRecDataSvrType    AppServType = 1
	DoyoRecCalcSvrType    AppServType = 2
	DoyoRecDisplaySvrType AppServType = 3
)

type AppServMsg struct {
	AppServCmdID   int    `json:"AppServCmdID"`   // 业务服务的命令字 = AppServType * 1000 + id
	AppServCmdData []byte `json:"AppServCmdData"` // 业务服务的命令数据，ffjson序列化后的buf
}
