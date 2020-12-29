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
	BATTLE_SHIP_TABLE_NAME = "battle_ship"
)

type BattleShipModel struct {
	Uin int
}

func (battleShipModel *BattleShipModel) TableName() string {
	index := GetTableSplitIndex(battleShipModel.Uin)
	return fmt.Sprintf("%s_%d", BATTLE_SHIP_TABLE_NAME, index)
}

func (battleShipModel *BattleShipModel) InitBattleShip(shipIDs []int) (affected int64, err error) {
	affected = -1
	err = nil

	if len(shipIDs) == 0 {
		err = custom_errors.New("user default battle ships can not be empty")
		return
	}

	engine := GetUinSetMysql(battleShipModel.Uin)
	if engine == nil {
		err = custom_errors.New("database engine is nil")
		return
	}

	tableName := battleShipModel.TableName()

	records := make([]*table.TblBattleShip, len(shipIDs))
	for index, protype_id := range shipIDs {
		// 默认是碎片状态
		record := table.NewDefaultBattleShip()
		record.Uin = battleShipModel.Uin
		record.ProtypeID = protype_id
		record.Status = table.BATTLE_SHIP_STATUS_SHIP
		record.Level = 1

		records[index] = record
	}

	base.GLog.Debug("database is [%s] table is [%s]", engine.DBEngineDataSouceName, tableName)

	affected, err = mysql.InsertRecord(engine, tableName, &records)

	return
}

func (battleShipModel *BattleShipModel) AddBattleShip(records []*table.TblBattleShip) (affected int64, err error) {
	affected = -1
	err = nil

	if len(records) == 0 {
		err = custom_errors.New("user default battle ships can not be empty")
		return
	}

	engine := GetUinSetMysql(battleShipModel.Uin)
	if engine == nil {
		err = custom_errors.New("database engine is nil")
		return
	}

	tableName := battleShipModel.TableName()

	base.GLog.Debug("database is [%s] table is [%s]", engine.DBEngineDataSouceName, tableName)

	affected, err = mysql.InsertRecord(engine, tableName, &records)

	return
}

func (battleShipModel *BattleShipModel) GetBattleShipList() ([]*table.TblBattleShip, error) {
	uin := battleShipModel.Uin

	if uin <= 0 {
		return nil, custom_errors.New("uin is invalid")
	}

	engine := GetUinSetMysql(uin)
	if engine == nil {
		return nil, custom_errors.New("database engine is nil")
	}

	tableName := battleShipModel.TableName()

	condtion := fmt.Sprintf("uin=%d", uin)

	records := make([]*table.TblBattleShip, 0)
	err := mysql.FindRecordsBySimpleCond(engine, tableName, condtion, 0, 0, &records)
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (battleShipModel *BattleShipModel) GetBattleShipInfoByID(id int, battleShipInfo *table.TblBattleShip) (int, error) {
	base.GLog.Debug("GetBattleShipInfoByID enter ship_id %d", id)

	if battleShipInfo == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("null point")
	}

	uin := battleShipModel.Uin

	if uin <= 0 {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("uin is invalid")
	}

	engine := GetUinSetMysql(uin)
	if engine == nil {
		return errorcode.ERROR_CODE_DEFAULT, custom_errors.New("database engine is nil")
	}

	tableName := battleShipModel.TableName()

	condtion := fmt.Sprintf("id=%d", id)

	exist, err := mysql.GetRecordByCond(engine, tableName, condtion, battleShipInfo)
	if err != nil {
		return errorcode.ERROR_CODE_DEFAULT, err
	}

	// 返回1表示数据存在
	if exist {
		return 1, nil
	} else {
		return 0, nil
	}
}

func (battleShipModel *BattleShipModel) UpgradeBattleShipInfoByID(battleShipInfo *table.TblBattleShip) error {
	if battleShipInfo == nil {
		return custom_errors.New("battleShipInfo null point")
	}

	base.GLog.Debug("UpgradeBattleShipByID enter ship_id %d level %d", battleShipInfo.Id, battleShipInfo.Level)

	uin := battleShipModel.Uin

	engine := GetUinSetMysql(uin)
	if engine == nil {
		return custom_errors.New("database engine is nil")
	}

	tableName := battleShipModel.TableName()

	PK := core.NewPK(battleShipInfo.Id)
	_, err := mysql.UpdateRecord(engine, tableName, PK, battleShipInfo)
	if err != nil {
		return err
	}

	return nil
}
