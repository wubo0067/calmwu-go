package handler

import (
	"encoding/json"
	"sailcraft/base"
	"sailcraft/fleetsvr_main/custom_errors"
	"sailcraft/fleetsvr_main/errorcode"
	"sailcraft/fleetsvr_main/handlerbase"
	"sailcraft/fleetsvr_main/proto"
	"time"

	"github.com/mitchellh/mapstructure"
)

const (
	// 事件名称
	EVENT_NAME_ON_LOGIN                               = "OnLogin"
	EVENT_NAME_ON_PLAYER_LEVEL_UP                     = "OnPlayerLevelUp"
	EVENT_NAME_ON_BATTLE_SHIP_LEVEL_UP                = "OnBattleShipLevelUp"
	EVENT_NAME_ON_BATTLE_SHIP_STAR_LEVEL_UP           = "OnBattleShipStarLevelUp"
	EVENT_NAME_ON_BATTLE_SHIP_MERGE                   = "OnBattleShipMerge"
	EVENT_NAME_ON_FRIGATE_LEVEL_UP                    = "OnFrigateLevelUp"
	EVENT_NAME_ON_FRIGATE_WEAPON_LEVEL_UP             = "OnFrigateWeaponLevelUp"
	EVENT_NAME_ON_FRIGATE_SKILL_LEVEL_UP              = "OnFrigateSkillLevelUp"
	EVENT_NAME_ON_CAMPAIGN_PASS                       = "OnCampaignPass"
	EVENT_NAME_ON_CAMPAIGN_EVENT_AWARD_RECEIVED       = "OnCampaignEventAwardReceived"
	EVENT_NAME_ON_CAMPAIGN_PRODUCE_RESOURCES_RECEIVED = "OnCampaignProduceResourcesReceived"
	EVENT_NAME_ON_LEAGUE_LEVEL_UP                     = "OnLeagueLevelUp"
	EVENT_NAME_ON_GUILD_POST_CHANGED                  = "OnGuildPostChanged"
	EVENT_NAME_ON_NEWBIE_TEACH_PASS                   = "OnNewbieTeachPass"
	EVENT_NAME_ON_NEWBIE_GUID_PASS                    = "OnNewbieGuidPass"
	EVENT_NAME_ON_CHARGE                              = "OnCharge"
	EVENT_NAME_ON_SHOP_PURCHASED                      = "OnShopPurchased"
	EVENT_NAME_ON_SHARE_TO_SOCIAL                     = "OnShareToSocial"
	EVENT_NAME_ON_CARD_BAG_OPEN                       = "OnCardBagOpen"
	EVENT_NAME_ON_BATTLE_END                          = "OnBattleEnd"
	EVENT_NAME_ON_RESOURCES_COST                      = "OnResourcesCost"
	EVENT_NAME_ON_DAILY_VITALITY_CHANGED              = "OnDailyVitalityChanged"
	EVENT_NAME_ON_BREAK_OUT_DONE                      = "OnBreakOutDone"
	EVENT_NAME_ON_SALVAGE                             = "OnSalvage"

	// 战斗类型
	BATTLE_TYPE_PVP       = "PVP"
	BATTLE_TYPE_PVE       = "PVE"
	BATTLE_TYPE_BREAK_OUT = "BreakOut"
	BATTLE_TYPE_GUILD_WAR = "GuildWar"

	// 战斗结果
	BATTLE_RESULT_FAILED  = 1
	BATTLE_RESULT_DRAW    = 2
	BATTLE_RESULT_SUCCESS = 3

	// 船类型
	SHIP_TYPE_FRIGATE      = "frigate"
	SHIP_TYPE_BATTLE_SHIP  = "battleship"
	SHIP_TYPE_INSTALLATION = "installation"

	// 公会职位
	GUILD_POST_CHAIRMAN      = "chairman"
	GUILD_POST_VICE_CHAIRMAN = "vice_chairman"
	GUILD_POST_ELDER         = "elder"
	GUILD_POST_MEMBER        = "member"

	// 商品类型
	GOODS_TYPE_CARDPACK = "cardpack"
)

// 道具
type EventPropItem struct {
	ProtypeId int `mapstructure:"ProtypeId"`
	Count     int `mapstructure:"Count"`
}

// 战舰卡片
type EventBattleShipCardItem struct {
	ProtypeId int `mapstructure:"ProtypeId"`
	Count     int `mapstructure:"Count"`
}

// 资源
type EventResourceItem struct {
	Type  string `json:"Type"`
	Count int    `json:"Count"`
}

type EventResourceAttr struct {
	Props           []EventPropItem           `json:"Props"`
	BattleShipCards []EventBattleShipCardItem `json:"BattleShipCards"`
	Resources       []EventResourceItem       `json:"Resources"`
}

// 登录事件
type OnLoginEventData struct {
	Continuous    int `mapstructure:"Continuouse"`
	LastLoginTime int `mapstructure:"LastLoginTime"`
}

// 玩家升级
type OnPlayerLevelUpEventData struct {
	OldLevel int `mapstructure:"OldLevel"`
	Level    int `mapstructure:"Level"`
}

// 战舰升级
type OnBattleShipLevelUpEventData struct {
	Level     int `mapstructure:"Level"`
	ProtypeId int `mapstructure:"ProtypeId"`
}

// 战舰升星
type OnBattleShipStarLevelUpEventData struct {
	StarLevel int `mapstructure:"StarLevel"`
	ProtypeId int `mapstructure:"ProtypeId"`
}

// 战舰合并
type OnBattleShipMergeEventData struct {
	ProtypeId int `mapstructure:"ProtypeId"`
}

// 护卫舰升级
type OnFrigateLevelUpEventData struct {
	Level int `mapstructure:"Level"`
}

// 护卫舰武器升级
type OnFrigateWeaponLevelUpEventData struct {
	Level     int `mapstructure:"Level"`
	ProtypeId int `mapstructure:"ProtypeId"`
}

// 护卫舰技能升级
type OnFrigateSkillLevelUpEventData struct {
	Level     int `mapstructure:"Level"`
	ProtypeId int `mapstructure:"ProtypeId"`
}

// PVE征服世界---通关
type OnCampaignPassEventData struct {
	CampaignId      int `mapstructure:"CampaignId"`
	FirstTimeToPass int `mapstructure:"FirstTimeToPass"`
}

// PVE征服世界---领取事件奖励
type OnCampaignEventAwardReceivedEventData struct {
	ProtypeId int `mapstructure:"ProtypeId"`
}

// PVE征服世界---领取生产奖励
type OnCampaignProduceResourcesReceivedEventData struct {
	Rewards EventResourceAttr `mapstructure:"Rewards"`
}

// 联赛升级
type OnLeagueLevelUpEventData struct {
	LeagueLevel    int `mapstructure:"LeagueLevel"`
	MaxLeagueLevel int `mapstructure:"MaxLeagueLevel"`
}

// 公会职位变化
type OnGuildPostChangedEventData struct {
	Post string `mapstructure:"Post"`
}

// 新手教学
type OnNewbieTechPassEventData struct {
	Step    int `mapstructure:"Step"`
	MaxStep int `mapstructure:"MaxStep"`
}

// 新手引导
type OnNewbieGuidPassEventData struct {
	IsAllDone int `mapstructure:"IsAllDone"`
}

// 充值
type OnChargeEventData struct {
	Gem      int `mapstructure:"Gem"`
	VipLevel int `mapstructure:"VipLevel"`
}

// 商店购买
type GoodsInfo struct {
	Type      string `mapstructure:"Type"`
	ProtypeId int    `mapstructure:"ProtypeId"`
	Count     int    `mapstructure:"Count"`
}

type OnShopPurchasedEventData struct {
	Goods *EventResourceAttr `mapstructure:"Goods"`
	Cost  *EventResourceAttr `mapstructure:"Cost"`
}

// 社交平台分享
type OnShareToSocialEventData struct {
	Type string `mapstructure:"Type"`
}

// 开卡包
type CardBagBattleShipCard struct {
	ProtypeId int `mapstructure:"ProtypeId"`
	Count     int `mapstructure:"Count"`
}

type OnCardBagOpenEventData struct {
	ProtypeId int                     `mapstructure:"ProtypeId"`
	ShipCards []CardBagBattleShipCard `mapstructure:"ShipCards"`
}

// 资源消耗
type OnResourceCostEventData struct {
	Resources EventResourceAttr `mapstructure:"Resources"`
}

type OnDailyVitalityChanged struct {
	Vitality int `mapstruct:"Vitality"`
}

// 战斗结束事件
type OnBattleEndEventData struct {
	BattleType           string  `mapstructure:"Type"`
	BattleResult         int     `mapstructure:"Result"`
	SinkShipCount        int     `mapstructure:"SinkShipCount"`
	WinStreak            int     `mapstructure:"WinStreak"`
	MaxCombos            int     `mapstructure:"MaxCombos"`
	SrcScore             float64 `mapstructure:"SrcScore"`
	RivalScore           float64 `mapstructure:"RivalScore"`
	FrigateWeaponKORival int     `mapstructure:"FrigateWeaponKORival"`
	PerishTogether       int     `mapstructure:"PerishTogether"`
	MainFormationAlive   int     `mapstructure:"MainFormationAlive"`
	MaxCombosHistory     int     `mapstructure:"MaxCombosHistory"`
	TotalPvpTimes        int     `mapstructure:"TotalPvpTimes"`
	GuildId              string  `mapstructure:"GuildId"`
	WarId                int     `mapstructure:"WarId"`
}

type OnBreakOutDone struct {
	IsAllDone int `mapstructure:"IsAllDone"`
}

type OnSalvage struct {
}

// 事件集合
type EventsHappenedDataSet struct {
	LoginEventData                            *OnLoginEventData
	DailyVitalityChangedEventData             *OnDailyVitalityChanged
	PlayerLevelUpEventData                    *OnPlayerLevelUpEventData
	BattleShipLevelUpEventData                *OnBattleShipLevelUpEventData
	BattleShipStarLevelUpEventData            *OnBattleShipStarLevelUpEventData
	BattleShipMergeEventData                  *OnBattleShipMergeEventData
	FrigateLevelUpEventData                   *OnFrigateLevelUpEventData
	FirgateWeaponLevelUpEventData             *OnFrigateWeaponLevelUpEventData
	FrigateSkillLevelUpEventData              *OnFrigateSkillLevelUpEventData
	CampaignPassEventData                     *OnCampaignPassEventData
	CampaignEventAwardReceivedEventData       *OnCampaignEventAwardReceivedEventData
	CampaignProduceResourcesReceivedEventData *OnCampaignProduceResourcesReceivedEventData
	LeagueLevelUpEventData                    *OnLeagueLevelUpEventData
	GuildPostChangedEventData                 *OnGuildPostChangedEventData
	NewbieTechPassEventData                   *OnNewbieTechPassEventData
	NewbieGuidPassEventData                   *OnNewbieGuidPassEventData
	ChargeEventData                           *OnChargeEventData
	ShopPurchasedEventData                    *OnShopPurchasedEventData
	ShareToSocialEventData                    *OnShareToSocialEventData
	CardBagOpenEventData                      *OnCardBagOpenEventData
	ResourcesCostEventData                    *OnResourceCostEventData
	BattleEndData                             *OnBattleEndEventData
	BreakOutDoneEventData                     *OnBreakOutDone
	SalvageData                               *OnSalvage
}

// 事件处理函数模板
type EventHandleFunc func(*base.ProtoRequestS, *map[string]interface{}, *EventsHappenedDataSet) (int, error)

// 事件处理函数数组
var eventHandlers []EventHandleFunc

// 事件处理函数数组初始化
func init() {
	RegisterEventHandler(HandleCampaignEventProgress)
	RegisterEventHandler(OnCampaignEventTriggerForPvpWin)
	RegisterEventHandler(HandleAchievementProgress)
	RegisterEventHandler(HandleActivityTaskProgress)
	RegisterEventHandler(HandleGrowupTaskProgress)
	RegisterEventHandler(HandleGuildMission)
}

type EventHandler struct {
	handlerbase.WebHandler
}

func (this *EventHandler) OnEventHappened() (int, error) {
	var reqParams proto.ProtoOnEventHappenedRequest
	err := this.UnpackParams(&reqParams)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	var events EventsHappenedDataSet

	retCode, err := composeDataSet(&events, &reqParams)
	if err != nil {
		return retCode, err
	}

	base.GLog.Debug("EventsHappenedDataSet:[%+v]", events)

	retCode, err = OnEventHappenedWithEventDataSet(this.Request, &events, this.Response)
	if err != nil {
		return retCode, err
	}

	return 0, nil
}

func OnEventHappenedWithEventDataSet(req *base.ProtoRequestS, events *EventsHappenedDataSet, res *base.ProtoResponseS) (int, error) {
	if req == nil || res == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.NullPoint()
	}

	resParams := make(map[string]interface{})

	for _, eventHandler := range eventHandlers {
		if eventHandler != nil {
			start := time.Now()
			// 蛋疼，在这里循环调用
			retCode, err := eventHandler(req, &resParams, events)
			base.GLog.Debug("eventHandler[%+v] Cost: %dns", eventHandler, time.Since(start).Nanoseconds())
			if err != nil {
				return retCode, err
			}
		}
	}

	oldParams := make(map[string]interface{})
	if res.ResData.Params != nil {
		data, err := json.Marshal(&res.ResData.Params)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		err = json.Unmarshal(data, &oldParams)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		for k, v := range oldParams {
			resParams[k] = v
		}
	}

	res.ResData.Params = resParams
	return 0, nil
}

// 注册事件处理函数
func RegisterEventHandler(hd EventHandleFunc) {
	if hd != nil {
		eventHandlers = append(eventHandlers, hd)
	}
}

func composeDataSet(dst *EventsHappenedDataSet, src *proto.ProtoOnEventHappenedRequest) (int, error) {

	if dst == nil || src == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.NullPoint()
	}

	for index := range src.Events {
		eventInfo := &src.Events[index]
		var err error
		switch eventInfo.EventName {
		case EVENT_NAME_ON_LOGIN:
			if dst.LoginEventData == nil {
				dst.LoginEventData = new(OnLoginEventData)
			}
			err = mapstructure.Decode(eventInfo.EventData, dst.LoginEventData)
		case EVENT_NAME_ON_PLAYER_LEVEL_UP:
			if dst.PlayerLevelUpEventData == nil {
				dst.PlayerLevelUpEventData = new(OnPlayerLevelUpEventData)
			}
			err = mapstructure.Decode(eventInfo.EventData, dst.PlayerLevelUpEventData)
		case EVENT_NAME_ON_BATTLE_SHIP_LEVEL_UP:
			if dst.BattleShipLevelUpEventData == nil {
				dst.BattleShipLevelUpEventData = new(OnBattleShipLevelUpEventData)
			}
			err = mapstructure.Decode(eventInfo.EventData, dst.BattleShipLevelUpEventData)
		case EVENT_NAME_ON_BATTLE_SHIP_STAR_LEVEL_UP:
			if dst.BattleShipStarLevelUpEventData == nil {
				dst.BattleShipStarLevelUpEventData = new(OnBattleShipStarLevelUpEventData)
			}
			err = mapstructure.Decode(eventInfo.EventData, dst.BattleShipStarLevelUpEventData)
		case EVENT_NAME_ON_BATTLE_SHIP_MERGE:
			if dst.BattleShipMergeEventData == nil {
				dst.BattleShipMergeEventData = new(OnBattleShipMergeEventData)
			}
			err = mapstructure.Decode(eventInfo.EventData, dst.BattleShipMergeEventData)
		case EVENT_NAME_ON_FRIGATE_LEVEL_UP:
			if dst.FrigateLevelUpEventData == nil {
				dst.FrigateLevelUpEventData = new(OnFrigateLevelUpEventData)
			}
			err = mapstructure.Decode(eventInfo.EventData, dst.FrigateLevelUpEventData)
		case EVENT_NAME_ON_FRIGATE_WEAPON_LEVEL_UP:
			if dst.FirgateWeaponLevelUpEventData == nil {
				dst.FirgateWeaponLevelUpEventData = new(OnFrigateWeaponLevelUpEventData)
			}
			err = mapstructure.Decode(eventInfo.EventData, dst.FirgateWeaponLevelUpEventData)
		case EVENT_NAME_ON_FRIGATE_SKILL_LEVEL_UP:
			if dst.FrigateSkillLevelUpEventData == nil {
				dst.FrigateSkillLevelUpEventData = new(OnFrigateSkillLevelUpEventData)
			}
			err = mapstructure.Decode(eventInfo.EventData, dst.FrigateSkillLevelUpEventData)
		case EVENT_NAME_ON_CAMPAIGN_PASS:
			if dst.CampaignPassEventData == nil {
				dst.CampaignPassEventData = new(OnCampaignPassEventData)
			}
			err = mapstructure.Decode(eventInfo.EventData, dst.CampaignPassEventData)
		case EVENT_NAME_ON_CAMPAIGN_EVENT_AWARD_RECEIVED:
			if dst.CampaignEventAwardReceivedEventData == nil {
				dst.CampaignEventAwardReceivedEventData = new(OnCampaignEventAwardReceivedEventData)
			}
			err = mapstructure.Decode(eventInfo.EventData, dst.CampaignEventAwardReceivedEventData)
		case EVENT_NAME_ON_CAMPAIGN_PRODUCE_RESOURCES_RECEIVED:
			if dst.CampaignProduceResourcesReceivedEventData == nil {
				dst.CampaignProduceResourcesReceivedEventData = new(OnCampaignProduceResourcesReceivedEventData)
			}
			err = mapstructure.Decode(eventInfo.EventData, dst.CampaignProduceResourcesReceivedEventData)
		case EVENT_NAME_ON_LEAGUE_LEVEL_UP:
			if dst.LeagueLevelUpEventData == nil {
				dst.LeagueLevelUpEventData = new(OnLeagueLevelUpEventData)
			}
			err = mapstructure.Decode(eventInfo.EventData, dst.LeagueLevelUpEventData)
		case EVENT_NAME_ON_GUILD_POST_CHANGED:
			if dst.GuildPostChangedEventData == nil {
				dst.GuildPostChangedEventData = new(OnGuildPostChangedEventData)
			}
			err = mapstructure.Decode(eventInfo.EventData, dst.GuildPostChangedEventData)
		case EVENT_NAME_ON_NEWBIE_TEACH_PASS:
			if dst.NewbieTechPassEventData == nil {
				dst.NewbieTechPassEventData = new(OnNewbieTechPassEventData)
			}
			err = mapstructure.Decode(eventInfo.EventData, dst.NewbieTechPassEventData)
		case EVENT_NAME_ON_NEWBIE_GUID_PASS:
			if dst.NewbieGuidPassEventData == nil {
				dst.NewbieGuidPassEventData = new(OnNewbieGuidPassEventData)
			}
			err = mapstructure.Decode(eventInfo.EventData, dst.NewbieGuidPassEventData)
		case EVENT_NAME_ON_CHARGE:
			if dst.ChargeEventData == nil {
				dst.ChargeEventData = new(OnChargeEventData)
			}
			err = mapstructure.Decode(eventInfo.EventData, dst.ChargeEventData)
		case EVENT_NAME_ON_SHOP_PURCHASED:
			if dst.ShopPurchasedEventData == nil {
				dst.ShopPurchasedEventData = new(OnShopPurchasedEventData)
			}
			err = mapstructure.Decode(eventInfo.EventData, dst.ShopPurchasedEventData)
		case EVENT_NAME_ON_SHARE_TO_SOCIAL:
			if dst.ShareToSocialEventData == nil {
				dst.ShareToSocialEventData = new(OnShareToSocialEventData)
			}
			err = mapstructure.Decode(eventInfo.EventData, dst.ShareToSocialEventData)
		case EVENT_NAME_ON_CARD_BAG_OPEN:
			if dst.CardBagOpenEventData == nil {
				dst.CardBagOpenEventData = new(OnCardBagOpenEventData)
			}
			err = mapstructure.Decode(eventInfo.EventData, dst.CardBagOpenEventData)
		case EVENT_NAME_ON_BATTLE_END:
			if dst.BattleEndData == nil {
				dst.BattleEndData = new(OnBattleEndEventData)
			}
			err = mapstructure.Decode(eventInfo.EventData, dst.BattleEndData)
		case EVENT_NAME_ON_RESOURCES_COST:
			if dst.ResourcesCostEventData == nil {
				dst.ResourcesCostEventData = new(OnResourceCostEventData)
			}
			err = mapstructure.Decode(eventInfo.EventData, dst.ResourcesCostEventData)
		case EVENT_NAME_ON_DAILY_VITALITY_CHANGED:
			if dst.DailyVitalityChangedEventData == nil {
				dst.DailyVitalityChangedEventData = new(OnDailyVitalityChanged)
			}
			err = mapstructure.Decode(eventInfo.EventData, dst.DailyVitalityChangedEventData)
		case EVENT_NAME_ON_BREAK_OUT_DONE:
			if dst.BreakOutDoneEventData == nil {
				dst.BreakOutDoneEventData = new(OnBreakOutDone)
			}
			err = mapstructure.Decode(eventInfo.EventData, dst.BreakOutDoneEventData)
		case EVENT_NAME_ON_SALVAGE:
			if dst.SalvageData == nil {
				dst.SalvageData = new(OnSalvage)
			}
			err = mapstructure.Decode(eventInfo.EventData, dst.SalvageData)
		}

		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}
	}

	return 0, nil
}
