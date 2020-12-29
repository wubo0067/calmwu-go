package main

import (
	"kp_proto"
)
import proto "github.com/golang/protobuf/proto"

func pack_kpmessage(cmd kp_proto.KPProtoCmd, cmd_body []byte) ([]byte, int) {
	var kp_msg kp_proto.KPMessage
	kp_msg.Cmd = &cmd
	kp_msg.RealMessageMarshaldata = cmd_body

	g_log.Debug(kp_msg.String())

	// nnd，接口就是方法的集合，golang会判断这个对象是否支持这些方法，关键就要看方法的recevier是什么了
	data, err := proto.Marshal(&kp_msg)
	if err != nil {
		g_log.Error("Marshaling KPMessage failed! reason[%s]", err.Error())
		return nil, -1
	}
	return data, 0
}

func unpack_kpmessage(data []byte) (*kp_proto.KPMessage, int) {
	var kp_msg kp_proto.KPMessage
	err := proto.Unmarshal(data, &kp_msg)
	if err != nil {
		g_log.Error("Unmarshal KPMessage failed! reason[%s]", err.Error())
		return nil, -1
	}
	return &kp_msg, 0
}

// 转变为interface是默认的，只要传入的对象实现了接口的方法，这里的关键是方法的recevier是什么，因为有些方法是指针
// 明白了不，不要用c的指针去思考问题
func pack_kpcmd(kp_cmd proto.Message) ([]byte, int) {
	if kp_cmd == nil {
		g_log.Error("input kp_cmd message is nil")
		return nil, -1
	}
	// 打印
	// g_log.Debug(kp_cmd.String())
	// 打包
	data, err := proto.Marshal(kp_cmd)
	if err != nil {
		g_log.Error("Marshaling kp_cmd failed! reason[%s]", err.Error())
		return nil, -1
	}
	return data, 0
}

// 解具体的命令包，这个kp_cmd对应的是&kp_proto.xxxx
func unpack_kpcmd(data []byte, kp_cmd proto.Message) int {
	if len(data) == 0 || data == nil {
		g_log.Error("input data is invalid!")
		return -1
	}

	err := proto.Unmarshal(data, kp_cmd)
	if err != nil {
		g_log.Error("Unmarshal kp_cmd failed! reason[%s]", err.Error())
		return -1
	}
	// 打印
	// g_log.Debug(kp_cmd.String())
	return 0
}
