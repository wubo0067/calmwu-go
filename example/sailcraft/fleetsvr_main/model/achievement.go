package model

import (
	"fmt"
	"sailcraft/base"
	"sailcraft/dataaccess/mysql"
	"sailcraft/fleetsvr_main/custom_errors"
	"sailcraft/fleetsvr_main/errorcode"
	"sailcraft/fleetsvr_main/table"
	"sailcraft/fleetsvr_main/utils"

	"github.com/go-xorm/builder"

	"github.com/go-xorm/core"
)

const (
	ACHIEVEMENT_TABLE_NAME = "achievement"

	ACHIEVEMENT_STATUS_UNCOMPLETED = 0
	ACHIEVEMENT_STATUS_COMPLETED   = 1

	ACHIEVEMENT_PROGRESS_CURRENT        = "current"
	ACHIEVEMENT_PROGRESS_TOTAL          = "total"
	ACHIEVEMENT_PROGRESS_LAST_DAY       = "last_day"
	ACHIEVEMENT_PROGRESS_DAY_CONTINUOUS = "day_continuous"
)

type AchievementModel struct {
	Uin int
}

func (this *AchievementModel) TableName() string {
	index := GetTableSplitIndex(this.Uin)
	return fmt.Sprintf("%s_%d", ACHIEVEMENT_TABLE_NAME, index)
}

func (this *AchievementModel) GetAchievementList() ([]*table.TblAchievement, error) {
	if this.Uin <= 0 {
		return nil, custom_errors.New("uin is invalid")
	}

	engine := GetUinSetMysql(this.Uin)
	if engine == nil {
		return nil, custom_errors.New("database engine is nil")
	}

	tableName := this.TableName()
	base.GLog.Debug("database is [%s] table is [%s]", engine.DBEngineDataSouceName, tableName)

	condition := fmt.Sprintf("uin=%d", this.Uin)

	records := make([]*table.TblAchievement, 0)
	err := mysql.FindRecordsBySimpleCond(engine, tableName, condition, 0, 0, &records)
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (this *AchievementModel) GetAchievementByProtypeId(protypeId int) (*table.TblAchievement, error) {
	if protypeId <= 0 {
		return nil, custom_errors.New("protype id is invalid")
	}

	if this.Uin <= 0 {
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
	xormCond = xormCond.And(builder.Expr(fmt.Sprintf("protype_id=%d", protypeId)))

	records := make([]*table.TblAchievement, 0)
	err := mysql.FindRecordsByMultiConds(engine, tableName, &xormCond, 0, 0, &records)
	if err != nil {
		return nil, err
	}

	if len(records) > 0 {
		return records[0], nil
	}

	return nil, nil
}

func (this *AchievementModel) UpdateAchievement(record *table.TblAchievement) (int, error) {
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
		base.GLog.Debug(err)
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	return 0, nil
}

func (this *AchievementModel) AddAchievement(record *table.TblAchievement) (int, error) {
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
		base.GLog.Error("insert record[%+v] into %s err[%s]", record, tableName, err)
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	return 0, nil
}

func (this *AchievementModel) AddMultiAchievements(records []*table.TblAchievement) (int, error) {
	if this.Uin <= 0 {
		base.GLog.Error("uin is invalid")
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("uin is invalid")
	}

	if len(records) <= 0 {
		return 0, nil
	}

	engine := GetUinSetMysql(this.Uin)
	if engine == nil {
		base.GLog.Error("database engine is nil")
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("database engine is nil")
	}

	tableName := this.TableName()
	base.GLog.Debug("database is [%s] table is [%s]", engine.DBEngineDataSouceName, tableName)

	_, err := mysql.InsertSliceRecordsToSameTable(engine, tableName, &records)
	if err != nil {
		base.GLog.Error("insert multi records[%+v] into %s err[%s]", records, tableName, err)
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	return 0, nil
}

func (this *AchievementModel) UpdateMultiAchievements(records []*table.TblAchievement) (int, error) {
	if records == nil {
		base.GLog.Error("null point")
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("null point")
	}

	if len(records) <= 0 {
		return 0, nil
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

	_, err := utils.UpdateMultiRecords(engine.RealEngine, tableName, records)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	return 0, nil
}
