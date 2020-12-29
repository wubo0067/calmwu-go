package model

import (
	"fmt"
	"sailcraft/dataaccess/mysql"
	"sailcraft/fleetsvr_main/custom_errors"
	"sailcraft/fleetsvr_main/errorcode"
	"sailcraft/fleetsvr_main/table"

	"github.com/go-xorm/builder"
	"github.com/go-xorm/core"
)

const (
	TABLE_NAME_GUILD_FRIGATE_SHIP = "guild_frigate_ship"
)

type GuildFrigateShipModel struct {
	Id      int
	Creator int
}

func (this *GuildFrigateShipModel) TableName() string {
	index := GetTableSplitIndex(this.Creator)
	return fmt.Sprintf("%s_%d", TABLE_NAME_GUILD_FRIGATE_SHIP, index)
}

func (this *GuildFrigateShipModel) Validate() (int, error) {
	if this.Creator <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("guild creator is invalid")
	}

	if this.Id <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("guild id is invalid")
	}

	return 0, nil
}

func (this *GuildFrigateShipModel) Query() (*table.TblGuildFrigateShip, error) {
	_, err := this.Validate()
	if err != nil {
		return nil, err
	}

	tableName := this.TableName()
	engine := GetUinSetMysql(this.Creator)
	if engine == nil {
		return nil, custom_errors.New("engine is nil")
	}

	xormCond := builder.NewCond()
	xormCond = xormCond.And(builder.Expr(fmt.Sprintf("guild_id=%d", this.Id)))
	xormCond = xormCond.And(builder.Expr(fmt.Sprintf("guild_creator=%d", this.Creator)))

	records := make([]*table.TblGuildFrigateShip, 0)
	err = mysql.FindRecordsByMultiConds(engine, tableName, &xormCond, 0, 0, &records)
	if err != nil {
		return nil, err
	}

	if len(records) > 0 {
		return records[0], nil
	}

	return nil, nil
}

func (this *GuildFrigateShipModel) Update(record *table.TblGuildFrigateShip, updateCols ...string) (int, error) {
	retCode, err := this.Validate()
	if err != nil {
		return retCode, err
	}

	if record == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("record is nil")
	}

	tableName := this.TableName()
	engine := GetUinSetMysql(this.Creator)
	if engine == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("engine is nil")
	}

	pk := core.NewPK(record.Id)
	if len(updateCols) > 0 {
		_, err = mysql.UpdateRecordCols(engine, tableName, pk, record, updateCols...)
	} else {
		_, err = mysql.UpdateRecord(engine, tableName, pk, record)
	}

	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	return 0, nil
}

func (this *GuildFrigateShipModel) Insert(record *table.TblGuildFrigateShip) (int, error) {
	retCode, err := this.Validate()
	if err != nil {
		return retCode, err
	}

	tableName := this.TableName()
	engine := GetUinSetMysql(this.Creator)
	if engine == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("engine is nil")
	}

	_, err = mysql.InsertRecord(engine, tableName, record)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	return 0, nil
}

func (this *GuildFrigateShipModel) Delete() (int, error) {
	_, err := this.Validate()
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	tableName := this.TableName()
	engine := GetUinSetMysql(this.Creator)
	if engine == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("engine is nil")
	}

	_, err = engine.RealEngine.Exec(fmt.Sprintf("DELETE FROM %s WHERE guild_id=%d AND guild_creator=%d", tableName, this.Id, this.Creator))
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	return 0, nil
}
