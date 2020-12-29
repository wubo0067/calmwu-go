package model

import (
	"fmt"
	"sailcraft/base"
	"sailcraft/dataaccess/mysql"
	"sailcraft/dataaccess/redistool"
	"sailcraft/fleetsvr_main/custom_errors"
	"sailcraft/fleetsvr_main/table"
)

type LeagueModel struct {
	Uin int
}

const (
	TABLE_NAME_LEAGUE       = "league"
	LEAGUE_REDIS_CACHED_KEY = "cache"
)

func (this *LeagueModel) TableName() string {
	index := GetTableSplitIndex(this.Uin)
	return fmt.Sprintf("%s_%d", TABLE_NAME_LEAGUE, index)
}

func (this *LeagueModel) CachedKey() string {
	return fmt.Sprintf("%s.%s.%d", TABLE_NAME_LEAGUE, LEAGUE_REDIS_CACHED_KEY, this.Uin)
}

func (this *LeagueModel) GetLeagueInfo() (*table.TblLeagueInfo, error) {
	if this.Uin <= 0 {
		return nil, custom_errors.New("uin is invalid")
	}

	engine := GetUinSetMysql(this.Uin)
	if engine == nil {
		return nil, custom_errors.New("engine is nil")
	}

	tableName := this.TableName()

	base.GLog.Debug("database is [%s] table is [%s]", engine.DBEngineDataSouceName, tableName)

	cond := fmt.Sprintf("uin=%d", this.Uin)

	records := make([]*table.TblLeagueInfo, 0)
	err := mysql.FindRecordsBySimpleCond(engine, tableName, cond, 0, 0, &records)
	if err != nil {
		return nil, err
	}

	if len(records) > 0 {
		return records[0], nil
	}

	return nil, nil
}

func GetMultiLeagueInfo(uinSlice ...int) ([]*table.TblLeagueInfo, error) {
	leagueInfoModel := LeagueModel{Uin: 0}

	keys := make([]string, 0, len(uinSlice))
	for _, uin := range uinSlice {
		if uin <= 0 {
			return nil, custom_errors.New("uin is invalid")
		}

		leagueInfoModel.Uin = uin
		keys = append(keys, leagueInfoModel.CachedKey())
	}

	groupKeys, err := GetKeysGroupByClusterRedis(keys...)
	if err != nil {
		return nil, err
	}

	leagueArr := make([]*table.TblLeagueInfo, 0, len(uinSlice))
	for _, single := range groupKeys {
		redisPipeLine := redistool.NewRedisPipeLine()
		for _, key := range single.Keys {
			redisPipeLine.Append(redistool.REDIS_CONTAINER_HASH, "HGETALL", key)
		}

		results, err := redisPipeLine.Run(single.RedisMgr)
		if err != nil {
			return nil, err
		}

		for _, res := range results {
			hashV, err := res.Hash()
			if err != nil {
				return nil, err
			}

			leagueInfo := new(table.TblLeagueInfo)
			err = redistool.ConvertRedisHashToObj(hashV, leagueInfo)
			if err != nil {
				return nil, err
			}

			leagueArr = append(leagueArr, leagueInfo)
		}
	}

	return leagueArr, nil
}
