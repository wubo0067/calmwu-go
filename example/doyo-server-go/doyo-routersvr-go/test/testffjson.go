/*
 * @Author: calmwu
 * @Date: 2018-09-26 16:32:50
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-09-26 17:00:38
 */

package main

import (
	"doyo-server-go/doyo-routersvr-go/proto"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pquerna/ffjson/ffjson"

	"github.com/satori/go.uuid"
)

func main() {

	for i := 0; i < 10; i++ {
		uid, _ := uuid.NewV4()
		fmt.Println(uid.String())
	}

	uid, _ := uuid.NewV4()
	req := &proto.RouterSvrDispatchMsg{
		MessageID:      uid.String(),
		FromTopic:      "DoyoAppServer-1",
		ToServType:     "DoyoRecommended",
		DispatchPolicy: proto.RouterSvrDispatchPolicyRandom,
	}

	fmt.Printf("req: %+v\n", req)

	serialData, err := json.Marshal(req)
	if err != nil {
		fmt.Println(err.Error())
	}

	req1 := new(proto.RouterSvrDispatchMsg)
	err = ffjson.Unmarshal(serialData, req1)
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Printf("req1: %+v\n", req1)

	ntfChan := make(chan struct{})
	go func() {
		fmt.Println("---------close ntfChan before")
		close(ntfChan)
		fmt.Println("---------close ntfChan after")
	}()
	<-ntfChan
	fmt.Println("---------receive ntfChan")
	time.Sleep(time.Second)
}
