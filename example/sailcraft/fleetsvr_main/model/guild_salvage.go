package model

import (
	"fmt"
	"sailcraft/dataaccess/mysql"
	"sailcraft/fleetsvr_main/custom_errors"
	"sailcraft/fleetsvr_main/errorcode"
	"sailcraft/fleetsvr_main/table"

	"github.com/go-xorm/core"
)

const (
	TABLE_NAME_GUILD_SALVAGE = "guild_salvage"
)

type GuildSalvageModel struct {
	Uin int
}

func (this *GuildSalvageModel) TableName() string {
	index := GetTableSplitIndex(this.Uin)
	return fmt.Sprintf("%s_%d", TABLE_NAME_GUILD_SALVAGE, index)
}

func (this *GuildSalvageModel) GetSalvage() (*table.TblGuildSalvage, error) {
	if this.Uin <= 0 {
		return nil, custom_errors.New("uin is invalid")
	}

	tableName := this.TableName()
	engine := GetUinSetMysql(this.Uin)
	if engine == nil {
		return nil, custom_errors.New("engine is nil")
	}

	cond := fmt.Sprintf("uin=%d", this.Uin)

	records := make([]*table.TblGuildSalvage, 0)
	err := mysql.FindRecordsBySimpleCond(engine, tableName, cond, 0, 0, &records)
	if err != nil {
		return nil, err
	}

	if len(records) > 0 {
		return records[0], nil
	}

	return nil, nil
}

func (this *GuildSalvageModel) AddSalvage(record *table.TblGuildSalvage) (int, error) {
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

func (this *GuildSalvageModel) UpdateSalvage(record *table.TblGuildSalvage) (int, error) {
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
