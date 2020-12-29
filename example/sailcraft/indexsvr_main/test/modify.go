/*
 * @Author: calmwu
 * @Date: 2017-09-23 10:57:57
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-10-11 10:34:53
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

// .\modify.exe --type=user --old=Captain66093 --new=kingsinger --id=66093

const (
	UrlModifyGuildFmt = "http://%s/sailcraft/api/v1/IndexSvr/ModifyGuildName"
	UrlModifyUserFmt  = "http://%s/sailcraft/api/v1/IndexSvr/ModifyUserName"
)

var (
	cmdParamsSvrIp      = flag.String("svrip", "123.59.40.19:505", "")
	cmdParamsOld        = flag.String("old", "", "")
	cmdParamsNew        = flag.String("new", "", "")
	cmdParamsID         = flag.String("id", "", "")
	cmdParamsModifyType = flag.String("type", "guild", "guild/user")
)

func SendRequest(url string, req *base.ProtoRequestS) {
	// 编码压缩
	serialData, err := json.Marshal(req)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Printf("serialData len[%d]\n", len(serialData))

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

	if *cmdParamsModifyType == "user" {
		UrlQuery := fmt.Sprintf(UrlModifyUserFmt, *cmdParamsSvrIp)

		req := base.ProtoRequestS{
			ProtoRequestHeadS: base.ProtoRequestHeadS{
				Version:    1,
				EventId:    998,
				TimeStamp:  time.Now().Unix(),
				ChannelUID: "21312",
				Uin:        65540,
				CsrfToken:  "02cf14994a3be74301657dbcd9c0a189",
			},
			ReqData: base.ProtoData{
				InterfaceName: "ModifyUserName",
				Params: map[string]interface{}{
					"UserName":    *cmdParamsOld,
					"Uin":         *cmdParamsID,
					"NewUserName": *cmdParamsNew,
				},
			},
		}

		SendRequest(UrlQuery, &req)
	} else if *cmdParamsModifyType == "guild" {
		UrlQuery := fmt.Sprintf(UrlModifyGuildFmt, *cmdParamsSvrIp)

		req := base.ProtoRequestS{
			ProtoRequestHeadS: base.ProtoRequestHeadS{
				Version:    1,
				EventId:    998,
				TimeStamp:  time.Now().Unix(),
				ChannelUID: "21312",
				Uin:        65540,
				CsrfToken:  "02cf14994a3be74301657dbcd9c0a189",
			},
			ReqData: base.ProtoData{
				InterfaceName: "ModifyGuildName",
				Params: map[string]interface{}{
					"GuildName":    *cmdParamsOld,
					"ID":           *cmdParamsID,
					"NewGuildName": *cmdParamsNew,
				},
			},
		}

		SendRequest(UrlQuery, &req)
	}
}
