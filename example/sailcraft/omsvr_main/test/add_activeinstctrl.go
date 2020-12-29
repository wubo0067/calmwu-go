/*
 * @Author: calmwu
 * @Date: 2018-05-18 14:37:37
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-05-19 14:10:53
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
	"sailcraft/omsvr_main/proto"
	"strings"
	"time"
)

const (
	UrlAddActiveInstsCtrlFmt      = "http://%s/sailcraft/api/v1/OMSvr/AddActiveInsts"
	UrlReloadActiveInstsCtrlFmt   = "http://%s/sailcraft/api/v1/OMSvr/ReloadActiveInsts"
	UrlQueryRunningActiveTypesFmt = "http://%s/sailcraft/api/v1/OMSvr/QueryRunningActiveTypes"
)

var (
	cmdParamsSvrIp  = flag.String("svrip", "123.59.40.19:2000", "")
	cmdParamsUin    = flag.Int("uin", 1, "")
	cmdParamsZoneID = flag.Int("zoneid", 1000, "")
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

	now := time.Now()
	startTime := now.Add(time.Minute).Format("2006-01-02 15:04:05")
	fmt.Println(startTime)

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
			InterfaceName: "AddActiveInsts",
			Params: proto.ProtoAddActiveInstCtrlsReq{
				Uin: uint64(*cmdParamsUin),
				ActiveInstCtrls: []proto.ProtoActiveInstControlS{
					proto.ProtoActiveInstControlS{
						ZoneID:          *cmdParamsZoneID,
						ActiveType:      1,
						ActiveID:        1,
						StartTimeName:   startTime,
						DurationMinutes: 1,
						ChannelName:     "NOAREA",
						TimeZone:        "Local",
						GroupID:         -1,
					},
					proto.ProtoActiveInstControlS{
						ZoneID:          *cmdParamsZoneID,
						ActiveType:      2,
						ActiveID:        2,
						StartTimeName:   startTime,
						DurationMinutes: 2,
						ChannelName:     "NOAREA",
						TimeZone:        "Local",
						GroupID:         -1,
					},
					proto.ProtoActiveInstControlS{
						ZoneID:          *cmdParamsZoneID,
						ActiveType:      3,
						ActiveID:        3,
						StartTimeName:   startTime,
						DurationMinutes: 3,
						ChannelName:     "NOAREA",
						TimeZone:        "Local",
						GroupID:         -1,
					},
				},
			},
		},
	}

	UrlQuery := fmt.Sprintf(UrlAddActiveInstsCtrlFmt, *cmdParamsSvrIp)
	SendRequest(UrlQuery, &req)

	req = base.ProtoRequestS{
		ProtoRequestHeadS: base.ProtoRequestHeadS{
			Version:    1,
			EventId:    998,
			TimeStamp:  time.Now().Unix(),
			ChannelUID: "21312",
			Uin:        *cmdParamsUin,
			CsrfToken:  "02cf14994a3be74301657dbcd9c0a189",
		},
		ReqData: base.ProtoData{
			InterfaceName: "ReloadActiveInsts",
			Params: proto.ProtoLoadWatingActiveInstCtrlsReq{
				Uin: uint64(*cmdParamsUin),
			},
		},
	}

	UrlQuery = fmt.Sprintf(UrlReloadActiveInstsCtrlFmt, *cmdParamsSvrIp)
	SendRequest(UrlQuery, &req)

	time.Sleep(70 * time.Second)

	req = base.ProtoRequestS{
		ProtoRequestHeadS: base.ProtoRequestHeadS{
			Version:    1,
			EventId:    998,
			TimeStamp:  time.Now().Unix(),
			ChannelUID: "21312",
			Uin:        *cmdParamsUin,
			CsrfToken:  "02cf14994a3be74301657dbcd9c0a189",
		},
		ReqData: base.ProtoData{
			InterfaceName: "QueryRunningActiveTypes",
			Params: proto.ProtoQueryRunningActiveTypesReq{
				Uin:    uint64(*cmdParamsUin),
				ZoneID: int32(*cmdParamsZoneID),
			},
		},
	}

	UrlQuery = fmt.Sprintf(UrlQueryRunningActiveTypesFmt, *cmdParamsSvrIp)
	SendRequest(UrlQuery, &req)
}
