/*
 * @Author: calmwu
 * @Date: 2018-03-28 15:54:22
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-03-28 18:59:29
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
	NewPlayerLoginBenefitConfigFile   = "Active/ActiveAccumulativeLogin.json"
	UrlNewPlayerLoginBenefitConfigFmt = "http://%s/sailcraft/api/v1/FinanceSvr/GMConfigNewPlayerLoginBenefit"
)

type benefitInfoS struct {
	Id                    int         `json:"Id"`
	PosterAssetBundleName string      `json:"PosterAssetBundleName"`
	PosterTextureName     string      `json:"PosterTextureName"`
	Reward                interface{} `json:"Reward"`
}

func configNewPlayerLoginBenefit(configPath string) {
	fileFullName := configPath + "/" + NewPlayerLoginBenefitConfigFile
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

	benefitInfos := make([]benefitInfoS, 0)
	err = json.Unmarshal(data, &benefitInfos)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unmarshal %s data failed! reason[%s]", fileFullName, err.Error())
		return
	}

	//fmt.Printf("benefitInfos:%+v\n", benefitInfos)
	var configReq proto.ProtoGMConfigNewPlayerLoginBenefitsReq
	configReq.Uin = *cmdParamsUin
	configReq.ZoneID = int32(*cmdParamsZoneID)
	configReq.Config.Benefits = make([]proto.ProtoNewPlayerBenefitInfo, len(benefitInfos))
	for index, _ := range benefitInfos {
		configReq.Config.Benefits[index].Id = benefitInfos[index].Id
		configReq.Config.Benefits[index].PosterAssetBundleName = benefitInfos[index].PosterAssetBundleName
		configReq.Config.Benefits[index].PosterTextureName = benefitInfos[index].PosterTextureName

		benefitJC, err := json.Marshal(benefitInfos[index].Reward)
		if err != nil {
			fmt.Fprintf(os.Stderr, "BenefitId[%d] marshal Reward failed! reason[%s]",
				benefitInfos[index].Id, err.Error())
			os.Exit(-1)
		}

		configReq.Config.Benefits[index].JsonContent = string(benefitJC)
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
			InterfaceName: "GMConfigNewPlayerLoginBenefit",
			Params:        configReq,
		},
	}

	UrlQuery := fmt.Sprintf(UrlNewPlayerLoginBenefitConfigFmt, *cmdParamsSvrIp)
	SendRequest(UrlQuery, &req)
	return
}
