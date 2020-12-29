/*
 * @Author: calmwu
 * @Date: 2018-03-29 11:01:37
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-05-19 11:35:14
 * @Comment:
 */
package proto

import (
	"encoding/json"
	"fmt"
	"sailcraft/base"
	"sailcraft/dataaccess/redistool"
	"sailcraft/financesvr_main/common"
	"time"
)

const (
	ActiveRunningKeyFmt    = "ActiveRunning-Zone%d"      // 活动开启活动的信息
	ActiveInstConfigKeyFmt = "ActiveConfig-Zone%d-%s-%d" // 活动的基础配置 zoneid-activetype-activeid
	ActiveKeyFmt           = "Active-Type%d"
	CDKeyCountKeyFmt       = "CDKeyCount-%s"
)

type ActiveType int

const (
	E_ACTIVETYPE_SUPERGIFTPACKAGE ActiveType = iota // 超级礼包
	E_ACTIVETYPE_MISSION                            // 活跃任务
	E_ACTIVETYPE_EXCHANGE                           // 兑换
	E_ACTIVETYPE_CDKEYEXCHANGE                      // 密令兑换
	E_ACTIVETYPE_NEWPLAYERBENEFIT                   // 新用户七日
	E_ACTIVETYPE_FIRSTRECHARGE                      // 首冲礼包 5
)

//----------------------------------------------------------------------------------
// 活动控制
type ProtoActiveControlInfoS struct {
	ActiveType   ActiveType `json:"ActiveType"`                         // 活动类型id，超值礼包、兑换
	ActiveID     int        `json:"ActiveID"`                           // 同类型下的活动id，例如不同的超值礼包
	ChannelID    string     `json:"ChannelID" mapstructure:"ChannelID"` // CN US AreaCode
	StartTime    int64      `json:"StartTime"`                          // 开启时间，单位秒
	DurationSecs int64      `json:"DurationSecs"`                       // 持续时间，单位秒
}

func (paci *ProtoActiveControlInfoS) BeginTime() time.Time {
	return time.Unix(paci.StartTime, 0)
}

func (paci *ProtoActiveControlInfoS) EndTime() time.Time {
	return time.Unix(paci.StartTime, 0).Add(time.Duration(paci.DurationSecs) * time.Second)
}

func (paci *ProtoActiveControlInfoS) IsExpired() bool {
	now := time.Now()
	endTime := paci.EndTime()
	if now.After(endTime) {
		base.GLog.Debug("Active Type[%s] Id[%d] EndTime[%s] expired, so reopen",
			paci.ActiveType.String(), paci.ActiveID, base.TimeName(endTime))
		return true
	}
	return false
}

type RunningActiveMgr struct {
	RunningActives []ProtoActiveControlInfoS `json:"runningActives"` // 运行中的活动
}

func CreateRunningActiveMgr(zoneID int32, redis *redistool.RedisNode) *RunningActiveMgr {
	runningActiveMgr := new(RunningActiveMgr)
	runningActiveMgr.RunningActives = make([]ProtoActiveControlInfoS, 0)
	runningActiveMgr.SyncRedis(zoneID, redis)
	return runningActiveMgr
}

// 控制信息写入redis
func (ram *RunningActiveMgr) SyncRedis(zoneID int32, redis *redistool.RedisNode) {
	redisData, err := json.Marshal(ram)
	if err != nil {
		base.GLog.Error("Marshal failed! reason[%s]", err.Error())
	}

	ActiveRunningKey := fmt.Sprintf(ActiveRunningKeyFmt, zoneID)
	err = common.GRedis.StringSet(ActiveRunningKey, redisData)
	if err != nil {
		base.GLog.Error("Set ActiveRunningKey[%s] data failed! reason[%s]",
			ActiveRunningKey, err.Error())
	}
	base.GLog.Debug("Set ActiveRunningKey[%s] data[%s] successed!", ActiveRunningKey, string(redisData))
}

// 开启一个活动
func (ram *RunningActiveMgr) OpenActive(addActive *ProtoActiveControlInfoS) error {
	for index, _ := range ram.RunningActives {
		activeCtrl := &ram.RunningActives[index]
		if activeCtrl.ActiveID == addActive.ActiveID && activeCtrl.ActiveType == addActive.ActiveType {
			if activeCtrl.IsExpired() {
				// 活动过期了
				base.GLog.Debug("Active Type[%s] Id[%d] had expired, so reopen",
					activeCtrl.ActiveType.String(), activeCtrl.ActiveID)
				*activeCtrl = *addActive
				return nil
			} else {
				err := fmt.Errorf("Active Type[%s] Id[%d] is running! cannot open!",
					activeCtrl.ActiveType.String(), activeCtrl.ActiveID)
				base.GLog.Error(err.Error())
				return err
			}

		}
	}
	base.GLog.Debug("Active Type[%s] Id[%d] append to the running queue", addActive.ActiveType.String(),
		addActive.ActiveID)
	ram.RunningActives = append(ram.RunningActives, *addActive)
	return nil
}

// 关闭一个活动
func (ram *RunningActiveMgr) CloseActive(activeType ActiveType, activeID int) error {
	index := 0
	exists := false
	for index = range ram.RunningActives {
		activeCtrl := &ram.RunningActives[index]
		if activeCtrl.ActiveID == activeID && activeCtrl.ActiveType == activeType {
			exists = true
			break
		}
	}
	if exists {
		base.GLog.Debug("Active Type[%s] Id[%d] index[%d] remove from running queue", activeType.String(),
			activeID, index)
		ram.RunningActives = append(ram.RunningActives[:index], ram.RunningActives[index+1:]...)
	} else {
		err := fmt.Errorf("Active Type[%s] Id[%d] not in running queue!", activeType.String(),
			activeID)
		base.GLog.Error(err.Error())
		return err
	}
	return nil
}

func (ram *RunningActiveMgr) IsEmpty() bool {
	return len(ram.RunningActives) == 0
}

//----------------------------------------------------------------------------------
// 开启活动命令
type ProtoOpenActiveReq struct {
	Uin                  uint64                    `json:"Uin"`
	ZoneID               int32                     `json:"ZoneID"`
	ActiveControlConfigs []ProtoActiveControlInfoS `json:"ActiveControlConfigs"`
}

// 关闭活动命令
type ProtoCloseActiveReq struct {
	Uin        uint64     `json:"Uin"`
	ZoneID     int32      `json:"ZoneID"`
	ActiveType ActiveType `json:"ActiveType"` // 活动类型id，超值礼包、兑换
	ActiveIDs  []int      `json:"ActiveIDs"`  // 同类型下的活动id，例如不同的超值礼包
}

// 活动开启关闭的控制结果
type ProtoControlActiveRes struct {
	Uin            uint64  `json:"Uin"`
	ZoneID         int32   `json:"ZoneID"`
	ActiveIDs      []int   `json:"ActiveIDs"`      //
	ControlResults []int32 `json:"ControlResults"` // =1成功 =0失败
}

//----------------------------------------------------------------------------------
// 活动
type ActiveBaseConfigInfoS struct {
	ActiveID      int    `json:"ID" mapstructure:"ID"`   // 同类型下的活动id，例如不同的超值礼包
	ReceiveLimit  int32  `json:"RL" mapstructure:"RL"`   // 活动领取次数限制
	ReceiveCond   int32  `json:"RC" mapstructure:"RC"`   // 活动领取条件，这里就是判断accumulate值
	ResetEveryDay int32  `json:"RED" mapstructure:"RED"` // 是否每天开启 1：隔天重置 0：整个活动有效期
	ChannelID     string `json:"C" mapstructure:"C"`     // CN US AreaCode
}

// 超级礼包
type ActiveSuperGiftInfoS struct {
	Base                  ActiveBaseConfigInfoS `json:"base"`                 // 活动的基本配置
	FakePrice             float32               `json:"FP" mapstructure:"FP"` //
	FakePriceDesc         string                `json:"FPD" mapstructure:"FPD"`
	DescKey               string                `json:"DK" mapstructure:"DK"`
	InnerGoods            string                `json:"IG" mapstructure:"IG"` // 对应的奖品
	NameKey               string                `json:"NK" mapstructure:"NK"`
	PosterAssetBundleName string                `json:"PAN" mapstructure:"PAN"`
	PosterTextureName     string                `json:"PIN" mapstructure:"PIN"`
	Price                 float32               `json:"P" mapstructure:"P"`
	PriceDesc             string                `json:"PD" mapstructure:"PD"`
	ProductId             string                `json:"PI" mapstructure:"PI"`
	Discount              int                   `json:"DC" mapstructure:"DC"`
	Type                  string                `json:"Type" mapstructure:"Type"`
}

// 活跃任务
type ActiveMissionInfoS struct {
	Base       ActiveBaseConfigInfoS `json:"base"`                 // 活动的基本配置
	InnerGoods string                `json:"IG" mapstructure:"IG"` // 对应的奖品
	TaskType   string                `json:"TP" mapstructure:"TP"`
	TitleKey   string                `json:"TK" mapstructure:"TK"`
	Parameter  string                `json:"PA" mapstructure"PA"`
}

// 兑换任务
type ActiveExchangeInfoS struct {
	Base         ActiveBaseConfigInfoS `json:"base"`                 // 活动的基本配置
	InnerGoods   string                `json:"IG" mapstructure:"IG"` // 兑换目标
	TitleKey     string                `json:"TK" mapstructure:"TK"`
	ExchangeCost string                `json:"EC" mapstructure:"EC"` // 兑换资源
}

// CDKEY兑换
type ActiveCDKeyExchangeInfoS struct {
	Base       ActiveBaseConfigInfoS `json:"base"` // 活动的基本配置
	CDKey      string                `json:"CDKey"`
	InnerGoods string                `json:"IG" mapstructure:"IG`
}

// GM导入超级礼包配置
type ProtoGMConfigActiveSuperGiftReq struct {
	Uin              uint64                 `json:"Uin"`
	ZoneID           int32                  `json:"ZoneID"`
	SuperGiftConfigs []ActiveSuperGiftInfoS `json:"SuperGiftConfigs"`
}

// GM导入活跃任务配置
type ProtoGMConfigActiveMissionReq struct {
	Uin            uint64               `json:"Uin"`
	ZoneID         int32                `json:"ZoneID"`
	ActiveMissions []ActiveMissionInfoS `json:"ActiveMissions"`
}

// GM导入兑换任务配置
type ProtoGMConfigActiveExchangeReq struct {
	Uin             uint64                `json:"Uin"`
	ZoneID          int32                 `json:"ZoneID"`
	ActiveExchanges []ActiveExchangeInfoS `json:"ActiveExchanges"`
}

// GM导入CDKEY兑换配置
type ProtoGMConfigActiveCDKeyExchangeReq struct {
	Uin                  uint64                     `json:"Uin"`
	ZoneID               int32                      `json:"ZoneID"`
	ActiveCDKeyExchanges []ActiveCDKeyExchangeInfoS `json:"ActiveCDKeyExchanges"`
}

// 活动获取通用接口
type ProtoGetPlayerActiveReq struct {
	Uin        uint64     `json:"Uin"`
	ZoneID     int32      `json:"ZoneID"`
	ActiveType ActiveType `json:"ActiveType" mapstructure:"ActiveType"` // 活动类型id，超值礼包、兑换
	ChannelID  string     `json:"ChannelID" mapstructure:"ChannelID"`   // CN US NOAREA
}

type ProtoActiveInstanceInfo struct {
	ActiveID             int         `json:"AID"` // 活动实例id
	RemainderSeconds     int64       `json:"RS"`  // 活动剩余秒数
	AccumulateCount      int32       `json:"AC"`  // 运行时数据，玩家在该活动的累积数量，超级礼包该字段无效，兑换活动是作为领取的判断条件
	ReceiveCount         int32       `json:"RC"`  // 运行时数据，领取的次数，有的可以领取多次，有的一天只能领取一次
	ActiveinstanceConfig interface{} `json:"AIC"` // 各种活动的配置
	RefreshRemainderSecs int64       `json:"RRS"` // 刷新剩余秒数
}

type ProtoGetPlayerActiveRes struct {
	Uin               uint64                    `json:"Uin"`
	ZoneID            int32                     `json:"ZoneID"`
	ActiveType        ActiveType                `json:"ActiveType" mapstructure:"ActiveType"` // 活动类型id，超值礼包、兑换
	ActiveInstanceLst []ProtoActiveInstanceInfo `json:"AILST" mapstructure:"AILST"`
}

// 活动领取通用接口
type ProtoPlayerActiveReceiveReq struct {
	Uin        uint64     `json:"Uin"`
	ZoneID     int32      `json:"ZoneID"`
	ActiveType ActiveType `json:"ActiveType" mapstructure:"ActiveType"` // 活动类型id，超值礼包、兑换
	ActiveID   int        `json:"ActiveID" mapstructure:"ActiveID"`     // 同类型下的活动id，例如不同的超值礼包，兑换，这些都是可以在活动期间多次领取的
	ChannelID  string     `json:"ChannelID" mapstructure:"ChannelID"`   // CN US NOAREA(LUA层调用的都填写NOAREA)
}

type ProtoPlayerActiveReceiveRes struct {
	Uin             uint64     `json:"Uin"`
	ZoneID          int32      `json:"ZoneID"`
	ActiveType      ActiveType `json:"Type" mapstructure:"Type"` // 活动类型id，超值礼包、兑换
	ActiveID        int        `json:"ID" mapstructure:"ID"`     // 同类型下的活动id，例如不同的超值礼包
	InnerGoods      string     `json:"IG" mapstructure:"IG"`     // 领取的奖励
	AccumulateCount int32      `json:"AC" mapstructure:"AC`      // 玩家在该活动的累积数量，超级礼包该字段无效，兑换活动是作为领取的判断条件
	ReceiveCount    int32      `json:"RC" mapstructure:"RC`      // 领取的次数，有的可以领取多次，有的一天只能领取一次
}

// 活动累积参数通知接口，由内部系统调用
type ProtoActiveDataS struct {
	ActiveTaskType     string `json:"ActiveTaskType"`
	TaskOpType         int32  `json:"TaskOpType"`           // =1: SET, =2: ADD
	AccumalateParamter int32  `json:"AP" mapstructure:"AP"` // 该活动继承的船数量，获得排位赛胜利，只要完整活动的行为都调用该接口
}

type ProtoActiveAccumulateParameterNtf struct {
	Uin         uint64             `json:"Uin"`
	ZoneID      int32              `json:"ZoneID"`
	ActiveType  ActiveType         `json:"ActiveType" mapstructure:"ActiveType"`   // 活动类型，活跃任务
	ActiveDatas []ProtoActiveDataS `json:"ActiveDatas" mapstructure:"ActiveDatas"` // 相关参数
}
type ProtoActiveCanReceiveNtf struct {
	Uin        uint64     `json:"Uin"`
	ZoneID     int32      `json:"ZoneID"`
	ActiveType ActiveType `json:"ActiveType" mapstructure:"ActiveType"` // 活动类型，活跃任务
	IsDone     int32      `json:"IsDone" mapstructure:"IsDone"`         // =1 任务可以领取 =0 任务不可领取
}

// 获得兑换花费
type ProtoGetActiveExchangeCostReq struct {
	Uin        uint64     `json:"Uin"`
	ZoneID     int32      `json:"ZoneID"`
	ActiveType ActiveType `json:"ActiveType" mapstructure:"ActiveType"` // 活动类型id，超值礼包、兑换
	ActiveID   int        `json:"ActiveID" mapstructure:"ActiveID"`     // 同类型下的活动id，例如不同的超值礼包
}

type ProtoGetActiveExchangeCostRes struct {
	Uin          uint64     `json:"Uin"`
	ZoneID       int32      `json:"ZoneID"`
	ActiveType   ActiveType `json:"ActiveType" mapstructure:"ActiveType"` // 活动类型id，超值礼包、兑换
	ActiveID     int        `json:"ActiveID" mapstructure:"ActiveID"`     // 同类型下的活动id，例如不同的超值礼包
	ExchangeCost string     `json:"EC" mapstructure:"EC"`                 // 兑换条件，json content
}

// 检查活动是否配置是否存在
type ProtoCheckActiveConfigReq struct {
	Uin        uint64     `json:"Uin"`
	ZoneID     int32      `json:"ZoneID"`
	ActiveType ActiveType `json:"ActiveType" mapstructure:"ActiveType"` // 活动类型id，超值礼包、兑换
	ActiveIDs  []int      `json:"ActiveIDs" mapstructure:"ActiveIDs"`   // 同类型下的活动id列表，例如不同的超值礼包
}

type ProtoCheckActiveConfigRes struct {
	Uin        uint64     `json:"Uin"`
	ZoneID     int32      `json:"ZoneID"`
	ActiveType ActiveType `json:"ActiveType" mapstructure:"ActiveType"` // 活动类型id，超值礼包、兑换
	ActiveIDs  []int      `json:"ActiveIDs" mapstructure:"ActiveIDs"`   // 同类型下的活动id列表，例如不同的超值礼包
	IsExists   []int32    `json:"IsExists" mapstructure:"IsExists"`     // 存在列表 =1 配置存在，=0 配置不存在
}

// cdkey兑换
type ProtoPlayerExchangeCDKeyReq struct {
	Uin    uint64 `json:"Uin"`
	ZoneID int32  `json:"ZoneID"`
	CDKey  string `json:"CDKey"`
}

type ProtoPlayerExchangeCDKeyRes struct {
	Uin        uint64 `json:"Uin"`
	ZoneID     int32  `json:"ZoneID"`
	InnerGoods string `json:"IG" mapstructure:"IG"` // 兑换的奖励
}

// 检查一次活动是否完成
type ProtoCheckPlayerActiveIsCompletedReq struct {
	Uin    uint64 `json:"Uin"`
	ZoneID int32  `json:"ZoneID"`
}

type ActiveIsCompleteS struct {
	ActiveType  ActiveType `json:"ActiveType"`
	IsCompleted int32      `json:"IsCompleted"` // 领取是否结束，1：结束, 0: 未结束
}

type ProtoCheckPlayerActiveIsCompletedRes struct {
	Uin               uint64              `json:"Uin"`
	ZoneID            int32               `json:"ZoneID"`
	ActiveCompleteLst []ActiveIsCompleteS `json:"ACLst"`
}

type ProtoQueryRechargeCommodityPricesReq struct {
	ZoneID                int32                 `json:"ZoneID"`
	RechargeCommodityType RechargeCommodityType `json:"RCType" mapstructure:"RCType"`
	ChannelID             string                `json:"ChannelID"` // CN US
	RechargeCommodityID   int                   `json:"ID" mapstructure:"ID"`
}

type ProtoQueryRechargeCommodityPricesRes struct {
	ZoneID                int32                 `json:"ZoneID"`
	RechargeCommodityType RechargeCommodityType `json:"RCType" mapstructure:"RCType"`
	ChannelID             string                `json:"ChannelID"` // CN US
	RechargeCommodityID   int                   `json:"ID" mapstructure:"ID"`
	Price                 float32               `json:"Price" mapstructure:"Price"`
}
