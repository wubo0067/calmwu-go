package model

import (
	"fmt"
	"sailcraft/base"
	"sailcraft/dataaccess/mysql"
	"sailcraft/dataaccess/redistool"
	"sailcraft/fleetsvr_main/custom_errors"
	"sailcraft/fleetsvr_main/errorcode"
	"sailcraft/fleetsvr_main/table"
	"sailcraft/fleetsvr_main/utils"
	"strconv"
	"strings"

	"github.com/go-xorm/core"
)

const (
	USER_INFO_TABLE_NAME       = "user_info"
	USER_INFO_REDIS_CACHED_KEY = ".cache."

	USER_INFO_TABLE_ATTR_LEVEL             = "level"
	USER_INFO_TABLE_ATTR_EXP               = "exp"
	USER_INFO_TABLE_ATTR_STAR              = "star"
	USER_INFO_TABLE_ATTR_GOLD              = "gold"
	USER_INFO_TABLE_ATTR_MINERAl           = "mineral"
	USER_INFO_TABLE_ATTR_WOOD              = "wood"
	USER_INFO_TABLE_ATTR_GEM               = "gem"
	USER_INFO_TABLE_ATTR_PURCHASE_GEM      = "purchase_gem"
	USER_INFO_TABLE_ATTR_STONE             = "stone"
	USER_INFO_TABLE_ATTR_IRON              = "iron"
	USER_INFO_TABLE_ATTR_CHANGE_NAME_COUNT = "change_name_count"
	USER_INFO_TABLE_ATTR_OCEAN_DUST        = "ocean_dust"
	USER_INFO_TABLE_ATTR_GUILD_ID          = "guild_id"
	USER_INFO_TABLE_ATTR_VITALITY          = "vitality"
)

type UserInfoModel struct {
	Uin int
}

func (userInfoModel *UserInfoModel) TableName() string {
	index := GetTableSplitIndex(userInfoModel.Uin)
	return fmt.Sprintf("%s_%d", USER_INFO_TABLE_NAME, index)
}

func (userInfoModel *UserInfoModel) CachedKey() string {
	return fmt.Sprintf("%s%s%d", USER_INFO_TABLE_NAME, USER_INFO_REDIS_CACHED_KEY, userInfoModel.Uin)
}

func (userInfoModel *UserInfoModel) GetUserInfo(userInfo *table.TblUserInfo) (int, error) {
	if userInfo == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("null point")
	}

	if userInfoModel.Uin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("uin is invalid")
	}

	engine := GetUinSetMysql(userInfoModel.Uin)
	if engine == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("database engine is nil")
	}

	tableName := userInfoModel.TableName()
	condtion := fmt.Sprintf("uin=%d", userInfoModel.Uin)

	exist, err := mysql.GetRecordByCond(engine, tableName, condtion, userInfo)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if exist {
		return 1, nil
	} else {
		return 0, nil
	}
}

// 这个是结构体全字段更新，如果只更新部分字段，请用第二个接口
func (userInfoModel *UserInfoModel) UpdateUserInfo(userInfo *table.TblUserInfo) (int, error) {
	if userInfo == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("null point")
	}

	if userInfoModel.Uin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("uin is invalid")
	}

	engine := GetUinSetMysql(userInfoModel.Uin)
	if engine == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("database engine is nil")
	}

	tableName := userInfoModel.TableName()

	PK := core.NewPK(userInfoModel.Uin)
	_, err := mysql.UpdateRecord(engine, tableName, PK, userInfo)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	// 更新完数据库，需要更新redis
	redisKey := userInfoModel.CachedKey()
	redisMgr := GetClusterRedis(redisKey)
	if redisMgr != nil {
		redisMap, err := redistool.ConvertObjToRedisHash(userInfo)
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

func (userInfoModel *UserInfoModel) UpdateUserInfoChanged(attrMap map[string]interface{}) (int, error) {
	if attrMap == nil || len(attrMap) == 0 {
		base.GLog.Debug("UpdateUserInfoChanged input params is empty")
		return 0, nil
	}

	if userInfoModel.Uin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("uin is invalid")
	}

	engine := GetUinSetMysql(userInfoModel.Uin)
	if engine == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("database engine is nil")
	}

	tableName := userInfoModel.TableName()

	cond := fmt.Sprintf("uin=%d", userInfoModel.Uin)
	_, err := mysql.UpdateRecordSpecifiedFieldsByCond(engine, tableName, cond, attrMap)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	// 更新完数据库，需要更新redis
	redisKey := userInfoModel.CachedKey()
	redisMgr := GetClusterRedis(redisKey)
	if redisMgr != nil {
		redisHashMap, err := utils.ConvertMapInterfaceToMapString(attrMap)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		err = redisMgr.HashSet(redisKey, redisHashMap)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}
	}

	return 0, nil
}

func GetMultiUserInfo(uinSlice ...int) ([]*table.TblUserInfo, error) {
	userInfoModel := UserInfoModel{Uin: 0}

	keys := make([]string, 0, len(uinSlice))
	for _, uin := range uinSlice {
		if uin <= 0 {
			return nil, custom_errors.New("uin is invalid")
		}

		userInfoModel.Uin = uin
		keys = append(keys, userInfoModel.CachedKey())
	}

	groupKeys, err := GetKeysGroupByClusterRedis(keys...)
	if err != nil {
		return nil, err
	}

	userArr := make([]*table.TblUserInfo, 0, len(uinSlice))
	for _, single := range groupKeys {
		redisPipeLine := redistool.NewRedisPipeLine()
		for _, key := range single.Keys {
			redisPipeLine.Append(redistool.REDIS_CONTAINER_HASH, "HGETALL", key)
		}

		results, err := redisPipeLine.Run(single.RedisMgr)
		if err != nil {
			return nil, err
		}

		base.GLog.Debug("Pipeline Results Len: %d\n", len(results))
		for _, res := range results {
			hashV, err := res.Hash()
			if err != nil {
				return nil, err
			}

			userInfo := new(table.TblUserInfo)
			err = redistool.ConvertRedisHashToObj(hashV, userInfo)
			if err != nil {
				return nil, err
			}

			userArr = append(userArr, userInfo)
		}
	}

	return userArr, nil
}

func ValidGuildId(guildId string) bool {
	_, _, ok := ConvertGuildIdToUinAndId(guildId)
	return ok
}

func FormatGuildId(creatorUin int, id int) string {
	return fmt.Sprintf("%dg%d", creatorUin, id)
}

func ConvertGuildIdToUinAndId(guildId string) (int, int, bool) {
	attrs := strings.Split(guildId, "g")

	if len(attrs) != 2 {
		return 0, 0, false
	}

	creator, err := strconv.Atoi(attrs[0])
	if err != nil || creator <= 0 {
		return 0, 0, false
	}

	id, err := strconv.Atoi(attrs[1])
	if err != nil && id <= 0 {
		return 0, 0, false
	}

	return creator, id, true
}
