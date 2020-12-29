/*
 * @Author: calmwu
 * @Date: 2018-05-19 15:21:33
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-05-19 16:27:27
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
	UrlMaintainFmt = "http://%s/sailcraft/api/v1/GuideSvr/SetMaintainInfo"
)

var (
	cmdParamsSvrIp = flag.String("svrip", "123.59.40.19:8000", "")
	cmdParamsUin   = flag.Int("uin", 1, "")
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
			InterfaceName: "SetMaintainInfo",
			Params: map[string]interface{}{
				"gm_tool_game_maintain_key": map[string]interface{}{
					"white_flag":         1,
					"game_maintain_flag": 0,
					"white_list":         []string{"6497b35f-3e69-4481-9987-a737b62bea61", "21312", "698742", "asdasd-sdfs-sdfs"},
					"main_dead_line":     1498740020,
				},
			},
		},
	}

	UrlQuery := fmt.Sprintf(UrlMaintainFmt, *cmdParamsSvrIp)
	SendRequest(UrlQuery, &req)
}
