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
	GROWUP_TASK_TABLE_NAME = "growup_task"

	GROWUP_TASK_STATUS_UNCOMPLETED      = 0
	GROWUP_TASK_STATUS_WAIT_FOR_RECEIVE = 1
	GROWUP_TASK_STATUS_COMPLETED        = 2
)

type GrowupTaskModel struct {
	Uin int
}

func (this *GrowupTaskModel) TableName() string {
	index := GetTableSplitIndex(this.Uin)
	return fmt.Sprintf("%s_%d", GROWUP_TASK_TABLE_NAME, index)
}

func (this *GrowupTaskModel) GetGrowupTaskList() ([]*table.TblGrowupTask, error) {
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

	records := make([]*table.TblGrowupTask, 0)
	err := mysql.FindRecordsBySimpleCond(engine, tableName, condition, 0, 0, &records)
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (this *GrowupTaskModel) UpdateGrowupTask(record *table.TblGrowupTask) (int, error) {
	if record == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("null point")
	}

	if this.Uin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("uin is invalid")
	}

	engine := GetUinSetMysql(this.Uin)
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

	return 0, nil
}

func (this *GrowupTaskModel) AddGrowupTask(record *table.TblGrowupTask) (int, error) {
	if record == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("null point")
	}

	if this.Uin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("uin is invalid")
	}

	engine := GetUinSetMysql(this.Uin)
	if engine == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("database engine is nil")
	}

	tableName := this.TableName()
	base.GLog.Debug("database is [%s] table is [%s]", engine.DBEngineDataSouceName, tableName)

	_, err := mysql.InsertRecord(engine, tableName, record)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	return 0, nil
}

func (this *GrowupTaskModel) UpdateMultiGrowupTask(records []*table.TblGrowupTask) (int, error) {
	if this.Uin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("uin is invalid")
	}

	engine := GetUinSetMysql(this.Uin)
	if engine == nil {
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

func (this *GrowupTaskModel) AddMultiGrowupTask(records []*table.TblGrowupTask) (int, error) {
	if len(records) <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("null point")
	}

	if this.Uin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("uin is invalid")
	}

	engine := GetUinSetMysql(this.Uin)
	if engine == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("database engine is nil")
	}

	tableName := this.TableName()
	base.GLog.Debug("database is [%s] table is [%s]", engine.DBEngineDataSouceName, tableName)

	_, err := mysql.InsertSliceRecordsToSameTable(engine, tableName, &records)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	return 0, nil
}

func (this *GrowupTaskModel) GetGrowupTaskByProtypeId(protypeId string) (*table.TblGrowupTask, error) {

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
	xormCond = xormCond.And(builder.Expr(fmt.Sprintf("protype_id=\"%s\"", protypeId)))

	records := make([]*table.TblGrowupTask, 0)
	err := mysql.FindRecordsByMultiConds(engine, tableName, &xormCond, 0, 0, &records)
	if err != nil {
		return nil, err
	}

	if len(records) > 0 {
		return records[0], nil
	}

	return nil, nil
}
