package model

import (
	"fmt"
	"sailcraft/base"
	"sailcraft/dataaccess/mysql"
	"sailcraft/dataaccess/redistool"
	"sailcraft/fleetsvr_main/custom_errors"
	"sailcraft/fleetsvr_main/table"
	"strconv"
)

const (
	TABLE_NAME_USER_HEAD_FRAME = "user_head_frame"
)

type UserHeadFrameModel struct {
	Uin int
}

func (this *UserHeadFrameModel) TableName() string {
	index := GetTableSplitIndex(this.Uin)
	return fmt.Sprintf("%s_%d", TABLE_NAME_USER_HEAD_FRAME, index)
}

func (this *UserHeadFrameModel) CachedKey() string {
	return fmt.Sprintf("%s.%d", TABLE_NAME_USER_HEAD_FRAME, this.Uin)
}

func (this *UserHeadFrameModel) GetUserHeadFrame() (*table.TblUserHeadFrame, error) {
	if this.Uin <= 0 {
		return nil, custom_errors.New("uin is invalid")
	}

	tableName := this.TableName()
	engine := GetUinSetMysql(this.Uin)
	if engine == nil {
		return nil, custom_errors.New("engine is nil")
	}

	cond := fmt.Sprintf("uin=%d", this.Uin)

	records := make([]*table.TblUserHeadFrame, 0)
	err := mysql.FindRecordsBySimpleCond(engine, tableName, cond, 0, 0, &records)
	if err != nil {
		return nil, err
	}

	if len(records) > 0 {
		return records[0], nil
	}

	return nil, nil
}

func GetMultiUserHeadFrame(uinSlice ...int) (map[int]*table.TblUserHeadFrame, error) {
	userHeadFrameModel := UserHeadFrameModel{Uin: 0}

	keys := make([]string, 0, len(uinSlice))
	for _, uin := range uinSlice {
		if uin <= 0 {
			return nil, custom_errors.New("uin is invalid")
		}

		userHeadFrameModel.Uin = uin
		keys = append(keys, userHeadFrameModel.CachedKey())
	}

	groupKeys, err := GetKeysGroupByClusterRedis(keys...)
	if err != nil {
		return nil, err
	}

	userHeadFrameMap := make(map[int]*table.TblUserHeadFrame)
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

			var userHeadInfo table.TblUserHeadFrame

			if uinStr, ok := hashV["uin"]; ok {
				uin, err := strconv.Atoi(uinStr)
				if err != nil {
					continue
				}

				userHeadInfo.Uin = uin

				if protypeIdStr, ok := hashV["protype_id"]; ok {
					protypeId, err := strconv.Atoi(protypeIdStr)
					if err != nil {
						continue
					}

					userHeadInfo.CurHeadFrame = protypeId
				}

				if headIdStr, ok := hashV["head_id"]; ok {
					userHeadInfo.HeadId = headIdStr
				}

				if headTypeStr, ok := hashV["head_type"]; ok {
					headType, err := strconv.Atoi(headTypeStr)
					if err != nil {
						continue
					}

					userHeadInfo.HeadType = headType
				}

				userHeadFrameMap[uin] = &userHeadInfo
			}
		}
	}

	return userHeadFrameMap, nil
}

func GetUserHeadFrame(uin int) (*table.TblUserHeadFrame, error) {
	if uin <= 0 {
		return nil, custom_errors.New("uin is invalid")
	}

	userHeadFrameModel := UserHeadFrameModel{Uin: uin}
	redisKey := userHeadFrameModel.CachedKey()
	redisMgr := GetClusterRedis(redisKey)
	if redisMgr == nil {
		return nil, custom_errors.New("redisMgr is nil")
	}

	mapKV, err := redisMgr.HashGet(redisKey)
	if err != nil {
		return nil, err
	}

	var userHeadInfo table.TblUserHeadFrame
	userHeadInfo.Uin = uin

	if protypeIdStr, ok := mapKV["protype_id"]; ok {
		protypeId, err := strconv.Atoi(protypeIdStr)
		if err != nil {
			return nil, err
		}

		userHeadInfo.CurHeadFrame = protypeId

		if headIdStr, ok := mapKV["head_id"]; ok {
			userHeadInfo.HeadId = headIdStr
		}

		if headTypeStr, ok := mapKV["head_type"]; ok {
			headType, err := strconv.Atoi(headTypeStr)
			if err != nil {
				return nil, err
			}

			userHeadInfo.HeadType = headType
		}
	}

	return &userHeadInfo, nil
}
