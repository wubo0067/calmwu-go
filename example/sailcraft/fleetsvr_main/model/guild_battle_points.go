package model

import (
	"sailcraft/fleetsvr_main/custom_errors"
	"sailcraft/fleetsvr_main/errorcode"
)

const (
	TABLE_NAME_GUILD_BATTLE_POINTS = "guild_battle.points"
)

type GuildBattlePointsModel struct {
}

func (this *GuildBattlePointsModel) CachedKey() string {
	return TABLE_NAME_GUILD_BATTLE_POINTS
}

func (this *GuildBattlePointsModel) Query(guildId string) (points int, err error) {
	points = 0
	err = nil

	if !ValidGuildId(guildId) {
		err = custom_errors.New("guild id format error")
		return
	}

	redisKey := this.CachedKey()
	redisMgr := GetSingletonRedis()
	if redisMgr == nil {
		err = custom_errors.New("engine is nil")
		return
	}

	points, err = redisMgr.ZScore(redisKey, guildId)
	if err != nil {
		return
	}

	return
}

func (this *GuildBattlePointsModel) IncreasePoints(guildId string, incr int) (points int, err error) {
	points = 0
	err = nil

	redisKey := this.CachedKey()
	redisMgr := GetSingletonRedis()
	if redisMgr == nil {
		err = custom_errors.New("engine is nil")
	}

	points, err = redisMgr.ZIncrBy(redisKey, guildId, incr)
	if err != nil {
		return
	}

	return
}

func (this *GuildBattlePointsModel) Delete(guildId string) (int, error) {
	redisKey := this.CachedKey()
	redisMgr := GetSingletonRedis()
	if redisMgr == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("redisMgr is nil")
	}

	_, err := redisMgr.ZRem(redisKey, guildId)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	return 0, nil
}
