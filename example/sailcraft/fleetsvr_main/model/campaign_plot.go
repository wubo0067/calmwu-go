package model

import (
	"fmt"
	"sailcraft/base"
	"sailcraft/dataaccess/mysql"
	"sailcraft/fleetsvr_main/custom_errors"
	"sailcraft/fleetsvr_main/errorcode"
	"sailcraft/fleetsvr_main/table"

	"github.com/go-xorm/builder"
)

const (
	TABLE_NAME_CAMPAIGN_PLOT = "campaign_plot"
)

type CampaignPlotModel struct {
	Uin int
}

func (this *CampaignPlotModel) TableName() string {
	index := GetTableSplitIndex(this.Uin)
	return fmt.Sprintf("%s_%d", TABLE_NAME_CAMPAIGN_PLOT, index)
}

func (this *CampaignPlotModel) GetPlotsInfo() ([]*table.TblCampaignPlot, error) {
	if this.Uin <= 0 {
		return nil, custom_errors.New("uin is invalid")
	}

	engine := GetUinSetMysql(this.Uin)
	if engine == nil {
		base.GLog.Error("database engine is nil")
		return nil, custom_errors.New("database engine is nil")
	}

	tableName := this.TableName()
	base.GLog.Debug("database is [%s] table is [%s]", engine.DBEngineDataSouceName, tableName)

	cond := fmt.Sprintf("uin=%d", this.Uin)

	records := make([]*table.TblCampaignPlot, 0)

	err := mysql.FindRecordsBySimpleCond(engine, tableName, cond, 0, 0, &records)
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (this *CampaignPlotModel) GetPlotInfo(protypeId int) (*table.TblCampaignPlot, error) {
	if this.Uin <= 0 {
		return nil, custom_errors.New("uin is invalid")
	}

	if protypeId <= 0 {
		return nil, custom_errors.New("protype id is invalid")
	}

	engine := GetUinSetMysql(this.Uin)
	if engine == nil {
		base.GLog.Error("database engine is nil")
		return nil, custom_errors.New("database engine is nil")
	}

	tableName := this.TableName()
	base.GLog.Debug("database is [%s] table is [%s]", engine.DBEngineDataSouceName, tableName)

	xormCond := builder.NewCond()
	xormCond = xormCond.And(builder.Expr(fmt.Sprintf("uin=%d", this.Uin)))
	xormCond = xormCond.And(builder.Expr(fmt.Sprintf("protype_id=%d", protypeId)))

	records := make([]*table.TblCampaignPlot, 0)

	err := mysql.FindRecordsByMultiConds(engine, tableName, &xormCond, 0, 0, &records)
	if err != nil {
		return nil, err
	}

	if len(records) > 0 {
		return records[0], nil
	}

	return nil, nil
}

func (this *CampaignPlotModel) AddPlotInfo(record *table.TblCampaignPlot) (int, error) {
	if this.Uin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("uin is invalid")
	}

	if record == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("null record")
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
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	return 0, nil
}
