/*
 * @Author: calmwu
 * @Date: 2018-03-23 16:58:43
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-04-16 16:47:19
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

func GMConfigVIPPrivilege(req *proto.ProtoGMConfigVIPPrivilegeReq) error {
	redisData, err := json.Marshal(req.VIPPrivilegeConfig)
	if err != nil {
		base.GLog.Error("Uin[%d] ZoneID[%d] Marshal failed! reason[%s]", req.Uin, req.ZoneID, err.Error())
		return err
	}

	vipPrivilegeKey := fmt.Sprintf(common.VIPPrivilegeKeyFmt, req.ZoneID)

	base.GLog.Debug("vipPrivilegeKey[%s]", vipPrivilegeKey)

	err = common.GRedis.StringSet(vipPrivilegeKey, redisData)
	if err != nil {
		err := fmt.Errorf("Set vipPrivilegeKey[%s] data failed! reason[%s]", vipPrivilegeKey, err.Error())
		base.GLog.Error(err.Error())
		return err
	}
	base.GLog.Debug("Uin[%d] ZoneID[%d] set vipPrivilegeKey[%s] successed!", req.Uin, req.ZoneID,
		vipPrivilegeKey)
	return nil
}

func getVIPPrivilegeConfig(zoneID int32) *proto.ProtoVIPPrivilegeConfigS {
	vipPrivilegeKey := fmt.Sprintf(common.VIPPrivilegeKeyFmt, zoneID)

	base.GLog.Debug("vipPrivilegeKey[%s]", vipPrivilegeKey)

	var configInfo proto.ProtoVIPPrivilegeConfigS
	err := common.GetStrDataFromRedis(vipPrivilegeKey, &configInfo)
	if err != nil {
		return nil
	}

	return &configInfo
}

func GetPlayerVIPInfo(req *proto.ProtoGetPlayerVIPInfoReq) (*proto.ProtoGetPlayerVIPInfoRes, error) {
	var err error
	// 查询玩家信息
	userFinance, _ := QueryFinanceUser(req.Uin)
	if userFinance == nil {
		err = fmt.Errorf("Uin[%d] is not exist!", req.Uin)
		base.GLog.Error(err.Error())
		return nil, err
	}

	vipPrivilegeConfig := getVIPPrivilegeConfig(req.ZoneID)
	if vipPrivilegeConfig == nil {
		err = fmt.Errorf("Uin[%d] ZoneID[%d] VIPPrivilegeConfig is not configured!", req.Uin, req.ZoneID)
		return nil, err
	}

	location, _ := time.LoadLocation(userFinance.TimeZone)
	now := time.Now().In(location)
	currDate := int32(base.GetDateNum(location))

	vipInfoRes := new(proto.ProtoGetPlayerVIPInfoRes)
	vipInfoRes.Uin = req.Uin
	vipInfoRes.ZoneID = req.ZoneID
	vipInfoRes.VipType = GetFinanceUserVIPType(userFinance)
	if vipInfoRes.VipType&proto.E_USER_VIP_NORMALMONTHLY != 0 {
		vipInfoRes.NormalMonthVIPCollectPrizeCount = userFinance.VipInfo.NormalMonthVIPCollectPrizeCount
		vipInfoRes.NormalMonthVIPRemainderSeconds = int64(userFinance.VipInfo.NormalMonthVIPExpireTime.Sub(now).Seconds())
		vipInfoRes.NormalMonthVIPCollectPrizeExpireDate = userFinance.VipInfo.NormalMonthVIPCollectPrizeExpireDate
		if userFinance.VipInfo.NormalMonthVIPCollectPrizeDate == currDate {
			vipInfoRes.NormalMonthVIPDayCollected = 1
		}
	}

	if vipInfoRes.VipType&proto.E_USER_VIP_LUXURYMONTHLY != 0 {
		vipInfoRes.LuxuryMonthVIPCollectPrizeCount = userFinance.VipInfo.LuxuryMonthVIPCollectPrizeCount
		vipInfoRes.LuxuryMonthVIPRemainderSeconds = int64(userFinance.VipInfo.LuxuryMonthVIPExpireTime.Sub(now).Seconds())
		vipInfoRes.LuxuryMonthVIPCollectPrizeExpireDate = userFinance.VipInfo.LuxuryMonthVIPCollectPrizeExpireDate
		if userFinance.VipInfo.LuxuryMonthVIPCollectPrizeDate == currDate {
			vipInfoRes.LuxuryMonthVIPDayCollected = 1
		}
	}

	vipInfoRes.PrivilegeInfo = make([]proto.ProtoVIPPrivilegeInfoS, 0)
	for index, _ := range vipPrivilegeConfig.VIPPrivilegeInfos {
		if vipPrivilegeConfig.VIPPrivilegeInfos[index].ChannelID == req.ChannelID {
			vipInfoRes.PrivilegeInfo = append(vipInfoRes.PrivilegeInfo, vipPrivilegeConfig.VIPPrivilegeInfos[index])
		}
	}

	return vipInfoRes, nil
}

func VIPPlayerCollectPrize(req *proto.ProtoVIPPlayerCollectPrizeReq) (*proto.ProtoVIPPlayerCollectPrizeRes, error) {
	var err error
	// 查询玩家信息
	userFinance, _ := QueryFinanceUser(req.Uin)
	if userFinance == nil {
		err = fmt.Errorf("Uin[%d] is not exist!", req.Uin)
		base.GLog.Error(err.Error())
		return nil, err
	}

	vipPrivilegeConfig := getVIPPrivilegeConfig(req.ZoneID)
	if vipPrivilegeConfig == nil {
		err = fmt.Errorf("Uin[%d] get ZoneID[%d] VIPPrivilegeConfig failed!", req.Uin, req.ZoneID)
		return nil, err
	}

	location, _ := time.LoadLocation(userFinance.TimeZone)
	now := time.Now().In(location)
	currDate := int32(base.GetDateNum(location))
	vipType := GetFinanceUserVIPType(userFinance)

	res := new(proto.ProtoVIPPlayerCollectPrizeRes)
	res.Uin = req.Uin
	res.ZoneID = req.ZoneID
	res.VipType = req.VipType

	base.GLog.Debug("Uin[%d] now[%s] currDate[%d] vipType[%s]", req.Uin, base.TimeName(now), currDate,
		vipType.String())

	if req.VipType == proto.E_USER_VIP_NORMALMONTHLY && vipType&proto.E_USER_VIP_NORMALMONTHLY != 0 {
		// 领取普通月会员的每日奖励

		// 判断当天是否领取过
		if userFinance.VipInfo.NormalMonthVIPCollectPrizeDate >= currDate {
			err = fmt.Errorf("Uin[%d] NormalMonthVIP currDate[%d] has been collected prize!", req.Uin,
				currDate)
			base.GLog.Error(err.Error())
			return nil, err
		}

		if userFinance.VipInfo.NormalMonthVIPCollectPrizeCount >= common.C_MAX_MONTHVIP_COLLECTPRIZEDAYS {
			err = fmt.Errorf("Uin[%d] NormalMonthVIP collectCount[%d] reached limited[%d]!", req.Uin,
				userFinance.VipInfo.NormalMonthVIPCollectPrizeCount, common.C_MAX_MONTHVIP_COLLECTPRIZEDAYS)
			base.GLog.Error(err.Error())
			return nil, err
		}

		// 判断是否过期
		// 修改：如果过了领取日期，但是vip还没过期，且领取次数没满30，则当天可以领取，其实下面这个过期日期不用判断了
		// if currDate > userFinance.VipInfo.NormalMonthVIPCollectPrizeExpireDate {
		// 	err = fmt.Errorf("Uin[%d] NormalMonthVIPCollectPrizeExpireDate[%d] expired!", req.Uin,
		// 		userFinance.VipInfo.NormalMonthVIPCollectPrizeExpireDate)
		// 	base.GLog.Error(err.Error())
		// 	return nil, err
		// }

		userFinance.VipInfo.NormalMonthVIPCollectPrizeDate = currDate
		userFinance.VipInfo.NormalMonthVIPCollectPrizeCount++
		userFinance.VipInfo.SyncDB(common.GDBEngine, req.Uin)

		vipPrivilegeInfo := vipPrivilegeConfig.FindVIPPrivilege(req.ChannelID, req.Id)
		if vipPrivilegeInfo != nil {
			res.CollectDiamonds = vipPrivilegeInfo.DailyGem
		}
		res.CollectPrizeCount = userFinance.VipInfo.NormalMonthVIPCollectPrizeCount
	} else if req.VipType == proto.E_USER_VIP_LUXURYMONTHLY && vipType&proto.E_USER_VIP_LUXURYMONTHLY != 0 {
		// 领取月会员每日奖励
		// if currDate > userFinance.VipInfo.LuxuryMonthVIPCollectPrizeExpireDate {
		// 	err = fmt.Errorf("Uin[%d] LuxuryMonthVIPCollectPrizeExpireDate[%d] expired!", req.Uin,
		// 		userFinance.VipInfo.LuxuryMonthVIPCollectPrizeExpireDate)
		// 	base.GLog.Error(err.Error())
		// 	return nil, err
		// }

		// 判断当天是否领取过
		if userFinance.VipInfo.LuxuryMonthVIPCollectPrizeDate >= currDate {
			err = fmt.Errorf("Uin[%d] LuxuryMonthVIP currDate[%d] has been collected prize!", req.Uin,
				currDate)
			base.GLog.Error(err.Error())
			return nil, err
		}

		if userFinance.VipInfo.LuxuryMonthVIPCollectPrizeCount >= common.C_MAX_MONTHVIP_COLLECTPRIZEDAYS {
			err = fmt.Errorf("Uin[%d] LuxuryMonthVIP collectCount[%d] reached limited[%d]!", req.Uin,
				userFinance.VipInfo.LuxuryMonthVIPCollectPrizeCount, common.C_MAX_MONTHVIP_COLLECTPRIZEDAYS)
			base.GLog.Error(err.Error())
			return nil, err
		}

		userFinance.VipInfo.LuxuryMonthVIPCollectPrizeDate = currDate
		userFinance.VipInfo.LuxuryMonthVIPCollectPrizeCount++
		userFinance.VipInfo.SyncDB(common.GDBEngine, req.Uin)

		vipPrivilegeInfo := vipPrivilegeConfig.FindVIPPrivilege(req.ChannelID, req.Id)
		if vipPrivilegeInfo != nil {
			res.CollectDiamonds = vipPrivilegeInfo.DailyGem
		}
		res.CollectPrizeCount = userFinance.VipInfo.LuxuryMonthVIPCollectPrizeCount
	} else {
		err = fmt.Errorf("Uin[%d] vipType[%s] is not vip!", req.Uin, vipType.String())
		base.GLog.Error(err.Error())
		return nil, err
	}
	return res, nil
}
