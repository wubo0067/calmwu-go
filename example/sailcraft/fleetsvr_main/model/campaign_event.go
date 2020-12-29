package model

import (
	"fmt"
	"sailcraft/base"
	"sailcraft/dataaccess/mysql"
	"sailcraft/fleetsvr_main/custom_errors"
	"sailcraft/fleetsvr_main/errorcode"
	"sailcraft/fleetsvr_main/table"

	"github.com/go-xorm/builder"

	"sailcraft/fleetsvr_main/utils"

	"github.com/go-xorm/core"
)

const (
	CAMPAIGN_EVENT_PREFIX               = "campaign_event"
	CAMPAIGN_EVENT_TRIGGER_TYPE_TIME    = 0
	CAMPAIGN_EVENT_TRIGGER_TYPE_PVP_WIN = 1
)

const (
	CAMPAIGN_EVENT_STATUS_UNFINISHED = 0 // 未完成
	CAMPAIGN_EVENT_STATUS_RECEIVED   = 1 // 已领取
	CAMPAIGN_EVENT_STATUS_CLOSED     = 2 // 已关闭
)

type CampaignEventModel struct {
	Uin int
}

func (this *CampaignEventModel) TableName() string {
	index := GetTableSplitIndex(this.Uin)
	return fmt.Sprintf("%s_%d", CAMPAIGN_EVENT_PREFIX, index)
}

// 添加事件信息
func (this *CampaignEventModel) AddCampaignEvent(eventInfo *table.TblCampaignEvent) (int, error) {
	if eventInfo == nil {
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

	eventInfo.Uin = this.Uin
	_, err := mysql.InsertRecord(engine, tableName, eventInfo)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	return 0, nil
}

// 更新事件信息
func (this *CampaignEventModel) UpdateMultiCampaignEvents(records []*table.TblCampaignEvent) (int, error) {
	if records == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("null point")
	}

	if len(records) <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("campaign event records is empty")
	}

	if this.Uin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("uin is invalid")
	}

	engine := GetUinSetMysql(this.Uin)

	if engine == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("database engine is nil")
	}

	tableName := this.TableName()

	base.GLog.Debug("Database engine: [%s], TableName: [%s]", engine.DBEngineDataSouceName, tableName)

	_, err := utils.UpdateMultiRecords(engine.RealEngine, tableName, &records)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	for index := range records {
		base.GLog.Debug("Uin[%d] record:%+v", this.Uin, *records[index])
	}

	return 0, nil
}

func (this *CampaignEventModel) UpdateCampaignEvent(record *table.TblCampaignEvent) (int, error) {
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

	base.GLog.Debug("Database engine: [%s], TableName: [%s]", engine.DBEngineDataSouceName, tableName)
	pk := core.NewPK(record.Id)
	_, err := mysql.UpdateRecord(engine, tableName, pk, record)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	base.GLog.Debug("Uin[%d] record:%+v", this.Uin, *record)

	return 0, nil
}

// 查询未完成事件
func (this *CampaignEventModel) QueryCampaignEventUnfinished() ([]*table.TblCampaignEvent, error) {
	if this.Uin <= 0 {
		return nil, custom_errors.New("uin is invalid")
	}

	engine := GetUinSetMysql(this.Uin)
	if engine == nil {
		return nil, custom_errors.New("database engine is nil")
	}

	tableName := this.TableName()

	base.GLog.Debug("Database engine: [%s], TableName: [%s]", engine.DBEngineDataSouceName, tableName)

	condition := fmt.Sprintf("uin=%d", this.Uin)
	records := make([]*table.TblCampaignEvent, 0)

	err := mysql.FindRecordsBySimpleCond(engine, tableName, condition, -1, 0, &records)
	if err != nil {
		base.GLog.Debug("FindRecordsByMultiConds error![%s]", err)
		return nil, err
	}

	validRecords := make([]*table.TblCampaignEvent, 0, len(records))
	for _, record := range records {
		if record.EventStatus == CAMPAIGN_EVENT_STATUS_UNFINISHED {
			validRecords = append(validRecords, record)
			base.GLog.Debug("Uin[%d] UnFinished Record:%+v", this.Uin, *record)
		}
	}

	return validRecords, nil
}

// 获取所有事件
func (this *CampaignEventModel) GetAllEvents() ([]*table.TblCampaignEvent, error) {
	if this.Uin <= 0 {
		return nil, custom_errors.New("uin is invalid")
	}

	engine := GetUinSetMysql(this.Uin)
	if engine == nil {
		return nil, custom_errors.New("database engine is nil")
	}

	tableName := this.TableName()

	base.GLog.Debug("Database engine: [%s], TableName: [%s]", engine.DBEngineDataSouceName, tableName)

	condition := fmt.Sprintf("uin=%d", this.Uin)
	records := make([]*table.TblCampaignEvent, 0)

	err := mysql.FindRecordsBySimpleCond(engine, tableName, condition, 0, 0, &records)
	if err != nil {
		base.GLog.Debug("FindRecordsByMultiConds error![%s]", err)
		return nil, err
	}

	return records, nil
}

// 通过事件的id查询事件内容
// 该id事件不存在返回“0, nil”，否则返回“1, nil”
func (this *CampaignEventModel) QueryCampaignEventById(id int, eventInfo *table.TblCampaignEvent) (int, error) {
	if eventInfo == nil {
		return 0, custom_errors.New("null point")
	}

	if this.Uin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("uin is invalid")
	}

	if id <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("id(pk) is invalid")
	}

	engine := GetUinSetMysql(this.Uin)
	if engine == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("database engine is nil")
	}

	tableName := this.TableName()

	base.GLog.Debug("Database engine: [%s], TableName: [%s]", engine.DBEngineDataSouceName, tableName)

	condition := fmt.Sprintf("id=%d", id)
	records := make([]*table.TblCampaignEvent, 0)

	err := mysql.FindRecordsBySimpleCond(engine, tableName, condition, 1, 0, &records)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if rcLen := len(records); rcLen > 0 {
		*eventInfo = *records[0]
		return 1, nil
	} else {
		return 0, nil
	}
}

// 删除事件
func (this *CampaignEventModel) DeleteCampaignEvents(eventInfos []*table.TblCampaignEvent) (int, error) {
	if len(eventInfos) <= 0 {
		return 0, nil
	}

	if this.Uin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("uin is invalid")
	}

	engine := GetUinSetMysql(this.Uin)
	if engine == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("database engine is nil")
	}

	tableName := this.TableName()

	base.GLog.Debug("Database engine: [%s], TableName: [%s]", engine.DBEngineDataSouceName, tableName)

	eventIds := make([]int, len(eventInfos))
	for index, eventInfo := range eventInfos {
		eventIds[index] = eventInfo.Id
	}

	var retEventInfo table.TblCampaignEvent
	xormCond := builder.In("id", eventIds)

	_, err := mysql.DeleteRecordsByMultiConds(engine, tableName, &xormCond, &retEventInfo)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	return 0, nil
}

func (this *CampaignEventModel) GetUnifinishedEventsByMissionId(missionIds []int) ([]*table.TblCampaignEvent, error) {
	maxIndex := len(missionIds) - 1
	if maxIndex < 0 {
		return nil, nil
	}

	if this.Uin <= 0 {
		return nil, custom_errors.New("uin is invalid")
	}

	engine := GetUinSetMysql(this.Uin)
	if engine == nil {
		return nil, custom_errors.New("database engine is nil")
	}

	tableName := this.TableName()

	base.GLog.Debug("Database engine: [%s], TableName: [%s]", engine.DBEngineDataSouceName, tableName)

	// xormConds := builder.NewCond()
	// for _, missionId := range missionIds {
	// 	xormConds = xormConds.Or(builder.Expr(fmt.Sprintf("mission_id=%d", missionId)))
	// }
	xormConds := builder.Expr(fmt.Sprintf("uin=%d", this.Uin))

	records := make([]*table.TblCampaignEvent, 0)
	err := mysql.FindRecordsByMultiConds(engine, tableName, &xormConds, 0, 0, &records)
	if err != nil {
		return nil, err
	}

	validRecords := make([]*table.TblCampaignEvent, 0, len(records))
	for _, record := range records {
		if record.EventStatus == CAMPAIGN_EVENT_STATUS_UNFINISHED {
			// 还要判断record的missionId是否在列表中
			for _, missionID := range missionIds {
				if record.MissionId == missionID {
					validRecords = append(validRecords, record)
					base.GLog.Debug("Uin[%d] UnFinished Record:%+v", this.Uin, *record)
					break
				}
			}
		}
	}

	return validRecords, nil
}
