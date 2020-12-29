package model

import (
	"fmt"
	"sailcraft/dataaccess/mysql"
	"sailcraft/fleetsvr_main/custom_errors"
	"sailcraft/fleetsvr_main/errorcode"
	"sailcraft/fleetsvr_main/table"
	"sailcraft/fleetsvr_main/utils"

	"github.com/go-xorm/core"
)

const (
	TABLE_NAME_GUILD_DAILY_VITALITY_REWARD = "guild_daily_vitality_reward"
)

type GuildDailyVitalityRewardModel struct {
	Uin int
}

func (this *GuildDailyVitalityRewardModel) TableName() string {
	index := GetTableSplitIndex(this.Uin)
	return fmt.Sprintf("%s_%d", TABLE_NAME_GUILD_DAILY_VITALITY_REWARD, index)
}

func (this *GuildDailyVitalityRewardModel) GetRewardList() ([]*table.TblGuildDailyVitalityReward, error) {
	if this.Uin <= 0 {
		return nil, custom_errors.New("uin is invalid")
	}

	tableName := this.TableName()
	engine := GetUinSetMysql(this.Uin)
	if engine == nil {
		return nil, custom_errors.New("engine is nil")
	}

	cond := fmt.Sprintf("uin=%d", this.Uin)

	records := make([]*table.TblGuildDailyVitalityReward, 0)
	err := mysql.FindRecordsBySimpleCond(engine, tableName, cond, 0, 0, &records)
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (this *GuildDailyVitalityRewardModel) UpdateReward(record *table.TblGuildDailyVitalityReward) (int, error) {
	if this.Uin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("uin is invalid")
	}

	if record == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("null point")
	}

	tableName := this.TableName()
	engine := GetUinSetMysql(this.Uin)
	if engine == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("engine is nil")
	}

	pk := core.NewPK(record.Id)
	_, err := mysql.UpdateRecord(engine, tableName, pk, record)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	return 0, nil
}

func (this *GuildDailyVitalityRewardModel) InsertReward(record *table.TblGuildDailyVitalityReward) (int, error) {
	if this.Uin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("uin is invalid")
	}

	if record == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("null point")
	}

	tableName := this.TableName()
	engine := GetUinSetMysql(this.Uin)
	if engine == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("engine is nil")
	}

	_, err := mysql.InsertRecord(engine, tableName, record)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	return 0, nil
}

func (this *GuildDailyVitalityRewardModel) InsertMultiReward(recordSlice []*table.TblGuildDailyVitalityReward) (int, error) {
	if this.Uin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("uin is invalid")
	}

	if len(recordSlice) <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("record slice is empty")
	}

	tableName := this.TableName()
	engine := GetUinSetMysql(this.Uin)
	if engine == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("engine is nil")
	}

	_, err := mysql.InsertSliceRecordsToSameTable(engine, tableName, recordSlice)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	return 0, nil
}

func (this *GuildDailyVitalityRewardModel) UpdateMultiRewards(recSlice []*table.TblGuildDailyVitalityReward) (int, error) {
	if this.Uin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("uin is invalid")
	}

	if len(recSlice) <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("record slice is empty")
	}

	tableName := this.TableName()
	engine := GetUinSetMysql(this.Uin)
	if engine == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("engine is nil")
	}

	_, err := utils.UpdateMultiRecords(engine.RealEngine, tableName, &recSlice)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	return 0, nil
}

func (this *GuildDailyVitalityRewardModel) ResetStatus() (int, error) {
	if this.Uin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("uin is invalid")
	}

	tableName := this.TableName()
	engine := GetUinSetMysql(this.Uin)
	if engine == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("engine is nil")
	}

	_, err := engine.RealEngine.Exec(fmt.Sprintf("UPDATE %s SET status = 0 WHERE uin=%d", tableName, this.Uin))
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	return 0, nil
}
