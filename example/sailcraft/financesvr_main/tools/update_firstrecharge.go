/*
 * @Author: calmwu
 * @Date: 2018-04-16 17:24:59
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-04-16 17:40:27
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
	ActiveFirstRechargeConfigFile   = "Active/ActiveFirstRecharge.json"
	UrlActiveFirstRechargeConfigFmt = "http://%s/sailcraft/api/v1/FinanceSvr/GMConfigFirstRecharge"
)

type firstRechargeInfo struct {
	ActiveID int         `json:"Id"`
	Target   int32       `json:"Target"`
	Reward   interface{} `json:"Reward"`
	TitleKey string      `json:"TitleKey"`
	Value    int32       `json:"Value"`
}

func configActiveFirstRecharge(configPath string) {
	fileFullName := configPath + "/" + ActiveFirstRechargeConfigFile
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

	firstRechargeDatas := make([]firstRechargeInfo, 0)
	err = json.Unmarshal(data, &firstRechargeDatas)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unmarshal %s data failed! reason[%s]", fileFullName, err.Error())
		return
	}

	fmt.Printf("firstRecharge:%+v\n", firstRechargeDatas)

	var configReq proto.ProtoGMConfigFirstRechargeReq
	configReq.Uin = *cmdParamsUin
	configReq.ZoneID = int32(*cmdParamsZoneID)

	for i := range firstRechargeDatas {
		frData := &firstRechargeDatas[i]
		frLevel := new(proto.ProtoFirstRechargeLevelConfS)

		frLevel.Id = frData.ActiveID
		frLevel.Target = frData.Target
		frLevel.TitleKey = frData.TitleKey
		frLevel.Value = frData.Value

		jc, err := json.Marshal(frData.Reward)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ActiveID[%d] marshal InnerGoods failed! reason[%s]",
				frData.ActiveID, err.Error())
			os.Exit(-1)
		}

		frLevel.Reward = string(jc)

		configReq.Config.FRLevelConfLst = append(configReq.Config.FRLevelConfLst, *frLevel)
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
			InterfaceName: "GMConfigFirstRecharge",
			Params:        configReq,
		},
	}

	UrlQuery := fmt.Sprintf(UrlActiveFirstRechargeConfigFmt, *cmdParamsSvrIp)
	SendRequest(UrlQuery, &req)
	return
}
