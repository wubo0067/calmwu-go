/*
 * @Author: calmwu
 * @Date: 2018-03-30 16:32:56
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-04-14 11:25:39
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
	"sailcraft/financesvr_main/proto"
	"strings"
	"time"
)

const (
	UrlOpenActiveFmt  = "http://%s/sailcraft/api/v1/FinanceSvr/OpenActive"
	UrlCloseActiveFmt = "http://%s/sailcraft/api/v1/FinanceSvr/CloseActive"
)

var (
	cmdParamsSvrIp      = flag.String("svrip", "123.59.40.19:400", "")
	cmdParamsUin        = flag.Int("uin", 1, "")
	cmdParamsZoneID     = flag.Int("zoneid", 1, "")
	cmdParamType        = flag.String("type", "open", "open/close")
	cmdParamsActiveID   = flag.Int("activeid", 1001, "")
	cmdParamsActiveType = flag.Int("activetype", 0, "0,1,2,3")
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

	var realReq interface{}
	var interfaceName string
	var urlFmt string

	channelID := proto.C_CHANNELNAME_NOAREA
	if *cmdParamsActiveType == 0 {
		channelID = proto.C_CHANNELNAME_CN
	}

	if *cmdParamType == "open" {
		realReq = proto.ProtoOpenActiveReq{
			Uin:    uint64(*cmdParamsUin),
			ZoneID: int32(*cmdParamsZoneID),
			ActiveControlConfigs: []proto.ProtoActiveControlInfoS{
				proto.ProtoActiveControlInfoS{
					ActiveType:   proto.ActiveType(*cmdParamsActiveType),
					ActiveID:     *cmdParamsActiveID,
					ChannelID:    channelID,
					StartTime:    time.Now().Unix(),
					DurationSecs: 3600 * 24 * 3,
				},
			},
		}
		interfaceName = "OpenActive"
		urlFmt = UrlOpenActiveFmt
	}

	if *cmdParamType == "close" {
		realReq = proto.ProtoCloseActiveReq{
			Uin:        uint64(*cmdParamsUin),
			ZoneID:     int32(*cmdParamsZoneID),
			ActiveType: proto.ActiveType(*cmdParamsActiveType),
			ActiveIDs:  []int{*cmdParamsActiveID},
		}
		interfaceName = "CloseActive"
		urlFmt = UrlCloseActiveFmt
	}

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
