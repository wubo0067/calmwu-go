/*
 * @Author: calmwu
 * @Date: 2018-03-27 15:55:31
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-03-27 16:23:13
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
	VipPrivilegeConfigFile   = "Active/ActivePrivilegeCard.json"
	UrlVipPrivilegeConfigFmt = "http://%s/sailcraft/api/v1/FinanceSvr/GMConfigVIPPrivilege"
)

type vip_card struct {
	ChannelID string  `json:"AreaCode"`
	Id        int     `json:"Id"`
	DailyGem  int32   `json:"DailyGem"`
	Duration  int32   `json:"Duration"`
	GemCount  int32   `json:"GemCount"`
	Price     float32 `json:"Price"`
	PriceDesc string  `json:"PriceDesc"`
	ProductId string  `json:"ProductId"`
	Type      string  `json:"Type"`
	NameKey   string  `json:"NameKey"`
}

type vip_cards struct {
	Cards []vip_card `json:"Cards"`
}

func configVIPPrivilege(configPath string) {
	fileFullName := configPath + "/" + VipPrivilegeConfigFile
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

	vipPrivileges := make([]vip_cards, 0)
	err = json.Unmarshal(data, &vipPrivileges)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unmarshal %s data failed! reason[%s]", fileFullName, err.Error())
		return
	}

	//fmt.Printf("vipPrivileges:%+v\n", vipPrivileges)
	var configReq proto.ProtoGMConfigVIPPrivilegeReq
	configReq.Uin = *cmdParamsUin
	configReq.ZoneID = int32(*cmdParamsZoneID)
	configReq.VIPPrivilegeConfig.VIPPrivilegeInfos = make([]proto.ProtoVIPPrivilegeInfoS, 4)

	for index, _ := range configReq.VIPPrivilegeConfig.VIPPrivilegeInfos {
		configReq.VIPPrivilegeConfig.VIPPrivilegeInfos[index].Id =
			vipPrivileges[index/2].Cards[index%2].Id
		configReq.VIPPrivilegeConfig.VIPPrivilegeInfos[index].DailyGem =
			vipPrivileges[index/2].Cards[index%2].DailyGem
		configReq.VIPPrivilegeConfig.VIPPrivilegeInfos[index].Duration =
			vipPrivileges[index/2].Cards[index%2].Duration
		configReq.VIPPrivilegeConfig.VIPPrivilegeInfos[index].GemCount =
			vipPrivileges[index/2].Cards[index%2].GemCount
		configReq.VIPPrivilegeConfig.VIPPrivilegeInfos[index].Price =
			vipPrivileges[index/2].Cards[index%2].Price
		configReq.VIPPrivilegeConfig.VIPPrivilegeInfos[index].PriceDesc =
			vipPrivileges[index/2].Cards[index%2].PriceDesc
		configReq.VIPPrivilegeConfig.VIPPrivilegeInfos[index].ProductId =
			vipPrivileges[index/2].Cards[index%2].ProductId
		configReq.VIPPrivilegeConfig.VIPPrivilegeInfos[index].ChannelID =
			vipPrivileges[index/2].Cards[index%2].ChannelID
		configReq.VIPPrivilegeConfig.VIPPrivilegeInfos[index].Type =
			vipPrivileges[index/2].Cards[index%2].Type
		configReq.VIPPrivilegeConfig.VIPPrivilegeInfos[index].NameKey =
			vipPrivileges[index/2].Cards[index%2].NameKey
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
			InterfaceName: "GMConfigVIPPrivilege",
			Params:        configReq,
		},
	}

	UrlQuery := fmt.Sprintf(UrlVipPrivilegeConfigFmt, *cmdParamsSvrIp)
	SendRequest(UrlQuery, &req)
}
