/*
 * @Author: calmwu
 * @Date: 2018-01-31 16:12:16
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-05-29 14:31:56
 * @Comment:
 */

package proto

import (
	"encoding/json"
	"sailcraft/base"
)

// 小红点，Finance中可更新的业务类型
type FinanceUpdateBusinessType int32

const (
	FinanceUpdateBusinessTypeCount = 3

	C_CHANNELNAME_NOAREA = "NOAREA"
	C_CHANNELNAME_CN     = "CN"
	C_CHANNELNAME_US     = "US"
)

const (
	E_UPDATEBUSINESS_COMMONSHOP FinanceUpdateBusinessType = iota
	E_UPDATEBUSINESS_BREAKOUTSHOP
	E_UPDATEBUSINESS_SIGIN
)

type FinanceBusinessRedLightInfo struct {
	BusinessType     FinanceUpdateBusinessType `json:"BusinessType" mapstructure:"BusinessType"`
	RedPointIsLight  int32                     `json:"RedPointIsLight" mapstructure:"RedPointIsLight"`   // =1：亮着
	RemainderSeconds int64                     `json:"RemainderSeconds" mapstructure:"RemainderSeconds"` // 剩余时间，倒计时秒
}

//-----------------------------------------------------------------------------------------

type ShopCommodityType int32

const (
	E_SHOPCOMMODITY_RECHARGE ShopCommodityType = iota
	E_SHOPCOMMODITY_RESOURCE
	E_SHOPCOMMODITY_CARDPACK
)

// 用户的消费类型，普通、普通月卡、月卡，月卡和普通月卡可以同时激活
type UserVIPType int32

const (
	E_USER_VIP_NO            UserVIPType = 0x00
	E_USER_VIP_NORMALMONTHLY UserVIPType = 0x01
	E_USER_VIP_LUXURYMONTHLY UserVIPType = 0x02
	E_USER_VIP_ALL           UserVIPType = 0x03
)

// 充值商品类型
type RechargeCommodityType int32

const (
	E_RECHARGECOMMODITY_DIAMONDS          RechargeCommodityType = iota // 购买钻石
	E_RECHARGECOMMODITY_NORMALMONTHLYCARD                              // 购买普通月卡
	E_RECHARGECOMMODITY_LUXURYMONTHLYCARD                              // 购买豪华月卡
	E_RECHARGECOMMODITY_SUPERGIFT                                      // 购买超级礼包
	E_RECHARGECOMMODITY_FIRSTRECHARGE                                  // 首充
)

// 创建用户
type ProtoNewFinanceUserReq struct {
	Uin      uint64 `json:"Uin" mapstructure:"Uin"`
	ZoneID   int32  `json:"ZoneID" mapstructure:"ZoneID"`
	TimeZone string `json:"TimeZone" mapstructure:"TimeZone"`
}

type ProtoNewFinanceUserRes struct {
	Uin    uint64 `json:"Uin" mapstructure:"Uin"`
	ZoneID int32  `json:"ZoneID" mapstructure:"ZoneID"`
	Result int    `json:"Result" mapstructure:"Result"`
}

// 充值商品
type RechargeCommodityInfoS struct {
	RechargeCommodityID               int     `json:"ID" mapstructure:"ID"`
	BuyDiamonds                       int32   `json:"BDiamonds" mapstructure:"BDiamonds"` // 默认购买获得的钻石数量
	FirstRechargePresentDiamonds      int32   `json:"FDiamonds" mapstructure:"FDiamonds"` // 首冲获得的钻石数量
	PresentDiamonds                   int32   `json:"PDiamonds" mapstructure:"PDiamonds"` // 额外赠送的钻石数量
	FirstTimePurchaseRewardCaptionKey string  `json:"FTPRCKey" mapstructure:"FTPRCKey"`   //FirstTimePurchaseRewardCaptionKey
	Price                             float32 `json:"Price" mapstructure:"Price"`
	PriceDesc                         string  `json:"PriceDesc" mapstructure:"PriceDesc"`
	ProductId                         string  `json:"ProductId" mapstructure:"ProductId"`
	PurchasedFlag                     int32   `json:"PF" mapstructure:"PF"` // 购买过标识，=1 购买过，=0 没有购买过
}

// 渠道商品列表
type ChannelRechargeCommodityInfoS struct {
	ChannelID           string                   `json:"ChannelID" mapstructure:"ChannelID"`
	RechargeCommodities []RechargeCommodityInfoS `json:"RechargeCommodities" mapstructure:"RechargeCommodities"`
}

type ShopRechargeCommoditiesInfoS struct {
	VersionID                     string                          `json:"VersionID" mapstructure:"VersionID"`
	ChannelRechargeCommodityInfos []ChannelRechargeCommodityInfoS `json:"ChannelRechargeCommodityInfos" mapstructure:"ChannelRechargeCommodityInfos"`
}

func (src ShopRechargeCommoditiesInfoS) String() string {
	val, err := json.Marshal(src)
	if err == nil {
		return string(val)
	}
	return ""
}

func (src ShopRechargeCommoditiesInfoS) FindChannelRechargeCommodities(channelID string) *ChannelRechargeCommodityInfoS {
	for j := range src.ChannelRechargeCommodityInfos {
		if channelID == src.ChannelRechargeCommodityInfos[j].ChannelID {
			return &src.ChannelRechargeCommodityInfos[j]
		}
	}
	base.GLog.Error("channelID[%s] is invalid!", channelID)
	return nil
}

func (src ShopRechargeCommoditiesInfoS) FindRechargeCommodity(channelID string, rechargeCommodityID int) *RechargeCommodityInfoS {
	for j := range src.ChannelRechargeCommodityInfos {
		if channelID == src.ChannelRechargeCommodityInfos[j].ChannelID {
			for i := range src.ChannelRechargeCommodityInfos[j].RechargeCommodities {
				if src.ChannelRechargeCommodityInfos[j].RechargeCommodities[i].RechargeCommodityID == rechargeCommodityID {
					return &src.ChannelRechargeCommodityInfos[j].RechargeCommodities[i]
				}
			}
		}
	}
	base.GLog.Error("channelID[%s] rechargeCommodityID[%d] is invalid!", channelID, rechargeCommodityID)
	return nil
}

// 查询商店商品请求
type ProtoQueryShopCommoditiesReq struct {
	Uin       uint64 `json:"Uin" mapstructure:"Uin"`
	ZoneID    int32  `json:"ZoneID" mapstructure:"ZoneID"`
	ClientIP  string `json:"ClientIP" mapstructure:"ClientIP"`
	ChannelID string `json:"ChannelID" mapstructure:"ChannelID"`
}

// 查询充值商品请求
type ProtoQueryRechargeCommoditiesReq = ProtoQueryShopCommoditiesReq

type ProtoQueryRechargeCommoditiesRes struct {
	Uin                 uint64                   `json:"Uin" mapstructure:"Uin"`
	ZoneID              int32                    `json:"ZoneID" mapstructure:"ZoneID"`
	VersionID           string                   `json:"VersionID" mapstructure:"VersionID"`
	ChannelID           string                   `json:"ChannelID" mapstructure:"ChannelID"`
	RechargeCommodities []RechargeCommodityInfoS `json:"Commodities" mapstructure:"Commodities"`
}

// 支付后tsapi回调发货请求
type ProtoDeliveryRechargeCommodityReq struct {
	Uin                   uint64                `json:"Uin" mapstructure:"Uin"`
	ZoneID                int32                 `json:"ZoneID" mapstructure:"ZoneID"`
	VersionID             string                `json:"VersionID" mapstructure:"VersionID"`
	RechargeCommdityID    int                   `json:"ID" mapstructure:"ID"`
	RechargeCommodityType RechargeCommodityType `json:"RCType" mapstructure:"RCType"`       // RechargeCommodityType
	ChannelID             string                `json:"ChannelID" mapstructure:"ChannelID"` // CN US
	PlatForm              string                `json:"PlatForm"`                           // IOS ANDROID
}

type ProtoDeliveryRechargeCommodityRes struct {
	Uin                   uint64                `json:"Uin" mapstructure:"Uin"`
	ZoneID                int32                 `json:"ZoneID" mapstructure:"ZoneID"`
	RechargeCommdityID    int                   `json:"ID" mapstructure:"ID"`
	RechargeCommodityType RechargeCommodityType `json:"RCType" mapstructure:"RCType"`                   // RechargeCommodityType
	BuyDiamonds           int32                 `json:"BDiamonds" mapstructure:"BDiamonds"`             // 默认购买获得的钻石数量
	PresentDiamonds       int32                 `json:"PDiamonds" mapstructure:"PDiamonds"`             // 赠送的钻石
	IsFristPurchase       int32                 `json:"IsFristPurchase" mapstructure:"IsFristPurchase"` // 首次购买标志 = 1
	InnerGoods            string                `json:"IG"`                                             // 返回的奖品，json数据，例如超级礼包
	NameKey               string                `json:"NK"`
}

// GM工具刷新充值商品
type ProtoRefreshRechargeCommoditiesReq struct {
	Uin                         uint64                       `json:"Uin" mapstructure:"Uin"`
	ZoneID                      int32                        `json:"ZoneID" mapstructure:"ZoneID"`
	ShopRechargeCommoditiesInfo ShopRechargeCommoditiesInfoS `json:"ShopRechargeCommoditiesInfo" mapstructure:"ShopRechargeCommoditiesInfo"`
}

// 查询用户消费卡类型
type ProtoUserVIPTypeReq struct {
	Uin    uint64 `json:"Uin" mapstructure:"Uin"`
	ZoneID int32  `json:"ZoneID" mapstructure:"ZoneID"`
}

type ProtoUserVIPTypeRes struct {
	Uin         uint64      `json:"Uin"`
	ZoneID      int32       `json:"ZoneID"`
	UserVIPType UserVIPType `json:"Type" mapstructure:"Type"`
	TimeZone    string      `json:"TimeZone"`
	Result      int         `json:"Result"`
}

//------------------------------------------------------------------------------------------------------------
// 资源商品
type ResourceCommodityType int32

const (
	E_RESOURCECOMMODITY_GOLD ResourceCommodityType = iota
	E_RESOURCECOMMODITY_WOOD
	E_RESOURCECOMMODITY_STONE
	E_RESOURCECOMMODITY_IRON
)

type ResourceCommodityInfoS struct {
	ResourceCommodityID          int                   `json:"ID" mapstructure:"ID"`
	ResourceCommodityType        ResourceCommodityType `json:"RCType" mapstructure:"RCType"`               // 资源商品类型
	ResourceCommodityStackCount  int32                 `json:"RCStackCount" mapstructure:"RCStackCount"`   // 资源堆叠数量
	ResourceCommodityDiamondCost int32                 `json:"RCDiamondCost" mapstructure:"RCDiamondCost"` // 购买所需钻石
}

type ResourceShopConfigS struct {
	VersionID               string                   `json:"VersionID" mapstructure:"VersionID"`
	WeeklyCardDiscountRate  int32                    `json:"WCDR" mapstructure:"WCDR"`
	MonthlyCardDiscountRate int32                    `json:"MCDR" mapstructure:"MCDR"`
	Count                   int                      `json:"Count" mapstructure:"Count"` // 资源商品数量
	ResourceCommodities     []ResourceCommodityInfoS `json:"ResourceCommodities" mapstructure:"ResourceCommodities"`
}

func (src ResourceShopConfigS) String() string {
	val, err := json.Marshal(src)
	if err == nil {
		return string(val)
	}
	return ""
}

func (src ResourceShopConfigS) Find(resourceCommodityID int) *ResourceCommodityInfoS {
	for i := range src.ResourceCommodities {
		if src.ResourceCommodities[i].ResourceCommodityID == resourceCommodityID {
			return &src.ResourceCommodities[i]
		}
	}
	base.GLog.Error("resourceCommodityID[%d] is invalid!", resourceCommodityID)
	return nil
}

// 查询商店资源商品请求
type ProtoQueryResourceCommoditiesReq = ProtoQueryShopCommoditiesReq

type ProtoQueryResourceCommoditiesRes struct {
	Uin                    uint64               `json:"Uin"`
	ZoneID                 int32                `json:"ZoneID"`
	ResourceShopConfigInfo *ResourceShopConfigS `json:"RSCInfo" mapstructure:"RSCInfo"`
}

// GM工具刷新资源商店配置
type ProtoRefreshResourceShopConfigReq struct {
	Uin                    uint64              `json:"Uin" mapstructure:"Uin"`
	ZoneID                 int32               `json:"ZoneID" mapstructure:"ZoneID"`
	ResourceShopConfigInfo ResourceShopConfigS `json:"RSCInfo" mapstructure:"RSCInfo"`
}

// 购买资源商品
type ProtoBuyResourceCommodityReq struct {
	Uin                   uint64                `json:"Uin"`
	ZoneID                int32                 `json:"ZoneID"`
	VersionID             string                `json:"VersionID"`
	ResourceCommdityID    int                   `json:"ID" mapstructure:"ID"`
	ResourceCommodityType ResourceCommodityType `json:"RCType" mapstructure:"RCType"`
}

type ProtoBuyResourceCommodityRes struct {
	Uin                   uint64                `json:"Uin"`
	ZoneID                int32                 `json:"ZoneID"`
	ResourceCommodityType ResourceCommodityType `json:"RCType" mapstructure:"RCType"`
	CostDiamonds          int32                 `json:"Cost" mapstructure:"Cost"`
	ResourceCount         int32                 `json:"ResourceCount"` // 获得的资源数量
}

//------------------------------------------------------------------------------------------------------------
// 卡包商品
type CardPackCommdityInfoS struct {
	CardPackCommodityID int    `json:"ID" mapstructure:"ID"` // 商品id
	CardPackDiamondCost int32  `json:"CPDCost" mapstructure:"CPDCost"`
	CardPackJsonContent string `json:"CPJsonContent" mapstructure:"CPJsonContent"` // 卡包礼包的具体内容，json描述
	CardPackRecommend   int32  `json:"CPRecommend" mapstructure:"CPRecommend"`     // 卡包是否推荐
	CardPackGiftDescKey string `json:"CPGiftDescKey" mapstructure:"CPGiftDescKey"`
	CardPackCountKey    string `json:"CPCountKey" mapstructure:"CPCountKey"`
}

type ShopCardPackCommoditiesInfoS struct {
	VersionID               string                  `json:"VersionID" mapstructure:"VersionID"`
	WeeklyCardDiscountRate  int32                   `json:"WCDR" mapstructure:"WCDR"`
	MonthlyCardDiscountRate int32                   `json:"MCDR" mapstructure:"MCDR"`
	Count                   int                     `json:"Count" mapstructure:"Count"`
	CardPackCommodities     []CardPackCommdityInfoS `json:"CPCommodities" mapstructure:"CPCommodities"`
}

func (spc *ShopCardPackCommoditiesInfoS) String() string {
	val, err := json.Marshal(spc)
	if err == nil {
		return string(val)
	}
	return ""
}

func (spc *ShopCardPackCommoditiesInfoS) Find(cardPackCommodityID int) *CardPackCommdityInfoS {
	for i := range spc.CardPackCommodities {
		if spc.CardPackCommodities[i].CardPackCommodityID == cardPackCommodityID {
			return &spc.CardPackCommodities[i]
		}
	}
	base.GLog.Error("cardPackCommodityID[%d] is invalid!", cardPackCommodityID)
	return nil
}

func (spc *ShopCardPackCommoditiesInfoS) Remove(cardPackCommodityID int) {
	index := 0
	exists := false
	for index, _ = range spc.CardPackCommodities {
		cardPackInfo := &spc.CardPackCommodities[index]
		if cardPackInfo.CardPackCommodityID == cardPackCommodityID {
			exists = true
			break
		}
	}
	if exists {
		spc.CardPackCommodities = append(spc.CardPackCommodities[:index],
			spc.CardPackCommodities[index+1:]...)
		spc.Count--
	}
}

// 查询卡包商品
type ProtoQueryCardPackCommoditiesReq = ProtoQueryShopCommoditiesReq

type ProtoQueryCardPackCommoditiesRes struct {
	Uin                         uint64                        `json:"Uin"`
	ZoneID                      int32                         `json:"ZoneID"`
	ShopCardPackCommoditiesInfo *ShopCardPackCommoditiesInfoS `json:"SCPCommoditiesInfo" mapstructure:"SCPCommoditiesInfo"`
}

// GM工具刷新卡包商品
type ProtoRefreshCardPackShopReq struct {
	Uin                         uint64                        `json:"Uin"`
	ZoneID                      int32                         `json:"ZoneID"`
	ShopCardPackCommoditiesInfo *ShopCardPackCommoditiesInfoS `json:"SCPCommoditiesInfo" mapstructure:"SCPCommoditiesInfo"`
}

// 购买卡包商品
type ProtoBuyCardPackCommodityReq struct {
	Uin                uint64 `json:"Uin"`
	ZoneID             int32  `json:"ZoneID"`
	VersionID          string `json:"VersionID"`
	CardPackCommdityID int    `json:"ID" mapstructure:"ID"`
}

type ProtoBuyCardPackCommodityRes struct {
	Uin                 uint64 `json:"Uin"`
	ZoneID              int32  `json:"ZoneID"`
	CostDiamonds        int32  `json:"CostDiamonds"`                               // 花费的钻石数量
	CardPackJsonContent string `json:"CPJsonContent" mapstructure:"CPJsonContent"` // 卡包礼包的具体内容，json描述
}

//------------------------------------------------------------------------------------------------------------
// 刷新商店类型
type RefreshShopType = string

// 支付类型
type GamePayType int32

// 商品池类型
type GameShopCommodityPoolType int32

const (
	C_REFRESHSHOPTYPE_NORMAL   RefreshShopType = "commonshop"
	C_REFRESHSHOPTYPE_BREAKOUT RefreshShopType = "breakoutshop"
)

const (
	E_GAMEPAY_GOLD    GamePayType = iota // 金币
	E_GAMEPAY_DIAMOND                    // 钻石
	E_GAMEPAY_HONOR                      // 勋章
	E_GAMEPAY_SHIPSOUL
)

const C_REFRESHSHOP_COMMODITY_DISPLAY_COUNT = 6 // 刷新商店中商品显示的数量

type CommodityPoolS struct {
	PoolName          string  `json:"PoolName"`          // 商品池的名字，CommonShopPool_1---->CommonShopPool_1-Zone1-VersionID
	DisplaySlotIndexs []int32 `json:"DisplaySlotIndexs"` // 对应的显示槽位 {0,1,2}, 从0开始
}

// 刷新商店的商品配置
type RefreshShopCommodityS struct {
	CommodityID          int         `json:"ID" mapstructure:"ID"`               // 商品id
	GamePayType          GamePayType `json:"GamePayType"`                        // 商品支付类型
	CommodityPrice       int32       `json:"Price" mapstructure:"Price"`         // 商品价格
	PoolKey              string      `json:"PoolKey"`                            // 商品归属的那个pool， CommonShopPool_1-Zone1-VersionID， BreakoutShopPool_1-Zone1-VersionID
	CommodityJsonContent string      `json:"CJContent" mapstructure:"CJContent"` // 商品包含的内容，对应配置文件中的InnerGoods
	CommodityChance      int         `json:"Chance" mapstructure:"Chance"`       // 显示几率
}

// 刷新商店刷新周期内商品购买的信息，已经购买的次数
type AlreadyPurchasedCommodityInfoS struct {
	CommodityID    int   `json:"ID" mapstructure:"ID"`       // 商品id
	PurchasedCount int32 `json:"Count" mapstructure:"Count"` // 商品购买次数
}

// 刷新商品配置
type RefreshShopConfigS struct {
	ShopAutoRefreshIntervalHours int32 `json:"ShopAutoRefreshIntervalHours"` // 自动刷新商店的时间间隔，单位小时

	CommonManualRefreshCosts          []int32          `json:"CommonManualRefreshCosts"`          // 普通刷新商店每次刷新的价格，每次价格不同
	CommonManualRefreshPayType        GamePayType      `json:"CommonManualRefreshPayType"`        // 普通刷新商店刷新的支付类型
	CommonShopCommodityPools          []CommodityPoolS `json:"CommonShopCommodityPools"`          // 商品池和slot的对应关系
	DailyCommonManualRefreshCount     int32            `json:"DailyCommonManualRefreshCount"`     // 每天可刷新次数
	CommonShopPresentCommodityCount   int32            `json:"CommonShopPresentCommodityCount"`   // 普通商店展示商品的个数
	CommonShopWeeklyCardDiscountRate  int32            `json:"CommonShopWeeklyCardDiscountRate"`  // 普通月卡用户折扣率，购买商品
	CommonShopMonthlyCardDiscountRate int32            `json:"CommonShopMonthlyCardDiscountRate"` // 月卡用户折扣率，购买商品
	CommonShopCommodityDailyBuyCount  int32            `json:"CommonShopCommodityDailyBuyCount"`  // 商品每天每个人购买的次数，现在为一次

	BreakoutManualRefreshCosts        []int32          `json:"BreakoutManualRefreshCosts"`        // breakout刷新商店每次刷新的价格，每次价格不同
	BreakoutManualRefreshPayType      GamePayType      `json:"BreakoutManualRefreshPayType"`      // breakout刷新商店刷新的支付类型
	BreakoutShopCommodityPools        []CommodityPoolS `json:"BreakoutShopCommodityPools"`        // 商品池和slot的对应关系，这只是一级索引，在更新pool是，会建立二级索引（有版本号）
	DailyBreakoutManualRefreshCount   int32            `json:"DailyBreakoutManualRefreshCount"`   // 每天可刷新次数
	BreakoutShopPresentCommodityCount int32            `json:"BreakoutShopPresentCommodityCount"` // breakout商店展示商品的个数
	BreakoutWeeklyCardDiscountRate    int32            `json:"BreakoutWeeklyCardDiscountRate"`    // 普通月卡用户折扣率，购买商品
	BreakoutMonthlyCardDiscountRate   int32            `json:"BreakoutMonthlyCardDiscountRate"`   // 月卡用户折扣率，购买商品
	BreakoutCommodityDailyBuyCount    int32            `json:"BreakoutCommodityDailyBuyCount"`    // 商品每天每个人购买的次数，现在为一次
}

type RefreshShopCommodityPoolConfigS struct {
	PoolName        string                  `json:"PoolName"`        // 商品池的名字，CommonShopPool_1
	VersionID       string                  `json:"VersionID"`       // 整个池子的版本，只要对应池子里的商品有变更，该版本号必须递增
	PoolCommodities []RefreshShopCommodityS `json:"PoolCommodities"` // 池子中的商品
}

// 用户商店行为
type PlayerRefreshShopInfoS struct { // 商店类型
	ManualRefreshCount          int32                            `json:"ManualRefreshCount"`          // 刷新的次数
	AlreadyPurchasedCommodities []AlreadyPurchasedCommodityInfoS `json:"AlreadyPurchasedCommodities"` // 已经购买的商品id
	CurrentDisplayCommodities   []RefreshShopCommodityS          `json:"CurrentDisplayCommodities"`   // 当前刷新展示的商品信息
}

// 判断是否可以手工刷新
type ProtoCheckManualRefreshReq struct {
	Uin             uint64          `json:"Uin"`
	ZoneID          int32           `json:"ZoneID"`
	RefreshShopType RefreshShopType `json:"RSType" mapstructure:"RSType"` // 商店类型
}

type ProtoCheckManualRefreshRes struct {
	Uin                         uint64      `json:"Uin"`
	ZoneID                      int32       `json:"ZoneID"`
	ManualRefreshRemainderCount int32       `json:"ManualRefreshRemainderCount"` // 刷新的剩余次数
	ManualRefreshPayType        GamePayType `json:"ManualRefreshPayType"`        // 刷新的支付类型
	ManualRefreshCost           int32       `json:"ManualRefreshCost"`           // 手动刷新的花费
}

// 玩家获取刷新商店商品列表
type ProtoGetRefreshShopCommoditiesReq struct {
	Uin             uint64          `json:"Uin"`
	ZoneID          int32           `json:"ZoneID"`
	ClientIP        string          `json:"ClientIP"`
	RefreshShopType RefreshShopType `json:"RSType" mapstructure:"RSType"` // 商店类型
	IsManualRefresh int32           `json:"IsManualRefresh"`              // 1：手动刷新商品列表
}

type ProtoGetRefreshShopCommoditiesRes struct {
	Uin                           uint64                           `json:"Uin"`
	ZoneID                        int32                            `json:"ZoneID"`
	RefreshShopType               RefreshShopType                  `json:"RSType" mapstructure:"RSType"`                     // 商店类型
	AlreadyPurchasedCommodities   []AlreadyPurchasedCommodityInfoS `json:"APCommodities" mapstructure:"APCommodities"`       // 已经购买的商品id
	RefreshShopDisplayCommodities []RefreshShopCommodityS          `json:"RSDCommodities" mapstructure:"RSDCommodities"`     // 刷新商店展示的商品信息
	AutoRefreshRemainderSeconds   int64                            `json:"RemainderSeconds" mapstructure:"RemainderSeconds"` // 距离自动刷新剩余的秒数
	IsManualRefresh               int32                            `json:"IsManualRefresh"`                                  // 1：手动刷新商品列表
	ManualRefreshRemainderCount   int32                            `json:"MRRCount" mapstructure:"MRRCount"`                 // 手动刷新的剩余次数
	ManualRefreshCount            int32                            `json:"MRCount" mapstructure:"MRCount"`                   // 刷新间隔期内的总的刷新次数
	ManualRefreshPrice            int32                            `json:"MRPrice" mapstructure:"MRPrice"`                   // 手动刷新价格，客户端展示
	ManualRefreshDiscountPrice    int32                            `json:"MRDPrice" mapstructure:"MRDPrice"`                 // 手动刷新折扣价格，客户端展示
	ManualRefreshPayType          GamePayType                      `json:"MRPType" mapstructure:"MRPType"`                   // 刷新的支付类型
	ManualRefreshCost             int32                            `json:"MRCost" mapstructure:"MRCost"`                     // 手动刷新的花费，上次
	CommodityBuyCountInPeriod     int32                            `json:"BuyCountInPeriod" mapstructure:"BuyCountInPeriod"` // 在刷新期间内可购买的次数，配置数据
	WeeklyCardDiscountRate        int32                            `json:"WCDR" mapstructure:"WCDR"`
	MonthlyCardDiscountRate       int32                            `json:"MCDR" mapstructure:"MCDR"`
}

// GM工具更新刷新商店的配置
type ProtoUpdateRefreshShopConfigReq struct {
	Uin               uint64              `json:"Uin"`
	ZoneID            int32               `json:"ZoneID"`
	RefreshShopConfig *RefreshShopConfigS `json:"RefreshShopConfig"` // 具体的配置信息
}

// GM工具刷新商品池
type ProtoUpdateRefreshShopCommodityPoolReq struct {
	Uin                            uint64                           `json:"Uin"`
	ZoneID                         int32                            `json:"ZoneID"`
	RefreshShopCommodityPoolConfig *RefreshShopCommodityPoolConfigS `json:"RefreshShopCommodityPoolConfig"`
}

// 得到刷新商店商品的花费
type ProtoGetRefreshShopCommodityCostReq = ProtoBuyRefreshShopCommodityReq

type ProtoGetRefreshShopCommodityCostRes struct {
	Uin            uint64      `json:"Uin"`
	ZoneID         int32       `json:"ZoneID"`
	CommodityID    int         `json:"ID" mapstructure:"ID"`           // 商品id
	GamePayType    GamePayType `json:"PayType" mapstructure:"PayType"` // 商品支付类型
	CommodityPrice int32       `json:"Price" mapstructure:"Price"`     // 商品价格，lua逻辑要减去用户的花费
}

// 购买刷新商店中的商品
type ProtoBuyRefreshShopCommodityReq struct {
	Uin             uint64          `json:"Uin"`
	ZoneID          int32           `json:"ZoneID"`
	CommodityID     int             `json:"ID" mapstructure:"ID"`         // 商品id
	RefreshShopType RefreshShopType `json:"RSType" mapstructure:"RSType"` // 商店类型
	PoolKey         string          `json:"PoolKey"`                      // 商品归属的那个pool， CommonShopPool_1-Zone1-VersionID， BreakoutShopPool_1-Zone1-VersionID
}

type ProtoBuyRefreshShopCommodityRes struct {
	Uin                  uint64      `json:"Uin"`
	ZoneID               int32       `json:"ZoneID"`
	CommodityID          int         `json:"ID" mapstructure:"ID"`                   // 商品id
	GamePayType          GamePayType `json:"PayType" mapstructure:"PayType"`         // 商品支付类型
	CommodityPrice       int32       `json:"Price" mapstructure:"Price"`             // 商品价格，lua逻辑要减去用户的花费
	CommodityJsonContent string      `json:"JsonContent" mapstructure:"JsonContent"` // 商品包含的内容，对应配置文件中的InnerGoods
}

//-----------------------------------------------------------------------------------------------------------
const (
	C_SIGNINPRIZE_COUNT = 31
)

// 活动每日签到奖品
type SignInPrizeS struct {
	PrizeID int `json:"ID"` // 奖品id
	//PrizeName        string `json:"Name"`                 // 名字
	PrizeJsonContent string `json:"JC" mapstructure:"JC"` // 奖品具体内容，json内容
}

// 月签到配置数据，存放redis
type MonthlySignInConfigS struct {
	PrizeLst                 []SignInPrizeS `json:"PrizeLst"`                 // 自然月31天奖品列表
	VipMultiplePrizeDays     []int32        `json:"ViMPDays"`                 // vip用户连续多少天是double奖励。[1,3,5......]
	VipMultipleNum           int32          `json:"Num"`                      // vip用户奖励具体倍数，默认是2倍
	RessiueActivityThreshold int32          `json:"RessiueActivityThreshold"` // 阈值：补签活跃值
}

// GM配置每月登录领奖的配置信息
type ProtoGMConfigMonthlySignInReq struct {
	Uin                 uint64               `json:"Uin"`
	ZoneID              int32                `json:"ZoneID"`
	MonthlySignInConfig MonthlySignInConfigS `json:"MonthlySignInConfig"`
}

// 用户获取签到信息
type ProtoGetMonthlySigninInfoReq struct {
	Uin      uint64 `json:"Uin"`
	ZoneID   int32  `json:"ZoneID"`
	ClientIP string `json:"ClientIP"`
}

type ProtoGetMonthlySigninInfoRes struct {
	Uin                  uint64         `json:"Uin"`
	ZoneID               int32          `json:"ZoneID"`
	MonthName            int            `json:"MonthName"` // month标识: 201809
	CurrDate             int            `json:"CurrDate"`  // 签到当天日期
	MonthlySignInCount   int32          `json:"Count"`     // 本月累计签到次数
	WeeklyPrizeLst       []SignInPrizeS `json:"Prizes"`    // 本月奖品列表
	VipMultiplePrizeDays []int32        `json:"ViMPDays"`  // vip用户连续多少天是double奖励。[1,3,5......]
	VipMultipleNum       int32          `json:"VMM"`       // vip用户奖励具体倍数，默认是2倍
	ToDaySignIn          int32          `json:"TdSi"`      // 今天是否签到，=0没有签到，非0签到
	ToDayReSignIn        int32          `json:"TdRSi"`     // 今天是否补签到，=0没有签到，非0签到
	ActivityThreshold    int32          `json:"AT"`        // 补签活跃值
}

// 用户签到
type ProtoPlayerSignInReq struct {
	Uin          uint64 `json:"Uin"`
	ZoneID       int32  `json:"ZoneID"`
	SignInDayNum int32  `json:"SignInDayNum"` // 签到第几天
	Activity     int32  `json:"Activity"`     // 用户的活跃值，在补签的时候需要
}

type ProtoPlayerSignInRes struct {
	Uin              uint64 `json:"Uin"`
	ZoneID           int32  `json:"ZoneID"`
	SignInDayNum     int32  `json:"SignInDayNum"` // 签到第几天
	PrizeID          int    `json:"ID"`           // 奖品id
	PrizeJsonContent string `json:"JC"`           // 奖品具体内容
	VipMultipleNum   int32  `json:"VMM"`          // vip的倍数，lua层需要对奖品内容进行增倍处理
}

//-------------------------------------------------------------------------------------------
type ProtoGetFinanceBusinessRedLightsReq struct {
	Uin      uint64 `json:"Uin"`
	ZoneID   int32  `json:"ZoneID"`
	Activity int32  `json:"Activity"` // 用户的活跃值，在补签的时候需要
}

type ProtoGetFinanceBusinessRedLightsRes struct {
	Uin                          uint64                        `json:"Uin"`
	ZoneID                       int32                         `json:"ZoneID"`
	FinanceBusinessRedLightInfos []FinanceBusinessRedLightInfo `json:"RedLightInfos"`
}

//-----------------------------------------------------------------------------------------
// VIP
type ProtoGetPlayerVIPInfoReq struct {
	Uin       uint64 `json:"Uin"`
	ZoneID    int32  `json:"ZoneID"`
	ClientIP  string `json:"ClientIP"`
	ChannelID string `json:"ChannelID"` // CN US
}

type ProtoGetPlayerVIPInfoRes struct {
	Uin     uint64      `json:"Uin"`
	ZoneID  int32       `json:"ZoneID"`
	VipType UserVIPType `json:"VIPType"`

	NormalMonthVIPCollectPrizeCount      int32 `json:"WVCPCount"`  // 普通会员已经领取的次数
	NormalMonthVIPRemainderSeconds       int64 `json:"WVRSeconds"` // 剩余多少秒
	NormalMonthVIPCollectPrizeExpireDate int32 `json:"WVCPEDate"`  // 普通会员领取奖励过期日期
	NormalMonthVIPDayCollected           int32 `json:"WVDC"`       // 普通月卡会员当天是否已经领取 1: 领取过，0：没有领取

	LuxuryMonthVIPCollectPrizeCount      int32 `json:"MVCPCount"`  // 高级会员已经领取的次数
	LuxuryMonthVIPRemainderSeconds       int64 `json:"MVRSeconds"` // 剩余多少秒
	LuxuryMonthVIPCollectPrizeExpireDate int32 `json:"MVCPEDate"`  // 高级会员领取奖励过期日期
	LuxuryMonthVIPDayCollected           int32 `json:"MVDC"`       // 月卡会员当天是否已经领取 1: 领取过，0：没有领取

	PrivilegeInfo []ProtoVIPPrivilegeInfoS `json:"VIPINFO"` // 基本配置信息
}

type ProtoVIPPlayerCollectPrizeReq struct {
	Uin       uint64      `json:"Uin"`
	ZoneID    int32       `json:"ZoneID"`
	ClientIP  string      `json:"ClientIP"`
	VipType   UserVIPType `json:"VIPType"`   // vip类型
	ChannelID string      `json:"ChannelID"` // CN US
	Id        int         `json:"Id"`        //
}

type ProtoVIPPlayerCollectPrizeRes struct {
	Uin               uint64      `json:"Uin"`
	ZoneID            int32       `json:"ZoneID"`
	VipType           UserVIPType `json:"VIPType"`         // vip类型
	CollectPrizeCount int32       `json:"CPCount"`         // 领取次数
	CollectDiamonds   int32       `json:"CollectDiamonds"` // 获得钻石数
}

// GM工具配置vip特权
type ProtoVIPPrivilegeInfoS struct {
	Id        int     `json:"ID" mapstructure:"ID"`
	DailyGem  int32   `json:"DG" mapstructure:"DG"` // 每天领取的钻石数量
	Duration  int32   `json:"Duration"`             // 连续领取多少天
	GemCount  int32   `json:"GC" mapstructure:"GC"` // 激活后获得钻石数量
	Price     float32 `json:"Price"`
	PriceDesc string  `json:"PD" mapstructure:"PD"`
	ProductId string  `json:"PID" mapstructure:"PID"`
	ChannelID string  `json:"ChannelID"`
	Type      string  `json:"Type"`
	NameKey   string  `json:"NK" mapstructure:"NK"`
}

type ProtoVIPPrivilegeConfigS struct {
	VIPPrivilegeInfos []ProtoVIPPrivilegeInfoS `json:"VIPPrivilegeInfos"`
}

func (src ProtoVIPPrivilegeConfigS) FindVIPPrivilege(channelID string, rechargeCommodityID int) *ProtoVIPPrivilegeInfoS {
	for j := range src.VIPPrivilegeInfos {
		if channelID == src.VIPPrivilegeInfos[j].ChannelID &&
			rechargeCommodityID == src.VIPPrivilegeInfos[j].Id {
			return &src.VIPPrivilegeInfos[j]
		}
	}
	base.GLog.Error("channelID[%s] rechargeCommodityID[%d] is invalid!", channelID, rechargeCommodityID)
	return nil
}

type ProtoGMConfigVIPPrivilegeReq struct {
	Uin                uint64                   `json:"Uin"`
	ZoneID             int32                    `json:"ZoneID"`
	VIPPrivilegeConfig ProtoVIPPrivilegeConfigS `json:"VIPPrivilegeConfig"`
}

//-----------------------------------------------------------------------------------------
const (
	C_NEWPLAYER_BENEFIT_DAYS = 7
)

// 新用户七天登录
type ProtoNewPlayerBenefitInfo struct {
	Id                    int    `json:"ID"`
	PosterAssetBundleName string `json:"PAN" mapstructure:"PAN"`
	PosterTextureName     string `json:"PIN" mapstructure:"PIN"`
	JsonContent           string `json:"JC" mapstructure:"JC"` // 奖品具体内容，json内容
}

type ProtoNewPlayerBenefitConfigS struct {
	Benefits []ProtoNewPlayerBenefitInfo `json:"Benefits"`
}

func (src ProtoNewPlayerBenefitConfigS) FindBenefit(id int) *ProtoNewPlayerBenefitInfo {
	for j := range src.Benefits {
		if id == src.Benefits[j].Id {
			return &src.Benefits[j]
		}
	}
	base.GLog.Error("BenefitID[%d] is invalid!", id)
	return nil
}

type ProtoGMConfigNewPlayerLoginBenefitsReq struct {
	Uin    uint64                       `json:"Uin"`
	ZoneID int32                        `json:"ZoneID"`
	Config ProtoNewPlayerBenefitConfigS `json:"Config"`
}

type ProtoGetNewPlayerLoginBenefitReq struct {
	Uin      uint64 `json:"Uin"`
	ZoneID   int32  `json:"ZoneID"`
	ClientIP string `json:"ClientIP"`
}

type ProtoGetNewPlayerLoginBenefitRes struct {
	Uin              uint64                      `json:"Uin"`
	ZoneID           int32                       `json:"ZoneID"`
	IsCompleted      int32                       `json:"IsCompleted"`      // 领取是否结束，1：结束, 0: 未结束
	LoginDays        int32                       `json:"LoginDays"`        // 登录天数，一天一次
	ReceiveAwardTags []int32                     `json:"ReceiveAwardTags"` // 领取标识，1：领取过，0：没有领取
	Benefits         []ProtoNewPlayerBenefitInfo `json:"Benefits"`         // 配置信息
}

type ProtoReceiveLoginBenefitReq struct {
	Uin           uint64 `json:"Uin"`
	ZoneID        int32  `json:"ZoneID"`
	Id            int    `json:"ID"`            // 奖品id
	ReceiveDayNum int32  `json:"ReceiveDayNum"` // 领取第几天
}

type ProtoReceiveLoginBenefitRes struct {
	Uin              uint64                     `json:"Uin"`
	ZoneID           int32                      `json:"ZoneID"`
	BenefitInfo      *ProtoNewPlayerBenefitInfo `json:"Benefit"`          // 奖品信息
	ReceiveAwardTags []int32                    `json:"ReceiveAwardTags"` // 领取标识，1：领取过，0：没有领取
	LoginDays        int32                      `json:"LoginDays"`        // 登录天数，一天一次
	IsCompleted      int32                      `json:"IsCompleted"`      // 领取是否结束，1：结束, 0: 未结束
}

//-----------------------------------------------------------------------------------------

type ProtoFirstRechargeLevelConfS struct {
	Id       int    `json:"ID"`                   // 任务id
	Reward   string `json:"Reward"`               // 领取的奖励
	Target   int32  `json:"Target"`               // 达成目标
	TitleKey string `json:"TK" mapstructure:"TK"` //
	Value    int32  `json:"V" mapstructure:"V"`   //
}

type ProtoFirstRechargeConfigS struct {
	FRLevelConfLst []ProtoFirstRechargeLevelConfS `json:"FRLevelConfLst"`
}

func (src ProtoFirstRechargeConfigS) Find(id int) *ProtoFirstRechargeLevelConfS {
	for j := range src.FRLevelConfLst {
		if id == src.FRLevelConfLst[j].Id {
			return &src.FRLevelConfLst[j]
		}
	}
	base.GLog.Error("FirstRechargeActive[%d] is invalid!", id)
	return nil
}

type ProtoGMConfigFirstRechargeReq struct {
	Uin    uint64                    `json:"Uin"`
	ZoneID int32                     `json:"ZoneID"`
	Config ProtoFirstRechargeConfigS `json:"Config"`
}

// 拉取首冲任务信息
type ProtoGetFirstRechargeActiveReq struct {
	Uin      uint64 `json:"Uin"`
	ZoneID   int32  `json:"ZoneID"`
	ClientIP string `json:"ClientIP"`
}

type ProtoGetFirstRechargeActiveRes struct {
	Uin             uint64                         `json:"Uin"`
	ZoneID          int32                          `json:"ZoneID"`
	CurrBuyDiamonds int32                          `json:"CBD"`         // 当前购买的钻石数量
	IsCompleted     int32                          `json:"IsCompleted"` // 首冲是否全部完成 1：完成 0：未完成
	LevelInfos      []FirstRechargeLevelInfoS      `json:"LIS"`         // 等级完成情况
	LevelConfigs    []ProtoFirstRechargeLevelConfS `json:"LCS"`         // 每一级的静态配置，客户端展示所用
}

// 玩家领取首冲奖励
type ProtoReceiveFirstRechargeRewardReq struct {
	Uin    uint64 `json:"Uin"`
	ZoneID int32  `json:"ZoneID"`
	Id     int    `json:"ID"` // 奖品id
}

type ProtoReceiveFirstRechargeRewardRes struct {
	Uin         uint64 `json:"Uin"`
	ZoneID      int32  `json:"ZoneID"`
	Id          int    `json:"ID"`                   // 奖品id
	JsonContent string `json:"JC" mapstructure:"JC"` // 奖品具体内容，json内容
	IsCompleted int32  `json:"IsCompleted"`          // 首冲是否全部完成 1：完成 0：未完成
}
