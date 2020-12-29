package main

import (
	"bytes"
	"compress/zlib"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"sandmonk_main/proto"
	"sync"
	"time"

	"github.com/mitchellh/mapstructure"
)

//http://tonybai.com/2015/04/30/go-and-https/
// chkvsc.uqsoft.com
// ./client_logincheck --svrip=chkvsc.uqsoft.com --runtimes=2 --ssl=2
// ./client_logincheck --svrip=192.168.2.104 --method=post
// ./client_logincheck --svrip=192.168.2.104 --method=put
// ./client_logincheck --svrip=118.89.34.64 --ssl=2 --method=put/post

var (
	cmdParamsSvrIP    = flag.String("svrip", "192.168.12.3", "")
	cmdParamsRuntimes = flag.Int("runtimes", 1, "")
	cmdParamsSSL      = flag.Int("ssl", 0, "")
	processWaitGroup  sync.WaitGroup
	cmdMethod         = flag.String("method", "post", "post/put")
)

const (
	UrlLoginCheckFmt  = "http://%s:808/OperationalModule/LoginCheck"
	UrlLoginCheckFmtS = "https://%s:808/OperationalModule/LoginCheck"
)

func PostRequest(req *proto.ProtoSandMonkReqS) (error, *proto.ProtoSandMonkResS) {
	// 编码压缩
	serialData, err := json.Marshal(req)
	if err != nil {
		fmt.Println(err.Error())
		return err, nil
	}
	var compressBuf bytes.Buffer
	compressWriter := zlib.NewWriter(&compressBuf)
	// 原始数据写入压缩对象
	compressWriter.Write(serialData)
	compressWriter.Close()

	fmt.Printf("serialData len[%d] compressBuf len[%d]\n", len(serialData), compressBuf.Len())

	// 发送
	var res *http.Response = nil

	if *cmdParamsSSL == 0 {
		UrlTuitionStep := fmt.Sprintf(UrlLoginCheckFmt, *cmdParamsSvrIP)

		res, err = http.Post(UrlTuitionStep, "text/plain; charset=utf-8", &compressBuf)
		if err != nil {
			fmt.Printf("Post to %s failed! [%s]\n", UrlTuitionStep, err.Error())
			return err, nil
		}
	} else if *cmdParamsSSL == 2 {
		UrlTuitionStep := fmt.Sprintf(UrlLoginCheckFmtS, *cmdParamsSvrIP)

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		res, err = client.Post(UrlTuitionStep, "text/plain; charset=utf-8", &compressBuf)
		if err != nil {
			fmt.Printf("Post to %s failed! [%s]\n", UrlTuitionStep, err.Error())
			return err, nil
		}
	} else {
		UrlTuitionStep := fmt.Sprintf(UrlLoginCheckFmtS, *cmdParamsSvrIP)

		// tr := &http.Transport{
		// 	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		// }
		//client := &http.Client{Transport: tr}
		res, err = http.Post(UrlTuitionStep, "text/plain; charset=utf-8", &compressBuf)
		if err != nil {
			fmt.Printf("Post to %s failed! [%s]\n", UrlTuitionStep, err.Error())
			return err, nil
		}
	}

	// 解压
	dcompressReader, err := zlib.NewReader(res.Body)
	if err != nil {
		fmt.Printf("zlib NewReader failed! [%s]\n", err.Error())
		return err, nil
	}

	var resPkg proto.ProtoSandMonkResS
	err = json.NewDecoder(dcompressReader).Decode(&resPkg)
	if err != nil {
		fmt.Printf("Decode failed! [%s]\n", err.Error())
		return err, nil
	}

	return nil, &resPkg
}

func PutRequest(req *proto.ProtoSandMonkReqS) (error, *proto.ProtoSandMonkResS) {
	// 编码压缩
	serialData, err := json.Marshal(req)
	if err != nil {
		fmt.Println(err.Error())
		return err, nil
	}
	var compressBuf bytes.Buffer
	compressWriter := zlib.NewWriter(&compressBuf)
	// 原始数据写入压缩对象
	compressWriter.Write(serialData)
	compressWriter.Close()

	UrlTuitionStep2 := fmt.Sprintf(UrlLoginCheckFmt, *cmdParamsSvrIP)

	var client *http.Client = nil
	if *cmdParamsSSL == 0 {
		client = &http.Client{}
	} else if *cmdParamsSSL == 2 {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr}
		UrlTuitionStep2 = fmt.Sprintf(UrlLoginCheckFmtS, *cmdParamsSvrIP)
	} else {
		fmt.Println("ssl[%d] is invalid!", *cmdParamsSSL)
		return fmt.Errorf("ssl is invalid!"), nil
	}

	fmt.Printf("serialData len[%d] compressBuf len[%d] PUT TO[%s]\n", len(serialData), compressBuf.Len(), UrlTuitionStep2)

	request, err := http.NewRequest("PUT", UrlTuitionStep2, &compressBuf)
	if err != nil {
		fmt.Println("http NewRequest failed! reason:", err.Error())
		return err, nil
	}
	request.Header.Set("Content-Type", "text/plain; charset=utf-8")
	request.ContentLength = int64(compressBuf.Len())
	response, err := client.Do(request)
	if err != nil {
		fmt.Println("http do Request failed! reason:", err.Error())
		return err, nil
	}

	// 解压
	dcompressReader, err := zlib.NewReader(response.Body)
	if err != nil {
		fmt.Printf("zlib NewReader failed! [%s]\n", err.Error())
		return err, nil
	}

	var resPkg proto.ProtoSandMonkResS
	err = json.NewDecoder(dcompressReader).Decode(&resPkg)
	if err != nil {
		fmt.Printf("Decode failed! [%s]\n", err.Error())
		return err, nil
	}

	return nil, &resPkg
}

func doLoginCheck(index int) {
	defer processWaitGroup.Done()

	req := proto.ProtoSandMonkReqS{
		ProtoSandMonkHeadReqS: proto.ProtoSandMonkHeadReqS{
			Version:    1,
			EventId:    998,
			TimeStamp:  time.Now().Unix(),
			ChannelUID: "213121",
			Uin:        65535 + index,
			CsrfToken:  "02cf14994a3be74301657dbcd9c0a189",
		},
		ReqData: proto.ProtoSandMonkData{
			InterfaceName: "LoginCheck",
			Params: map[string]interface{}{
				"ClientVersion": "1.1.8",
				"PlatformName":  "Android",
				"ChannelName":   "GooglePlay",
			},
		},
	}

	var err error = nil
	var res *proto.ProtoSandMonkResS = nil

	if *cmdMethod == "post" {
		err, res = PostRequest(&req)
	} else if *cmdMethod == "put" {
		err, res = PutRequest(&req)
	} else {
		fmt.Printf("method[%s] is invalid!", *cmdMethod)
		return
	}

	fmt.Printf("[%v]\n", res)
	if err == nil {
		switch res.ProtoSandMonkHeadResS.ReturnCode {
		case -1:
			var failedInfo proto.ProtoSandMonkFailedInfoS
			mapstructure.Decode(res.ResData.Params, &failedInfo)
			fmt.Printf("[%+v]\n", failedInfo)
		case 0:
			var loginInfo proto.ProtoSandMonkLoginInfoS
			mapstructure.Decode(res.ResData.Params, &loginInfo)
			fmt.Printf("[%+v]\n", loginInfo)
		case 1:
			fmt.Println("Forbit login!!!!")
		case 2:
			var versionUpdate proto.ProtoSandMonkVersionUpdateS
			mapstructure.Decode(res.ResData.Params, &versionUpdate)
			fmt.Printf("[%+v]\n", versionUpdate)
		case 3:
			var maintain proto.ProtoSandMonkMaintenanceS
			mapstructure.Decode(res.ResData.Params, &maintain)
			fmt.Printf("公告信息：[%s]\n", maintain.Bulletin)
		}
	}
}

func main() {
	startTime := time.Now()
	flag.Parse()

	var index = 0
	for index < *cmdParamsRuntimes {
		go doLoginCheck(index)
		processWaitGroup.Add(1)
		index++
	}
	processWaitGroup.Wait()
	stopTime := time.Now()
	fmt.Printf("LoginCheck test over! startTime[%v] stopTime[%v]\n", startTime, stopTime)
}

/*
serialData len[229] compressBuf len[189]
[{FailureReason:Interface Name is invalid!}]

serialData len[229] compressBuf len[189]
[{FailureReason:Token is invalid!}]

serialData len[269] compressBuf len[215]
[{FailureReason:ClientVersion is invalid!}]

serialData len[268] compressBuf len[213]
[{NewVersion:1.1.1 ChannelName:GooglePlay UpdateUrl:http://googleplayer}]

[&{{1.0.0 1494665990 998 3} {LoginCheck map[Bulletin:服务器维护中。。。。。。]}}]
公告信息：[服务器维护中。。。。。。]

[&{{1.0.0 1494666074 998 1} {LoginCheck <nil>}}]
Forbit login!!!!

[2017/05/13 17:03:28 CST] [WARN] (sandmonk_main/operational.(*OperationalModule).LoginCheck:132) ChannelID[8119] not in WhiteList, Cannot login!!

黑名单用户
[&{{1.0.0 1494666658 998 1} {LoginCheck <nil>}}]
Forbit login!!!!


*/
