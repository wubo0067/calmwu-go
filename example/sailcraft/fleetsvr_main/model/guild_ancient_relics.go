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
	TABLE_NAME_GUILD_ANCIENT_RELICS = "guild_ancient_relics"

	ANCIENT_RELICS_STATUS_UNCOMPLETED = 0
	ANCIENT_RELICS_STATUS_COMPLETED   = 1
	ANCIENT_RELICS_STATUS_RECEIVED    = 2
)

type GuildAncientRelicsModel struct {
	Uin int
}

func (this *GuildAncientRelicsModel) TableName() string {
	index := GetTableSplitIndex(this.Uin)
	return fmt.Sprintf("%s_%d", TABLE_NAME_GUILD_ANCIENT_RELICS, index)
}

func (this *GuildAncientRelicsModel) GetRelicsList() ([]*table.TblGuildAncientRelicsInfo, error) {
	if this.Uin <= 0 {
		return nil, custom_errors.New("uin is invalid")
	}

	tableName := this.TableName()
	engine := GetUinSetMysql(this.Uin)
	if engine == nil {
		return nil, custom_errors.New("engine is nil")
	}

	cond := fmt.Sprintf("uin=%d", this.Uin)

	records := make([]*table.TblGuildAncientRelicsInfo, 0)
	err := mysql.FindRecordsBySimpleCond(engine, tableName, cond, 0, 0, &records)
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (this *GuildAncientRelicsModel) GetRelics(protypeId int) (*table.TblGuildAncientRelicsInfo, error) {
	if this.Uin <= 0 {
		return nil, custom_errors.New("uin is invalid")
	}

	if protypeId <= 0 {
		return nil, custom_errors.New("protype id is invalid")
	}

	tableName := this.TableName()
	engine := GetUinSetMysql(this.Uin)
	if engine == nil {
		return nil, custom_errors.New("engine is nil")
	}

	xormCond := builder.NewCond()
	xormCond = xormCond.And(builder.Expr(fmt.Sprintf("uin=%d", this.Uin)))
	xormCond = xormCond.And(builder.Expr(fmt.Sprintf("protype_id=%d", protypeId)))

	records := make([]*table.TblGuildAncientRelicsInfo, 0)
	err := mysql.FindRecordsByMultiConds(engine, tableName, &xormCond, 0, 0, &records)
	if err != nil {
		return nil, err
	}

	if len(records) > 0 {
		return records[0], nil
	}

	return nil, nil
}

func (this *GuildAncientRelicsModel) UpdateReclis(record *table.TblGuildAncientRelicsInfo) (int, error) {
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

func (this *GuildAncientRelicsModel) AddReclis(record *table.TblGuildAncientRelicsInfo) (int, error) {
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
