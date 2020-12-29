package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sailcraft/base"
)

type GuildBattleConfig struct {
	Duration                 int `json:"duration"`                    // 上/下半场持续时间
	FirstHalfInDay           int `json:"first_half_in_day"`           // 上半场在当天开始时间
	SecondHalfInDay          int `json:"second_half_in_day"`          // 下半场半场在当天开始时间
	IntervalDayBetweenBattle int `json:"interval_day_between_battle"` // 两次公会战间隔天数
	PlayerBattleTimes        int `json:"player_battle_times"`         // 公会成员战斗次数
	WinScore                 int `json:"win_score"`
	DrawScore                int `json:"draw_score"`
	LoseScore                int `json:"lose_score"`
	WinStreakScore           int `json:"win_streak_score"`
}

type ValueRatioConfig struct {
	Gold  float32 `json:"gold"`
	Wood  float32 `json:"wood"`
	Iron  float32 `json:"iron"`
	Stone float32 `json:"stone"`
}

type GlobalCampaignConfig struct {
	TimeTriggerInterval      int           `json:"time_trigger_interval"`
	TimeTriggerLimit         int           `json:"time_trigger_limit"`
	PVPWinTriggerLimit       int           `json:"pvp_win_trigger_limit"`
	PVPWinTriggerProbability float32       `json:"pvp_win_trigger_probability"`
	ProduceResource          ResourcesAttr `json:"produce_resource"`
	MaxUnfinishedEventLimit  []int         `json:"max_unfinished_event_limit"`
}

type GlobalGuildConfig struct {
	CreateGuildCost        ResourcesAttr     `json:"create_guild_cost"`
	MaxMemberCount         int               `json:"max_member_count"`
	JoinGuildColdTime      int               `json:"join_guild_cold_time"`
	LimitChatMsg           int               `json:"limit_chat_message"`
	LimitGuildInvite       int               `json:"limit_guild_invite"`
	LimitGuildOperationMsg int               `json:"limit_guild_operation_message"`
	LimitGuildRecommend    int               `json:"limit_guild_recommend"`
	LimitGuildSearch       int               `json:"limit_guild_search"`
	LimitGuildApply        int               `json:"limit_guild_apply"`
	LimitDailyVitality     int               `json:"limit_daily_vitality"`
	LimitPostMemberCount   map[string]int    `json:"limit_post_member_count"`
	LimitSalvageTimes      int               `json:"limit_salvage_times"`
	SalvageRecoverTime     int               `json:"salvage_recover_time"`
	ValueRatio             ValueRatioConfig  `json:"value_ratio"`
	BattleConfig           GuildBattleConfig `json:"battle"`
	ChairmanMaxOfflineTime int               `json:"chairman_max_offline_time"`
}

type GlobalFleetScoreConfig struct {
	ScoreDelta int `json:"score_delta"`
}

type GlobalConfig struct {
	Campaign GlobalCampaignConfig   `json:"campaign"`
	Guild    GlobalGuildConfig      `json:"guild"`
	Fleet    GlobalFleetScoreConfig `json:"fleet_score"`
}

var (
	GGlobalConfig *GlobalConfig = new(GlobalConfig)
)

func (conf *GlobalConfig) Init(configFile string) error {
	base.GLog.Debug("LoadConfig [%s]", configFile)

	hFile, err := os.Open(configFile)
	if err != nil {
		base.GLog.Error("open file %s failed err %s \n", configFile, err.Error())
		return err
	}
	defer hFile.Close()

	data, err := ioutil.ReadAll(hFile)
	if err != nil {
		base.GLog.Error("read file %s failed err %s \n", configFile, err.Error())
		return err
	}

	err = json.Unmarshal(data, conf)
	if err != nil {
		base.GLog.Error("parse json %s failed err %s \n", configFile, err.Error())
	}

	base.GLog.Debug("global config data is [%+v]", *conf)

	return nil
}
