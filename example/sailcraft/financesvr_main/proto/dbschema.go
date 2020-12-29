/*
 * @Author: calmwu
 * @Date: 2018-02-05 14:45:46
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-04-16 18:18:31
 * @Comment:
 */

package proto

import (
	"encoding/json"
	"fmt"
	"sailcraft/base"
	"sailcraft/dataaccess/mysql"
	"sailcraft/financesvr_main/common"
	"time"

	"github.com/go-xorm/builder"
	"github.com/go-xorm/core"
)

const TBNAME_FINANCEUSER = "tbl_UserFinance"
const TBNAME_NEWPLAYERLOGINBENEFITS = "tbl_NewPlayerLoginBenefits" // 新用户登录福利表
const TBNAME_PLAYERACTIVE = "tbl_PlayerActive"                     // 玩家活动表
const TBNAME_PLAYERCDKEYEXCHANGE = "tbl_PlayerCDKeyExchange"
const DEFAULT_TASKTYPE = "nil_type"

// 数据库表对象定义
type TblFinanceUserS struct {
	Uin                        uint64                        `xorm:"int pk 'Uin'"`                       //
	ZoneID                     int32                         `xorm:"int 'ZoneID'"`                       //
	TimeZone                   string                        `xorm:"string 'TimeZone'"`                  //
	FirstRecharge              TBPlayerFirstRechargeInfoS    `xorm:"jsonb 'FirstRecharge'"`              // 首冲任务情况
	VipInfo                    TBPlayerVIPInfoS              `xorm:"jsonb 'VipInfo'"`                    // 玩家消费卡类型
	ShopFirstPurchaseInfo      TBShopFirstPurchaseInfo       `xorm:"jsonb 'ShopFirstPurchaseInfo'"`      // 商店商品首次购买信息
	PlayerRefreshShopDailyInfo TBPlayerRefreshShopDailyInfoS `xorm:"jsonb 'PlayerRefreshShopDailyInfo'"` // 刷新商店信息
	SignInInfo                 TBPlayerMonthlySignInInfoS    `xorm:"jsonb 'SignInInfo'"`                 // 签到信息
}

func (TblUserFinance TblFinanceUserS) TableName() string {
	return TBNAME_FINANCEUSER
}

// 玩家当天刷新商店行为，数据库blob字段内容
type TBPlayerRefreshShopDailyInfoS struct {
	CommonShopAutoRefreshTime   time.Time              `json:"CSRefreshTime" mapstructure:"CSRefreshTime"` // 自动刷新的时间，这个是用户时区的时间
	BreakoutShopAutoRefreshTime time.Time              `json:"BSRefreshTime" mapstructure:"BSRefreshTime"` // 自动刷新的时间，这个是用户时区的时间
	PlayerCommonShopDailyInfo   PlayerRefreshShopInfoS `json:"PCSDailyInfo" mapstructure:"PCSDailyInfo"`   // 传统商店
	PlayerBreakoutShopDailyInfo PlayerRefreshShopInfoS `json:"PBSDailyInfo" mapstructure:"PBSDailyInfo"`   // 突出重围商店
}

func (rsd *TBPlayerRefreshShopDailyInfoS) Init(uin uint64, startHours int32, now time.Time) {
	rsd.CommonShopAutoRefreshTime =
		time.Date(now.Year(), now.Month(), now.Day(), int(startHours), 0, 0, 0, now.Location())
	rsd.BreakoutShopAutoRefreshTime = rsd.CommonShopAutoRefreshTime
	base.GLog.Debug("New Uin[%d] CommonShopAutoRefreshTime, BreakoutShopAutoRefreshTime[%s]", uin, base.TimeName(rsd.CommonShopAutoRefreshTime))
	rsd.PlayerCommonShopDailyInfo.AlreadyPurchasedCommodities =
		make([]AlreadyPurchasedCommodityInfoS, 0)
	rsd.PlayerCommonShopDailyInfo.CurrentDisplayCommodities =
		make([]RefreshShopCommodityS, 0)
	rsd.PlayerBreakoutShopDailyInfo.AlreadyPurchasedCommodities =
		make([]AlreadyPurchasedCommodityInfoS, 0)
	rsd.PlayerBreakoutShopDailyInfo.CurrentDisplayCommodities =
		make([]RefreshShopCommodityS, 0)
}

func (rsd *TBPlayerRefreshShopDailyInfoS) SyncDB(engine *mysql.DBEngineInfoS, uin uint64) error {
	fieldVal, _ := json.Marshal(rsd)
	cond := fmt.Sprintf("Uin=%d", uin)
	affected, err := mysql.UpdateRecordSpecifiedFieldsByCond(common.GDBEngine, TBNAME_FINANCEUSER, cond,
		map[string]interface{}{"PlayerRefreshShopDailyInfo": fieldVal})

	if err != nil {
		base.GLog.Critical("Update Uin[%d] TBPlayerRefreshShopDailyInfoS failed! reason[%s]", uin, err.Error())
	}
	base.GLog.Debug("Update Uin[%d] TBPlayerRefreshShopDailyInfoS:%+v successed! affected[%d]", uin, *rsd, affected)
	return err
}

type TbShopFirstPurchaseCommodity struct {
	CommodityType ShopCommodityType `json:"CommodityType"`
	CommodityID   int               `json:"CommodityID"`
}

type TBShopFirstPurchaseInfo struct {
	ShopFirstPurchaseCommodityLst []TbShopFirstPurchaseCommodity `json:"ShopFirstPurchaseCommodityLst"`
}

func (sfp *TBShopFirstPurchaseInfo) Init() {
	sfp.ShopFirstPurchaseCommodityLst = make([]TbShopFirstPurchaseCommodity, 0)
}

func (sfp *TBShopFirstPurchaseInfo) IsFirstPurchase(commodityType ShopCommodityType, commodityID int) bool {
	for index := range sfp.ShopFirstPurchaseCommodityLst {
		shopFirstPurchaseInfo := &sfp.ShopFirstPurchaseCommodityLst[index]
		if shopFirstPurchaseInfo.CommodityID == commodityID &&
			shopFirstPurchaseInfo.CommodityType == commodityType {
			base.GLog.Debug("ShopCommodityType[%s] commodityID[%d] had been purchased!",
				commodityType.String(), commodityID)
			return false
		}
	}
	base.GLog.Debug("ShopCommodityType[%s] commodityID[%d] is not purchased!", commodityType.String(), commodityID)
	return true
}

func (sfp *TBShopFirstPurchaseInfo) FirstPurchase(commodityType ShopCommodityType, commodityID int, uin uint64) {
	// 直接加入
	sfp.ShopFirstPurchaseCommodityLst = append(sfp.ShopFirstPurchaseCommodityLst, TbShopFirstPurchaseCommodity{
		CommodityType: commodityType,
		CommodityID:   commodityID,
	})

	fieldVal, _ := json.Marshal(sfp)
	_, err := mysql.UpdateRecordSpecifiedFieldsByCond(common.GDBEngine, TBNAME_FINANCEUSER,
		fmt.Sprintf("Uin=%d", uin),
		map[string]interface{}{
			"ShopFirstPurchaseInfo": fieldVal})
	if err != nil {
		base.GLog.Critical("Uin[%d] set ShopFirstPurchaseInfo failed! reason[%s]", uin, err.Error())
	} else {
		base.GLog.Debug("Uin[%d] set ShopFirstPurchaseInfo successed!", uin)
	}
	return
}

// 用户的每月签到信息，存放db
type TBPlayerMonthlySignInInfoS struct {
	SigninDate         int   `json:"SigninDate"`   // 签到当前的日期 20180505
	ReSigninDate       int   `json:"ReSigninDate"` // 补签的日期
	MonthName          int   `json:"MonthName"`    // month标识，用于标识本月,201801 201812
	MonthlySignInCount int32 `json:"SignInCount"`  // 本月签到次数
}

func (si *TBPlayerMonthlySignInInfoS) Init(location *time.Location) {
	si.MonthName = base.GetMonthName(location)
	si.MonthlySignInCount = 0
	si.SigninDate = 0
	si.ReSigninDate = 0
}

func (si *TBPlayerMonthlySignInInfoS) Month(location *time.Location) {
	currMonthName := base.GetMonthName(location)
	if si.MonthName != currMonthName {
		// 新的月份到了
		base.GLog.Debug("New month[%d] oldMonth[%d]", currMonthName, si.MonthName)
		si.MonthName = currMonthName
		si.MonthlySignInCount = 0
		si.SigninDate = 0
		si.ReSigninDate = 0
	}
}

func (si *TBPlayerMonthlySignInInfoS) SyncDB(engine *mysql.DBEngineInfoS, uin uint64) error {
	fieldVal, _ := json.Marshal(si)
	cond := fmt.Sprintf("Uin=%d", uin)
	affected, err := mysql.UpdateRecordSpecifiedFieldsByCond(common.GDBEngine, TBNAME_FINANCEUSER, cond,
		map[string]interface{}{"SignInInfo": fieldVal})

	if err != nil {
		base.GLog.Critical("Update Uin[%d] TBPlayerMonthlySignInInfoS failed! reason[%s]", uin, err.Error())
	}
	base.GLog.Debug("Update Uin[%d] TBPlayerMonthlySignInInfoS:%+v successed! affected[%d]", uin, *si, affected)
	return err
}

type TBPlayerVIPInfoS struct {
	VIPType UserVIPType `json:"VIPType" mapstructure:"VIPType"` // 会员类型

	NormalMonthVIPExpireTime             time.Time `json:"WVIPExpireTime" mapstructure:"WVIPExpireTime"` // 普通月卡过期时间
	NormalMonthVIPCollectPrizeCount      int32     `json:"WVCPCount" mapstructure:"WVCPCount"`           // 领取次数
	NormalMonthVIPCollectPrizeDate       int32     `json:"WVCPDate" mapstructure:"WVCPDate"`             // 领取的时间，一天就一次
	NormalMonthVIPCollectPrizeExpireDate int32     `json:"WVCPEDate" mapstructure:"WVCPEDate"`           // 领取的过期日期

	LuxuryMonthVIPExpireTime             time.Time `json:"MVIPExpireTime" mapstructure:"MVIPExpireTime"` // 月卡过期时间
	LuxuryMonthVIPCollectPrizeCount      int32     `json:"MVCPCount" mapstructure:"MVCPCount"`           // 领取次数
	LuxuryMonthVIPCollectPrizeDate       int32     `json:"MVCPDate" mapstructure:"MVCPDate"`             // 领取的日期
	LuxuryMonthVIPCollectPrizeExpireDate int32     `json:"MVCPEDate" mapstructure:"MVCPEDate"`           // 领取的过期日期
}

func (vip *TBPlayerVIPInfoS) Init(now time.Time) {
	vip.VIPType = E_USER_VIP_NO
	vip.NormalMonthVIPExpireTime = now
	vip.LuxuryMonthVIPExpireTime = now
}

func (vip *TBPlayerVIPInfoS) SyncDB(engine *mysql.DBEngineInfoS, uin uint64) error {
	cond := fmt.Sprintf("Uin=%d", uin)
	fieldVal, _ := json.Marshal(vip)
	affected, err := mysql.UpdateRecordSpecifiedFieldsByCond(engine, TBNAME_FINANCEUSER, cond,
		map[string]interface{}{"VipInfo": fieldVal})
	//affected, err := mysql.UpdateRecord(common.GDBEngine, proto.TBNAME_FINANCEUSER, PK, userFinance)
	if err != nil {
		base.GLog.Critical("Update Uin[%d] TBPlayerVIPInfoS failed! reason[%s]", uin, err.Error())
	}
	base.GLog.Debug("Update Uin[%d] TBPlayerVIPInfoS:%+v successed! affected[%d]", uin, *vip, affected)
	return err
}

type FirstRechargeLevelInfoS struct {
	ActiveID int   `json:"ID"`
	Received int32 `json:"Received"` // 是否领取过，1：已经领取 0：还没领取
}

type TBPlayerFirstRechargeInfoS struct {
	CurrBuyDiamonds int32                     `json:"CurrBuyDiamonds"` // 当前购买的钻石数量
	ReceiveCount    int                       `json:"ReceiveCount"`    // 领取的次数,=0的时候领取完毕
	LevelInfoLst    []FirstRechargeLevelInfoS `json:"LevelInfoLst"`    // 等级完成情况
}

func (fr *TBPlayerFirstRechargeInfoS) Init(config *ProtoFirstRechargeConfigS) {
	fr.CurrBuyDiamonds = 0
	fr.ReceiveCount = len(config.FRLevelConfLst)
	fr.LevelInfoLst = make([]FirstRechargeLevelInfoS, len(config.FRLevelConfLst))
	for index := range fr.LevelInfoLst {
		fr.LevelInfoLst[index].ActiveID = config.FRLevelConfLst[index].Id
	}
}

func (fr *TBPlayerFirstRechargeInfoS) SyncDB(engine *mysql.DBEngineInfoS, uin uint64) error {
	cond := fmt.Sprintf("Uin=%d", uin)
	fieldVal, _ := json.Marshal(fr)
	affected, err := mysql.UpdateRecordSpecifiedFieldsByCond(engine, TBNAME_FINANCEUSER, cond,
		map[string]interface{}{"FirstRecharge": fieldVal})
	//affected, err := mysql.UpdateRecord(common.GDBEngine, proto.TBNAME_FINANCEUSER, PK, userFinance)
	if err != nil {
		base.GLog.Critical("Update Uin[%d] FirstRecharge failed! reason[%s]", uin, err.Error())
	}
	base.GLog.Debug("Update Uin[%d] FirstRecharge:%+v successed! affected[%d]", uin, *fr, affected)
	return err
}

//----------------------------------------------------------------------------------

type TblNewPlayerLoginBenefits struct {
	Uin               uint64  `xorm:"int pk 'Uin'"`             //
	ZoneID            int32   `xorm:"int index 'ZoneID'"`       //
	TimeZone          string  `xorm:"string 'TimeZone'"`        //
	CreateDate        int     `xorm:"int 'CreateDate'"`         // 账号创建时间
	LoginDays         int32   `xorm:"int 'LoginDays'"`          // 登录天数，一天一次
	LastLoginDate     int     `xorm:"int 'LastLoginDate'"`      // 最后登录的日期
	ReceiveAwardTags  []int32 `xorm:"jsonb 'ReceiveAwardTags'"` // 领奖标识
	ReceiveAwardCount int32   `xorm:"int 'ReceiveAwardCount'"`  // 领奖次数
	IsCompleted       int32   `xorm:"int 'IsCompleted'"`        // 领取是否全部结束，1：结束，0：未完成
}

func (lb *TblNewPlayerLoginBenefits) Init(uin uint64, zoneID int32, timeZone string, location *time.Location) {
	lb.Uin = uin
	lb.ZoneID = zoneID
	lb.TimeZone = timeZone
	lb.CreateDate = base.GetDateNum(location)
	lb.LastLoginDate = lb.CreateDate
	lb.LoginDays = 1
	lb.ReceiveAwardTags = make([]int32, C_NEWPLAYER_BENEFIT_DAYS)
	lb.ReceiveAwardCount = 0
	lb.IsCompleted = 0
}

func (lb *TblNewPlayerLoginBenefits) SyncDB(engine *mysql.DBEngineInfoS) error {

	PK := core.NewPK(lb.Uin)
	affected, err := mysql.UpdateRecord(engine, TBNAME_NEWPLAYERLOGINBENEFITS, PK, lb)
	if err != nil {
		base.GLog.Error("Update %s[%d] failed! reason[%s]", TBNAME_NEWPLAYERLOGINBENEFITS, lb.Uin, err.Error())
		return err
	} else {
		base.GLog.Debug("Update %s[%d] successed! affected[%d]", TBNAME_NEWPLAYERLOGINBENEFITS, lb.Uin, affected)
	}
	return nil
}

//----------------------------------------------------------------------------------
// 玩家活动信息
type TblPlayerActiveInfo struct {
	Uin             uint64     `xorm:"int pk 'Uin'"`               //
	ZoneID          int32      `xorm:"int index 'ZoneID'"`         //
	ActiveType      ActiveType `xorm:"int pk 'ActiveType'"`        // 活动类型id，超值礼包、兑换
	ActiveID        int        `xorm:"int pk 'ActiveID'"`          // 同类型下的活动id
	ChannelID       string     `xorm:"string 'ChannelID'"`         // CN US
	AccumulateCount int32      `xorm:"int 'AccumulateCount'"`      // 该活动累积数量，有的次数，有的是钻石数量
	ReceiveCount    int32      `xorm:"int 'ReceiveCount'"`         // 领取的数量，有的活动只能领取一次，有的可以多次
	ActiveResetTime time.Time  `xorm:"DateTime 'ActiveResetTime'"` // 玩家参与活动时间
	ActiveStartTime time.Time  `xorm:"DateTime 'ActiveStartTime'"` // 该活动开始时间
	ActiveEndTime   time.Time  `xorm:"DateTime 'ActiveEndTime'"`   // 该活动结束时间
	TaskType        string     `xorm:"string index 'TaskType'"`    // 活动类型，活跃任务需要
}

func CreatePlayerActiveInfo(uin uint64, zoneID int32, activeType ActiveType, activeID int, channelID string,
	activeStartTime, activeEndTime, activeResetTime *time.Time, taskType string, accumulateCount int32,
	engine *mysql.DBEngineInfoS) *TblPlayerActiveInfo {
	playerActiveInfo := new(TblPlayerActiveInfo)
	playerActiveInfo.Uin = uin
	playerActiveInfo.ZoneID = zoneID
	playerActiveInfo.ActiveType = activeType
	playerActiveInfo.ActiveID = activeID
	playerActiveInfo.ChannelID = channelID
	playerActiveInfo.AccumulateCount = accumulateCount
	playerActiveInfo.ReceiveCount = 0
	playerActiveInfo.ActiveResetTime = *activeResetTime
	playerActiveInfo.ActiveStartTime = *activeStartTime
	playerActiveInfo.ActiveEndTime = *activeEndTime
	playerActiveInfo.TaskType = taskType
	affected, err := mysql.InsertRecord(common.GDBEngine, TBNAME_PLAYERACTIVE, playerActiveInfo)
	if err != nil {
		base.GLog.Error("Insert %s[%d] activeType[%s] activeID[%d] failed! reason[%s]",
			TBNAME_PLAYERACTIVE, uin, activeType.String(), activeID, err.Error())
		return nil
	}
	base.GLog.Debug("Insert %s[%d] activeType[%s] activeID[%d] successed! affected[%d]",
		TBNAME_PLAYERACTIVE, uin, activeType.String(), activeID, affected)
	return playerActiveInfo
}

func (pa *TblPlayerActiveInfo) ResetActive(activeStartTime, activeEndTime, activeResetTime *time.Time, engine *mysql.DBEngineInfoS) error {
	pa.ReceiveCount = 0
	pa.AccumulateCount = 0
	pa.ActiveStartTime = *activeStartTime
	pa.ActiveEndTime = *activeEndTime
	pa.ActiveResetTime = *activeResetTime
	return pa.SyncDB(engine)
}

func (pa *TblPlayerActiveInfo) SyncDB(engine *mysql.DBEngineInfoS) error {
	PK := core.NewPK(pa.Uin, pa.ActiveType, pa.ActiveID)
	affected, err := mysql.UpdateRecord(engine, TBNAME_PLAYERACTIVE, PK, pa)
	if err != nil {
		base.GLog.Error("Update %s[%d] activeType[%s] activeID[%d] failed! reason[%s]",
			TBNAME_PLAYERACTIVE, pa.Uin, pa.ActiveType.String(), pa.ActiveID, err.Error())
		return err
	} else {
		base.GLog.Debug("Update %s[%d] successed! affected[%d]", TBNAME_PLAYERACTIVE, pa.Uin, affected)
	}
	return nil
}

func QueryPlayerActiveInfo(uin uint64, activeType ActiveType, activeID int, engine *mysql.DBEngineInfoS) *TblPlayerActiveInfo {
	playerActiveInfo := new(TblPlayerActiveInfo)
	playerActiveInfo.Uin = uin
	playerActiveInfo.ActiveType = activeType
	playerActiveInfo.ActiveID = activeID

	exists, err := mysql.GetRecord(engine, TBNAME_PLAYERACTIVE, playerActiveInfo)
	if err != nil {
		base.GLog.Error("Query %s[%d] activeType[%s] activeID[%d] failed! reason[%s]",
			TBNAME_PLAYERACTIVE, uin, activeType.String(), activeID, err.Error())
		return nil
	}

	if !exists {
		err := fmt.Errorf("Query %s[%d] activeType[%s] activeID[%d] not exists!",
			TBNAME_PLAYERACTIVE, uin, activeType.String(), activeID)
		base.GLog.Error(err.Error())
		return nil
	}
	base.GLog.Debug("playerActiveInfo:%+v", playerActiveInfo)
	return playerActiveInfo
}

func QueryPlayerMissionActionInfos(uin uint64, taskType string, engine *mysql.DBEngineInfoS) ([]TblPlayerActiveInfo, error) {
	result := make([]TblPlayerActiveInfo, 0)
	var err error

	cond := builder.Expr(fmt.Sprintf("Uin=%d", uin)).And(builder.Expr(fmt.Sprintf("TaskType='%s'", taskType)))
	err = mysql.FindRecordsByMultiConds(engine, TBNAME_PLAYERACTIVE, &cond, 0, 0, &result)
	if err != nil {
		base.GLog.Error("Query %s[%d] MissionActionInfo taskType[%s] failed! reason[%s]",
			TBNAME_PLAYERACTIVE, uin, taskType, err.Error())
		return nil, err
	}

	return result, nil
}

//----------------------------------------------------------------------------------
// 玩家cdkey兑换
type TblPlayerCDKeyExchange struct {
	Uin    uint64                 `xorm:"int pk 'Uin'"`       //
	ZoneID int32                  `xorm:"int index 'ZoneID'"` //
	KHLst  TblFieldExchangeKeyLst `xorm:"jsonb 'KHLst'"`
}

type TblFieldExchangeKeyLst struct {
	CDKeyHashValLst []uint32 `json:"KHLst" mapstructure:"KHLst"` // 已经兑换过的key的hash列表
}

func CreatePlayerCDKeyExchange(uin uint64, zoneID int32, engine *mysql.DBEngineInfoS) *TblPlayerCDKeyExchange {
	playerCDKeyExchange := new(TblPlayerCDKeyExchange)
	playerCDKeyExchange.Uin = uin
	playerCDKeyExchange.ZoneID = zoneID
	playerCDKeyExchange.KHLst.CDKeyHashValLst = make([]uint32, 0)
	affected, err := mysql.InsertRecord(common.GDBEngine, TBNAME_PLAYERCDKEYEXCHANGE, playerCDKeyExchange)
	if err != nil {
		base.GLog.Error("Insert %s[%d] failed! reason[%s]",
			TBNAME_PLAYERCDKEYEXCHANGE, uin, err.Error())
		return nil
	}
	base.GLog.Debug("Insert %s[%d]successed! affected[%d]", TBNAME_PLAYERCDKEYEXCHANGE, uin, affected)
	return playerCDKeyExchange
}

func QueryPlayerCDKeyExchange(uin uint64, engine *mysql.DBEngineInfoS) *TblPlayerCDKeyExchange {
	playerCDKeyExchange := new(TblPlayerCDKeyExchange)
	playerCDKeyExchange.Uin = uin

	exists, err := mysql.GetRecord(engine, TBNAME_PLAYERCDKEYEXCHANGE, playerCDKeyExchange)
	if err != nil {
		base.GLog.Error("Query %s[%d] failed! reason[%s]", TBNAME_PLAYERCDKEYEXCHANGE, uin, err.Error())
		return nil
	}

	if !exists {
		err := fmt.Errorf("Query %s[%d] not exists!", TBNAME_PLAYERCDKEYEXCHANGE, uin)
		base.GLog.Error(err.Error())
		return nil
	}
	base.GLog.Debug("playerCDKeyExchange:%+v", playerCDKeyExchange)
	return playerCDKeyExchange
}

func (ke *TblPlayerCDKeyExchange) SyncDB(engine *mysql.DBEngineInfoS) error {
	PK := core.NewPK(ke.Uin)
	affected, err := mysql.UpdateRecord(engine, TBNAME_PLAYERCDKEYEXCHANGE, PK, ke)
	if err != nil {
		base.GLog.Error("Update %s[%d]failed! reason[%s]",
			TBNAME_PLAYERCDKEYEXCHANGE, ke.Uin, err.Error())
		return err
	} else {
		base.GLog.Debug("Update %s[%d] successed! affected[%d]", TBNAME_PLAYERCDKEYEXCHANGE, ke.Uin, affected)
	}
	return nil
}

func (ke *TblPlayerCDKeyExchange) ExchangeCDKey(cdkey string) bool {
	cdKeyHashVal := base.HashStr2Uint32(cdkey)
	for _, val := range ke.KHLst.CDKeyHashValLst {
		if val == cdKeyHashVal {
			base.GLog.Error("Uin[%d] cdkey[%s]:[%d] had exchanged!", ke.Uin, cdkey, val)
			return false
		}
	}
	// 加入
	ke.KHLst.CDKeyHashValLst = append(ke.KHLst.CDKeyHashValLst, cdKeyHashVal)
	return true
}
