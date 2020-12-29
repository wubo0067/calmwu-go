/*
 * @Author: calmwu
 * @Date: 2018-03-23 15:24:58
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-03-23 15:59:37
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
	MonthSignInConfigFile          = "Active/ActiveDailySign.json"
	VipDoubleConfigFile            = "Active/PrivilegeCardDouble.json"
	UrlRefreshMonthSignInConfigFmt = "http://%s/sailcraft/api/v1/FinanceSvr/GMUpdateMonthlySigninConfigInfo"
)

// type si_prize struct {
// 	Count     int32  `json:"count"`
// 	CountType string `json:"count_type"`
// 	Type      string `json:"type"`
// }

// type si_prize_list struct {
// 	Resources []si_prize `json:"resources"`
// }

type si_dailysignininfo struct {
	PrizeID int         `json:"Id"`
	Reward  interface{} `json:"Reward"`
}

type si_vipdoubledays struct {
	ViMPDays []int32 `json:"DailySignDouble"`
}

func configMonthSignInActive(configPath string) {
	fileFullName := configPath + "/" + MonthSignInConfigFile
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

	dailySignInfos := make([]si_dailysignininfo, 0)
	err = json.Unmarshal(data, &dailySignInfos)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unmarshal %s data failed! reason[%s]", fileFullName, err.Error())
		return
	}

	//---------------------------------------------------------------------------------------------

	fileFullName = configPath + "/" + VipDoubleConfigFile
	vip_file, err := os.Open(fileFullName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open [%s] failed! err[%s]\n", fileFullName, err.Error())
		return
	}
	defer vip_file.Close()

	data, err = ioutil.ReadAll(vip_file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read %s failed, reason:%s:\n", fileFullName, err.Error())
		return
	}

	vipdays := make([]si_vipdoubledays, 0)
	err = json.Unmarshal(data, &vipdays)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unmarshal %s data failed! reason[%s]", fileFullName, err.Error())
		return
	}

	fmt.Printf("vipdays:%+v\n", vipdays[0].ViMPDays)

	var configReq proto.ProtoGMConfigMonthlySignInReq
	configReq.Uin = *cmdParamsUin
	configReq.ZoneID = int32(*cmdParamsZoneID)
	configReq.MonthlySignInConfig.VipMultiplePrizeDays = vipdays[0].ViMPDays
	configReq.MonthlySignInConfig.VipMultipleNum = 2
	configReq.MonthlySignInConfig.RessiueActivityThreshold = 80
	configReq.MonthlySignInConfig.PrizeLst = make([]proto.SignInPrizeS, len(dailySignInfos))

	for index, _ := range dailySignInfos {
		configReq.MonthlySignInConfig.PrizeLst[index].PrizeID = dailySignInfos[index].PrizeID

		signinPrizeJC, err := json.Marshal(dailySignInfos[index].Reward)
		if err != nil {
			fmt.Fprintf(os.Stderr, "PrizeId[%d] marshal Reward failed! reason[%s]",
				dailySignInfos[index].PrizeID, err.Error())
			os.Exit(-1)
		}

		configReq.MonthlySignInConfig.PrizeLst[index].PrizeJsonContent = string(signinPrizeJC)
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
			InterfaceName: "GMUpdateMonthlySigninConfigInfo",
			Params:        configReq,
		},
	}

	UrlQuery := fmt.Sprintf(UrlRefreshMonthSignInConfigFmt, *cmdParamsSvrIp)
	SendRequest(UrlQuery, &req)
	return
}
