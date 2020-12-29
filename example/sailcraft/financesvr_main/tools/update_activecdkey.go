/*
 * @Author: calmwu
 * @Date: 2018-04-14 11:06:13
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-04-14 11:23:34
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
	ActiveCDKeyConfigFile           = "Active/ActiveCDKey.json"
	UrlActiveCDKeyExchangeConfigFmt = "http://%s/sailcraft/api/v1/FinanceSvr/GMConfigCDKeyExchangeActive"
)

type cdkeyInfo struct {
	ActiveID int         `json:"Id"`
	CDKey    string      `json:"Password"`
	Reward   interface{} `json:"Reward"`
}

func configActiveCDKeyExchange(configPath string) {
	fileFullName := configPath + "/" + ActiveCDKeyConfigFile
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

	cdkeys := make([]cdkeyInfo, 0)
	err = json.Unmarshal(data, &cdkeys)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unmarshal %s data failed! reason[%s]", fileFullName, err.Error())
		return
	}

	//fmt.Printf("missions:%+v\n", cdkeys)
	var configReq proto.ProtoGMConfigActiveCDKeyExchangeReq
	configReq.Uin = *cmdParamsUin
	configReq.ZoneID = int32(*cmdParamsZoneID)
	for i := range cdkeys {
		cdkey := &cdkeys[i]

		activeCDKey := new(proto.ActiveCDKeyExchangeInfoS)
		activeCDKey.Base.ActiveID = cdkey.ActiveID
		activeCDKey.Base.ChannelID = "NOAREA"
		activeCDKey.Base.ReceiveCond = 0
		activeCDKey.Base.ReceiveLimit = 0
		activeCDKey.Base.ResetEveryDay = 0

		activeCDKey.CDKey = cdkey.CDKey

		jc, err := json.Marshal(cdkey.Reward)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ActiveID[%d] marshal InnerGoods failed! reason[%s]",
				cdkey.ActiveID, err.Error())
			os.Exit(-1)
		}

		activeCDKey.InnerGoods = string(jc)

		configReq.ActiveCDKeyExchanges = append(configReq.ActiveCDKeyExchanges, *activeCDKey)
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
			InterfaceName: "GMConfigCDKeyExchangeActive",
			Params:        configReq,
		},
	}

	UrlQuery := fmt.Sprintf(UrlActiveCDKeyExchangeConfigFmt, *cmdParamsSvrIp)
	SendRequest(UrlQuery, &req)
	return
}
