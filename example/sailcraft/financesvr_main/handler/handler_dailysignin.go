/*
 * @Author: calmwu
 * @Date: 2018-03-06 14:51:54
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-03-17 20:25:49
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

func GMUpdateMonthlySigninConfigInfo(req *proto.ProtoGMConfigMonthlySignInReq) error {
	// 参数校验
	var err error
	prizeLstSize := len(req.MonthlySignInConfig.PrizeLst)
	if prizeLstSize != proto.C_SIGNINPRIZE_COUNT {
		err = fmt.Errorf("MonthlySignInConfig.PrizeLst count[%d] are not equal 31", prizeLstSize)
		base.GLog.Error(err.Error())
		return err
	}

	redisData, err := json.Marshal(req.MonthlySignInConfig)
	if err != nil {
		base.GLog.Error("Uin[%d] ZoneID[%d] Marshal failed! reason[%s]", req.Uin, req.ZoneID, err.Error())
		return err
	}

	monthlySigninKey := fmt.Sprintf(common.MonthlySigninKeyFmt, req.ZoneID)

	base.GLog.Debug("monthlySigninKey[%s]", monthlySigninKey)

	err = common.GRedis.StringSet(monthlySigninKey, redisData)
	if err != nil {
		err := fmt.Errorf("Set monthlySigninKey[%s] data failed! reason[%s]", monthlySigninKey, err.Error())
		base.GLog.Error(err.Error())
		return err
	}
	base.GLog.Debug("Uin[%d] ZoneID[%d] set monthlySigninKey[%s] successed!", req.Uin, req.ZoneID,
		monthlySigninKey)
	return nil
}

func getMonthlySigninConfigInfo(zoneID int32) *proto.MonthlySignInConfigS {
	monthlySigninKey := fmt.Sprintf(common.MonthlySigninKeyFmt, zoneID)
	base.GLog.Debug("monthlySigninKey[%s]", monthlySigninKey)

	var configInfo proto.MonthlySignInConfigS
	err := common.GetStrDataFromRedis(monthlySigninKey, &configInfo)
	if err != nil {
		return nil
	}

	return &configInfo
}

func GetMonthlySigninInfo(req *proto.ProtoGetMonthlySigninInfoReq) (*proto.ProtoGetMonthlySigninInfoRes, error) {
	var err error
	// 查询玩家信息
	userFinance, _ := QueryFinanceUser(req.Uin)
	if userFinance == nil {
		err = fmt.Errorf("Uin[%d] is not exist!", req.Uin)
		base.GLog.Error(err.Error())
		return nil, err
	}

	// 查询签到配置
	signInConfig := getMonthlySigninConfigInfo(req.ZoneID)
	if signInConfig == nil {
		err = fmt.Errorf("Uin[%d] zoneID[%d] is invalid!", req.Uin, req.ZoneID)
		base.GLog.Error(err.Error())
		return nil, err
	}

	// 计算时间
	location, _ := time.LoadLocation(userFinance.TimeZone)
	now := time.Now().In(location)
	// 当月有多少天
	monthDayCount := base.GetMonthlyDayCount(now.Year(), int(now.Month()))
	// 月份标识
	monthName := base.GetMonthName(location)
	// 当天日期
	currDate := base.GetDateNum(location)

	userFinance.SignInInfo.Month(location)

	var res proto.ProtoGetMonthlySigninInfoRes
	res.Uin = req.Uin
	res.ZoneID = req.ZoneID
	res.MonthName = monthName
	res.CurrDate = currDate
	res.WeeklyPrizeLst = signInConfig.PrizeLst[:monthDayCount]
	res.VipMultipleNum = signInConfig.VipMultipleNum
	res.VipMultiplePrizeDays = signInConfig.VipMultiplePrizeDays
	res.MonthlySignInCount = userFinance.SignInInfo.MonthlySignInCount
	res.ActivityThreshold = signInConfig.RessiueActivityThreshold

	if currDate > userFinance.SignInInfo.SigninDate {
		res.ToDaySignIn = 0
	} else {
		res.ToDaySignIn = 1
	}

	if currDate > userFinance.SignInInfo.ReSigninDate {
		res.ToDayReSignIn = 0
	} else {
		res.ToDayReSignIn = 1
	}

	return &res, nil
}

func PlayerSignIn(req *proto.ProtoPlayerSignInReq) (*proto.ProtoPlayerSignInRes, error) {
	var err error

	// 判断用户签到的周天是否合法
	if req.SignInDayNum < 1 || req.SignInDayNum > proto.C_SIGNINPRIZE_COUNT {
		err = fmt.Errorf("Uin[%d] SignInDayNum[%d] is invalid!", req.Uin, req.SignInDayNum)
		base.GLog.Error(err.Error())
		return nil, err
	}

	// 查询玩家信息
	userFinance, _ := QueryFinanceUser(req.Uin)
	if userFinance == nil {
		err = fmt.Errorf("Uin[%d] is not exist!", req.Uin)
		base.GLog.Error(err.Error())
		return nil, err
	}

	// 查询签到配置
	signInConfig := getMonthlySigninConfigInfo(req.ZoneID)
	if signInConfig == nil {
		err = fmt.Errorf("Uin[%d] zoneID[%d] is invalid!", req.Uin, req.ZoneID)
		base.GLog.Error(err.Error())
		return nil, err
	}

	// 计算时间
	location, _ := time.LoadLocation(userFinance.TimeZone)
	now := time.Now().In(location)
	// 当月几号
	dayNum := int32(now.Day())
	// 当月有多少天
	monthDayCount := int32(base.GetMonthlyDayCount(now.Year(), int(now.Month())))
	// 当天日期
	currDate := base.GetDateNum(location)

	userFinance.SignInInfo.Month(location)

	base.GLog.Debug("dayNum[%d] monthDayCount[%d] currDate[%d]", dayNum, monthDayCount, currDate)

	// 签到天数必须小于等于当月几号
	if req.SignInDayNum > dayNum {
		err = fmt.Errorf("Uin[%d] SignInDayNum[%d] larger than dayNum[%d]", req.Uin, req.SignInDayNum, dayNum)
		base.GLog.Error(err.Error())
		return nil, err
	}

	// 签到天数必须等于已经签到次数+1
	if req.SignInDayNum != (userFinance.SignInInfo.MonthlySignInCount + 1) {
		err = fmt.Errorf("Uin[%d] SignInDayNum[%d] not equal MonthlySignInCount[%d]+1", req.Uin, req.SignInDayNum,
			userFinance.SignInInfo.MonthlySignInCount)
		base.GLog.Error(err.Error())
		return nil, err
	}

	// 用户的当月签到数量不能操作当月几号
	if userFinance.SignInInfo.MonthlySignInCount > dayNum {
		err = fmt.Errorf("Uin[%d] MonthlySignInCount[%d] larger than dayNum[%d]", req.Uin, req.SignInDayNum, dayNum)
		base.GLog.Error(err.Error())
		return nil, err
	}

	// 判断是否当天首次签到还是补签，设置签到时间
	if userFinance.SignInInfo.SigninDate < currDate {
		base.GLog.Debug("Uin[%d] SignIn in the day[%d]", req.Uin, currDate)
		userFinance.SignInInfo.SigninDate = currDate
	} else {
		if userFinance.SignInInfo.ReSigninDate < currDate {
			// 判断活跃值是否满足配置
			if req.Activity >= signInConfig.RessiueActivityThreshold {
				base.GLog.Debug("Uin[%d] ReSignIn in the day[%d] ", req.Uin, currDate)
				userFinance.SignInInfo.ReSigninDate = currDate
			} else {
				err = fmt.Errorf("Uin[%d] ReSignIn be rejected! activity[%d] less than activityThreshold[%d]",
					req.Uin, req.Activity, signInConfig.RessiueActivityThreshold)
				base.GLog.Debug(err.Error())
				return nil, err
			}
		} else {
			err = fmt.Errorf("Uin[%d] SigninDate[%d] ReSigninDate[%d] in date[%d]", req.Uin, userFinance.SignInInfo.SigninDate,
				userFinance.SignInInfo.ReSigninDate, currDate)
			base.GLog.Error(err.Error())
			return nil, err
		}
	}

	// 得到会员类型
	vipType := GetFinanceUserVIPType(userFinance)

	// 递增签到次数
	userFinance.SignInInfo.MonthlySignInCount++

	var res proto.ProtoPlayerSignInRes
	res.Uin = req.Uin
	res.ZoneID = req.ZoneID
	res.SignInDayNum = req.SignInDayNum
	res.PrizeID = signInConfig.PrizeLst[req.SignInDayNum-1].PrizeID
	res.PrizeJsonContent = signInConfig.PrizeLst[req.SignInDayNum-1].PrizeJsonContent
	res.VipMultipleNum = 1

	// 判断vip是否多倍
	for _, vipDay := range signInConfig.VipMultiplePrizeDays {
		if req.SignInDayNum == vipDay && vipType&proto.E_USER_VIP_LUXURYMONTHLY != 0 {
			res.VipMultipleNum = signInConfig.VipMultipleNum
		}
	}

	// 更新数据库
	userFinance.SignInInfo.SyncDB(common.GDBEngine, req.Uin)
	return &res, nil
}
