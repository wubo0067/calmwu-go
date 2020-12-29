package model

import (
	"encoding/json"
	"fmt"
	"sailcraft/base"
	"sailcraft/dataaccess/mysql"
	"sailcraft/fleetsvr_main/custom_errors"
	"sailcraft/fleetsvr_main/errorcode"
	"sailcraft/fleetsvr_main/table"
	"time"

	"github.com/go-xorm/core"
)

const (
	CAMPAIGN_EVENT_STATISTICS_PREFIX = "campaign_event_statistics"
)

type CampaignDailyFinishedData struct {
	TimeTriggerEventCount   int `json:"time_trigger_event"`
	PVPWinTriggerEventCount int `json:"pvp_win_trigger_event"`
}

type CampaignEventStatisticsModel struct {
	Uin int
}

func (this *CampaignEventStatisticsModel) TableName() string {
	index := GetTableSplitIndex(this.Uin)
	return fmt.Sprintf("%s_%d", CAMPAIGN_EVENT_STATISTICS_PREFIX, index)
}

func (this *CampaignEventStatisticsModel) GetStatistics(statistics *table.TblCampaignEventStatistics) (int, error) {
	if statistics == nil {
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

	condition := fmt.Sprintf("uin=%d", this.Uin)

	records := make([]*table.TblCampaignEventStatistics, 0)

	err := mysql.FindRecordsBySimpleCond(engine, tableName, condition, 1, 0, &records)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	if rcLen := len(records); rcLen > 0 {
		*statistics = *records[0]
		return 1, nil
	} else {
		return 0, nil
	}
}

func (this *CampaignEventStatisticsModel) AddTimeTriggerCountToday(count int, statistics *table.TblCampaignEventStatistics) (int, error) {
	retCode, err := this.GetStatistics(statistics)
	if err != nil {
		return retCode, err
	}

	engine := GetUinSetMysql(this.Uin)
	if engine == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("database engine is nil")
	}

	tableName := this.TableName()

	if retCode == 1 {
		var todayFinishedData CampaignDailyFinishedData

		// 判断统计时间是否是今天
		freshTime := time.Unix(int64(statistics.LastFreshTime), 0)
		todayTime := time.Now()

		if freshTime.Year() == todayTime.Year() && freshTime.YearDay() == todayTime.YearDay() {
			err = json.Unmarshal([]byte(statistics.DailyFinishedData), &todayFinishedData)
			if err != nil {
				return errorcode.ERROR_CODE_DEFAULT, err
			}

			todayFinishedData.TimeTriggerEventCount += 1
		} else {
			statistics.LastFreshTime = int(todayTime.Unix())
			todayFinishedData.TimeTriggerEventCount = 1
			todayFinishedData.PVPWinTriggerEventCount = 0
		}

		dailyFinishedData, err := json.Marshal(&todayFinishedData)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}
		statistics.DailyFinishedData = string(dailyFinishedData)

		statistics.TotalFinishedCount += 1

		pk := core.NewPK(statistics.Uin)
		_, err = mysql.UpdateRecord(engine, tableName, pk, statistics)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		return 0, nil
	} else {
		var todayFinishData CampaignDailyFinishedData
		todayFinishData.TimeTriggerEventCount = 1
		todayFinishData.PVPWinTriggerEventCount = 0

		dailyFinishedData, err := json.Marshal(todayFinishData)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		statistics.Uin = this.Uin
		statistics.LastFreshTime = int(time.Now().Unix())
		statistics.TotalFinishedCount = 1
		statistics.DailyFinishedData = string(dailyFinishedData)

		_, err = mysql.InsertRecord(engine, tableName, statistics)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		return 0, nil
	}
}

func (this *CampaignEventStatisticsModel) AddPVPWinTriggerCountToday(count int, statistics *table.TblCampaignEventStatistics) (int, error) {
	retCode, err := this.GetStatistics(statistics)
	if err != nil {
		return retCode, err
	}

	engine := GetUinSetMysql(this.Uin)
	if engine == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("database engine is nil")
	}

	tableName := this.TableName()

	if retCode == 1 {
		var todayFinishedData CampaignDailyFinishedData

		// 判断统计时间是否是今天
		freshTime := time.Unix(int64(statistics.LastFreshTime), 0)
		todayTime := time.Now()

		if freshTime.Year() == todayTime.Year() && freshTime.YearDay() == todayTime.YearDay() {
			err = json.Unmarshal([]byte(statistics.DailyFinishedData), &todayFinishedData)
			if err != nil {
				return errorcode.ERROR_CODE_DEFAULT, err
			}

			todayFinishedData.PVPWinTriggerEventCount += 1
		} else {
			statistics.LastFreshTime = int(todayTime.Unix())
			todayFinishedData.TimeTriggerEventCount = 0
			todayFinishedData.PVPWinTriggerEventCount = 1
		}

		dailyFinishedData, err := json.Marshal(&todayFinishedData)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}
		statistics.DailyFinishedData = string(dailyFinishedData)

		statistics.TotalFinishedCount += 1

		pk := core.NewPK(statistics.Uin)
		_, err = mysql.UpdateRecord(engine, tableName, pk, statistics)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		return 0, nil
	} else {
		var todayFinishData CampaignDailyFinishedData
		todayFinishData.PVPWinTriggerEventCount = 1
		todayFinishData.TimeTriggerEventCount = 0

		dailyFinishedData, err := json.Marshal(todayFinishData)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		statistics.Uin = this.Uin
		statistics.LastFreshTime = int(time.Now().Unix())
		statistics.TotalFinishedCount = 1
		statistics.DailyFinishedData = string(dailyFinishedData)

		_, err = mysql.InsertRecord(engine, tableName, statistics)
		if err != nil {
			return errorcode.ERROR_CODE_DEFAULT, err
		}

		return 0, nil
	}
}
