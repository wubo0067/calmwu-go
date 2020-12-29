/*
 * @Author: calmwu
 * @Date: 2017-11-29 11:47:49
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-12-08 17:29:41
 */

package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"sailcraft/base"

	"github.com/golang/protobuf/proto"
)

type ProtoC2SCmd int16

const (
	E_C2S_CMD_SYN ProtoC2SCmd = iota
	E_S2C_CMD_ASYN
	E_C2S_CMD_ACK // 握手协议，主要用来交换密钥，验证加密用
	E_TRANSFER    // 业务传输
	E_HEARTBEAT
)

var (
	ErrUnPackProtoHead = errors.New("unpackprotohead failed")
	ErrPackProtoSyn    = errors.New("pack ProtoSyn failed")
	ErrPackProtoAsyn   = errors.New("pack ProtoAsyn failed")
	ErrPackProtoAck    = errors.New("pack ProtoAck failed")
)

// 这里的数据类型不能是int，必须明确位宽，否则binary操作会报错
type ProtoC2SHead struct {
	MagicValue uint32      // 魔术字
	MsgId      uint32      // 消息序号
	Cmd        ProtoC2SCmd //
	DataLen    int32       // 数据长度
}

const (
	ProtoC2SMagic = uint32(0x98651210)
	//ProtoC2SHeadSize = int(unsafe.Sizeof(head_82732831712))  这里无法使用sizeof，应为有align，可序列化时是pack 1的
	ProtoC2SHeadSize int = 14
)

func UnPackProtoHead(data []byte) (*ProtoC2SHead, error) {
	// make read buf
	var err error
	unPackBuf := bytes.NewReader(data)
	head := new(ProtoC2SHead)

	if err = binary.Read(unPackBuf, binary.BigEndian, &head.MagicValue); err != nil {
		goto FAILED
	}
	if err = binary.Read(unPackBuf, binary.BigEndian, &head.MsgId); err != nil {
		goto FAILED
	}
	if err = binary.Read(unPackBuf, binary.BigEndian, &head.Cmd); err != nil {
		goto FAILED
	}
	if err = binary.Read(unPackBuf, binary.BigEndian, &head.DataLen); err != nil {
		goto FAILED
	}

	return head, nil

FAILED:
	base.GLog.Error("UnPackProtoHead failed! reason[%s]", err.Error())

	return nil, ErrUnPackProtoHead
}

func PackSynMsg(msgId uint32, protoSyn *ProtoSyn) ([]byte, error) {
	if protoSyn == nil {
		return nil, fmt.Errorf("input args is invalid!")
	}

	// 对pb对象打包序列化
	serialBuf, err := proto.Marshal(protoSyn)
	if err != nil {
		base.GLog.Error("PackSynMsg failed! reason[%s]", err.Error())
		return nil, ErrPackProtoSyn
	}

	// make write buf
	packBuf := bytes.NewBuffer(nil)
	// write head
	binary.Write(packBuf, binary.BigEndian, ProtoC2SMagic)
	binary.Write(packBuf, binary.BigEndian, msgId)
	binary.Write(packBuf, binary.BigEndian, E_C2S_CMD_SYN)
	binary.Write(packBuf, binary.BigEndian, int32(len(serialBuf)))

	packBuf.Write(serialBuf)

	return packBuf.Bytes(), nil
}

func PackAsynMsg(msgId uint32, protoAsyn *ProtoAsyn) ([]byte, error) {
	if protoAsyn == nil {
		return nil, fmt.Errorf("input args is invalid!")
	}
	serialBuf, err := proto.Marshal(protoAsyn)
	if err != nil {
		base.GLog.Error("PackAsynMsg failed! reason[%s]", err.Error())
		return nil, ErrPackProtoSyn
	}

	// make write buf
	packBuf := bytes.NewBuffer(nil)
	// write head
	binary.Write(packBuf, binary.BigEndian, ProtoC2SMagic)
	binary.Write(packBuf, binary.BigEndian, msgId)
	binary.Write(packBuf, binary.BigEndian, E_S2C_CMD_ASYN)
	binary.Write(packBuf, binary.BigEndian, int32(len(serialBuf)))

	packBuf.Write(serialBuf)
	return packBuf.Bytes(), nil
}

func PackAckMsg(msgId uint32, protoAck *ProtoAck) ([]byte, error) {
	if protoAck == nil {
		return nil, fmt.Errorf("input args is invalid!")
	}

	serialBuf, err := proto.Marshal(protoAck)
	if err != nil {
		base.GLog.Error("PackAckMsg failed! reason[%s]", err.Error())
		return nil, ErrPackProtoSyn
	}

	// make write buf
	packBuf := bytes.NewBuffer(nil)
	// write head
	binary.Write(packBuf, binary.BigEndian, ProtoC2SMagic)
	binary.Write(packBuf, binary.BigEndian, msgId)
	binary.Write(packBuf, binary.BigEndian, E_C2S_CMD_ACK)
	binary.Write(packBuf, binary.BigEndian, int32(len(serialBuf)))

	packBuf.Write(serialBuf)
	return packBuf.Bytes(), nil
}

func PackHeartBeat(msgId uint32, unix int64) ([]byte, error) {
	heartbeat := new(ProtoHeartBeat)
	heartbeat.Timestamp = unix
	serialBuf, err := proto.Marshal(heartbeat)
	if err != nil {
		base.GLog.Error("PackAckMsg failed! reason[%s]", err.Error())
		return nil, ErrPackProtoSyn
	}

	// make write buf
	packBuf := bytes.NewBuffer(nil)
	// write head
	binary.Write(packBuf, binary.BigEndian, ProtoC2SMagic)
	binary.Write(packBuf, binary.BigEndian, msgId)
	binary.Write(packBuf, binary.BigEndian, E_HEARTBEAT)
	binary.Write(packBuf, binary.BigEndian, int32(len(serialBuf)))
	packBuf.Write(serialBuf)

	return packBuf.Bytes(), nil
}

func PackTransferData(msgId uint32, payLoad []byte) ([]byte, error) {
	// make write buf
	packBuf := bytes.NewBuffer(nil)
	// write head
	binary.Write(packBuf, binary.BigEndian, ProtoC2SMagic)
	binary.Write(packBuf, binary.BigEndian, msgId)
	binary.Write(packBuf, binary.BigEndian, E_TRANSFER)
	binary.Write(packBuf, binary.BigEndian, int32(len(payLoad)))
	packBuf.Write(payLoad)

	return packBuf.Bytes(), nil
}
