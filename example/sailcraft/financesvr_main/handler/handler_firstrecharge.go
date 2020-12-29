/*
 * @Author: calmwu
 * @Date: 2018-04-16 16:27:15
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-04-20 18:31:28
 * @Comment:
 */
package handler

import (
	"encoding/json"
	"fmt"
	"sailcraft/base"
	"sailcraft/financesvr_main/common"
	"sailcraft/financesvr_main/proto"
)

func GMConfigFirstRecharge(req *proto.ProtoGMConfigFirstRechargeReq) error {
	var err error

	redisData, err := json.Marshal(req.Config)
	if err != nil {
		base.GLog.Error("Uin[%d] ZoneID[%d] Marshal failed! reason[%s]", req.Uin, req.ZoneID, err.Error())
		return err
	}

	firstRechargeKey := fmt.Sprintf(common.FirstRechargeKeyFmt, req.ZoneID)

	base.GLog.Debug("firstRechargeKey[%s]", firstRechargeKey)

	err = common.GRedis.StringSet(firstRechargeKey, redisData)
	if err != nil {
		err := fmt.Errorf("Set firstRechargeKey[%s] data failed! reason[%s]", firstRechargeKey, err.Error())
		base.GLog.Error(err.Error())
		return err
	}
	base.GLog.Debug("Uin[%d] ZoneID[%d] set firstRechargeKey[%s] successed!", req.Uin, req.ZoneID,
		firstRechargeKey)
	return nil
}

func getFirstRechargeConfigInfo(zoneID int32) *proto.ProtoFirstRechargeConfigS {
	firstRechargeKey := fmt.Sprintf(common.FirstRechargeKeyFmt, zoneID)
	base.GLog.Debug("firstRechargeKey[%s]", firstRechargeKey)

	var configInfo proto.ProtoFirstRechargeConfigS
	err := common.GetStrDataFromRedis(firstRechargeKey, &configInfo)
	if err != nil {
		return nil
	}
	base.GLog.Debug("%s %+v", firstRechargeKey, configInfo)
	return &configInfo
}

func GetFirstRechargeActive(req *proto.ProtoGetFirstRechargeActiveReq) (*proto.ProtoGetFirstRechargeActiveRes, error) {
	var err error
	userFinance, _ := QueryFinanceUser(req.Uin)
	if userFinance == nil {
		err = fmt.Errorf("Uin[%d] is not exist!", req.Uin)
		base.GLog.Error(err.Error())
		return nil, err
	}

	firstRechargeConfig := getFirstRechargeConfigInfo(req.ZoneID)
	if firstRechargeConfig == nil {
		err = fmt.Errorf("Uin[%d] ZoneID[%d] firstRechargeConfig is not configured!", req.Uin, req.ZoneID)
		return nil, err
	}

	var res proto.ProtoGetFirstRechargeActiveRes
	res.Uin = req.Uin
	res.ZoneID = req.ZoneID
	res.CurrBuyDiamonds = userFinance.FirstRecharge.CurrBuyDiamonds
	if userFinance.FirstRecharge.ReceiveCount == 0 {
		res.IsCompleted = 1
	}
	res.LevelInfos = userFinance.FirstRecharge.LevelInfoLst
	res.LevelConfigs = firstRechargeConfig.FRLevelConfLst

	return &res, nil
}

func ReceiveFirstRechargeReward(req *proto.ProtoReceiveFirstRechargeRewardReq) (*proto.ProtoReceiveFirstRechargeRewardRes, error) {
	var err error

	userFinance, _ := QueryFinanceUser(req.Uin)
	if userFinance == nil {
		err = fmt.Errorf("Uin[%d] is not exist!", req.Uin)
		base.GLog.Error(err.Error())
		return nil, err
	}

	firstRechargeConfig := getFirstRechargeConfigInfo(req.ZoneID)
	if firstRechargeConfig == nil {
		err = fmt.Errorf("Uin[%d] ZoneID[%d] firstRechargeConfig is not configured!", req.Uin, req.ZoneID)
		return nil, err
	}

	firstRechargeLevelConf := firstRechargeConfig.Find(req.Id)
	if firstRechargeLevelConf == nil {
		return nil, fmt.Errorf("FirstRechargeActive[%d] is invalid!", req.Id)
	}

	base.GLog.Debug("Uin[%d] ZoneID[%d] CurrBuyDiamonds[%d] Target[%d] FirstRechargeID[%d]",
		req.Uin, req.ZoneID, userFinance.FirstRecharge.CurrBuyDiamonds, firstRechargeLevelConf.Target,
		firstRechargeLevelConf.Id)

	var res proto.ProtoReceiveFirstRechargeRewardRes
	res.Uin = req.Uin
	res.ZoneID = req.ZoneID
	res.Id = req.Id

	// 判断钻石是足够
	if userFinance.FirstRecharge.CurrBuyDiamonds >= firstRechargeLevelConf.Target {
		for index := range userFinance.FirstRecharge.LevelInfoLst {
			levelInfo := &userFinance.FirstRecharge.LevelInfoLst[index]
			base.GLog.Debug("Uin[%d] ZoneID[%d] LevelIndex[%d] ActiveID[%d] Received[%d]",
				req.Uin, req.ZoneID, index, levelInfo.ActiveID, levelInfo.Received)
			if levelInfo.ActiveID == req.Id {
				if levelInfo.Received == 0 {
					base.GLog.Debug("Uin[%d] ZoneID[%d] FirstRechargeID[%d] can receive!",
						req.Uin, req.ZoneID, firstRechargeLevelConf.Id)
					levelInfo.Received = 1
					userFinance.FirstRecharge.ReceiveCount--
					userFinance.FirstRecharge.SyncDB(common.GDBEngine, req.Uin)
					// 得到奖品
					res.JsonContent = firstRechargeLevelConf.Reward
					break
				} else {
					err = fmt.Errorf("FirstRechargeActive[%d] had already been received!", req.Id)
					base.GLog.Error(err.Error())
					return nil, err
				}
			}
		}
	}

	if userFinance.FirstRecharge.ReceiveCount == 0 {
		res.IsCompleted = 1
	}

	return &res, nil
}
