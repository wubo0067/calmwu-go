/*
 * @Author: calmwu
 * @Date: 2017-09-21 15:26:19
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-10-11 10:34:36
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

// .\find.exe --type=user --match=like --name=66093
// .\find.exe --type=user --match=match --name=Captain66093

const (
	UrlQueryGuildFmt = "http://%s/sailcraft/api/v1/IndexSvr/FindGuidsByName"
	UrlQueryUserFmt  = "http://%s/sailcraft/api/v1/IndexSvr/FindUsersByName"
)

var (
	cmdParamsSvrIp      = flag.String("svrip", "123.59.40.19:505", "")
	cmdParamsName       = flag.String("name", "", "")
	cmdParamsQueryType  = flag.String("type", "guild", "guild/user")
	cmdParamsMatchType  = flag.String("match", "match", "match/like")
	cmdParamsQueryCount = flag.Int("count", 1, "")
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

	if *cmdParamsQueryType == "user" {
		UrlQuery := fmt.Sprintf(UrlQueryUserFmt, *cmdParamsSvrIp)

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
				InterfaceName: "FindGuidsByName",
				Params: map[string]interface{}{
					"UserName":   *cmdParamsName,
					"QueryType":  *cmdParamsMatchType,
					"QueryCount": *cmdParamsQueryCount,
				},
			},
		}

		SendRequest(UrlQuery, &req)
	} else if *cmdParamsQueryType == "guild" {
		UrlQuery := fmt.Sprintf(UrlQueryGuildFmt, *cmdParamsSvrIp)

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
				InterfaceName: "FindGuidsByName",
				Params: map[string]interface{}{
					"GuildName":  *cmdParamsName,
					"QueryType":  *cmdParamsMatchType,
					"QueryCount": *cmdParamsQueryCount,
				},
			},
		}

		SendRequest(UrlQuery, &req)
	}
}
