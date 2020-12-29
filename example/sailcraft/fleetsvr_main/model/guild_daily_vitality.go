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
	TABLE_NAME_GUILD_DAILY_VITALITY = "guild_daily_vitality"
)

type GuildDailyVitalityModel struct {
	Uin int
}

func (this *GuildDailyVitalityModel) TableName() string {
	index := GetTableSplitIndex(this.Uin)
	return fmt.Sprintf("%s_%d", TABLE_NAME_GUILD_DAILY_VITALITY, index)
}

func (this *GuildDailyVitalityModel) GetVitality() (*table.TblGuildDailyVitality, error) {
	if this.Uin <= 0 {
		return nil, custom_errors.New("uin is invalid")
	}

	tableName := this.TableName()

	engine := GetUinSetMysql(this.Uin)
	if engine == nil {
		return nil, custom_errors.New("engine is nil")
	}

	cond := fmt.Sprintf("uin=%d", this.Uin)

	records := make([]*table.TblGuildDailyVitality, 0)
	err := mysql.FindRecordsBySimpleCond(engine, tableName, cond, 0, 0, &records)
	if err != nil {
		return nil, err
	}

	if len(records) > 0 {
		return records[0], nil
	}

	return nil, nil
}

func (this *GuildDailyVitalityModel) UpdateGuildDailyVitality(record *table.TblGuildDailyVitality, updateCols ...string) (int, error) {
	if this.Uin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("uin is invalid")
	}

	if record == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("null point")
	}

	tableName := this.TableName()
	engine := GetUinSetMysql(this.Uin)
	if engine == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("null point")
	}

	pk := core.NewPK(record.Id)
	_, err := mysql.UpdateRecordCols(engine, tableName, pk, record)
	if err != nil {
		base.GLog.Error("UpdateGuildDailyVitality failed[%s]", err)
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	return 0, nil
}

func (this *GuildDailyVitalityModel) AddGuildDialyVitality(record *table.TblGuildDailyVitality) (int, error) {
	if this.Uin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("uin is invalid")
	}

	if record == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("null point")
	}

	tableName := this.TableName()
	engine := GetUinSetMysql(this.Uin)
	if engine == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("null point")
	}

	_, err := mysql.InsertRecord(engine, tableName, record)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	return 0, nil
}
