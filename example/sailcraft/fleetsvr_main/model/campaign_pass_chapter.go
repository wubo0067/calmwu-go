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
	CAMPAIGN_PASS_CHAPTER_NAME = "campaign_pass_chapter"
)

type CampaignPassChapterModel struct {
	Uin int
}

func (model *CampaignPassChapterModel) TableName() string {
	index := GetTableSplitIndex(model.Uin)
	return fmt.Sprintf("%s_%d", CAMPAIGN_PASS_CHAPTER_NAME, index)
}

func (model *CampaignPassChapterModel) QueryChapterInfoByCampaignId(campaignId int, chapterInfo *table.TblCampaignPassChapter) (int, error) {
	base.GLog.Debug("QueryChapterInfoByCampaignId enter campaignId %d", campaignId)

	if chapterInfo == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("null point")
	}

	uin := model.Uin

	if uin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("uin is invalid")
	}

	engine := GetUinSetMysql(uin)
	if engine == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("database engine is nil")
	}

	tableName := model.TableName()

	xormCond := builder.Expr(fmt.Sprintf("uin=%d", uin)).And(builder.Expr(fmt.Sprintf("campaign_id=%d", campaignId)))
	records := make([]*table.TblCampaignPassChapter, 0)

	err := mysql.FindRecordsByMultiConds(engine, tableName, &xormCond, 1, 0, &records)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if recordsLen := len(records); recordsLen > 0 {
		for _, info := range records {
			*chapterInfo = *info
			break
		}

		return 1, nil
	}

	return 0, nil
}

func (model *CampaignPassChapterModel) QueryMaxChapterInfoByUin(chapterInfo *table.TblCampaignPassChapter) (int, error) {
	if chapterInfo == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("null point")
	}

	uin := model.Uin

	if uin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("uin is invalid")
	}

	engine := GetUinSetMysql(uin)
	if engine == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("database engine is nil")
	}

	tableName := model.TableName()

	condtion := fmt.Sprintf("uin=%d", uin)

	orderbyLst := make([]string, 0)
	orderbyLst = append(orderbyLst, "campaign_id desc")

	records := make([]*table.TblCampaignPassChapter, 0)
	err := mysql.FindRecordsBySimpleCondWithOrderBy(engine, tableName, condtion, 1, 0, orderbyLst, &records)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if recordsLen := len(records); recordsLen <= 0 {
		base.GLog.Debug("without campaign chapter passed info for user[uin=%d]", uin)
	}

	if rcLen := len(records); rcLen > 0 {
		*chapterInfo = *records[0]
		return 1, nil
	}

	return 0, nil
}

func (model *CampaignPassChapterModel) AddCampaignPassChapter(records []*table.TblCampaignPassChapter) (affected int64, err error) {
	affected = -1
	err = nil

	if len(records) == 0 {
		err = custom_errors.New("input params is empty")
		return
	}

	engine := GetUinSetMysql(model.Uin)
	if engine == nil {
		err = custom_errors.New("database engine is nil")
		return
	}

	tableName := model.TableName()

	base.GLog.Debug("database is [%s] table is [%s]", engine.DBEngineDataSouceName, tableName)

	affected, err = mysql.InsertRecord(engine, tableName, &records)

	return
}
