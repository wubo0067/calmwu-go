/*
 * @Author: calmwu
 * @Date: 2018-05-02 14:10:35
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-05-02 14:24:09
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
	"sailcraft/base"
	"time"
)

const (
	UrlLoginCheckFmt  = "http://%s/sailcraft/api/v1/GuideSvr/LoginCheck"
	UrlLoginCheckFmtS = "https://%s/sailcraft/api/v1/GuideSvr/LoginCheck"
)

var (
	cmdParamsSvrIP   = flag.String("svrip", "123.59.40.19:800", "")
	cmdParamsUin     = flag.Int("uin", 1, "")
	cmdParamsVersion = flag.String("version", "1.2.2", "")
	cmdParamsSSL     = flag.Int("ssl", 0, "")
	cmdMethod        = flag.String("method", "post", "post/put")
)

func PostRequest(req *base.ProtoRequestS) (error, *base.ProtoResponseS) {
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

	var resPkg base.ProtoResponseS
	err = json.NewDecoder(dcompressReader).Decode(&resPkg)
	if err != nil {
		fmt.Printf("Decode failed! [%s]\n", err.Error())
		return err, nil
	}

	return nil, &resPkg
}

func PutRequest(req *base.ProtoRequestS) (error, *base.ProtoResponseS) {
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

	var resPkg base.ProtoResponseS
	err = json.NewDecoder(dcompressReader).Decode(&resPkg)
	if err != nil {
		fmt.Printf("Decode failed! [%s]\n", err.Error())
		return err, nil
	}

	return nil, &resPkg
}

func main() {
	flag.Parse()

	req := base.ProtoRequestS{
		ProtoRequestHeadS: base.ProtoRequestHeadS{
			Version:    1,
			EventId:    998,
			TimeStamp:  time.Now().Unix(),
			ChannelUID: "21312",
			Uin:        *cmdParamsUin,
			CsrfToken:  "02cf14994a3be74301657dbcd9c0a189",
		},
		ReqData: base.ProtoData{
			InterfaceName: "LoginCheck",
			Params: map[string]interface{}{
				"ClientVersion": "1.2.2",
				"PlatformName":  "Android",
				"ChannelName":   "GooglePlay",
			},
		},
	}

	var res *base.ProtoResponseS = nil
	if *cmdMethod == "post" {
		_, res = PostRequest(&req)
	} else if *cmdMethod == "put" {
		_, res = PutRequest(&req)
	}

	fmt.Printf("[%v]\n", res)
}
