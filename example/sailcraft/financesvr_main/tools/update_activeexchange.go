/*
 * @Author: calmwu
 * @Date: 2018-04-02 16:02:00
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-04-16 12:31:06
 * @Comment:
 */

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sailcraft/base"
	"sailcraft/financesvr_main/proto"
	"time"
)

const (
	ActiveExchangeConfigFile   = "Active/ActiveExchange.json"
	UrlActiveExchangeConfigFmt = "http://%s/sailcraft/api/v1/FinanceSvr/GMConfigExchangeActive"
)

type exchangeLimit struct {
	LimitTimes int32  `json:"limit_times"`
	LimitType  string `json:"limit_type"`
}

type exchangeInfo struct {
	ActiveID int          `json:"Id"`
	Limit    missionLimit `json:"Limit"`
	Cost     interface{}  `json:"Cost"`
	Reward   interface{}  `json:"Reward"`
	TitleKey string       `json:"TitleKey"`
}

type ExchangeData struct {
	Cost   string `json:"Cost"`
	Reward string `json:"Reward"`
}

func configActiveExchange(configPath string) {
	fileFullName := configPath + "/" + ActiveExchangeConfigFile
	conf_file, err := os.Open(fileFullName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open [%s] failed! err[%s]\n", fileFullName, err.Error())
		return
	}
	defer conf_file.Close()

	data, err := ioutil.ReadAll(conf_file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read %s failed, reason:%s:\n", fileFullName, err.Error())
		return
	}

	exchanges := make([]exchangeInfo, 0)
	err = json.Unmarshal(data, &exchanges)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unmarshal %s data failed! reason[%s]", fileFullName, err.Error())
		return
	}

	fmt.Printf("missions:%+v\n", exchanges)

	var configReq proto.ProtoGMConfigActiveExchangeReq
	configReq.Uin = *cmdParamsUin
	configReq.ZoneID = int32(*cmdParamsZoneID)
	for i := range exchanges {
		exchange := &exchanges[i]

		activeExchange := new(proto.ActiveExchangeInfoS)
		activeExchange.Base.ActiveID = exchange.ActiveID
		activeExchange.Base.ChannelID = "NOAREA"
		activeExchange.Base.ReceiveCond = 0
		activeExchange.Base.ReceiveLimit = exchange.Limit.LimitTimes

		if exchange.Limit.LimitType == "daily" {
			activeExchange.Base.ResetEveryDay = 1
		} else {
			activeExchange.Base.ResetEveryDay = 0
		}

		jc, err := json.Marshal(exchange.Reward)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ActiveID[%d] marshal InnerGoods failed! reason[%s]",
				exchange.ActiveID, err.Error())
			os.Exit(-1)
		}

		activeExchange.InnerGoods = string(jc)

		jc, err = json.Marshal(exchange.Cost)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ActiveID[%d] marshal Cost failed! reason[%s]",
				exchange.ActiveID, err.Error())
			os.Exit(-1)
		}

		activeExchange.ExchangeCost = string(jc)
		activeExchange.TitleKey = exchange.TitleKey

		configReq.ActiveExchanges = append(configReq.ActiveExchanges, *activeExchange)
	}

	req := base.ProtoRequestS{
		ProtoRequestHeadS: base.ProtoRequestHeadS{
			Version:    1,
			EventId:    998,
			TimeStamp:  time.Now().Unix(),
			ChannelUID: "21312",
			Uin:        int(*cmdParamsUin),
			CsrfToken:  "02cf14994a3be74301657dbcd9c0a189",
		},
		ReqData: base.ProtoData{
			InterfaceName: "GMConfigExchangeActive",
			Params:        configReq,
		},
	}

	UrlQuery := fmt.Sprintf(UrlActiveExchangeConfigFmt, *cmdParamsSvrIp)
	SendRequest(UrlQuery, &req)
	return
}
