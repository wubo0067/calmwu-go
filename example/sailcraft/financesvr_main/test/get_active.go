/*
 * @Author: calmwu
 * @Date: 2018-03-30 16:56:02
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-04-13 10:45:03
 * @Comment:
 */
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"sailcraft/base"
	"sailcraft/financesvr_main/proto"
	"strings"
	"time"
)

const (
	UrlGetPlayerActiveFmt = "http://%s:400/sailcraft/api/v1/FinanceSvr/GetPlayerActive"
)

var (
	cmdParamsSvrIp  = flag.String("svrip", "123.59.40.19", "")
	cmdParamsUin    = flag.Int("uin", 1, "")
	cmdParamsZoneID = flag.Int("zoneid", 1, "")
	cmdParamType    = flag.Int("type", 0, "")
)

type OurCustomTransport struct {
	//Transport http.RoundTripper
}

func (t *OurCustomTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	dump, _ := httputil.DumpRequestOut(req, true)
	fmt.Printf("+++++++%q\n", dump)

	return http.DefaultTransport.RoundTrip(req)
}

func SendRequest(url string, req *base.ProtoRequestS) {
	// 编码压缩
	serialData, err := json.Marshal(req)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Printf("url[%s] serialData len[%d]\n", url, len(serialData))

	t := &OurCustomTransport{}
	httpClient := &http.Client{Transport: t}
	httpReq, _ := http.NewRequest("POST", url, strings.NewReader(string(serialData)))
	httpReq.Header.Set("Content-Type", "text/plain; charset=utf-8")
	// dump, _ := httputil.DumpRequestOut(httpReq, true)
	// fmt.Printf("%q\n", dump)
	res, _ := httpClient.Do(httpReq)

	dump, _ := httputil.DumpResponse(res, true)
	fmt.Printf("------%q\n", dump)

	// 发送
	// res, err := http.Post(url, "text/plain; charset=utf-8", strings.NewReader(string(serialData)))
	// if err != nil {
	// 	fmt.Printf("Post to %s failed! [%s]\n", url, err.Error())
	// 	return
	// }
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("Read body failed! reason[%s]\n", err.Error())
		return
	}
	fmt.Printf("%s\n", body)
}

func main() {
	flag.Parse()

	var realReq interface{}
	var interfaceName string
	var urlFmt string
	channelID := "CN"
	if *cmdParamType != 0 {
		channelID = "NOAREA"
	}

	realReq = proto.ProtoGetPlayerActiveReq{
		Uin:        uint64(*cmdParamsUin),
		ZoneID:     int32(*cmdParamsZoneID),
		ActiveType: proto.ActiveType(*cmdParamType),
		ChannelID:  channelID,
	}
	interfaceName = "GetPlayerActive"
	urlFmt = UrlGetPlayerActiveFmt

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
			InterfaceName: interfaceName,
			Params:        realReq,
		},
	}

	UrlQuery := fmt.Sprintf(urlFmt, *cmdParamsSvrIp)
	SendRequest(UrlQuery, &req)
}
