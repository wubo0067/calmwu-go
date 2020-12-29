package model

import (
	"fmt"
	"sailcraft/base"
	"sailcraft/dataaccess/mysql"
	"sailcraft/dataaccess/redistool"
	"sailcraft/fleetsvr_main/custom_errors"
	"sailcraft/fleetsvr_main/errorcode"
	"sailcraft/fleetsvr_main/table"

	"github.com/go-xorm/core"
)

const (
	GUILD_INFO_TABLE_NAME       = "guild_info"
	GUILD_INFO_REDIS_CACHED_KEY = "cache"
)

type GuildInfoModel struct {
	Creator int
	Id      int
}

func (this *GuildInfoModel) TableName() string {
	index := GetTableSplitIndex(this.Creator)
	return fmt.Sprintf("%s_%d", GUILD_INFO_TABLE_NAME, index)
}

func (this *GuildInfoModel) CachedKey() string {
	return fmt.Sprintf("%s.%s.%s", GUILD_INFO_TABLE_NAME, GUILD_INFO_REDIS_CACHED_KEY, FormatGuildId(this.Creator, this.Id))
}

func (this *GuildInfoModel) AddNewGuild(record *table.TblGuildInfo) (int, error) {
	if record == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("null point")
	}

	if record.Creator <= 0 || this.Creator != record.Creator {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("creator is invalid")
	}

	engine := GetUinSetMysql(record.Creator)
	if engine == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("database engine is nil")
	}

	tableName := this.TableName()
	base.GLog.Debug("database is [%s] table is [%s]", engine.DBEngineDataSouceName, tableName)

	_, err := mysql.InsertRecord(engine, tableName, record)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	this.Id = record.Id
	redisKey := this.CachedKey()
	redisMgr := GetClusterRedis(redisKey)
	if redisMgr != nil {
		redisMap, err := redistool.ConvertObjToRedisHash(record)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		err = redisMgr.HashSet(redisKey, redisMap)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}
	}

	return 0, nil
}

func (this *GuildInfoModel) GetGuildInfo() (*table.TblGuildInfo, error) {
	if this.Creator <= 0 {
		return nil, custom_errors.New("creator is invalid")
	}

	if this.Id <= 0 {
		return nil, custom_errors.New("guild id is invalid")
	}

	engine := GetUinSetMysql(this.Creator)
	if engine == nil {
		return nil, custom_errors.New("database engine is nil")
	}

	tableName := this.TableName()
	base.GLog.Debug("database is [%s] table is [%s]", engine.DBEngineDataSouceName, tableName)

	cond := fmt.Sprintf("id=%d", this.Id)

	records := make([]*table.TblGuildInfo, 0)
	err := mysql.FindRecordsBySimpleCond(engine, tableName, cond, 0, 0, &records)
	if err != nil {
		return nil, err
	}

	if len(records) > 0 {
		return records[0], nil
	}

	return nil, nil
}

func (this *GuildInfoModel) UpdateGuildInfo(record *table.TblGuildInfo) (int, error) {
	if record == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("null point")
	}

	if this.Creator <= 0 || this.Creator != record.Creator {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("creator is invalid")
	}

	if this.Id <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("guild id is invalid")
	}

	engine := GetUinSetMysql(record.Creator)
	if engine == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("database engine is nil")
	}

	tableName := this.TableName()
	base.GLog.Debug("database is [%s] table is [%s]", engine.DBEngineDataSouceName, tableName)

	pk := core.NewPK(record.Id)
	_, err := mysql.UpdateRecord(engine, tableName, pk, record)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	redisKey := this.CachedKey()
	redisMgr := GetClusterRedis(redisKey)
	if redisMgr != nil {
		redisMap, err := redistool.ConvertObjToRedisHash(record)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		err = redisMgr.HashSet(redisKey, redisMap)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}
	}

	return 0, nil
}

func (this *GuildInfoModel) DeleteGuildInfo(record *table.TblGuildInfo) (int, error) {
	if record == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("null point")
	}

	if this.Creator <= 0 || this.Creator != record.Creator {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("guild creator is invalid")
	}

	if this.Id <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("guild id is invalid")
	}

	engine := GetUinSetMysql(this.Creator)
	if engine == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("database engine is nil")
	}

	tableName := this.TableName()
	base.GLog.Debug("database is [%s] table is [%s]", engine.DBEngineDataSouceName, tableName)

	pk := core.NewPK(record.Id)
	_, err := mysql.DeleteRecord(engine, tableName, pk, record)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	redisKey := this.CachedKey()
	redisMgr := GetClusterRedis(redisKey)
	if redisMgr != nil {
		err = redisMgr.DelKey(redisKey)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}
	}

	return 0, nil
}

func GetMultiGuildInfo(guildIdSlice ...string) ([]*table.TblGuildInfo, error) {
	guildInfoModel := GuildInfoModel{}

	keys := make([]string, 0, len(guildIdSlice))
	for _, guildId := range guildIdSlice {
		creator, id, ok := ConvertGuildIdToUinAndId(guildId)
		if !ok {
			continue
		}

		guildInfoModel.Creator = creator
		guildInfoModel.Id = id
		keys = append(keys, guildInfoModel.CachedKey())
	}

	groupKeys, err := GetKeysGroupByClusterRedis(keys...)
	if err != nil {
		return nil, err
	}

	guildInfoList := make([]*table.TblGuildInfo, 0, len(guildIdSlice))
	for _, group := range groupKeys {
		redisPipeLine := redistool.NewRedisPipeLine()
		for _, key := range group.Keys {
			redisPipeLine.Append(redistool.REDIS_CONTAINER_HASH, "HGETALL", key)
		}

		results, err := redisPipeLine.Run(group.RedisMgr)
		if err != nil {
			return nil, err
		}

		base.GLog.Debug("Pipeline Results Len: %d\n", len(results))
		for _, res := range results {
			hashV, err := res.Hash()
			if err != nil {
				return nil, err
			}

			if hashV == nil || len(hashV) <= 0 {
				continue
			}

			guildInfo := new(table.TblGuildInfo)
			err = redistool.ConvertRedisHashToObj(hashV, guildInfo)
			if err != nil {
				return nil, err
			}

			guildInfoList = append(guildInfoList, guildInfo)
		}
	}

	return guildInfoList, nil
}

func GenerateGuildPerformId(serverId int) (int, error) {
	if serverId <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("zone id is invalid")
	}

	redisKey := fmt.Sprintf("guild.id.server.%d", serverId)
	redisMgr := GetClusterRedis(redisKey)
	if redisMgr == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("redisMgr is nil")
	}

	id, err := redisMgr.Incr(redisKey)
	if err != nil {
		return 0, err
	}

	sId := 311 + id
	zs := 10000
	for {
		if sId >= zs {
			zs *= 10
		} else {
			break
		}
	}

	guildId := (4211+serverId)*zs + sId

	return guildId, nil
}
