/*
 * @Author: calmwu
 * @Date: 2017-11-29 15:49:09
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-11-29 17:15:27
 * @Comment:
 */

package protocol

import (
	"sailcraft/base"
	"strings"
	"testing"

	"github.com/golang/protobuf/proto"
)

func TestPackUnPack(t *testing.T) {

	protoSyn := new(ProtoSyn)
	protoSyn.DHClientPubKey = "12222222221"
	protoSyn.VerifyStr = "verify string"

	data, _ := PackSynMsg(998, protoSyn)

	t.Logf("%v", data)

	head, err := UnPackProtoHead(data)
	if err != nil {
		t.Error(err.Error())
		return
	}
	t.Logf("head:%+v\n", head)

	t.Logf("%v, %d", data, ProtoC2SHeadSize)

	var protoSyn1 ProtoSyn
	err = proto.Unmarshal(data[ProtoC2SHeadSize:], &protoSyn1)
	if err != nil {
		t.Error(err.Error())
		return
	}
	t.Logf("protoSyn1:%+v\n", protoSyn1)
}

func TestProtoFlod(t *testing.T) {
	cryptoKey := "1234567887654321"

	cipherBlock, err := base.NewCipherBlock(cryptoKey)
	if err != nil {
		t.Error(err.Error())
		return
	}

	rawData := `123456780-=qwertyuiopasdfghjklzxcvbnm!@#$%^&*()
	123456780-=qwertyuiopasdfghjklzxcvbnm!@#$%^&*()
	123456780-=qwertyuiopasdfghjklzxcvbnm!@#$%^&*()
	123456780-=qwertyuiopasdfghjklzxcvbnm!@#$%^&*()
	123456780-=qwertyuiopasdfghjklzxcvbnm!@#$%^&*()
	123456780-=qwertyuiopasdfghjklzxcvbnm!@#$%^&*()
	123456780-=qwertyuiopasdfghjklzxcvbnm!@#$%^&*()
	123456780-=qwertyuiopasdfghjklzxcvbnm!@#$%^&*()
	123456780-=qwertyuiopasdfghjklzxcvbnm!@#$%^&*()
	123456780-=qwertyuiopasdfghjklzxcvbnm!@#$%^&*()
	123456780-=qwertyuiopasdfghjklzxcvbnm!@#$%^&*()
	123456780-=qwertyuiopasdfghjklzxcvbnm!@#$%^&*()`

	t.Logf("rawData len:%d, %v", len(rawData), []byte(rawData))

	flodData, err := FlodPayloadData([]byte(rawData), cipherBlock)
	if err != nil {
		t.Error(err.Error())
		return
	}

	t.Logf("flodData len:%d", len(flodData))

	unFlodData, err := UnFlodPayloadData(flodData, cipherBlock)
	if err != nil {
		t.Error(err.Error())
		return
	}

	t.Logf("unFlodData len:%d, %v", len(unFlodData), unFlodData)

	if strings.Compare(rawData, string(unFlodData)) == 0 {
		t.Log("Test ok")
	} else {
		t.Errorf("Test failed, unFlodData[%s]", string(unFlodData))
	}
}
