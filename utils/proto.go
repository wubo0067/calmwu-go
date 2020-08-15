/*
 * @Author: calmwu
 * @Date: 2017-10-11 10:31:16
 * @Last Modified by: calmwu
 * @Last Modified time: 2020-08-15 19:48:02
 * @Comment:
 */

package utils

type ProtoReturnCode int

type ProtoRequestHeadS struct {
	Version    int    `json:"Version"`
	EventID    int    `json:"EventID"`
	TimeStamp  int64  `json:"TimeStamp"`
	CsrfToken  string `json:"CsrfToken"`
	ChannelUID string `json:"ChannelUID"`
	Uin        int    `json:"Uin"`
}

type ProtoResponseHeadS struct {
	Version    int             `json:"Version"`
	TimeStamp  int64           `json:"TimeStamp"`
	EventID    int             `json:"EventID"`
	ReturnCode ProtoReturnCode `json:"ReturnCode"`
}

type ProtoData struct {
	InterfaceName string      `json:"InterfaceName"`
	Params        interface{} `json:"Params"`
}

type ProtoRequestS struct {
	ProtoRequestHeadS
	ReqData ProtoData `json:"ReqData"`
}

type ProtoResponseS struct {
	ProtoResponseHeadS
	ResData ProtoData `json:"ResData"`
}

type ProtoFailInfoS struct {
	FailureReason string `json:"FailureReason"`
}
