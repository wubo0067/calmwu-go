/*
 * @Author: calmwu
 * @Date: 2017-12-05 15:11:59
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-01-04 18:59:27
 * @Comment:
 */

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"calmwu-go/utils"
	"sailcraft/network/protocol"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/monnand/dhkx"
)

var (
	wg           sync.WaitGroup
	sessionCount = flag.Int("count", 1, "")
)

func sendSyn(session *net.TCPConn, msgId uint32, pubKey []byte) error {
	synPkg := new(protocol.ProtoSyn)
	synPkg.VerifyBuf = []byte("123456qwaszx")
	synPkg.DHClientPubKey = pubKey

	fmt.Printf("synPkg:%+v\n", *synPkg)
	data, _ := protocol.PackSynMsg(msgId, synPkg)
	n, err := session.Write(data)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	fmt.Printf("send bytes[%d] to server\n", n)
	return nil
}

func sendAck(session *net.TCPConn, msgId uint32, cipherText []byte) error {
	ackPkg := new(protocol.ProtoAck)
	ackPkg.CipherText = cipherText
	data, _ := protocol.PackAckMsg(msgId, ackPkg)
	n, err := session.Write(data)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	fmt.Printf("send bytes[%d] to server\n", n)
	return nil
}

func startSession() {
	var err error
	var msgId uint32 = 1
	var dhKey *dhkx.DHKey

	defer wg.Done()

	//addr, _ := net.ResolveTCPAddr("tcp", "118.89.34.64:1003")
	addr, _ := net.ResolveTCPAddr("tcp", "192.168.12.3:8003")

	session, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	dhKey, err = utils.GenerateDHKey()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	sendSyn(session, msgId, dhKey.Bytes())
	msgId++

	readBuf := make([]byte, 1024)
	n, err := session.Read(readBuf)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("read %d bytes\n", n)
	readBuf = readBuf[:n]

	protoHead, err := protocol.UnPackProtoHead(readBuf)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("protoHead:%+v\n", *protoHead)

	var synAck protocol.ProtoAsyn
	err = proto.Unmarshal(readBuf[protocol.ProtoC2SHeadSize:], &synAck)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("asyn:%+v\n", synAck)

	key, err := utils.GenerateEncryptionKey(synAck.DHServerPubKey, dhKey)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("key:%v\n", key)
	fmt.Printf("key:%x\n", key)

	cipherBlock, err := utils.NewCipherBlock(key[:32])
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// 加密
	cipherBuf, err := utils.EncryptPlainText(cipherBlock, []byte("123456qwaszx"))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	sendAck(session, msgId, cipherBuf)
	msgId++

	content := []byte("Hello Session!")
	cipherBuf, _ = protocol.FlodPayloadData(content, cipherBlock)

	reader := bufio.NewReader(session)
	protoHeadBuf := make([]byte, protocol.ProtoC2SHeadSize)

	for {
		payLoad, _ := protocol.PackTransferData(msgId, cipherBuf)

		n, err = session.Write(payLoad)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		//fmt.Printf("send transfer data to server, %d:bytes\n", n)

		_, err = io.ReadFull(reader, protoHeadBuf)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		protoHead, err = protocol.UnPackProtoHead(protoHeadBuf)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		dataBuf := make([]byte, protoHead.DataLen)
		_, err = io.ReadFull(reader, dataBuf)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		content, err = protocol.UnFlodPayloadData(dataBuf, cipherBlock)

		//fmt.Println(string(content))

		time.Sleep(1 * time.Second)
	}
}

func main() {
	flag.Parse()

	for i := 0; i < *sessionCount; i++ {
		wg.Add(1)
		go startSession()
	}

	wg.Wait()

	fmt.Println("----------exit------------")
	return
}
