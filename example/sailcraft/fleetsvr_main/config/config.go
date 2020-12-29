package config

const (
	PATH = "assets/Data/"

	GLOBAL_CONFIG_PATH = "Common/GlobalConfig.json"

	BATTLE_SHIP_PROTYPE_CONFIG_PATH   = "Ship/BattleShipProtype.json"
	BATTLE_SHIP_UPGRADE_CONFIG_PATH   = "Ship/BattleShipUpgrade.json"
	BATTLE_SHIP_STENGTHEN_CONFIG_PATH = "Ship/BattleShipStrengthen.json"
	BATTLE_SHIP_RECLAIM_CONFIG_PATH   = "Ship/ShipReclaim.json"

	RESOURCE_GEM_EXCHANGE_CONFIG_PATH = "Shop/ResourceGemExchangeConfig.json"

	CAMPAIGN_PROTYPE_PATH       = "PVE/Campaign.json"
	CAMPAIGN_AREA_PROTYPE_PATH  = "PVE/CampaignArea.json"
	CAMPAIGN_EVENT_PROTYPE_PATH = "PVE/CampaignEvent.json"
	CAMPAIGN_MISSON_PATH        = "PVE/CampaignMission.json"

	MISSION_TYPE_PATH = "Achievement/MissionType.json"

	ACHIEVEMENT_PATH = "Achievement/Achievement.json"

	ACTIVITY_TASK_PATH         = "Task/ActivityTask.json"
	DAILY_ACTIVITY_REWARD_PATH = "Task/DailyActivityReward.json"
	GROW_UP_TASK_PATH          = "Task/GrowupTask.json"

	USER_LEVEL_EXP_CONFIG_PATH = "Exp/LevelExp.json"

	GUILD_LEVEL_CONFIG_PATH        = "Guild/GuildLevel.json"
	GUILD_TASK_CONFIG_PATH         = "Guild/GuildTask.json"
	GUILD_TASK_REWARD_CONFIG_PATH  = "Guild/GuildTaskReward.json"
	GUILD_SALVAGE_CONFIG_PATH      = "Guild/Salvage.json"
	GUILD_SALVAGE_POOL_CONFIG_PATH = "Guild/SalvagePool.json"
	GUILD_EXCHANGE_CONFIG_PATH     = "Guild/Exchange.json"

	FRIGATE_GUILD_CONFIG_PATH  = "Frigate/GuildFrigate.json"
	FRIGATE_WEAPON_CONFIG_PATH = "Frigate/Weapon.json"

	CHAT_CONFIG_PATH       = "Common/Chat.json"
	ICON_FRAME_CONFIG_PATH = "Common/IconFrame.json"

	PROP_CONFIG_PATH = "Props/Props.json"
)

const (
	SHOP_ITEM_TYPE_CARDPACK = "Cardpack"
)

func Initialize(filePath string) error {
	strFilePath := filePath + PATH

	err := GBattleShipProtypeConfig.Init(strFilePath + BATTLE_SHIP_PROTYPE_CONFIG_PATH)
	if err != nil {
		return err
	}

	err = GBattleShipUpgradeConfig.Init(strFilePath + BATTLE_SHIP_UPGRADE_CONFIG_PATH)
	if err != nil {
		return err
	}

	err = GBattleShipStrengthenConfig.Init(strFilePath + BATTLE_SHIP_STENGTHEN_CONFIG_PATH)
	if err != nil {
		return err
	}

	err = GResourceGemExchangeConfig.Init(strFilePath + RESOURCE_GEM_EXCHANGE_CONFIG_PATH)
	if err != nil {
		return err
	}

	err = GLevelExpConfig.Init(strFilePath + USER_LEVEL_EXP_CONFIG_PATH)
	if err != nil {
		return err
	}

	err = GCampaignConfig.Init(strFilePath + CAMPAIGN_PROTYPE_PATH)
	if err != nil {
		return err
	}

	err = GCampaignAreaConfig.Init(strFilePath + CAMPAIGN_AREA_PROTYPE_PATH)
	if err != nil {
		return err
	}

	err = GCampaignEventConfig.Init(strFilePath + CAMPAIGN_EVENT_PROTYPE_PATH)
	if err != nil {
		return err
	}

	err = GCampaignMissionConfig.Init(strFilePath + CAMPAIGN_MISSON_PATH)
	if err != nil {
		return err
	}

	err = GGlobalConfig.Init(strFilePath + GLOBAL_CONFIG_PATH)
	if err != nil {
		return err
	}

	err = GAchievementConfig.Init(strFilePath + ACHIEVEMENT_PATH)
	if err != nil {
		return err
	}

	err = GActivityTaskConfig.Init(strFilePath + ACTIVITY_TASK_PATH)
	if err != nil {
		return err
	}

	err = GActivityScoreRewardConfig.Init(strFilePath + DAILY_ACTIVITY_REWARD_PATH)
	if err != nil {
		return err
	}

	err = GGrowupTaskConfig.Init(strFilePath + GROW_UP_TASK_PATH)
	if err != nil {
		return err
	}

	err = GGuildLevelConfig.Init(strFilePath + GUILD_LEVEL_CONFIG_PATH)
	if err != nil {
		return err
	}

	err = GGuildTaskConfig.Init(strFilePath + GUILD_TASK_CONFIG_PATH)
	if err != nil {
		return err
	}

	err = GGuildTaskRewardConfig.Init(strFilePath + GUILD_TASK_REWARD_CONFIG_PATH)
	if err != nil {
		return err
	}

	err = GGuildSalvageConfig.Init(strFilePath + GUILD_SALVAGE_CONFIG_PATH)
	if err != nil {
		return err
	}

	err = GGuildSalvagePoolConfig.Init(strFilePath + GUILD_SALVAGE_POOL_CONFIG_PATH)
	if err != nil {
		return err
	}

	err = GAncientRelicsConfig.Init(strFilePath + GUILD_EXCHANGE_CONFIG_PATH)
	if err != nil {
		return err
	}

	err = GPropConfig.Init(strFilePath + PROP_CONFIG_PATH)
	if err != nil {
		return err
	}

	err = GGuildFrigateConfig.Init(strFilePath + FRIGATE_GUILD_CONFIG_PATH)
	if err != nil {
		return err
	}

	err = GFrigateWeaponConfig.Init(strFilePath + FRIGATE_WEAPON_CONFIG_PATH)
	if err != nil {
		return err
	}

	err = GChatMessageConfig.Init(strFilePath + CHAT_CONFIG_PATH)
	if err != nil {
		return err
	}

	err = GIconFrameConfig.Init(strFilePath + ICON_FRAME_CONFIG_PATH)
	if err != nil {
		return err
	}

	err = GShipReclaimConfig.Init(strFilePath + BATTLE_SHIP_RECLAIM_CONFIG_PATH)
	if err != nil {
		return err
	}

	return nil
}
