/*
 * @Author: calmwu
 * @Date: 2018-04-16 11:17:43
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-04-16 11:27:21
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
	UrlAddVIPFmt = "http://%s:400/sailcraft/api/v1/FinanceSvr/PlayerExchangeCDKey"
)

var (
	cmdParamsSvrIp  = flag.String("svrip", "123.59.40.19", "")
	cmdParamsUin    = flag.Int("uin", 1, "")
	cmdParamsZoneID = flag.Int("zoneid", 1, "")
	cmdCDKey        = flag.String("cdkey", "sailcraft2", "")
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
			InterfaceName: "PlayerExchangeCDKey",
			Params: map[string]interface{}{
				"Uin":    uint64(*cmdParamsUin),
				"ZoneID": int32(*cmdParamsZoneID),
				"CDKey":  *cmdCDKey,
			},
		},
	}

	UrlQuery := fmt.Sprintf(UrlAddVIPFmt, *cmdParamsSvrIp)
	SendRequest(UrlQuery, &req)
}
