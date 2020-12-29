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
	GUILD_MEMBER_INFO_TABLE_NAME = "guild_member_info"
)

type GuildMemberInfoModel struct {
	Creator int
	GId     int
}

func (this *GuildMemberInfoModel) TableName() string {
	index := GetTableSplitIndex(this.Creator)
	return fmt.Sprintf("%s_%d", GUILD_MEMBER_INFO_TABLE_NAME, index)
}

func (this *GuildMemberInfoModel) Validate() error {
	if this.Creator <= 0 {
		return custom_errors.New("guild creator[%d] is invalid", this.Creator)
	}

	if this.GId <= 0 {
		return custom_errors.New("guild id[%d] is invalid", this.GId)
	}

	return nil
}

func (this *GuildMemberInfoModel) ValidateRecord(record *table.TblGuildMemberInfo) error {
	if record == nil {
		return custom_errors.New("record is nil")
	}

	if this.Creator != record.Creator {
		return custom_errors.New("record creator[%d] is not equal to current guild creator[%d]", record.Creator, this.Creator)
	}

	if this.GId != record.GuildId {
		return custom_errors.New("record guild id[%d] is not equal to current guild id[%d]", record.GuildId, this.GId)
	}

	return nil
}

func (this *GuildMemberInfoModel) GetGuildMemberInfo(uin int) (*table.TblGuildMemberInfo, error) {
	err := this.Validate()
	if err != nil {
		return nil, err
	}

	if uin <= 0 {
		return nil, custom_errors.New("uin is invalid")
	}

	engine := GetUinSetMysql(this.Creator)
	if engine == nil {
		return nil, custom_errors.New("database engine is nil")
	}

	tableName := this.TableName()

	xormCond := builder.NewCond()
	xormCond = xormCond.And(builder.Expr(fmt.Sprintf("creator=%d", this.Creator)))
	xormCond = xormCond.And(builder.Expr(fmt.Sprintf("guild_id=%d", this.GId)))
	xormCond = xormCond.And(builder.Expr(fmt.Sprintf("member_uin=%d", uin)))

	records := make([]*table.TblGuildMemberInfo, 0)

	err = mysql.FindRecordsByMultiConds(engine, tableName, &xormCond, 0, 0, &records)
	if err != nil {
		return nil, err
	}

	if len(records) > 0 {
		return records[0], nil
	}

	return nil, nil
}

func (this *GuildMemberInfoModel) GetAllGuildMember() ([]*table.TblGuildMemberInfo, error) {
	err := this.Validate()
	if err != nil {
		return nil, err
	}

	engine := GetUinSetMysql(this.Creator)
	if engine == nil {
		return nil, custom_errors.New("database engine is nil")
	}

	tableName := this.TableName()
	base.GLog.Debug("database is [%s] table is [%s]", engine.DBEngineDataSouceName, tableName)

	xormCond := builder.NewCond()
	xormCond = xormCond.And(builder.Expr(fmt.Sprintf("creator=%d", this.Creator)))
	xormCond = xormCond.And(builder.Expr(fmt.Sprintf("guild_id=%d", this.GId)))

	records := make([]*table.TblGuildMemberInfo, 0)

	err = mysql.FindRecordsByMultiConds(engine, tableName, &xormCond, 0, 0, &records)
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (this *GuildMemberInfoModel) AddNewGuildMemberInfo(record *table.TblGuildMemberInfo) (int, error) {
	err := this.Validate()
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	err = this.ValidateRecord(record)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	engine := GetUinSetMysql(this.Creator)
	if engine == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("database engine is nil")
	}

	tableName := this.TableName()

	base.GLog.Debug("database is [%s] table is [%s]", engine.DBEngineDataSouceName, tableName)

	_, err = mysql.InsertRecord(engine, tableName, record)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	return 0, nil
}

func (this *GuildMemberInfoModel) UpdateGuildMemberInfo(record *table.TblGuildMemberInfo) (int, error) {
	err := this.Validate()
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	err = this.ValidateRecord(record)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	engine := GetUinSetMysql(record.Creator)
	if engine == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("database engine is nil")
	}

	tableName := this.TableName()
	base.GLog.Debug("database is [%s] table is [%s]", engine.DBEngineDataSouceName, tableName)

	pk := core.NewPK(record.Id)
	_, err = mysql.UpdateRecord(engine, tableName, pk, record)

	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	return 0, nil
}

func (this *GuildMemberInfoModel) DeleteMember(record *table.TblGuildMemberInfo) (int, error) {
	err := this.Validate()
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	err = this.ValidateRecord(record)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	engine := GetUinSetMysql(this.Creator)
	if engine == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("database engine is nil")
	}

	tableName := this.TableName()

	pk := core.NewPK(record.Id)
	_, err = mysql.DeleteRecord(engine, tableName, pk, record)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	return 0, nil
}

func (this *GuildMemberInfoModel) GetGuildMemberByPost(post string) ([]*table.TblGuildMemberInfo, error) {
	if this.Creator <= 0 {
		return nil, custom_errors.New("guild creator uin is invalid")
	}

	if this.GId <= 0 {
		return nil, custom_errors.New("guild id is invalid")
	}

	engine := GetUinSetMysql(this.Creator)
	if engine == nil {
		return nil, custom_errors.New("database engine is nil")
	}

	tableName := this.TableName()
	base.GLog.Debug("database is [%s] table is [%s]", engine.DBEngineDataSouceName, tableName)

	xormCond := builder.NewCond()
	xormCond = xormCond.And(builder.Expr(fmt.Sprintf("creator=%d", this.Creator)))
	xormCond = xormCond.And(builder.Expr(fmt.Sprintf("guild_id=%d", this.GId)))
	xormCond = xormCond.And(builder.Expr(fmt.Sprintf("post='%s'", post)))

	records := make([]*table.TblGuildMemberInfo, 0)
	err := mysql.FindRecordsByMultiConds(engine, tableName, &xormCond, 0, 0, &records)
	if err != nil {
		return nil, err
	}

	return records, nil
}
