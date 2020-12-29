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
	CAMPAIGN_PRODUCE_RESOURCE = "campaign_produce_resource"
)

type CampaignProduceResourceModel struct {
	Uin int
}

func (model *CampaignProduceResourceModel) TableName() string {
	index := GetTableSplitIndex(model.Uin)
	return fmt.Sprintf("%s_%d", CAMPAIGN_PRODUCE_RESOURCE, index)
}

// 查询玩家生产资源信息，有数据则返回1，无数据返回0，有错误返回其他
func (model *CampaignProduceResourceModel) QueryCampaignProductResourceInfo(produceInfo *table.TblCampaignProduceResource) (int, error) {
	base.GLog.Debug("QueryCampaignProductResourceInfo enter uin %d", model.Uin)

	if produceInfo == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("null point")
	}

	if model.Uin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("uin is invalid")
	}

	engine := GetUinSetMysql(model.Uin)
	if engine == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("database engine is nil")
	}

	tableName := model.TableName()

	condition := fmt.Sprintf("uin = %d", model.Uin)
	records := make([]*table.TblCampaignProduceResource, 0)

	err := mysql.FindRecordsBySimpleCond(engine, tableName, condition, 1, 0, &records)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if rcLen := len(records); rcLen > 0 {
		*produceInfo = *records[0]

		return 1, nil
	}

	return 0, nil
}

// 刷新玩家领取资源时间
func (model *CampaignProduceResourceModel) RefreshResourceReceivedTime(receivedTime int, produceInfo *table.TblCampaignProduceResource, coverOld bool) (int, error) {

	if produceInfo == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("null point")
	}

	if model.Uin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("uin is invalid")
	}

	engine := GetUinSetMysql(model.Uin)

	if engine == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("database engine is nil")
	}

	tableName := model.TableName()

	base.GLog.Debug("database is [%s] table is [%s]", engine.DBEngineDataSouceName, tableName)

	retCode, err := model.QueryCampaignProductResourceInfo(produceInfo)

	if err != nil {
		return retCode, err
	}

	produceInfo.Uin = model.Uin
	produceInfo.LastReceiveTime = receivedTime

	// 有数据则刷新，没有数据则添加一条新数据
	if retCode == 1 {
		if coverOld {

			PK := core.NewPK(produceInfo.Uin)

			_, err := mysql.UpdateRecord(engine, tableName, PK, produceInfo)

			if err != nil {
				return errorcode.ERROR_CODE_DEFAULT, err
			}

			return 0, nil
		} else {
			return errorcode.ERROR_CODE_CAMPAIGN_PRODUCE_RESOURCE_ALREAY_INIT, custom_errors.New("already have one data")
		}
	} else {

		records := make([]*table.TblCampaignProduceResource, 1)
		records[0] = produceInfo

		_, err := mysql.InsertRecord(engine, tableName, &records)

		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		return 0, nil
	}
}
