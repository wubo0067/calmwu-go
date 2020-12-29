package model

import (
	"fmt"
	"sailcraft/base"
	"sailcraft/dataaccess/mysql"
	"sailcraft/fleetsvr_main/custom_errors"
	"sailcraft/fleetsvr_main/errorcode"
	"sailcraft/fleetsvr_main/table"

	"github.com/go-xorm/core"
)

const (
	ACTIVITY_TASK_FRESH_TABLE_NAME = "activity_task_fresh"
)

type ActivityTaskFreshModel struct {
	Uin int
}

func (this *ActivityTaskFreshModel) TableName() string {
	index := GetTableSplitIndex(this.Uin)
	return fmt.Sprintf("%s_%d", ACTIVITY_TASK_FRESH_TABLE_NAME, index)
}

func (this *ActivityTaskFreshModel) GetActivityTaskFresh() (*table.TblActivityTaskFresh, error) {
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

	records := make([]*table.TblActivityTaskFresh, 0)
	mysql.FindRecordsBySimpleCond(engine, tableName, condition, 0, 0, &records)

	if len(records) > 0 {
		return records[0], nil
	}

	return nil, nil
}

func (this *ActivityTaskFreshModel) UpdateActivityTaskFresh(record *table.TblActivityTaskFresh) (int, error) {
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

	pk := core.NewPK(record.Uin)
	_, err := mysql.UpdateRecord(engine, tableName, pk, record)
	if err != nil {
		base.GLog.Error("update record[%+v] from table[%s] error[%s]", record, tableName, err)
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	return 0, nil
}

func (this *ActivityTaskFreshModel) AddActivityTaskFresh(record *table.TblActivityTaskFresh) (int, error) {
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
		base.GLog.Error("insert record[%+v] to table[%s] error[%s]", record, tableName, err)
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	return 0, nil
}
