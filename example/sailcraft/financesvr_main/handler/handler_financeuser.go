/*
 * @Author: calmwu
 * @Date: 2018-02-05 15:50:08
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-04-16 17:22:26
 * @Comment:
 */

package handler

import (
	"fmt"
	"sailcraft/base"
	"sailcraft/dataaccess/mysql"
	"sailcraft/financesvr_main/common"
	"sailcraft/financesvr_main/proto"
	"strconv"
	"time"
)

func AddNewFinanceUser(req *proto.ProtoNewFinanceUserReq) error {
	userFinance := new(proto.TblFinanceUserS)
	userFinance.Uin = req.Uin
	userFinance.ZoneID = req.ZoneID
	userFinance.TimeZone = req.TimeZone

	if len(req.TimeZone) == 0 {
		base.GLog.Warn("Uin[%d] ZoneID[%d] TimeZone is emtpy!", req.Uin, req.ZoneID)
		userFinance.TimeZone = "Local"
	}

	// 用户时区
	localtion, err := time.LoadLocation(req.TimeZone)
	if err != nil {
		base.GLog.Error("Uin[%d] LoadLocation[%s] failed! reason[%s]", req.Uin, req.TimeZone, err.Error())
		// 无效的时区就用utc取代
		localtion, _ = time.LoadLocation("")
		userFinance.TimeZone = "Local"
	}

	// 用户的本地时区时间
	now := time.Now().In(localtion)

	// 计算新用户的刷新起始时间
	refreshShopConfig := getRefreshShopConfig(req.ZoneID)
	if refreshShopConfig == nil {
		err = fmt.Errorf("Uin[%d] ZoneID[%d] RefreshShop is not configured!", req.Uin, req.ZoneID)
		base.GLog.Error(err.Error())
		return err
	}

	// 首冲任务
	firstRechargeConfig := getFirstRechargeConfigInfo(req.ZoneID)
	if firstRechargeConfig == nil {
		err = fmt.Errorf("Uin[%d] ZoneID[%d] FirstRechargeActive is not configured!", req.Uin, req.ZoneID)
		base.GLog.Error(err.Error())
		return err
	}

	// 刷新商店信息
	startHours := common.CalcRefreshStartHours(refreshShopConfig.ShopAutoRefreshIntervalHours, int32(now.Hour()))
	userFinance.PlayerRefreshShopDailyInfo.Init(req.Uin, startHours, now)
	// 签到信息
	userFinance.SignInInfo.Init(localtion)
	// 商品首次购买信息
	userFinance.ShopFirstPurchaseInfo.Init()
	// 会员信息
	userFinance.VipInfo.Init(now)
	// 首冲任务
	userFinance.FirstRecharge.Init(firstRechargeConfig)
	affected, err := mysql.InsertRecord(common.GDBEngine, proto.TBNAME_FINANCEUSER, userFinance)
	if err != nil {
		base.GLog.Error("Insert FinanceUser[%d] failed! reason[%s]", req.Uin, err.Error())
		return err
	}

	base.GLog.Debug("Insert FinanceUser[%d] successed! affected[%d]", req.Uin, affected)
	//--------------------------------------------------------------------------------

	newPlayerLoginBenefits := new(proto.TblNewPlayerLoginBenefits)
	newPlayerLoginBenefits.Init(req.Uin, req.ZoneID, userFinance.TimeZone, localtion)
	affected, err = mysql.InsertRecord(common.GDBEngine, proto.TBNAME_NEWPLAYERLOGINBENEFITS, newPlayerLoginBenefits)
	if err != nil {
		base.GLog.Error("Insert %s[%d] failed! reason[%s]", proto.TBNAME_NEWPLAYERLOGINBENEFITS, req.Uin, err.Error())
		return err
	}
	base.GLog.Debug("Insert %s[%d] successed! affected[%d]", proto.TBNAME_NEWPLAYERLOGINBENEFITS, req.Uin, affected)

	return nil
}

func QueryFinanceUser(uin uint64) (*proto.TblFinanceUserS, error) {
	userFinance := new(proto.TblFinanceUserS)
	userFinance.Uin = uin
	exists, err := mysql.GetRecord(common.GDBEngine, proto.TBNAME_FINANCEUSER, userFinance)
	if err != nil {
		base.GLog.Error("Query FinanceUser[%d] failed! reason[%s]", uin, err.Error())
		return nil, err
	}

	if !exists {
		err := fmt.Errorf("Query player[%d] does not exist!", uin)
		base.GLog.Error(err.Error())
		return nil, err
	}

	//base.GLog.Debug("%+v", userFinance)
	return userFinance, nil
}

func calcCollectPrizeExpireDate(vipType proto.UserVIPType, now time.Time) int32 {
	durationDays := common.C_MAX_MONTHVIP_COLLECTPRIZEDAYS - 1
	year, month, day := now.Date()
	expireTime := time.Date(year, month, day+durationDays, 0, 0, 0, 0, now.Location())
	expireDate, _ := strconv.ParseInt(
		fmt.Sprintf("%d%02d%02d", expireTime.Year(), expireTime.Month(), expireTime.Day()), 10, 32)
	return int32(expireDate)
}

func FinanceUserAddVip(userFinance *proto.TblFinanceUserS, memberType proto.UserVIPType) error {

	location, _ := time.LoadLocation(userFinance.TimeZone)
	now := time.Now().In(location)

	vipDuration := common.MonthlyCardDuration

	switch memberType {
	case proto.E_USER_VIP_LUXURYMONTHLY:
		userFinance.VipInfo.VIPType |= proto.E_USER_VIP_LUXURYMONTHLY
		userFinance.VipInfo.LuxuryMonthVIPExpireTime = now.Add(vipDuration)
		userFinance.VipInfo.LuxuryMonthVIPCollectPrizeExpireDate = calcCollectPrizeExpireDate(proto.E_USER_VIP_LUXURYMONTHLY,
			now)
		userFinance.VipInfo.LuxuryMonthVIPCollectPrizeCount = 0
		userFinance.VipInfo.LuxuryMonthVIPCollectPrizeDate = 0

		base.GLog.Debug("Uin[%d] memberType[%s] LuxuryMonthVIPExpireTime[%s] LuxuryMonthVIPCollectPrizeExpireDate[%d] LuxuryMonthVIPCollectPrizeCount[%d]",
			userFinance.Uin, memberType.String(),
			base.TimeName(userFinance.VipInfo.LuxuryMonthVIPExpireTime), userFinance.VipInfo.LuxuryMonthVIPCollectPrizeExpireDate,
			userFinance.VipInfo.LuxuryMonthVIPCollectPrizeCount)

	case proto.E_USER_VIP_NORMALMONTHLY:
		userFinance.VipInfo.VIPType |= proto.E_USER_VIP_NORMALMONTHLY
		userFinance.VipInfo.NormalMonthVIPExpireTime = now.Add(vipDuration)
		userFinance.VipInfo.NormalMonthVIPCollectPrizeExpireDate = calcCollectPrizeExpireDate(proto.E_USER_VIP_NORMALMONTHLY,
			now)
		userFinance.VipInfo.NormalMonthVIPCollectPrizeCount = 0
		userFinance.VipInfo.NormalMonthVIPCollectPrizeDate = 0

		base.GLog.Debug("Uin[%d] memberType[%s] NormalMonthVIPExpireTime[%s] NormalMonthVIPCollectPrizeExpireDate[%d] NormalMonthVIPCollectPrizeCount[%d]",
			userFinance.Uin, memberType.String(),
			base.TimeName(userFinance.VipInfo.NormalMonthVIPExpireTime), userFinance.VipInfo.NormalMonthVIPCollectPrizeExpireDate,
			userFinance.VipInfo.NormalMonthVIPCollectPrizeCount)
	default:
		err := fmt.Errorf("Uin[%d] memberType[%s] is invalid!", userFinance.Uin, memberType.String())
		base.GLog.Error(err.Error())
		return err
	}

	// 更新数据库
	err := userFinance.VipInfo.SyncDB(common.GDBEngine, userFinance.Uin)
	return err
}

// 得到用户VIP类型
func GetFinanceUserVIPType(player *proto.TblFinanceUserS) proto.UserVIPType {
	location, _ := time.LoadLocation(player.TimeZone)
	now := time.Now().In(location)

	var vipExpired bool = false

	//base.GLog.Debug("Uin[%d] is %s player", player.Uin, player.VipInfo.VIPType.String())

	if player.VipInfo.VIPType&proto.E_USER_VIP_NORMALMONTHLY != 0 && now.After(player.VipInfo.NormalMonthVIPExpireTime) {
		player.VipInfo.VIPType ^= proto.E_USER_VIP_NORMALMONTHLY
		base.GLog.Warn("Uin[%d] vipType[%s] expired! vip[%s] now[%s] NormalMonthVIPExpireTime[%s]", player.Uin, proto.E_USER_VIP_NORMALMONTHLY.String(),
			player.VipInfo.VIPType.String(), base.TimeName(now), base.TimeName(player.VipInfo.NormalMonthVIPExpireTime))
		vipExpired = true
	}

	if player.VipInfo.VIPType&proto.E_USER_VIP_LUXURYMONTHLY != 0 && now.After(player.VipInfo.LuxuryMonthVIPExpireTime) {
		player.VipInfo.VIPType ^= proto.E_USER_VIP_LUXURYMONTHLY
		base.GLog.Warn("Uin[%d] vipType[%s] expired! vip[%s] now[%s] LuxuryMonthVIPExpireTime[%s]", player.Uin, proto.E_USER_VIP_LUXURYMONTHLY.String(),
			player.VipInfo.VIPType.String(), base.TimeName(now), base.TimeName(player.VipInfo.LuxuryMonthVIPExpireTime))
		vipExpired = true
	}

	if vipExpired {
		// 更新db
		player.VipInfo.SyncDB(common.GDBEngine, player.Uin)
	}
	base.GLog.Debug("Uin[%d] vipType[%s]", player.Uin, player.VipInfo.VIPType.String())
	return player.VipInfo.VIPType
}

func GetUserFinanceBusinessRedLights(req *proto.ProtoGetFinanceBusinessRedLightsReq) (*proto.ProtoGetFinanceBusinessRedLightsRes, error) {
	userFinance, err := QueryFinanceUser(req.Uin)
	if err != nil {
		return nil, err
	}
	location, _ := time.LoadLocation(userFinance.TimeZone)
	now := time.Now().In(location)
	currDate := base.GetDateNum(location)
	currMonthName := base.GetMonthName(location)

	signInConfig := getMonthlySigninConfigInfo(req.ZoneID)
	if signInConfig == nil {
		err = fmt.Errorf("Uin[%d] zoneID[%d] is invalid!", req.Uin, req.ZoneID)
		base.GLog.Error(err.Error())
		return nil, err
	}

	var res proto.ProtoGetFinanceBusinessRedLightsRes
	res.Uin = req.Uin
	res.ZoneID = req.ZoneID
	res.FinanceBusinessRedLightInfos = make([]proto.FinanceBusinessRedLightInfo, proto.FinanceUpdateBusinessTypeCount)

	// 判断时间
	// commonshop
	autoRefreshTime := getAutoRefreshShopTime(userFinance, proto.C_REFRESHSHOPTYPE_NORMAL)
	res.FinanceBusinessRedLightInfos[0].BusinessType = proto.E_UPDATEBUSINESS_COMMONSHOP
	if now.After(autoRefreshTime) {
		res.FinanceBusinessRedLightInfos[0].RedPointIsLight = 1
	}
	res.FinanceBusinessRedLightInfos[0].RemainderSeconds = int64(autoRefreshTime.Sub(now).Seconds())

	// breakoutshop
	autoRefreshTime = getAutoRefreshShopTime(userFinance, proto.C_REFRESHSHOPTYPE_BREAKOUT)
	res.FinanceBusinessRedLightInfos[1].BusinessType = proto.E_UPDATEBUSINESS_BREAKOUTSHOP
	if now.After(autoRefreshTime) {
		res.FinanceBusinessRedLightInfos[1].RedPointIsLight = 1
	}
	res.FinanceBusinessRedLightInfos[1].RemainderSeconds = int64(autoRefreshTime.Sub(now).Seconds())

	// signin
	res.FinanceBusinessRedLightInfos[2].BusinessType = proto.E_UPDATEBUSINESS_SIGIN
	if userFinance.SignInInfo.MonthName != currMonthName {
		res.FinanceBusinessRedLightInfos[2].RedPointIsLight = 1
	} else {
		if userFinance.SignInInfo.SigninDate < currDate {
			res.FinanceBusinessRedLightInfos[2].RedPointIsLight = 1
		} else if userFinance.SignInInfo.SigninDate == currDate {
			if userFinance.SignInInfo.ReSigninDate < currDate && req.Activity > signInConfig.RessiueActivityThreshold {
				res.FinanceBusinessRedLightInfos[2].RedPointIsLight = 1
			}
		}
	}
	if res.FinanceBusinessRedLightInfos[2].RedPointIsLight == 1 {
		nextDayStartTime := base.GetNextDayStartTimeByLocation(location)
		res.FinanceBusinessRedLightInfos[2].RemainderSeconds = int64(nextDayStartTime.Sub(now).Seconds())
	}

	return &res, nil
}

func QueryNewPlayerLoginBenefit(uin uint64) (*proto.TblNewPlayerLoginBenefits, error) {
	newPlayerLoginBenefits := new(proto.TblNewPlayerLoginBenefits)
	newPlayerLoginBenefits.Uin = uin
	exists, err := mysql.GetRecord(common.GDBEngine, proto.TBNAME_NEWPLAYERLOGINBENEFITS, newPlayerLoginBenefits)
	if err != nil {
		base.GLog.Error("Query TblNewPlayerLoginBenefits[%d] failed! reason[%s]", uin, err.Error())
		return nil, err
	}

	if !exists {
		err := fmt.Errorf("Query TblNewPlayerLoginBenefits[%d] does not exist!", uin)
		base.GLog.Error(err.Error())
		return nil, err
	}

	base.GLog.Debug("%+v", newPlayerLoginBenefits)
	return newPlayerLoginBenefits, nil
}
