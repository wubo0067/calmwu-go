## 说明
***
联调环境：http://123.59.40.19:400

### 基本数据结构
***
1. 充值商品
```go
type RechargeCommodityInfoS struct {
	RechargeCommodityID               int     `json:"ID"`
	BuyDiamonds                       int32   `json:"BDiamonds"` // 默认购买获得的钻石数量
	FirstRechargePresentDiamonds      int32   `json:"FDiamonds"` // 首冲获得的钻石数量
	PresentDiamonds                   int32   `json:"PDiamonds"` // 额外赠送的钻石数量
	FirstTimePurchaseRewardCaptionKey string  `json:"FTPRCKey"`  //FirstTimePurchaseRewardCaptionKey
	Price                             float32 `json:"Price"`
	PriceDesc                         string  `json:"PriceDesc"`
	ProductId                         string  `json:"ProductId"`
}
```

3. 资源商品类型
```go
const (
	E_RESOURCECOMMODITY_GOLD ResourceCommodityType = iota
	E_RESOURCECOMMODITY_WOOD
	E_RESOURCECOMMODITY_STONE
	E_RESOURCECOMMODITY_IRON
)
```

4. 资源商品
```go
type ResourceCommodityInfoS struct {
	ResourceCommodityID          int                   `json:"ID"`
	ResourceCommodityType        ResourceCommodityType `json:"RCType"`        // 资源商品类型
	ResourceCommodityStackCount  int32                 `json:"RCStackCount"`  // 资源堆叠数量
	ResourceCommodityDiamondCost int32                 `json:"RCDiamondCost"` // 购买所需钻石
}
```

5. 资源商品列表
```json
  "ResourceShopConfigS" : {
    "VersionID" : "1.0.0",
    "WCDR" : 90,
    "MCDR" : 80,
    "Count" : 6
    "ResourceCommodities" : [ResourceCommodityInfoS, ResourceCommodityInfoS, ResourceCommodityInfoS, ...]
  }
```

6. 刷新商店商品
```go
// 刷新商店的商品配置
type RefreshShopCommodityS struct {
	CommodityID          int         `json:"ID"`          // 商品id
	GamePayType          GamePayType `json:"GamePayType"` // 商品支付类型
	CommodityPrice       int32       `json:"Price"`       // 商品价格
	PoolKey              string      `json:"PoolKey"`     // 商品归属的那个pool， CommonShopPool_1-Zone1-VersionID， BreakoutShopPool_1-Zone1-VersionID
	CommodityJsonContent string      `json:"CJContent"`   // 商品包含的内容，对应配置文件中的InnerGoods
	CommodityChance      int         `json:"Chance"`      // 显示几率
}
```

7. 卡包商品
```go
type CardPackCommdityInfoS struct {
	CardPackCommodityID int    `json:"ID"` // 商品id
	CardPackDiamondCost int32  `json:"CPDCost"`
	CardPackJsonContent string `json:"CPJsonContent"` // 卡包礼包的具体内容，json描述
	CardPackRecommend   int32  `json:"CPRecommend"`   // 卡包是否推荐
	CardPackGiftDescKey string `json:"CPGiftDescKey"`
	CardPackCountKey    string `json:"CPCountKey"`
}
type ShopCardPackCommoditiesInfoS struct {
	VersionID               string                  `json:"VersionID"`
	WeeklyCardDiscountRate  int32                   `json:"WCDR"`
	MonthlyCardDiscountRate int32                   `json:"MCDR"`
	Count                   int                     `json:"Count"`
	CardPackCommodities     []CardPackCommdityInfoS `json:"CPCommodities"`
}
```

8. 游戏币类型
```go
const (
	E_GAMEPAY_GOLD    GamePayType = iota // 金币
	E_GAMEPAY_DIAMOND                         // 钻石
	E_GAMEPAY_HONOR                      // 勋章
	E_GAMEPAY_SHIPSOUL
)
```

9. 刷新商店名称
```go
const (
	C_REFRESHSHOPTYPE_NORMAL   RefreshShopType = "COMMONSHOP"
	C_REFRESHSHOPTYPE_BREAKOUT RefreshShopType = "BREAKOUTSHOP"
)
```

10. 已经购买过的商品
```go
type AlreadyPurchasedCommodityInfoS struct {
	CommodityID    int   `json:"ID"`    // 商品id
	PurchasedCount int32 `json:"Count"` // 商品购买次数
}
```

11. 小红点相关类型
```go
const (
	E_UPDATEBUSINESS_COMMONSHOP FinanceUpdateBusinessType = iota
	E_UPDATEBUSINESS_BREAKOUTSHOP
	E_UPDATEBUSINESS_SIGIN
)
type ProtoGetFinanceBusinessRedLightsRes struct {
	Uin                          uint64                        `json:"Uin"`
	ZoneID                       int32                         `json:"ZoneID"`
	FinanceBusinessRedLightInfos []FinanceBusinessRedLightInfo `json:"RedLightInfos"`
}
```

12. 资源商品类型
```go
const (
	E_RESOURCECOMMODITY_GOLD ResourceCommodityType = iota
	E_RESOURCECOMMODITY_WOOD
	E_RESOURCECOMMODITY_STONE
	E_RESOURCECOMMODITY_IRON
)
```

13. 每日签到奖品
```go
type SignInPrizeS struct {
	PrizeID int `json:"ID"` // 奖品id
	PrizeJsonContent string `json:"jc" mapstructure:"jc"` // 奖品具体内容，json内容
}
```

14. VIP类型
```go
const (
	E_USER_VIP_NO      UserVIPType = 0x00
	E_USER_VIP_NORMALMONTHLY  UserVIPType = 0x01
	E_USER_VIP_LUXURYMONTHLY UserVIPType = 0x02
	E_USER_VIP_ALL     UserVIPType = 0x03
)
```

15. vip特权信息
```go
type ProtoVIPPrivilegeInfoS struct {
	Id        int     `json:"ID"`
	DailyGem  int32   `json:"DG"`       // 每天领取的钻石数量
	Duration  int32   `json:"Duration"` // 连续领取多少天
	GemCount  int32   `json:"GC"`       // 激活后获得钻石数量
	Price     float32 `json:"Price"`
	PriceDesc string  `json:"PD"`
	ProductId string  `json:"PID"`
  ChannelID string  `json:"ChannelID"`
	Type      string  `json:"Type"`  
}
```

16. 新玩家福利信息
```go
type ProtoNewPlayerBenefitInfo struct {
	Id              int    `json:"ID"`
	PosterAssetBundleName string `json:"PAN" mapstructure:"PAN"`
	PosterTextureName  string `json:"PIN" mapstructure:"PIN"`
	JsonContent     string `json:"JC" mapstructure:"JC"` // 奖品具体内容，json内容
}
```

17. 活动类型
```go
const (
	E_ACTIVETYPE_SUPERGIFTPACKAGE ActiveType = iota // 超级礼包
	E_ACTIVETYPE_MISSION                            // 活跃任务
	E_ACTIVETYPE_EXCHANGE                           // 兑换
	E_ACTIVETYPE_CDKEYEXCHANGE                      // 密令兑换
	E_ACTIVETYPE_NEWPLAYERBENEFIT                   // 新用户七日
	E_ACTIVETYPE_FIRSTRECHARGE                      // 首冲礼包
)
```

18. 活动控制
```go
type ProtoActiveControlInfoS struct {
	ActiveType   ActiveType `json:"ActiveType"`         // 活动类型id，超值礼包、兑换
	ActiveID     int        `json:"ActiveID"`           // 同类型下的活动id，例如不同的超值礼包
	ChannelID    string     `json:"C" mapstructure:"C"` // CN US AreaCode
	StartTime    int64      `json:"StartTime"`          // 开启时间，单位秒
	DurationSecs int64      `json:"DurationSecs"`       // 持续时间，单位秒
}
```

19. 活动实例，**运行时数据**+静态配置
```go
type ProtoActiveInstanceInfo struct {
	ActiveID             int         `json:"AID"` // 活动实例id
	RemainderSeconds     int64       `json:"RS"`  // 活动剩余秒数
	AccumulateCount      int32       `json:"AC"`  // 运行时数据。玩家在该活动的累积数量，超级礼包该字段无效，兑换活动是作为领取的判断条件
	ReceiveCount         int32       `json:"RC"`  // 运行时数据。领取的次数，有的可以领取多次，有的一天只能领取一次
	ActiveinstanceConfig interface{} `json:"AIC"` // 各种活动的配置，每种配置有对应的结构体，超级礼包ActiveSuperGiftInfoS、活跃任务ActiveMissionInfoS、兑换任务ActiveExchangeInfoS
}
```

20. 超级礼包活动配置
```go
type ActiveSuperGiftInfoS struct {
	Base            ActiveBaseConfigInfoS `json:"base"` // 活动的基本配置
	FakePrice       float32               `json:"FP" mapstructure:"FP"`
	FakePriceDesc   string                `json:"FPD" mapstructure:"FPD"`
	DescKey         string                `json:"DK" mapstructure:"DK"`
	InnerGoods      string                `json:"IG" mapstructure:"IG"`
	NameKey         string                `json:"NK" mapstructure:"NK"`
	PosterAssetBundleName string                `json:"PAN" mapstructure:"PAN"`
	PosterTextureName  string                `json:"PIN" mapstructure:"PIN"`
	Price           float32               `json:"P" mapstructure:"P"`
	PriceDesc       string                `json:"PD" mapstructure:"PD"`
	ProductId       string                `json:"PI" mapstructure:"PI"`
  Discount       int                   `json:"DC" mapstructure:"DC"`
	Type           string                `json:"Type" mapstructure:"Type"`  
}
```

21. 活跃任务配置
```go
type ActiveMissionInfoS struct {
	Base       ActiveBaseConfigInfoS `json:"base"`                 // 活动的基本配置
	InnerGoods string                `json:"IG" mapstructure:"IG"` // 对应的奖品
	TaskType   string                `json:"TP" mapstructure:"TP"`
  TitleKey   string                `json:"TK" mapstructure:"TK"`
  Parameter  string                `json:"PA" mapstructure"PA"`
}
```

22. 兑换任务配置
```go
type ActiveExchangeInfoS struct {
	Base         ActiveBaseConfigInfoS `json:"base"`                 // 活动的基本配置
	InnerGoods   string                `json:"IG" mapstructure:"IG"` // 兑换目标
	TitleKey     string                `json:"TK" mapstructure:"TK"`
	ExchangeCost string                `json:"EC" mapstructure:"EC"` // 兑换资源
}
```

23. 活动基本配置 **配置文件数据**
```go
type ActiveBaseConfigInfoS struct {
	ActiveID      int    `json:"ID" mapstructure:"ID"`   // 同类型下的活动id，例如不同的超值礼包
	ReceiveLimit  int32  `json:"RL" mapstructure:"RL"`   // 静态配置数据。活动领取次数限制
	ReceiveCond   int32  `json:"RC" mapstructure:"RC"`   // 静态配置数据。活动领取条件
	ResetEveryDay int32  `json:"RED" mapstructure:"RED"` // 是否每天开启 1：隔天重置 0：整个活动有效期
	ChannelID     string `json:"C" mapstructure:"C"`     // CN US AreaCode
}
```

24. LUA活动数据上报
```go
type ProtoActiveDataS struct {
	ActiveTaskType     string `json:"ActiveTaskType"`
	TaskOpType         int32  `json:"TaskOpType"`           // =1: SET, =2: ADD
	AccumalateParamter int32  `json:"AP" mapstructure:"AP"` // 该活动继承的船数量，获得排位赛胜利，只要完整活动的行为都调用该接口
}
```

25. 首冲礼包等级领取情况
```go
type FirstRechargeLevelInfoS struct {
	ActiveID int   `json:"ActiveID"`
	Received int32 `json:"Received"` // 是否领取过，1：已经领取 0：还没领取
}
```

26. 首冲礼包各等级配置
```go
type ProtoFirstRechargeLevelConfS struct {
	Id       int    `json:"ID"`       // 任务id
	Reward   string `json:"Reward"`   // 领取的奖励
	Target   int32  `json:"Target"`   // 达成目标
	TitleKey string `json:"TK"` //
	Value    int32  `json:"V"`    //
}
```

27. 活动类型是否结束
```go
type ActiveIsCompleteS struct {
	ActiveType  ActiveType `json:"ActiveType"`
	IsCompleted int32      `json:"IsCompleted"` // 领取是否结束，1：结束, 0: 未结束
}
```

28. 小红点
```go
type FinanceBusinessRedLightInfo struct {
	BusinessType     FinanceUpdateBusinessType `json:"BusinessType" mapstructure:"BusinessType"`
	RedPointIsLight  int32                     `json:"RedPointIsLight" mapstructure:"RedPointIsLight"`   // =1：亮着
	RemainderSeconds int64                     `json:"RemainderSeconds" mapstructure:"RemainderSeconds"` // 剩余时间，倒计时秒
}
```

### 接口
***
1. 新建用户
```
  InterfaceName: NewFinanceUser
  Url: /sailcraft/api/v1/FinanceSvr/NewFinanceUser
  Request:
  "Params" : {
     "Uin" : 3232,
     "ZoneID" : 23232,
     "TimeZone" : "America/New_York",
  }
  Response:
  "Params" : {
     "Uin" : 3232,
     "ZoneID" : 23232,
     "Result" : 0/-1
  }  
```

2. 获取充值商品
```
  InterfaceName: QueryRechargeCommodities
  Url: /sailcraft/api/v1/FinanceSvr/QueryRechargeCommodities
  Request:
  "Params" : {
     "Uin" : 3232,
     "ZoneID" : 23232,
     "ClientIP" : "x.x.x.x",
     "ChannelID" : "CN",
  }
  Response:
  "Params" : {
     "Uin" : 3232,
     "ZoneID" : 23232,
     "VersionID" : "1",
     "ChannelID" : "CN",
     "Commodities" : [RechargeCommodityInfoS, RechargeCommodityInfoS, RechargeCommodityInfoS, ...]
  }  
```

3. 充值商品发货，tsapi-->lua--->调用该接口
```json
  InterfaceName: DeliveryRechargeCommodity
  Url: /sailcraft/api/v1/FinanceSvr/DeliveryRechargeCommodity
  Request:
  "Params" : {
     "Uin" : 3232,
     "ZoneID" : 23232,
     "ID" : 2323232, 商品ID
     "RCType" : 0（商店钻石）、1（普通月卡）、2（月卡）充值商品类型、3超级礼包
     "VersionID" : "xxxxx",
     "ChannelID" : "CN",
     "PlatForm" : "ANDROID" // IOS
  }  
  Response:
  "Params" : {
     "Uin" : 3232,
     "ZoneID" : 23232,
     "RCType" : 0（商店钻石）、1（普通月卡）、2（月卡）
     "BDiamonds" : 9999, 购买的钻石
     "PDiamonds" : 8888, 赠送的钻石
     "IsFristPurchase"  :  1,  首次购买标志 = 1
     "InnerGoods" : "json_content", // 购买超级礼包返回的奖品
     "NK" : "namekey",
  }   
```

4. GM工具调用该接口刷新充值商品
```json
  InterfaceName: RefreshRechargeCommodities
  Url: /sailcraft/api/v1/FinanceSvr/RefreshRechargeCommodities
  "Params" : {
     "Uin" : 3232,
     "ZoneID" : 1,
     "ShopRechargeCommodityInfo" : ShopRechargeCommoditiesInfoS
  }  
  Response:
  "Params" : {
     "Result" : 0/-1
  }   
```

5. 查询用户消费类型，普通用户、普通月卡用户、月卡用户
```json
  InterfaceName: QueryUserVIPType
  Url: /sailcraft/api/v1/FinanceSvr/QueryUserVIPType
  "Params" : {
     "Uin" : 232323,
     "ZoneID" : 1
  }  
  Response:
  "Params" : {
     "Uin" : 232323,
     "ZoneID" : 1,    
     "UCType" : 0（普通用户）、1（普通月卡用户）、2（月卡用户）, 玩家消费卡类型
     "Result" : 0/-1
  }   
```

6. 获取资源商品
```json
  InterfaceName: QueryResourceCommdities
  Url: /sailcraft/api/v1/FinanceSvr/QueryResourceCommdities
  Request:
  "Params" : {
     "Uin" : 3232,
     "ZoneID" : 23232,
     "ClientIP" : "x.x.x.x"
  }
  Response:
  "Params" : {
     "Uin" : 3232,
     "ZoneID" : 23232,
     "RSCInfo" : ResourceShopConfigS 资源商店配置信息，商品列表
  }  
```

7. 购买资源商品
```json
  InterfaceName: BuyResourceCommodity
  Url: /sailcraft/api/v1/FinanceSvr/BuyResourceCommodity
  Request:
  "Params" : {
     "Uin" : 3232,
     "ZoneID" : 23232,
     "VersionID" : "1.0.0", 
     "ID" : 1, 资源商品id
     "RCType" : ResourceCommodityType, 资源商品类型
  }
  Response:
  "Params" : {
     "Uin" : 3232,
     "ZoneID" : 23232,
     "RCType" : ResourceCommodityType, 资源商品类型
     "Cost" : 2323, 钻石花费
     "ResourceCount" : 23232
  }
```

8. GM工具调用该接口刷新资源商品
```json
  InterfaceName: RefreshResourceShopConfig
  Url: /sailcraft/api/v1/FinanceSvr/RefreshResourceShopConfig
  "Params" : {
     "Uin" : 3232,
     "ZoneID" : 1,
     "ResourceShopConfigInfo" : ResourceShopConfigS
  }  
  Response:
  "Params" : {
     "Result" : 0/-1
  }   
```

9. 请求刷新商店商品列表，客户端展示
```json
  InterfaceName: GetRefreshShopCommodities
  Url: /sailcraft/api/v1/FinanceSvr/GetRefreshShopCommodities
  "Params" : {
     "Uin" : 3232,
     "ZoneID" : 1,
     "ClientIP" "x.x.x.x",
     "RSType" : "commonshop"/"breakoutshop", 
     "IsManualRefresh" : 1/0,  1:表示手动刷新，0表示普通获取
  }  
  Response:
  "Params" : {
	  "Uin" : 3232,
	  "ZoneID" : 1,
	  "RSType" :  "commonshop"/"breakoutshop",                 // 商店类型
	  "APCommodities" : {AlreadyPurchasedCommodityInfo1,AlreadyPurchasedCommodityInfo2}                    // 已经购买的商品
	  "RSDCommodities" : {RefreshShopCommodityS, RefreshShopCommodityS},    // 刷新商店展示的商品信息, 商品类型见开头类型定义
	  "RemainderSeconds" : int64,                             // 距离自动刷新剩余的秒数
	  "IsManualRefresh" : 1/0,                                           // 1：手动刷新商品列表
	  "MRRCount" : int32,                             // 刷新的剩余次数
    "MRCount" : int32,                                      // 刷新间隔期内的总的刷新次数
    "MRPrice": int32,                                       // 手动刷新价格，客户端展示
    "MRDPrice" : int32,                              // 手动刷新折扣价格，客户端展示
	  "MRPType"  : 0/1/2,                                   // 0：金币 1：钻石 2：勋章 刷新的支付类型
	  "MRCost"  ：232,                                       // 手动刷新的花费，lua层判断IsManualRefresh==1时要扣除该数值
    "BuyCountInPeriod" : 2,    // 在刷新期间内可购买的次数，配置数据
    "WCDR" : int32,
    "MCDR" : int32,
  } 
```

10. GM工具更新刷新商店配置
```json
```

11. GM工具刷新商品池配置
```json
```

12. 购买刷新商店商品
```json
  InterfaceName: BuyRefreshShopCommodity
  Url: /sailcraft/api/v1/FinanceSvr/BuyRefreshShopCommodity
  "Params" : {
     "Uin" : 3232,
     "ZoneID" : 1,
     "ID" 12121,
     "RSType" : "commonshop"/"breakoutshop", 
     "PoolKey" : "商品列表中带有该字段",  
  }  
  Response:
  "Params" : {
    "Uin" : 332,
    "ZoneID" : 1
    "ID" : 12121                     // 商品id
    "PayType"  :  GamePayType   // 商品支付类型
    "Price"  : 21212                 // 商品价格，lua逻辑要减去用户的花费
    "JsonContent" : ""               // 商品包含的内容，对应配置文件中的InnerGoods
  } 
```

13. 请求卡包商品列表
```json
  InterfaceName: QueryCardPackCommdities
  Url: /sailcraft/api/v1/FinanceSvr/QueryCardPackCommdities
  Request:
  "Params" : {
     "Uin" : 3232,
     "ZoneID" : 23232,
     "ClientIP" : "x.x.x.x"
  }
  Response:
  "Params" : {
     "Uin" : 3232,
     "ZoneID" : 23232,
     "SCPCommoditiesInfo" : ShopCardPackCommoditiesInfoS（看数据描述）
  }  
```

14. GM工具刷新卡包商店配置
```
```

15. 购买卡包
```json
  InterfaceName: BuyCardPackCommodity
  Url: /sailcraft/api/v1/FinanceSvr/BuyCardPackCommodity
  Request:
  "Params" : {
     "Uin" : 3232,
     "ZoneID" : 23232,
     "VersionID" : "1.0.0", 
     "ID" : 1,
  }
  Response:
  "Params" : {
     "Uin" : 3232,
     "ZoneID" : 23232,
     "CostDiamonds" : 2323,
     "CPJsonContent" : "卡包礼包的具体内容，json描述, lua成unmarshal成对象即可"
  }
```

16. 获取手动刷新检查数据
```json
  InterfaceName: CheckRefreshShopManualRefresh
  Url: /sailcraft/api/v1/FinanceSvr/CheckRefreshShopManualRefresh
  Request:
  "Params" : {
     "Uin" : 3232,
     "ZoneID" : 1,
     "ClientIP" "x.x.x.x",
     "RSType" : "commonshop"/"breakoutshop", 
  }
  Response:
  "Params" : {
	  "Uin" : 3232
	  "ZoneID" : 1
    "ManualRefreshRemainderCount" : int32, // 刷新的剩余次数
	  "ManualRefreshPayType"  : 0/1/2                                   // 0：金币 1：钻石 2：勋章 刷新的支付类型
	  "ManualRefreshCost"  ：232                                       // 手动刷新的花费
  }
```

17. 获取可更新业务的小红点
```json
  InterfaceName: GetUserFinanceBusinessRedLights
  Url: /sailcraft/api/v1/FinanceSvr/GetUserFinanceBusinessRedLights
  Request:
  "Params" : {
     "Uin" : 3232,
     "ZoneID" : 1,
     "ClientIP" "x.x.x.x",
  }
  Response:
  "Params" : {
	  "Uin" : 3232,
	  "ZoneID" : 1,
	  "FinanceBusinessRedLightInfos"  : {FinanceBusinessRedLightInfo1, FinanceBusinessRedLightInfo2},
  }
```

18. GM工具更新每日签到配置
```json
  InterfaceName: GMUpdateMonthlySigninConfigInfo
  Url: /sailcraft/api/v1/FinanceSvr/GMUpdateMonthlySigninConfigInfo
```

19. 获得本月签到信息
```json
  InterfaceName: GetMonthlySigninInfo
  Url: /sailcraft/api/v1/FinanceSvr/GetMonthlySigninInfo
  Request:
  "Params" : {
     "Uin" : 3232,
     "ZoneID" : 1,
     "ClientIP" "x.x.x.x"
  }
  Response:
  "Params" : {
	  "Uin" : 3232,
	  "ZoneID" : 1,
    "MonthName" : 201801,  // month标识: 201809
    "CurrDate" : 20180327, // 当天日期
    "Count" : 2,          // 本月签到次数
    "Prizes" : [SignInPrizeS1, SignInPrizeS2], // 本普通月奖品列表
    "VIPMPDays" : [1,3, 5], // vip用户连续多少天是double奖励。[1,3,5......]
    "VMM" : 2, // vip用户奖励具体倍数，设定的是2倍
    "TdSi" : 0, // 今天是否签到，=0没有签到，非0签到
    "TdRSi" : 0, // 今天是否补签到，=0没有签到，非0签到
    "AT" : 80, // 补签活跃值
  }
```

20. 用户签到
```json
  InterfaceName: PlayerSignIn
  Url: /sailcraft/api/v1/FinanceSvr/PlayerSignIn
  Request:
  "Params" : {
     "Uin" : 3232,
     "ZoneID" : 1,
     "SignInDayNum" : int, //签到连续第几天
     "Activity" : int, // 用户的活跃值，在补签的时候需要
  }
  Response:
  "Params" : {
	  "Uin" : 3232,
	  "ZoneID" : 1,
    "SignInDayNum" : 3,  // // 签到第几天
    "ID" : 2323, // 奖品ID
    "JC" : "xxx", // 奖品内容json
    "VMM" : 2, // vip用户奖励具体倍数，设定的是2倍
  }
```

21. 玩家获取VIP信息
```json
  InterfaceName: GetPlayerVIPInfo
  Url: /sailcraft/api/v1/FinanceSvr/GetPlayerVIPInfo
  Request:
  "Params" : {
     "Uin" : 3232,
     "ZoneID" : 1,
     "ClientIP": "x.x.x.x",
     "ChannelID": "CN", // CN US
  }
  Response:
  "Params" : {
	  "Uin" : 3232,
	  "ZoneID" : 1,
    "VIPType" : UserVIPType,        // 会员类型
    "WVCPCount" : 3,                // 普通月会员领取的次数
    "WVRSeconds" : 23232,           // 普通月会员距离过期时间剩余秒数
    "WVCPEDate" : 20180316,         // 普通月会员领取奖励过期日期
    "MVCPCount" : 3,                // 月会员领取的次数
    "MVRSeconds" : 323232,          // 月会员距离过期时间剩余秒数
    "MVCPEDate" : 20180316,         // 月会员领取奖励过期日期
    "WVDC" : 1,                     // 普通月卡会员当天是否已经领取 1: 领取过，0：没有领取
    "MVDC" : 1,                     // 月卡会员当天是否已经领取 1: 领取过，0：没有领取
    "VIPINFO" : {ProtoVIPPrivilegeInfoS0, ProtoVIPPrivilegeInfoS1} //  基本配置信息
  }
```

22. VIP每日领取奖励
```json
  InterfaceName: VIPPlayerCollectPrize
  Url: /sailcraft/api/v1/FinanceSvr/VIPPlayerCollectPrize
  Request:
  "Params" : {
      "Uin" : 3232,
      "ZoneID" : 1,
      "ClientIP" "x.x.x.x",
      "VIPType" : UserVIPType,        // 会员类型     
      "ChannelID": "CN", // CN US      
      "Id" : int, //1, 2, 1001, 1002
  }
  Response:
  "Params" : {
	  "Uin" : 3232,
	  "ZoneID" : 1,
    "VIPType" : UserVIPType,        // 签到第几天
    "CPCount" : 3,                  // 会员领取的次数
    "CollectDiamonds" : 23232,      // 领取的钻石数量
  }
```

23. GM工具更新新用户七日登陆福利配置
```json
  InterfaceName: GMConfigNewPlayerLoginBenefit
  Url: /sailcraft/api/v1/FinanceSvr/GMConfigNewPlayerLoginBenefit
```

24. 获取新用户七日登陆福利信息，客户端每次登录后调用该接口
```json
  InterfaceName: GetNewPlayerLoginBenefitInfo
  Url: /sailcraft/api/v1/FinanceSvr/GetNewPlayerLoginBenefitInfo
  Request:
  "Params" : {
      "Uin" : 3232,
      "ZoneID" : 1,
      "ClientIP" : "x.x.x.x",
  }
  Response:
  "Params" : {
      "Uin" : 3232,
      "ZoneID" : 1,
      "ClientIP" "x.x.x.x",
      "IsCompleted" : 0,        // 领取是否结束，1：结束, 0: 未结束。结束后该页面无需显示    
      "LoginDays": "CN", // 新用户登录的天数      
      "ReceiveAwardTags" : [0,0,0,0,0,0,0], //领取标识，1：领取过，0：没有领取
      "Benefits" : [ProtoNewPlayerBenefitInfo1, ProtoNewPlayerBenefitInfo2,....], //配置信息
  }
```

25. 新用户领取7日福利
```json
  InterfaceName: ReceiveLoginBenefit
  Url: /sailcraft/api/v1/FinanceSvr/ReceiveLoginBenefit
  Request:
  "Params" : {
      "Uin" : 3232,
      "ZoneID" : 1,
      "Id" : int, // 奖品id
      "ReceiveDayNum" : int, // 领取第几天
  }
  Response:
  "Params" : {
      "Uin" : 3232,
      "ZoneID" : 1,
      "BenefitInfo" : ProtoNewPlayerBenefitInfo, // 福利信息，json
      "IsCompleted" : 0,        // 领取是否结束，1：结束, 0: 未结束。结束后该页面无需显示    
      "LoginDays": "CN", // 新用户登录的天数      
      "ReceiveAwardTags" : [1,0,0,0,0,0,0], //领取标识，1：领取过，0：没有领取
  }
```

25. 开启活动
```json
  InterfaceName: OpenActive
  Url: /sailcraft/api/v1/FinanceSvr/OpenActive
  Request:
  "Params" : {
      "Uin" : 3232,
      "ZoneID" : 1,
      "ActiveControlConfigs" : {ProtoActiveControlInfoS1, ProtoActiveControlInfoS2, ...} // 开启活动的控制参数列表
  }
  Response:
  "Params" : {
      "Uin" : 3232,
      "ZoneID" : 1,
      "ActiveIDs" : {int, int, ...}, // 活动类型实例id列表
      "ControlResults" : {int, int, ...}, // =1成功 =0失败
  }  
```

26. 关闭活动
```json
  InterfaceName: ProtoCloseActiveReq
  Url: /sailcraft/api/v1/FinanceSvr/ProtoCloseActiveReq
  Request:
  "Params" : {
      "Uin" : 3232,
      "ZoneID" : 1,
      "ActiveType" : ActiveType, // 活动类型
      "ActiveIDs" : {int, int, ...}, // 活动类型实例id列表
  }
  Response:
  "Params" : {
      "Uin" : 3232,
      "ZoneID" : 1,
      "ActiveIDs" : {int, int, ...}, // 活动类型实例id列表
      "ControlResults" : {int, int, ...}, // =1成功 =0失败
  }    
```

27. 获取活动展示信息
```json
  InterfaceName: GetPlayerActive
  Url: /sailcraft/api/v1/FinanceSvr/GetPlayerActive
  Request:
  "Params" : {
      "Uin" : 3232,
      "ZoneID" : 1,
      "ActiveType" : ActiveType, // 活动类型
      "ChannelID" : "CN", // CN US NOAREA
  }
  Response:
  "Params" : {
      "Uin" : 3232,
      "ZoneID" : 1,
      "ActiveType" : ActiveType, // 活动类型
      "AILST" : {ProtoActiveInstanceInfo1, ProtoActiveInstanceInfo2, ..},   // 活动实例信息    
  }
```

28. 活动累积数据通知（这个由lua业务层调用）
```json
  InterfaceName: ActiveAccumulateParameterNtf
  Url: /sailcraft/api/v1/FinanceSvr/ActiveAccumulateParameterNtf
  Request:
  "Params" : {
      "Uin" : 3232,
      "ZoneID" : 1,
      "ActiveType" : ActiveType, // 活动类型
      "ActiveDatas" : {ProtoActiveDataS1, ProtoActiveDataS2},
  }
  Response:
  "Params" : {
      "Uin" : 3232,
      "ZoneID" : 1,
      "ActiveType" : ActiveType, // 活动类型
      "IsDone" : int32,  // =1 任务可以领取 =0 任务不可领取
  }  
```


29. 活动领取
```json
  InterfaceName: PlayerActiveReceive
  Url: /sailcraft/api/v1/FinanceSvr/PlayerActiveReceive
  Request:
  "Params" : {
      "Uin" : 3232,
      "ZoneID" : 1,
      "ActiveType" : ActiveType, // 活动类型
      "ActiveID" : int32, //
      "ChannelID" : "CN", // CN US NOAREA
  }
  Response:
  "Params" : {
      "Uin" : 3232,
      "ZoneID" : 1,
      "Type" : ActiveType, // 活动类型
      "ID" : 1,  // 活动id
      "IG" : "json content", // 领取的奖励，兑换的物品 InnerGoods
      "AC" : int32, // 累积数量
      "RC" : int32, // 领取次数
  }
```

30. 获取兑换活动所需的资源
```json
  InterfaceName: GetActiveExchangeCost
  Url: /sailcraft/api/v1/FinanceSvr/GetActiveExchangeCost
  Request:
  "Params" : {
      "Uin" : 3232,
      "ZoneID" : 1,
      "ActiveType" : ActiveType, // 活动类型
      "ActiveID" : int32, //
  }
  Response:
  "Params" : {
      "Uin" : 3232,
      "ZoneID" : 1,
      "ActiveType" : ActiveType, // 活动类型
      "ActiveID" : 1,  
      "EC" : "json content", // 兑换所需的资源
  }
```

31. 检查活动配置是否存在
```json
  InterfaceName: CheckActiveConfig
  Url: /sailcraft/api/v1/FinanceSvr/CheckActiveConfig
  Request:
  "Params" : {
      "Uin" : 3232,
      "ZoneID" : 1,
      "ActiveType" : ActiveType, // 活动类型
      "ActiveIDs" : {int32, ...}, // 同类型下的活动id列表，例如不同的超值礼包 
  }
  Response:
  "Params" : {
      "Uin" : 3232,
      "ZoneID" : 1,
      "ActiveType" : ActiveType, // 活动类型
      "ActiveIDs" : {int32, ...}, // 同类型下的活动id列表，例如不同的超值礼包 
      "IsExists" : {int32, ...} // 存在列表 =1 配置存在，=0 配置不存在
  }
```

32. CDKey兑换活动
```json
  InterfaceName: PlayerExchangeCDKey
  Url: /sailcraft/api/v1/FinanceSvr/PlayerExchangeCDKey
  Request:
  "Params" : {
      "Uin" : 3232,
      "ZoneID" : 1,
      "CDKey" : string, // cdkey内容
  }
  Response:
  "Params" : {
      "Uin" : 3232,
      "ZoneID" : 1,
      "IG" : "json content", // 领取的奖励，兑换的物品 InnerGoods
  }
```

33. 拉取首冲礼包活动
```json
  InterfaceName: GetFirstRechargeActive
  Url: /sailcraft/api/v1/FinanceSvr/GetFirstRechargeActive
  Request:
  "Params" : {
      "Uin" : 3232,
      "ZoneID" : 1,
  }
  Response:
  "Params" : {
      "Uin" : 3232,
      "ZoneID" : 1,
      "CBD" : int, // CurrBuyDiamonds 当前购买的钻石数量
      "IsCompleted" : int, // 首冲礼包所有是否全部完成 1：完成 0：未完成
      "LIS" : {FirstRechargeLevelInfoS1, FirstRechargeLevelInfoS2}, // 每个等级的领取情况
      "LCS" : {ProtoFirstRechargeLevelConfS1, ProtoFirstRechargeLevelConfS2}, // 每个等级的配置
  }
```

33. 玩家领取首冲礼包
```json
  InterfaceName: ReceiveFirstRechargeReward
  Url: /sailcraft/api/v1/FinanceSvr/ReceiveFirstRechargeReward
  Request:
  "Params" : {
      "Uin" : 3232,
      "ZoneID" : 1,
      "Id" : int, // 礼包id
  }
  Response:
  "Params" : {
      "Uin" : 3232,
      "ZoneID" : 1,
      "Id" : int, // 礼包id
      "JC" : string, // 礼包对应的奖品，json
      "IsCompleted" : int, // 首冲礼包所有是否全部完成 1：完成 0：未完成      
  }
```

33. 判断玩家的活动类型是否完成
```json
  InterfaceName: CheckPlayerActiveIsCompleted
  Url: /sailcraft/api/v1/FinanceSvr/CheckPlayerActiveIsCompleted
  Request:
  "Params" : {
      "Uin" : 3232,
      "ZoneID" : 1,
  }
  Response:
  "Params" : {
      "Uin" : 3232,
      "ZoneID" : 1,
      "ACLst" : {ActiveIsCompleteS1, ActiveIsCompleteS2}, //
  }
```

34. 获得刷新商店商品的价格已经支付类型
```json
  InterfaceName: GetRefreshShopCommodityCost
  Url: /sailcraft/api/v1/FinanceSvr/GetRefreshShopCommodityCost
  "Params" : {
     "Uin" : 3232,
     "ZoneID" : 1,
     "ID" 12121,
     "RSType" : "commonshop"/"breakoutshop", 
     "PoolKey" : "商品列表中带有该字段",  
  }  
  Response:
  "Params" : {
    "Uin" : 332,
    "ZoneID" : 1
    "ID" : 12121                     // 商品id
    "PayType"  :  GamePayType   // 商品支付类型
    "Price"  : 21212                 // 商品价格，lua逻辑要减去用户的花费
  } 
```

35. 查询充值商品价格
```json
  InterfaceName: QueryRechargeCommodityPrices
  Url: /sailcraft/api/v1/FinanceSvr/QueryRechargeCommodityPrices
  "Params" : {
     "ZoneID" : 1,
     "ID" 12121,      // 充值商品id
     "RCType" : 0（商店钻石）、1（普通月卡）、2（月卡）充值商品类型、3超级礼包
     "ChannelID" : "CN", // US
  }  
  Response:
  "Params" : {
     "ZoneID" : 1,
     "ID" 12121,      // 充值商品id
     "RCType" : 0（商店钻石）、1（普通月卡）、2（月卡）充值商品类型、3超级礼包
     "ChannelID" : "CN", // US
     "Price"  : float32                 // 商品价格
  } 
```