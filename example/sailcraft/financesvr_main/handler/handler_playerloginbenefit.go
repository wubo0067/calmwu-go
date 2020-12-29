/*
 * @Author: calmwu
 * @Date: 2018-03-28 14:12:16
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-03-30 11:28:11
 * @Comment:
 */

package handler

import (
	"encoding/json"
	"fmt"
	"sailcraft/base"
	"sailcraft/financesvr_main/common"
	"sailcraft/financesvr_main/proto"
	"time"
)

func GMConfigNewPlayerLoginBenefit(req *proto.ProtoGMConfigNewPlayerLoginBenefitsReq) error {
	var err error
	benefitCount := len(req.Config.Benefits)
	if benefitCount != proto.C_NEWPLAYER_BENEFIT_DAYS {
		err = fmt.Errorf("NewPlayerLoginBenefitConfig.Benefits count[%d] are not equal 7", benefitCount)
		base.GLog.Error(err.Error())
		return err
	}

	redisData, err := json.Marshal(req.Config)
	if err != nil {
		base.GLog.Error("Uin[%d] ZoneID[%d] Marshal failed! reason[%s]", req.Uin, req.ZoneID, err.Error())
		return err
	}

	newPlayerLoginBenefitKey := fmt.Sprintf(common.NewPlayerLoginBenefitKeyFmt, req.ZoneID)

	base.GLog.Debug("newPlayerLoginBenefitKey[%s]", newPlayerLoginBenefitKey)

	err = common.GRedis.StringSet(newPlayerLoginBenefitKey, redisData)
	if err != nil {
		err := fmt.Errorf("Set newPlayerLoginBenefitKey[%s] data failed! reason[%s]", newPlayerLoginBenefitKey, err.Error())
		base.GLog.Error(err.Error())
		return err
	}
	base.GLog.Debug("Uin[%d] ZoneID[%d] set newPlayerLoginBenefitKey[%s] successed!", req.Uin, req.ZoneID,
		newPlayerLoginBenefitKey)
	return nil
}

func getNewPlayerLoginBenefitConfigInfo(zoneID int32) *proto.ProtoNewPlayerBenefitConfigS {
	newPlayerLoginBenefitKey := fmt.Sprintf(common.NewPlayerLoginBenefitKeyFmt, zoneID)
	base.GLog.Debug("newPlayerLoginBenefitKey[%s]", newPlayerLoginBenefitKey)

	var configInfo proto.ProtoNewPlayerBenefitConfigS
	err := common.GetStrDataFromRedis(newPlayerLoginBenefitKey, &configInfo)
	if err != nil {
		return nil
	}

	return &configInfo
}

func GetNewPlayerLoginBenefitInfo(req *proto.ProtoGetNewPlayerLoginBenefitReq) (*proto.ProtoGetNewPlayerLoginBenefitRes, error) {
	var res proto.ProtoGetNewPlayerLoginBenefitRes

	var err error
	// 查询玩家信息
	newPlayerLoginBenefits, _ := QueryNewPlayerLoginBenefit(req.Uin)
	if newPlayerLoginBenefits == nil {
		err = fmt.Errorf("Uin[%d] is not exist!", req.Uin)
		base.GLog.Error(err.Error())
		return nil, err
	}

	// 查询签到配置
	benefitConfig := getNewPlayerLoginBenefitConfigInfo(req.ZoneID)
	if benefitConfig == nil {
		err = fmt.Errorf("Uin[%d] zoneID[%d] is invalid!", req.Uin, req.ZoneID)
		base.GLog.Error(err.Error())
		return nil, err
	}

	// 计算时间
	location, _ := time.LoadLocation(newPlayerLoginBenefits.TimeZone)
	// 当天日期
	currDate := base.GetDateNum(location)

	syncDb := false
	if newPlayerLoginBenefits.LastLoginDate < currDate {
		newPlayerLoginBenefits.LoginDays++
		newPlayerLoginBenefits.LastLoginDate = currDate
		syncDb = true
	}

	res.LoginDays = newPlayerLoginBenefits.LoginDays
	res.ReceiveAwardTags = newPlayerLoginBenefits.ReceiveAwardTags
	res.Uin = req.Uin
	res.ZoneID = req.ZoneID
	res.IsCompleted = newPlayerLoginBenefits.IsCompleted
	res.Benefits = benefitConfig.Benefits

	if syncDb {
		newPlayerLoginBenefits.SyncDB(common.GDBEngine)
	}

	return &res, nil
}

func ReceiveLoginBenefit(req *proto.ProtoReceiveLoginBenefitReq) (*proto.ProtoReceiveLoginBenefitRes, error) {
	var err error
	if req.ReceiveDayNum < 1 || req.ReceiveDayNum > proto.C_NEWPLAYER_BENEFIT_DAYS {
		err = fmt.Errorf("Uin[%d] ReceiveDayNum[%d] is invalid!", req.Uin, req.ReceiveDayNum)
		base.GLog.Error(err.Error())
		return nil, err
	}
	// 查询玩家信息
	newPlayerLoginBenefits, _ := QueryNewPlayerLoginBenefit(req.Uin)
	if newPlayerLoginBenefits == nil {
		err = fmt.Errorf("Uin[%d] is not exist!", req.Uin)
		base.GLog.Error(err.Error())
		return nil, err
	}

	// 查询签到配置
	benefitConfig := getNewPlayerLoginBenefitConfigInfo(req.ZoneID)
	if benefitConfig == nil {
		err = fmt.Errorf("Uin[%d] zoneID[%d] is invalid!", req.Uin, req.ZoneID)
		base.GLog.Error(err.Error())
		return nil, err
	}

	var res proto.ProtoReceiveLoginBenefitRes
	res.Uin = req.Uin
	res.ZoneID = req.ZoneID
	res.LoginDays = newPlayerLoginBenefits.LoginDays

	if newPlayerLoginBenefits.ReceiveAwardCount < proto.C_NEWPLAYER_BENEFIT_DAYS &&
		newPlayerLoginBenefits.ReceiveAwardCount < newPlayerLoginBenefits.LoginDays {

		// 判断这天是否领取过
		if newPlayerLoginBenefits.ReceiveAwardTags[req.ReceiveDayNum-1] == 1 {
			err = fmt.Errorf("Uin[%d] ZoneID[%d] ReceiveDayNum[%d] already received!",
				req.Uin, req.ZoneID, req.ReceiveDayNum)
			base.GLog.Error(err.Error())
			return nil, err
		} else {
			newPlayerLoginBenefits.ReceiveAwardTags[req.ReceiveDayNum-1] = 1
		}
		newPlayerLoginBenefits.ReceiveAwardCount++

		if newPlayerLoginBenefits.ReceiveAwardCount == proto.C_NEWPLAYER_BENEFIT_DAYS {
			newPlayerLoginBenefits.IsCompleted = 1
		}
		err = newPlayerLoginBenefits.SyncDB(common.GDBEngine)
		if err != nil {
			return nil, err
		}

		res.IsCompleted = newPlayerLoginBenefits.IsCompleted
		res.ReceiveAwardTags = newPlayerLoginBenefits.ReceiveAwardTags

		benefitInfo := benefitConfig.FindBenefit(req.Id)
		if benefitConfig != nil {
			res.BenefitInfo = benefitInfo
		} else {
			err = fmt.Errorf("Uin[%d] zoneID[%d] benefitID[%d] is invalid!", req.Uin, req.ZoneID, req.Id)
			base.GLog.Error(err.Error())
			return nil, err
		}
	} else {
		err = fmt.Errorf("Uin[%d] zoneID[%d] ReceiveAwardCount[%d] exceed limit! LoginDays[%d]",
			req.Uin, req.ZoneID, newPlayerLoginBenefits.ReceiveAwardCount, newPlayerLoginBenefits.LoginDays)
		base.GLog.Error(err.Error())
		return nil, err
	}

	return &res, nil
}
