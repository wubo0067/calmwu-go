package model

import (
	"fmt"
	"sailcraft/fleetsvr_main/custom_errors"
	"sailcraft/fleetsvr_main/errorcode"
)

const (
	TABLE_NAME_GUILD_VITALITY = "guildvitality.rank"
)

type GuildVitalityModel struct {
}

func (this *GuildVitalityModel) CachedKey() string {
	return fmt.Sprintf("%s", TABLE_NAME_GUILD_VITALITY)
}

func (this *GuildVitalityModel) GetGuildVitality(guildId string) (int, error) {
	redisKey := this.CachedKey()
	redisMgr := GetSingletonRedis()
	if redisMgr == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("engine is nil")
	}

	vitality, err := redisMgr.ZScore(redisKey, guildId)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	return vitality, nil
}

func (this *GuildVitalityModel) IncreaseGuildVitality(guildId string, vitalityIncr int) (int, error) {
	redisKey := this.CachedKey()
	redisMgr := GetSingletonRedis()
	if redisMgr == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("engine is nil")
	}

	newVitality, err := redisMgr.ZIncrBy(redisKey, guildId, vitalityIncr)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	return newVitality, nil
}

func (this *GuildVitalityModel) DeleteGuildVitality(guildId string) (int, error) {
	redisKey := this.CachedKey()
	redisMgr := GetSingletonRedis()
	if redisMgr == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("engine is nil")
	}

	_, err := redisMgr.ZRem(redisKey, guildId)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	return 0, nil
}

func (this *GuildVitalityModel) GetRank(guildId string) (int, error) {
	redisKey := this.CachedKey()
	redisMgr := GetSingletonRedis()
	if redisMgr == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("engine is nil")
	}

	rank, err := redisMgr.ZReverseRank(redisKey, guildId)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	return rank, nil
}
