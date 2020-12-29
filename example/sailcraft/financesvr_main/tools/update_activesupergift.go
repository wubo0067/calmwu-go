/*
 * @Author: CALM.WU
 * @Date: 2018-03-31 14:39:01
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-04-02 15:49:21
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
	NewActiveSuperGiftConfigFile = "Active/ActiveSuperGift.json"
	UrlActiveSuperGiftConfigFmt  = "http://%s/sailcraft/api/v1/FinanceSvr/GMConfigSuperGiftActive"
)

type superGiftLimit struct {
	LimitTimes int32  `json:"limit_times"`
	LimitType  string `json:"limit_type"`
}

type superGift struct {
	ChannelID             string         `json:"AreaCode"`
	DescKey               string         `json:"DescKey"`
	FakePrice             float32        `json:"FakePrice"`
	FakePriceDesc         string         `json:"FakePriceDesc"`
	InnerGoods            interface{}    `json:"InnerGoods"`
	Limit                 superGiftLimit `json:"Limit"`
	NameKey               string         `json:"NameKey"`
	ActiveID              int            `json:"Id"`
	PosterAssetBundleName string         `json:"PosterAssetBundleName"`
	PosterTextureName     string         `json:"PosterTextureName"`
	Price                 float32        `json:"Price"`
	PriceDesc             string         `json:"PriceDesc"`
	ProductId             string         `json:"ProductId"`
	Discount              int            `json:"Discount"`
	Type                  string         `json:"Type"`
}

type superGiftInfo struct {
	Gifts []superGift `json:"Gift"`
}

func configActiveSuperGift(configPath string) {
	fileFullName := configPath + "/" + NewActiveSuperGiftConfigFile
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

	superGiftInfos := make([]superGiftInfo, 0)
	err = json.Unmarshal(data, &superGiftInfos)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unmarshal %s data failed! reason[%s]", fileFullName, err.Error())
		return
	}

	//fmt.Printf("superGiftInfos:%+v\n", superGiftInfos)
	var configReq proto.ProtoGMConfigActiveSuperGiftReq
	configReq.Uin = *cmdParamsUin
	configReq.ZoneID = int32(*cmdParamsZoneID)
	for i := range superGiftInfos {
		for j := range superGiftInfos[i].Gifts {
			superGift := &superGiftInfos[i].Gifts[j]

			activeSuperGiftInfo := new(proto.ActiveSuperGiftInfoS)
			activeSuperGiftInfo.Base.ActiveID = superGift.ActiveID
			activeSuperGiftInfo.Base.ChannelID = superGift.ChannelID
			activeSuperGiftInfo.Base.ReceiveCond = 0
			activeSuperGiftInfo.Base.ReceiveLimit = superGift.Limit.LimitTimes
			if superGift.Limit.LimitType == "daily" {
				activeSuperGiftInfo.Base.ResetEveryDay = 1
			} else {
				activeSuperGiftInfo.Base.ResetEveryDay = 0
			}
			activeSuperGiftInfo.DescKey = superGift.DescKey
			activeSuperGiftInfo.FakePrice = superGift.FakePrice
			activeSuperGiftInfo.FakePriceDesc = superGift.FakePriceDesc
			activeSuperGiftInfo.Type = superGift.Type

			jc, err := json.Marshal(superGift.InnerGoods)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ActiveID[%d] marshal InnerGoods failed! reason[%s]",
					superGift.ActiveID, err.Error())
				os.Exit(-1)
			}
			activeSuperGiftInfo.InnerGoods = string(jc)
			activeSuperGiftInfo.NameKey = superGift.NameKey
			activeSuperGiftInfo.PosterAssetBundleName = superGift.PosterAssetBundleName
			activeSuperGiftInfo.PosterTextureName = superGift.PosterTextureName
			activeSuperGiftInfo.Price = superGift.Price
			activeSuperGiftInfo.PriceDesc = superGift.PriceDesc
			activeSuperGiftInfo.ProductId = superGift.ProductId
			activeSuperGiftInfo.Discount = superGift.Discount

			configReq.SuperGiftConfigs = append(configReq.SuperGiftConfigs, *activeSuperGiftInfo)
		}
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
			InterfaceName: "GMConfigSuperGiftActive",
			Params:        configReq,
		},
	}

	UrlQuery := fmt.Sprintf(UrlActiveSuperGiftConfigFmt, *cmdParamsSvrIp)
	SendRequest(UrlQuery, &req)
	return
}
