package model

import (
	"fmt"
	"sailcraft/fleetsvr_main/custom_errors"
	"sailcraft/fleetsvr_main/errorcode"
)

const (
	CACHED_KEY_GUILD_WAR_RANK = "guildwarrank"
)

type GuildWarPointsModel struct {
	WarId int
}

func (this *GuildWarPointsModel) CachedKey() string {
	return fmt.Sprintf("%s.%d", CACHED_KEY_GUILD_WAR_RANK, this.WarId)
}

func (this *GuildWarPointsModel) GetGuildWarPoints(creator, gid int) (int, error) {
	redisKey := this.CachedKey()
	redisMgr := GetSingletonRedis()
	if redisMgr == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("redisMgr is nil")
	}

	score, err := redisMgr.ZScore(redisKey, FormatGuildId(creator, gid))
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	return score, nil
}

func (this *GuildWarPointsModel) IncrGuildWarPoints(creator, gId, pointsIncr int) (int, error) {
	redisKey := this.CachedKey()
	redisMgr := GetSingletonRedis()
	if redisMgr == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("redisMgr is nil")
	}

	newScore, err := redisMgr.ZIncrBy(redisKey, FormatGuildId(creator, gId), pointsIncr)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	redisMgr.Expire(redisKey, EXPIRE_TIME)

	return newScore, nil
}
