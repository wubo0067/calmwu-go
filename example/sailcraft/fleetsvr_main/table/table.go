package table

import (
	"sailcraft/dataaccess/mysql"
	"time"
)

const (
	BATTLE_SHIP_STATUS_CARD = 0
	BATTLE_SHIP_STATUS_SHIP = 1
)

type TblBattleShip struct {
	Id         int    `xorm:"int notnull pk autoincr 'id'"`
	Uin        int    `xorm:"int notnull index 'uin'"`
	ProtypeID  int    `xorm:"int notnull index 'protype_id'"`
	Level      int    `xorm:"int default(1) 'level'"`
	StarLevel  int    `xorm:"int default(0) 'star_level'"`
	CardNumber int    `xorm:"int default(0) 'card_number'"`
	Status     int    `xorm:"int default(0) 'status'"`
	Reserved0  int    `xorm:"int default(0) 'reserved_0'"`
	Reserved1  int    `xorm:"int default(0) 'reserved_1'"`
	Reserved2  string `xorm:"varchar(128) default('') 'reserved_2'"`
	Reserved3  string `xorm:"varchar(1024) default('') 'reserved_3'"`
	Reserved4  string `xorm:"varchar(4096) default('') 'reserved_4'"`
	Reserved5  string `xorm:"text default('') 'reserved_5'"`
}

type TblUserInfo struct {
	Uin             int       `xorm:"int notnull pk 'uin'"`
	UserName        string    `xorm:"varchar(128) default('') 'user_name'"`
	ISOCountryCode  string    `xorm:"char(64) default('Default') 'iso_country_code'"`
	ServerId        int       `xorm:"int default(0) 'server_id'"`
	Icon            string    `xorm:"varchar(64) default('') 'icon'"`
	Level           int       `xorm:"int default(1) 'level'"`
	Exp             int       `xorm:"int default(0) 'exp'"`
	Star            int       `xorm:"int default(0) 'star'"`
	Gold            int       `xorm:"int default(0) 'gold'"`
	Mineral         int       `xorm:"int default(0) 'mineral'"`
	Wood            int       `xorm:"int default(0) 'wood'"`
	Gem             int       `xorm:"int default(0) 'gem'"`
	PurchaseGem     int       `xorm:"int default(0) 'purchase_gem'"`
	Stone           int       `xorm:"int default(0) 'stone'"`
	Iron            int       `xorm:"int default(0) 'iron'"`
	ShipSoul        int       `xorm:"int default(0) 'shipsoul'"`
	ChangeNameCount int       `xorm:"int default(0) 'change_name_count'"`
	UserOfflineTime int       `xorm:"int default(0) 'user_offline_time'"`
	RegisterTime    int       `xorm:"int default(0) 'register_time'"`
	RegisterDate    time.Time `xorm:"timestamp 'register_date'"`
	GuildID         string    `xorm:"varchar(32) default('') 'guild_id'"`
	Reserved0       int       `xorm:"int default(0) 'reserved_0'"`
	Reserved1       int       `xorm:"int default(0) 'reserved_1'"`
	Reserved2       string    `xorm:"varchar(128) default('') 'reserved_2'"`
	Reserved3       string    `xorm:"varchar(1024) default('') 'reserved_3'"`
	Reserved4       string    `xorm:"varchar(4096) default('') 'reserved_4'"`
	Reserved5       string    `xorm:"text default('') 'reserved_5'"`
}

type TblCampaignPassChapter struct {
	Id         int    `xorm:"int notnull pk autoincr 'id'"`
	Uin        int    `xorm:"int notnull index 'uin'"`
	CampaignId int    `xorm:"int notnull index 'campaign_id'"`
	Reserved0  int    `xorm:"int default(0) 'reserved_0'"`
	Reserved1  int    `xorm:"int default(0) 'reserved_1'"`
	Reserved2  string `xorm:"varchar(128) default('') 'reserved_2'"`
	Reserved3  string `xorm:"varchar(1024) default('') 'reserved_3'"`
	Reserved4  string `xorm:"varchar(4096) default('') 'reserved_4'"`
	Reserved5  string `xorm:"text default('') 'reserved_5'"`
}

type TblCampaignEventProgress struct {
	Progress int `mapstructure:"Progress"`
	Total    int `mapstructure:"Total"`
}

type TblCampaignEvent struct {
	Id           int    `xorm:"int notnull pk autoincr 'id'"`
	Uin          int    `xorm:"int notnull index 'uin'"`
	CampaignId   int    `xorm:"int default(0) index 'campaign_id'"`
	EventId      int    `xorm:"int default(0) index 'event_id'"`
	StartTime    int    `xorm:"int default(0) 'start_time'"`
	MissionId    int    `xorm:"int notnull default(0) 'mission_id'"`
	ProgressData string `xorm:"varchar(1024) default('') 'progress_data'"`
	EventStatus  int    `xorm:"int default(0) 'event_status'"`
	TriggerType  int    `xorm:"int default(0) 'trigger_type'"`
	Reserved0    int    `xorm:"int default(0) 'reserved_0'"`
	Reserved1    int    `xorm:"int default(0) 'reserved_1'"`
	Reserved2    string `xorm:"varchar(128) default('') 'reserved_2'"`
	Reserved3    string `xorm:"varchar(1024) default('') 'reserved_3'"`
	Reserved4    string `xorm:"varchar(4096) default('') 'reserved_4'"`
	Reserved5    string `xorm:"text default('') 'reserved_5'"`
}

type TblCampaignEventFresh struct {
	Uin           int    `xorm:"int notnull pk 'uin'"`
	LastFreshTime int    `xorm:"int default(0) 'last_fresh_time'"`
	Reserved0     int    `xorm:"int default(0) 'reserved_0'"`
	Reserved1     int    `xorm:"int default(0) 'reserved_1'"`
	Reserved2     string `xorm:"varchar(128) default('') 'reserved_2'"`
	Reserved3     string `xorm:"varchar(1024) default('') 'reserved_3'"`
	Reserved4     string `xorm:"varchar(4096) default('') 'reserved_4'"`
	Reserved5     string `xorm:"text default('') 'reserved_5'"`
}

type TblCampaignProduceResource struct {
	Uin             int    `xorm:"int notnull pk 'uin'"`
	LastReceiveTime int    `xorm:"int default(0) 'last_receive_time'"`
	Reserved0       int    `xorm:"int default(0) 'reserved_0'"`
	Reserved1       int    `xorm:"int default(0) 'reserved_1'"`
	Reserved2       string `xorm:"varchar(128) default('') 'reserved_2'"`
	Reserved3       string `xorm:"varchar(1024) default('') 'reserved_3'"`
	Reserved4       string `xorm:"varchar(4096) default('') 'reserved_4'"`
	Reserved5       string `xorm:"text default('') 'reserved_5'"`
}

type TblCampaignEventStatistics struct {
	Uin                int    `xorm:"int notnull pk 'uin'"`
	TotalFinishedCount int    `xorm:"int default(0) 'total_finished_count'"`
	DailyFinishedData  string `xorm:"varchar(1024) default('') 'daily_finished_data'"`
	LastFreshTime      int    `xorm:"int default(0) 'last_fresh_time'"`
	Reserved0          int    `xorm:"int default(0) 'reserved_0'"`
	Reserved1          int    `xorm:"int default(0) 'reserved_1'"`
	Reserved2          string `xorm:"varchar(128) default('') 'reserved_2'"`
	Reserved3          string `xorm:"varchar(1024) default('') 'reserved_3'"`
	Reserved4          string `xorm:"varchar(4096) default('') 'reserved_4'"`
	Reserved5          string `xorm:"text default('') 'reserved_5'"`
}

type TblCampaignPlot struct {
	Id        int    `xorm:"int notnull pk autoincr 'id'"`
	Uin       int    `xorm:"int notnull index 'uin'"`
	ProtypeId int    `xorm:"int default(0) index 'protype_id'"`
	PassTime  int    `xorm:"int default(0) 'pass_time'"`
	Reserved0 int    `xorm:"int default(0) 'reserved_0'"`
	Reserved1 int    `xorm:"int default(0) 'reserved_1'"`
	Reserved2 string `xorm:"varchar(128) default('') 'reserved_2'"`
	Reserved3 string `xorm:"varchar(1024) default('') 'reserved_3'"`
	Reserved4 string `xorm:"varchar(4096) default('') 'reserved_4'"`
	Reserved5 string `xorm:"text default('') 'reserved_5'"`
}

type TblAchievement struct {
	Id           int    `xorm:"int notnull pk autoincr 'id'"`
	ProtypeId    int    `xorm:"int notnull unique(uin_protype_id) 'protype_id'"`
	Uin          int    `xorm:"int notnull unique(uin_protype_id) 'uin'"`
	Status       int    `xorm:"int default(0) 'status'"`
	ProgressData string `xorm:"varchar(1024) default('') 'progress_data'"`
	CompleteTime int    `xorm:"int default(0) 'complete_time'"`
	Reserved0    int    `xorm:"int default(0) 'reserved_0'"`
	Reserved1    int    `xorm:"int default(0) 'reserved_1'"`
	Reserved2    string `xorm:"varchar(128) default('') 'reserved_2'"`
	Reserved3    string `xorm:"varchar(1024) default('') 'reserved_3'"`
	Reserved4    string `xorm:"varchar(4096) default('') 'reserved_4'"`
	Reserved5    string `xorm:"text default('') 'reserved_5'"`
}

type TblActivityTask struct {
	Id           int    `xorm:"int notnull pk autoincr 'id'"`
	Uin          int    `xorm:"int notnull unique(uin_protype_id) 'uin'"`
	ProtypeId    int    `xorm:"int default(0) unique(uin_protype_id) 'protype_id'"`
	ProgressData string `xorm:"varchar(1024) default('') 'progress_data'"`
	Status       int    `xorm:"int notnull default(0) 'status'"`
	Reserved0    int    `xorm:"int default(0) 'reserved_0'"`
	Reserved1    int    `xorm:"int default(0) 'reserved_1'"`
	Reserved2    string `xorm:"varchar(128) default('') 'reserved_2'"`
	Reserved3    string `xorm:"varchar(1024) default('') 'reserved_3'"`
	Reserved4    string `xorm:"varchar(4096) default('') 'reserved_4'"`
	Reserved5    string `xorm:"text default('') 'reserved_5'"`
}

type TblActivityTaskFresh struct {
	Uin       int    `xorm:"int notnull pk 'uin'"`
	FreshTime int    `xorm:"int notnull 'fresh_time'"`
	Score     int    `xorm:"int default(0) 'score'"`
	Reserved0 int    `xorm:"int default(0) 'reserved_0'"`
	Reserved1 int    `xorm:"int default(0) 'reserved_1'"`
	Reserved2 string `xorm:"varchar(128) default('') 'reserved_2'"`
	Reserved3 string `xorm:"varchar(1024) default('') 'reserved_3'"`
	Reserved4 string `xorm:"varchar(4096) default('') 'reserved_4'"`
	Reserved5 string `xorm:"text default('') 'reserved_5'"`
}

type TblActivityScoreReward struct {
	Id        int    `xorm:"int notnull pk autoincr 'id'"`
	Uin       int    `xorm:"int notnull index unique(uin_reward_id) 'uin'"`
	RewardId  string `xorm:"varchar(128) default('') unique(uin_reward_id) 'reward_id'"`
	Status    int    `xorm:"int default(0) 'status'"`
	Reserved0 int    `xorm:"int default(0) 'reserved_0'"`
	Reserved1 int    `xorm:"int default(0) 'reserved_1'"`
	Reserved2 string `xorm:"varchar(128) default('') 'reserved_2'"`
	Reserved3 string `xorm:"varchar(1024) default('') 'reserved_3'"`
	Reserved4 string `xorm:"varchar(4096) default('') 'reserved_4'"`
	Reserved5 string `xorm:"text default('') 'reserved_5'"`
}

type TblGrowupTask struct {
	Id           int    `xorm:"int notnull pk autoincr 'id'"`
	Uin          int    `xorm:"int notnull unique(uin_protype_id) 'uin'"`
	ProtypeId    string `xorm:"varchar(128) default('') unique(uin_protype_id) 'protype_id'"`
	ProgressData string `xorm:"varchar(1024) default('') 'progress_data'"`
	Status       int    `xorm:"int notnull default(0) 'status'"`
	Reserved0    int    `xorm:"int default(0) 'reserved_0'"`
	Reserved1    int    `xorm:"int default(0) 'reserved_1'"`
	Reserved2    string `xorm:"varchar(128) default('') 'reserved_2'"`
	Reserved3    string `xorm:"varchar(1024) default('') 'reserved_3'"`
	Reserved4    string `xorm:"varchar(4096) default('') 'reserved_4'"`
	Reserved5    string `xorm:"text default('') 'reserved_5'"`
}

type TblGuildInfo struct {
	Id              int    `xorm:"int notnull pk autoincr 'id'"`
	PerformId       int    `xorm:"int notnull default(0) 'perform_id'"`
	Level           int    `xorm:"int notnull default(0) 'level'"`
	Creator         int    `xorm:"int notnull default(0) index 'creator'"`
	Vitality        int    `xorm:"int notnull default(0) 'vitality'"`
	Chairman        int    `xorm:"int notnull default(0) 'chairman'"`
	JoinType        int    `xorm:"int notnull default(0) 'join_type'"`
	CreateTime      int    `xorm:"int notnull default(0) 'create_time'"`
	MemberCount     int    `xorm:"int notnull default(0) 'member_count'"`
	CondLeagueLevel int    `xorm:"int notnull default(0) 'cond_league_level'"`
	Name            string `xorm:"varchar(128) default('') 'name'"`
	Symbol          string `xorm:"varchar(512) default('') 'symbol'"`
	Description     string `xorm:"varchar(512) default('') 'description'"`
	Reserved0       int    `xorm:"int default(0) 'reserved_0'"`
	Reserved1       int    `xorm:"int default(0) 'reserved_1'"`
	Reserved2       string `xorm:"varchar(128) default('') 'reserved_2'"`
	Reserved3       string `xorm:"varchar(1024) default('') 'reserved_3'"`
	Reserved4       string `xorm:"varchar(4096) default('') 'reserved_4'"`
	Reserved5       string `xorm:"text default('') 'reserved_5'"`
}

type TblGuildMemberInfo struct {
	Id        int    `xorm:"int notnull pk autoincr 'id'"`
	Creator   int    `xorm:"int notnull default(0) index 'creator'"`
	MemberUin int    `xorm:"int notnull default(0) 'member_uin'"`
	GuildId   int    `xorm:"int notnull default(0) 'guild_id'"`
	JoinTime  int    `xorm:"int notnull default(0) 'join_time'"`
	Vitality  int    `xorm:"int notnull default(0) 'vitality'"`
	Post      string `xorm:"varchar(255) default('') 'post'"`
	Reserved0 int    `xorm:"int default(0) 'reserved_0'"`
	Reserved1 int    `xorm:"int default(0) 'reserved_1'"`
	Reserved2 string `xorm:"varchar(128) default('') 'reserved_2'"`
	Reserved3 string `xorm:"varchar(1024) default('') 'reserved_3'"`
	Reserved4 string `xorm:"varchar(4096) default('') 'reserved_4'"`
	Reserved5 string `xorm:"text default('') 'reserved_5'"`
}

type TblGuildApplyInfo struct {
	GuildId   string `json:"guild_id"`
	ApplyUin  int    `json:"apply_uin"`
	ApplyTime int    `json:"apply_time"`
}

type TblGuildLeaveInfo struct {
	Uin       int    `json:"uin"`
	GuildId   string `json:"guild_id"`
	LeaveTime int    `json:"leave_time"`
}

type TblGuildInvite struct {
	Uin        int    `json:"uin"`
	GuildId    string `json:"guild_id"`
	InviteTime int    `json:"invite_time"`
	FromUin    int    `json:"from_uin"`
}

type TblGuildWeeklyVitality struct {
	Creator        int         `json:"creator"`
	GuildId        int         `json:"guild_id"`
	FreshTime      int         `json:"fresh_time"`
	MemberVitality map[int]int `json:"member_vitality"`
}

type TblGuildDailyVitality struct {
	Id        int    `xorm:"int notnull autoincr pk 'id'"`
	Uin       int    `xorm:"int notnull default(0) index 'uin'"`
	FreshTime int    `xorm:"int notnull default(0) 'fresh_time'"`
	Vitality  int    `xorm:"int notnull default(0) 'vitality'"`
	Reserved0 int    `xorm:"int default(0) 'reserved_0'"`
	Reserved1 int    `xorm:"int default(0) 'reserved_1'"`
	Reserved2 string `xorm:"varchar(128) default('') 'reserved_2'"`
	Reserved3 string `xorm:"varchar(1024) default('') 'reserved_3'"`
	Reserved4 string `xorm:"varchar(4096) default('') 'reserved_4'"`
	Reserved5 string `xorm:"text default('') 'reserved_5'"`
}

type TblGuildDailyVitalityReward struct {
	Id        int `xorm:"int notnull autoincr pk 'id'"`
	Uin       int `xorm:"int notnull default(0) index 'uin'"`
	ProtypeId int `xorm:"int notnull default(0) 'protype_id'"`
	Status    int `xorm:"int notnull default(0) 'status'"`
}

type TblGuildSalvage struct {
	Id                int    `xorm:"int notnull autoincr pk 'id'"`
	Uin               int    `xorm:"int notnull default(0) index 'uin'"`
	RestTimes         int    `xorm:"int notnull default(0) 'rest_times'"`
	LastNotEnoughTime int    `xorm:"int notnull default(0) 'last_not_enough_time'"`
	Reserved0         int    `xorm:"int default(0) 'reserved_0'"`
	Reserved1         int    `xorm:"int default(0) 'reserved_1'"`
	Reserved2         string `xorm:"varchar(128) default('') 'reserved_2'"`
	Reserved3         string `xorm:"varchar(1024) default('') 'reserved_3'"`
	Reserved4         string `xorm:"varchar(4096) default('') 'reserved_4'"`
	Reserved5         string `xorm:"text default('') 'reserved_5'"`
}

type TblGuildAncientRelicsPiece struct {
	PiecesId []int `json:"pieces_id"`
}

type TblGuildAncientRelicsInfo struct {
	Id        int                        `xorm:"int notnull autoincr pk 'id'"`
	Uin       int                        `xorm:"int notnull default(0) index 'uin'"`
	ProtypeId int                        `xorm:"int notnull default(0) index 'protype_id'"`
	Status    int                        `xorm:"int notnull default(0) 'status'"`
	Pieces    TblGuildAncientRelicsPiece `xorm:"jsonb default('') 'pieces'"`
	Reserved0 int                        `xorm:"int default(0) 'reserved_0'"`
	Reserved1 int                        `xorm:"int default(0) 'reserved_1'"`
	Reserved2 string                     `xorm:"varchar(128) default('') 'reserved_2'"`
	Reserved3 string                     `xorm:"varchar(1024) default('') 'reserved_3'"`
	Reserved4 string                     `xorm:"varchar(4096) default('') 'reserved_4'"`
	Reserved5 string                     `xorm:"text default('') 'reserved_5'"`
}

type TblPropInfo struct {
	Id        int    `xorm:"int notnull autoincr pk 'id'"`
	Uin       int    `xorm:"int notnull index 'uin'"`
	ProtypeId int    `xorm:"int notnull index 'protype_id'"`
	PropNum   int    `xorm:"int notnull 'prop_num'"`
	Reserved0 int    `xorm:"int default(0) 'reserved_0'"`
	Reserved1 int    `xorm:"int default(0) 'reserved_1'"`
	Reserved2 string `xorm:"varchar(128) default('') 'reserved_2'"`
	Reserved3 string `xorm:"varchar(1024) default('') 'reserved_3'"`
	Reserved4 string `xorm:"varchar(4096) default('') 'reserved_4'"`
	Reserved5 string `xorm:"text default('') 'reserved_5'"`
}

type TblLeagueSeasonInfo struct {
	StartTime       int `xorm:"'leagueseasonstarttime'"`
	AccountFlag     int `xorm:"'leagueseasonaccountflag'"`
	CurrentSeasonId int `xorm:"'currentleagueseasonid'"`
	AcountUserCount int `xorm:"'leagueseasonaccountusercount'"`
}

type TblGuildWar struct {
	StartTime     int `xorm:"'guild_war_start_time'"`   // 公会战开始时间，如果下半场已经结束，表示下一次公会战上半场开始时间
	GuildWardId   int `xorm:"'guild_ward_id'"`          // 公会战Id，如果下半场已经结束，表示下一次公会战的Id
	Phase         int `xorm:"'guild_war_phase'"`        // 公会战阶段, 0: 公会战没有开始，1：上半场进行中，2：上半场已结束，3：下半场进行中，4：下半场已结束，5：结算中
	RewardGroupId int `xorm:"'guild_war_reward_group'"` // 公会战奖励的组Id
}

type TblGuildWarInfo struct {
	Creator    int `xorm:"guild_creator"`
	GuildId    int `xorm:"guild_id"`
	WinTimes   int `xorm:"win_times"`
	LoseTimes  int `xorm:"lose_times"`
	TotalTimes int `xorm:"total_times"`
	UserCount  int `xorm:"user_count"`
	Score      int `xorm:"total_score"`
}

type TblGuildWarMemberInfo struct {
	GuildId          string                        `json:"guild_id"`
	Uin              int                           `json:"uin"`
	Score            int                           `json:"score"`
	RestBattleTimes  int                           `json:"rest_battle_times"`
	BattleRecordList []*TblGuildMemberBattleRecord `json:"battle_record_list"`
}

type TblGuildMemberBattleRecord struct {
	ScoreIncr    int `json:"score_incr"`
	BattleResult int `json:"battle_result"`
}

type TblLeagueInfo struct {
	Uin                       int    `xorm:"int notnull pk 'uin'"`
	CurrentLeagueLevel        int    `xorm:"int notnull default(0) 'current_league_level'"`
	BestLeagueLevel           int    `xorm:"int notnull default(0) 'best_league_level'"`
	LegendImpressCount        int    `xorm:"int notnull default(0) 'legend_impress_count'"`
	LastAccountLeagueSeasonId int    `xorm:"int notnull default(0) 'last_account_league_season_id'"`
	Reserved0                 int    `xorm:"int default(0) 'reserved_0'"`
	Reserved1                 int    `xorm:"int default(0) 'reserved_1'"`
	Reserved2                 string `xorm:"varchar(128) default('') 'reserved_2'"`
	Reserved3                 string `xorm:"varchar(1024) default('') 'reserved_3'"`
	Reserved4                 string `xorm:"varchar(4096) default('') 'reserved_4'"`
	Reserved5                 string `xorm:"text default('') 'reserved_5'"`
}

type TblGuildFrigateShip struct {
	Id           int    `xorm:"int notnull pk autoincr 'id'"`
	GuildId      int    `xorm:"int notnull default(0) index 'guild_id'"`
	GuildCreator int    `xorm:"int notnull default(0) index 'guild_creator'"`
	ProtypeId    int    `xorm:"int notnull default(0) 'protype_id'"`
	Level        int    `xorm:"int  default(1) 'level'"`
	Exp          int    `xorm:"int default(0) 'exp'"`
	Reserved0    int    `xorm:"int default(0) 'reserved_0'"`
	Reserved1    int    `xorm:"int default(0) 'reserved_1'"`
	Reserved2    string `xorm:"varchar(128) default('') 'reserved_2'"`
	Reserved3    string `xorm:"varchar(1024) default('') 'reserved_3'"`
	Reserved4    string `xorm:"varchar(4096) default('') 'reserved_4'"`
	Reserved5    string `xorm:"text default('') 'reserved_5'"`
}

type TblUserHeadFrame struct {
	Uin              int    `xorm:"int notnull pk 'uin'"`
	CurHeadFrame     int    `xorm:"int default(0) 'cur_head_frame'"`
	DefaultHeadFrame int    `xorm:"int default(0) 'default_head_frame'"`
	HeadId           string `xorm:"varchar(256) default('') 'head_id'"`
	HeadType         int    `xorm:"int default(0) 'head_type'"`
	Reserved0        int    `xorm:"int default(0) 'reserved_0'"`
	Reserved1        int    `xorm:"int default(0) 'reserved_1'"`
	Reserved2        string `xorm:"varchar(128) default('') 'reserved_2'"`
	Reserved3        string `xorm:"varchar(1024) default('') 'reserved_3'"`
	Reserved4        string `xorm:"varchar(4096) default('') 'reserved_4'"`
	Reserved5        string `xorm:"text default('') 'reserved_5'"`
}

type TblChatMessage struct {
	Uin      int    `xorm:"'uin'"`
	Message  string `xorm:"'message'"`
	SendTime int    `xorm:"'send_time'"`
	Type     string `xorm:"'type'"`
}

func init() {
	battleShip := new(TblBattleShip)
	mysql.RegisterTableObj(battleShip)

	userInfo := new(TblUserInfo)
	mysql.RegisterTableObj(userInfo)

	campaignPassChapter := new(TblCampaignPassChapter)
	mysql.RegisterTableObj(campaignPassChapter)

	campaignEvent := new(TblCampaignEvent)
	mysql.RegisterTableObj(campaignEvent)

	campaignEventFresh := new(TblCampaignEventFresh)
	mysql.RegisterTableObj(campaignEventFresh)

	campaignProduceResource := new(TblCampaignProduceResource)
	mysql.RegisterTableObj(campaignProduceResource)

	campaignEventStatistics := new(TblCampaignEventStatistics)
	mysql.RegisterTableObj(campaignEventStatistics)

	achievement := new(TblAchievement)
	mysql.RegisterTableObj(achievement)

	activityTask := new(TblActivityTask)
	mysql.RegisterTableObj(activityTask)

	activityTaskFresh := new(TblActivityTaskFresh)
	mysql.RegisterTableObj(activityTaskFresh)

	activityScoreReward := new(TblActivityScoreReward)
	mysql.RegisterTableObj(activityScoreReward)

	growupTask := new(TblGrowupTask)
	mysql.RegisterTableObj(growupTask)

	guildInfo := new(TblGuildInfo)
	mysql.RegisterTableObj(guildInfo)

	guildMemberInfo := new(TblGuildMemberInfo)
	mysql.RegisterTableObj(guildMemberInfo)

	guildFrigateShipInfo := new(TblGuildFrigateShip)
	mysql.RegisterTableObj(guildFrigateShipInfo)
}
