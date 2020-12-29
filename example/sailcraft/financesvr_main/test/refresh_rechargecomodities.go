/*
 * @Author: calmwu
 * @Date: 2018-02-06 16:41:29
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-02-06 17:48:06
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
	UrlRefreshRechargeCommoditiesFmt = "http://%s:400/sailcraft/api/v1/FinanceSvr/RefreshRechargeCommodities"
)

var (
	cmdParamsSvrIp   = flag.String("svrip", "123.59.40.19", "")
	cmdParamsUin     = flag.Int("uin", 1, "")
	cmdParamsZoneID  = flag.Int("zoneid", 1, "")
	cmdParamsVersion = flag.String("version", "1.0.0", "")
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

	var refreshReq proto.ProtoRefreshRechargeCommoditiesReq
	refreshReq.Uin = uint64(*cmdParamsUin)
	refreshReq.ZoneID = int32(*cmdParamsZoneID)

	refreshReq.ShopRechargeCommoditiesInfo.RechargeCommodities = make([]proto.RechargeCommodityInfoS, 6)
	refreshReq.ShopRechargeCommoditiesInfo.Count = 6
	refreshReq.ShopRechargeCommoditiesInfo.VersionID = *cmdParamsVersion

	i := 0
	for i < 6 {
		refreshReq.ShopRechargeCommoditiesInfo.RechargeCommodities[i].RechargeCommodityID = i
		refreshReq.ShopRechargeCommoditiesInfo.RechargeCommodities[i].ExchangeDiamonds = int32(100 * i)
		refreshReq.ShopRechargeCommoditiesInfo.RechargeCommodities[i].FirstRechargePresentDiamonds = int32(200 * i)
		refreshReq.ShopRechargeCommoditiesInfo.RechargeCommodities[i].PresentDiamonds = int32(200 * i)
		i++
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
			InterfaceName: "RefreshRechargeCommodities",
			Params:        refreshReq,
		},
	}

	UrlQuery := fmt.Sprintf(UrlRefreshRechargeCommoditiesFmt, *cmdParamsSvrIp)
	SendRequest(UrlQuery, &req)
}
