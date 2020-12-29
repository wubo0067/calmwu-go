package model

import (
	"fmt"
	"sailcraft/base"
	"sailcraft/dataaccess/mysql"
	"sailcraft/fleetsvr_main/custom_errors"
	"sailcraft/fleetsvr_main/errorcode"
	"sailcraft/fleetsvr_main/table"

	"github.com/go-xorm/builder"

	"github.com/go-xorm/core"
)

const (
	ACTIVITY_SCORE_REWARD_TABLE_NAME = "activity_score_reward"

	ACTIVITY_SCORE_REWARD_STATUS_UNACTIVE  = 0
	ACTIVITY_SCORE_REWARD_STATUS_UNRECEIVE = 1
	ACTIVITY_SCORE_REWARD_STATUS_RECEIVED  = 2
)

type ActivityScoreRewardModel struct {
	Uin int
}

func (this *ActivityScoreRewardModel) TableName() string {
	index := GetTableSplitIndex(this.Uin)
	return fmt.Sprintf("%s_%d", ACTIVITY_SCORE_REWARD_TABLE_NAME, index)
}

func (this *ActivityScoreRewardModel) GetActivityScoreRewardList() ([]*table.TblActivityScoreReward, error) {
	if this.Uin <= 0 {
		base.GLog.Error("uin is invalide")
		return nil, custom_errors.New("uin is invalid")
	}

	engine := GetUinSetMysql(this.Uin)
	if engine == nil {
		return nil, custom_errors.New("database engine is nil")
	}

	tableName := this.TableName()
	base.GLog.Debug("database is [%s] table is [%s]", engine.DBEngineDataSouceName, tableName)

	condition := fmt.Sprintf("uin=%d", this.Uin)

	records := make([]*table.TblActivityScoreReward, 0)
	err := mysql.FindRecordsBySimpleCond(engine, tableName, condition, 0, 0, &records)
	if err != nil {
		base.GLog.Error("find record from table[%s] error[%s]", tableName, err)
		return nil, nil
	}

	return records, nil
}

func (this *ActivityScoreRewardModel) UpdateActivityScoreReward(record *table.TblActivityScoreReward) (int, error) {
	if record == nil {
		base.GLog.Error("null point")
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("null point")
	}

	if this.Uin <= 0 {
		base.GLog.Error("uin is invalid")
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("uin is invalid")
	}

	engine := GetUinSetMysql(this.Uin)
	if engine == nil {
		base.GLog.Error("database engine is nil")
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("database engine is nil")
	}

	tableName := this.TableName()
	base.GLog.Debug("database is [%s] table is [%s]", engine.DBEngineDataSouceName, tableName)

	pk := core.NewPK(record.Id)
	_, err := mysql.UpdateRecord(engine, tableName, pk, record)
	if err != nil {
		base.GLog.Error("update record[%+v] from table[%s] error[%s]", record, tableName, err)
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	return 0, nil
}

func (this *ActivityScoreRewardModel) AddActivityScoreReward(record *table.TblActivityScoreReward) (int, error) {
	if record == nil {
		base.GLog.Error("null point")
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("null point")
	}

	if this.Uin <= 0 {
		base.GLog.Error("uin is invalid")
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("uin is invalid")
	}

	engine := GetUinSetMysql(this.Uin)
	if engine == nil {
		base.GLog.Error("database engine is nil")
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("database engine is nil")
	}

	tableName := this.TableName()
	base.GLog.Debug("database is [%s] table is [%s]", engine.DBEngineDataSouceName, tableName)

	_, err := mysql.InsertRecord(engine, tableName, record)
	if err != nil {
		base.GLog.Error("insert record[%+v] into table[%s] error[%s]", record, tableName, err)
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	return 0, nil
}

func (this *ActivityScoreRewardModel) GetActivityScoreRewardByRewardId(rewardId string) (*table.TblActivityScoreReward, error) {
	if this.Uin <= 0 {
		base.GLog.Error("uin is invalide")
		return nil, custom_errors.New("uin is invalid")
	}

	engine := GetUinSetMysql(this.Uin)
	if engine == nil {
		return nil, custom_errors.New("database engine is nil")
	}

	tableName := this.TableName()
	base.GLog.Debug("database is [%s] table is [%s]", engine.DBEngineDataSouceName, tableName)

	xormCond := builder.NewCond()
	xormCond = xormCond.And(builder.Expr(fmt.Sprintf("uin=%d", this.Uin)))
	xormCond = xormCond.And(builder.Expr(fmt.Sprintf("reward_id='%s'", rewardId)))

	records := make([]*table.TblActivityScoreReward, 0)
	err := mysql.FindRecordsByMultiConds(engine, tableName, &xormCond, 0, 0, &records)
	if err != nil {
		return nil, err
	}

	if len(records) > 0 {
		return records[0], nil
	}

	return nil, nil
}
