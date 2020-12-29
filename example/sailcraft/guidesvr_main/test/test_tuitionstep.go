/*
 * @Author: calmwu
 * @Date: 2018-01-11 16:54:31
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-01-11 16:56:05
 * @Comment:
 */

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
)

var (
	cmdParamsSvrIP    = flag.String("svrip", "123.59.40.19", "")
	cmdParamsRuntimes = flag.Int("runtimes", 1, "")
	cmdParamsSSL      = flag.Int("ssl", 0, "")
	cmdMethod         = flag.String("method", "post", "post/put")

	processWaitGroup sync.WaitGroup
)

const (
	UrlTuitionStepFmt  = "http://%s:808/sailcraft/api/v1/GuideSvr/TuitionStepReport"
	UrlTuitionStepFmtS = "https://%s:808//sailcraft/api/v1/GuideSvr/TuitionStepReport"
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
		UrlTuitionStep := fmt.Sprintf(UrlTuitionStepFmt, *cmdParamsSvrIP)

		res, err = http.Post(UrlTuitionStep, "text/plain; charset=utf-8", &compressBuf)
		if err != nil {
			fmt.Printf("Post to %s failed! [%s]\n", UrlTuitionStep, err.Error())
			return err, nil
		}
	} else if *cmdParamsSSL == 2 {
		UrlTuitionStep := fmt.Sprintf(UrlTuitionStepFmtS, *cmdParamsSvrIP)

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
		UrlTuitionStep := fmt.Sprintf(UrlTuitionStepFmtS, *cmdParamsSvrIP)

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

	UrlTuitionStep2 := fmt.Sprintf(UrlTuitionStepFmt, *cmdParamsSvrIP)

	var client *http.Client = nil
	if *cmdParamsSSL == 0 {
		client = &http.Client{}
	} else if *cmdParamsSSL == 2 {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr}
		UrlTuitionStep2 = fmt.Sprintf(UrlTuitionStepFmtS, *cmdParamsSvrIP)
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

func doTuitionStepReport(step int, index int) {
	defer processWaitGroup.Done()

	req := proto.ProtoSandMonkReqS{
		ProtoSandMonkHeadReqS: proto.ProtoSandMonkHeadReqS{
			Version:    1,
			EventId:    998,
			TimeStamp:  time.Now().Unix(),
			ChannelUID: "21312",
			Uin:        65535 + index,
			CsrfToken:  "02cf14994a3be74301657dbcd9c0a189",
		},
		ReqData: proto.ProtoSandMonkData{
			InterfaceName: "TuitionStepReport",
			Params: map[string]interface{}{
				"ClientVersion": "9.1.1",
				"StepId":        step,
				"PlatformName":  "Ios",
				"ChannelName":   "AppStore",
			},
		},
	}

	var res *proto.ProtoSandMonkResS = nil
	if *cmdMethod == "post" {
		_, res = PostRequest(&req)
	} else if *cmdMethod == "put" {
		_, res = PutRequest(&req)
	}

	fmt.Printf("[%v]\n", res)
}

func main() {
	startTime := time.Now()
	flag.Parse()

	stepAry := [5]int{1, 4, 11, 15, 23}
	var index = 0
	for index < *cmdParamsRuntimes {
		for _, v := range stepAry {
			fmt.Println(v)
			go doTuitionStepReport(v, index)
			processWaitGroup.Add(1)
		}
		index++
	}
	processWaitGroup.Wait()
	stopTime := time.Now()
	fmt.Printf("doTuitionStepReport test over! startTime[%v] stopTime[%v]\n", startTime, stopTime)
}
