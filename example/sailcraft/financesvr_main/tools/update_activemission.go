/*
 * @Author: calmwu
 * @Date: 2018-04-02 15:01:15
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-04-13 11:04:21
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
	ActiveMissionConfigFile   = "Active/ActiveMission.json"
	UrlActiveMissionConfigFmt = "http://%s/sailcraft/api/v1/FinanceSvr/GMConfigMissionActive"
)

type missionLimit struct {
	LimitTimes int32  `json:"limit_times"`
	LimitType  string `json:"limit_type"`
}

type missionParameter struct {
	Target int32 `json:"target"`
}

type missionInfo struct {
	ActiveID   int          `json:"Id"`
	Limit      missionLimit `json:"Limit"`
	InnerGoods interface{}  `json:"Reward"`
	//Parameter  missionParameter `json:"Parameter"`
	Parameter interface{} `json:"Parameter"`
	TaskType  string      `json:"TaskType"`
	TitleKey  string      `json:"TitleKey"`
}

func configActiveMission(configPath string) {
	fileFullName := configPath + "/" + ActiveMissionConfigFile
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

	missions := make([]missionInfo, 0)
	err = json.Unmarshal(data, &missions)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unmarshal %s data failed! reason[%s]", fileFullName, err.Error())
		return
	}

	//fmt.Printf("missions:%+v\n", missions)

	var configReq proto.ProtoGMConfigActiveMissionReq
	configReq.Uin = *cmdParamsUin
	configReq.ZoneID = int32(*cmdParamsZoneID)
	for i := range missions {
		mission := &missions[i]

		activeMission := new(proto.ActiveMissionInfoS)
		activeMission.Base.ActiveID = mission.ActiveID
		activeMission.Base.ChannelID = "NOAREA"
		activeMission.Base.ReceiveCond = int32(mission.Parameter.(map[string]interface{})["target"].(float64))
		activeMission.Base.ReceiveLimit = mission.Limit.LimitTimes

		if mission.Limit.LimitType == "daily" {
			activeMission.Base.ResetEveryDay = 1
		} else {
			activeMission.Base.ResetEveryDay = 0
		}

		jc, err := json.Marshal(mission.InnerGoods)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ActiveID[%d] marshal InnerGoods failed! reason[%s]",
				mission.ActiveID, err.Error())
			os.Exit(-1)
		}
		activeMission.InnerGoods = string(jc)

		jc, err = json.Marshal(mission.Parameter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ActiveID[%d] marshal Parameter failed! reason[%s]",
				mission.ActiveID, err.Error())
			os.Exit(-1)
		}
		activeMission.Parameter = string(jc)

		activeMission.TaskType = mission.TaskType
		activeMission.TitleKey = mission.TitleKey

		configReq.ActiveMissions = append(configReq.ActiveMissions, *activeMission)
	}

	fmt.Printf("configReq:%+v\n", configReq)

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
			InterfaceName: "GMConfigMissionActive",
			Params:        configReq,
		},
	}

	UrlQuery := fmt.Sprintf(UrlActiveMissionConfigFmt, *cmdParamsSvrIp)
	SendRequest(UrlQuery, &req)
	return
}
