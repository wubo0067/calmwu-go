/*
 * @Author: calmwu
 * @Date: 2018-02-06 16:09:07
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-02-06 16:26:13
 * @Comment:
 */

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"sailcraft/base"
	"strings"
	"time"
)

const (
	UrlNewFinanceUserFmt = "http://%s:400/sailcraft/api/v1/FinanceSvr/NewFinanceUser"
)

var (
	cmdParamsSvrIp  = flag.String("svrip", "123.59.40.19", "")
	cmdParamsUin    = flag.Int("uin", 1, "")
	cmdParamsZoneID = flag.Int("zoneid", 1, "")
	cmdParamsCount  = flag.Int("count", 1, "")
)

func SendRequest(url string, req *base.ProtoRequestS) {
	// 编码压缩
	serialData, err := json.Marshal(req)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Printf("url[%s] serialData len[%d]\n", url, len(serialData))

	// 发送
	res, err := http.Post(url, "text/plain; charset=utf-8", strings.NewReader(string(serialData)))
	if err != nil {
		fmt.Printf("Post to %s failed! [%s]\n", url, err.Error())
		return
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("Read body failed! reason[%s]\n", err.Error())
		return
	}
	fmt.Printf("%s\n", body)
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
			InterfaceName: "NewFinanceUser",
			Params: map[string]interface{}{
				"Uin":      uint64(*cmdParamsUin),
				"ZoneID":   int32(*cmdParamsZoneID),
				"TimeZone": "Asia/Shanghai",
			},
		},
	}

	UrlQuery := fmt.Sprintf(UrlNewFinanceUserFmt, *cmdParamsSvrIp)
	SendRequest(UrlQuery, &req)
}
