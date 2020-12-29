package main

import (
	base "doyo-server-go/doyo-base-go"

	pio "github.com/lincanli/go-pio"
)

/*
 * @Author: calmwu
 * @Date: 2018-10-20 15:46:26
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-10-20 18:21:25
 */

const (
	PioHost = "http://192.168.68.229:7070"
	AppKey  = "WAP0f5leqfLvxKZn2pSnW8QonwYZnh7AGkveCUekerxWlTGdUgme6KCCsdMjjFYF"
)

func main() {
	logger := base.NewSimpleLog(nil)
	// pioEventClient := pio.NewEventClient(PioHost, AppKey)

	// // 循环读取用户事件
	// fd, err := os.Open("sample_data.txt")
	// if err != nil {
	// 	logger.Println(err.Error())
	// 	return
	// }
	// defer fd.Close()

	// scanner := bufio.NewScanner(fd)
	// for scanner.Scan() {
	// 	fmt.Println(scanner.Text())
	// 	userEvtInfo := strings.Split(scanner.Text(), ",")

	// 	pioEvt := pio.NewEvent(userEvtInfo[1])

	// 	if userEvtInfo[1] != "$set" {
	// 		pioEvt.SetEntityType("user")
	// 		pioEvt.SetEntityID(userEvtInfo[0])
	// 		pioEvt.SetTargetEntityType("item")
	// 		pioEvt.SetTargetEntityID(userEvtInfo[2])
	// 	} else {
	// 		pioEvt.SetEvent("$set")
	// 		pioEvt.SetEntityType("item")
	// 		pioEvt.SetEntityID(userEvtInfo[0])
	// 		pioEvt.SetTargetEntityType("nil")
	// 		pioEvt.SetTargetEntityID("userEvtInfo[2]")
	// 		//pioEvt.SetProperties(map[string]interface{}{"categories": userEvtInfo[2]})
	// 	}
	// 	pioEvt.SetEventTime(time.Now())

	// 	evtRes, err := pioEventClient.SentClient(pioEvt)
	// 	if err != nil {
	// 		logger.Printf("user[%s] item[%s] Error, reason[%s]", userEvtInfo[0], userEvtInfo[2], err.Error())
	// 		continue
	// 	}
	// 	logger.Printf("Success, eventid[%s] user[%s] item[%s]", evtRes.EventID, userEvtInfo[0], userEvtInfo[2])
	// }

	// if err := scanner.Err(); err != nil {
	// 	logger.Fatal(err.Error())
	// }

	pioEngineClient := pio.NewEngineClient("http://192.168.68.229:8000")
	pioEngineClient.AccessKey = AppKey

	res, err := pioEngineClient.Query(map[string]string{"uid": "yosssi"})
	if err != nil {
		logger.Fatal(err.Error())
	} else {
		logger.Printf("query yosssi res[%s]", string(res))
	}
}
