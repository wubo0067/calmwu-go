package model

import (
	"fmt"
	"sailcraft/dataaccess/mysql"
	"sailcraft/fleetsvr_main/custom_errors"
	"sailcraft/fleetsvr_main/errorcode"
	"sailcraft/fleetsvr_main/table"
	"time"

	"github.com/go-xorm/core"
)

const (
	CAMPAIGN_EVENT_FRESH_PREFIX = "campaign_event_fresh"
)

type CampaignEventFreshModel struct {
	Uin int
}

func (this *CampaignEventFreshModel) TableName() string {
	index := GetTableSplitIndex(this.Uin)
	return fmt.Sprintf("%s_%d", CAMPAIGN_EVENT_FRESH_PREFIX, index)
}

func (this *CampaignEventFreshModel) RefreshCampaignEventFreshTime(eventFreshInfo *table.TblCampaignEventFresh) (int, error) {
	if eventFreshInfo == nil {
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

	condition := fmt.Sprintf("uin=%d", this.Uin)
	records := make([]*table.TblCampaignEventFresh, 0)

	err := mysql.FindRecordsBySimpleCond(engine, tableName, condition, 1, 0, &records)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	rcLen := len(records)
	if rcLen > 0 {
		*eventFreshInfo = *records[0]
		eventFreshInfo.LastFreshTime = int(time.Now().Unix())
		pk := core.NewPK(eventFreshInfo.Uin)
		_, err := mysql.UpdateRecord(engine, tableName, pk, eventFreshInfo)

		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		return 0, nil
	} else {
		eventFreshInfo.Uin = this.Uin
		eventFreshInfo.LastFreshTime = int(time.Now().Unix())

		records := make([]*table.TblCampaignEventFresh, 1)
		records[0] = eventFreshInfo

		_, err := mysql.InsertRecord(engine, tableName, &records)

		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}
	}

	return 0, nil
}

func (this *CampaignEventFreshModel) GetCampaignEventFreshInfo(eventFreshInfo *table.TblCampaignEventFresh) (int, error) {
	if eventFreshInfo == nil {
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

	condition := fmt.Sprintf("uin=%d", this.Uin)
	records := make([]*table.TblCampaignEventFresh, 0)

	err := mysql.FindRecordsBySimpleCond(engine, tableName, condition, 1, 0, &records)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if rcLen := len(records); rcLen > 0 {
		*eventFreshInfo = *records[0]
		return 1, nil
	}

	return 0, nil
}
