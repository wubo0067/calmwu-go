/*
 * @Author: calm.wu
 * @Date: 2018-10-03 20:44:41
 * @Last Modified by: calm.wu
 * @Last Modified time: 2018-10-03 22:03:30
 */

package main

const (
	CMD_REQ_REVERSESTRING int = iota
	CMD_RES_REVERSESTRING
	CMD_NTF_HELLO
)

type ReverseStringReq struct {
	NormalString string `json:"NormalString"`
}

type ReverseStringRes struct {
	ReverseString string `json:"ReverseString"`
}

type HelloNtf struct {
	HelloInfo string `json:"HelloInfo"`
}

type AppMsg struct {
	Cmd  int    `json:"Cmd"` // 业务定义的命令字
	Body []byte `json:"Body` // 序列化数据，根据命令字打包解包
}
