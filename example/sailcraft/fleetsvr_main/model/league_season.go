package model

import (
	"sailcraft/dataaccess/redistool"
	"sailcraft/fleetsvr_main/custom_errors"
	"sailcraft/fleetsvr_main/table"
)

const (
	TABLE_NAME_LEAGUE_SEASON = "leagueseason"

	LEAGUE_SEASON_FLAG_NORMAL     = 0
	LEAGUE_SEASON_FLAG_SETTLEMENT = 1
)

type LeagueSeasonModel struct {
}

func (this *LeagueSeasonModel) CachedKey() string {
	return TABLE_NAME_LEAGUE_SEASON
}

func (this *LeagueSeasonModel) GetLeagueSeasonInfo() (*table.TblLeagueSeasonInfo, error) {
	redisKey := this.CachedKey()
	redisMgr := GetSingletonRedis()
	if redisMgr == nil {
		return nil, custom_errors.New("redisMgr is nil")
	}

	leagueSeasonMapStr, err := redisMgr.HashGet(redisKey)
	if err != nil {
		return nil, err
	}

	var record table.TblLeagueSeasonInfo
	err = redistool.ConvertRedisHashToObj(leagueSeasonMapStr, &record)
	if err != nil {
		return nil, err
	}

	return &record, nil
}
