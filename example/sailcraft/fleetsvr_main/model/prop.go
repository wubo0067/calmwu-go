package model

import (
	"fmt"
	"sailcraft/dataaccess/mysql"
	"sailcraft/fleetsvr_main/custom_errors"
	"sailcraft/fleetsvr_main/table"
	"strconv"
	"strings"

	"github.com/go-xorm/builder"
)

const (
	TABLE_NAME_PROP_INFO = "prop_info"
)

type PropInfoModel struct {
	Uin int
}

func (this *PropInfoModel) TableName() string {
	index := GetTableSplitIndex(this.Uin)
	return fmt.Sprintf("%s_%d", TABLE_NAME_PROP_INFO, index)
}

func (this *PropInfoModel) GetMultiPropsByProtypeId(protypeIds ...int) ([]*table.TblPropInfo, error) {
	if this.Uin <= 0 {
		return nil, custom_errors.New("uin is invalid")
	}

	if len(protypeIds) <= 0 {
		return nil, custom_errors.New("protype id slice is empty")
	}

	tableName := this.TableName()
	engine := GetUinSetMysql(this.Uin)
	if engine == nil {
		return nil, custom_errors.New("engine is nil")
	}

	protypeIdSlice := make([]string, 0, len(protypeIds))
	for _, protypeId := range protypeIds {
		protypeIdSlice = append(protypeIdSlice, strconv.Itoa(protypeId))
	}
	protypeIdStr := strings.Join(protypeIdSlice, ",")

	xormCond := builder.NewCond()
	xormCond = xormCond.And(builder.Expr(fmt.Sprintf("uin=%d", this.Uin)))
	xormCond = xormCond.And(builder.Expr(fmt.Sprintf("protype_id in (%s)", protypeIdStr)))

	records := make([]*table.TblPropInfo, 0)
	err := mysql.FindRecordsByMultiConds(engine, tableName, &xormCond, 0, 0, &records)
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (this *PropInfoModel) GetPropByProtypeId(protypeId int) (*table.TblPropInfo, error) {
	if this.Uin <= 0 {
		return nil, custom_errors.New("uin is invalid")
	}

	if protypeId <= 0 {
		return nil, custom_errors.New("prop protype id is invalid")
	}

	tableName := this.TableName()
	engine := GetUinSetMysql(this.Uin)
	if engine == nil {
		return nil, custom_errors.New("engine is nil")
	}

	xormCond := builder.NewCond()
	xormCond = xormCond.And(builder.Expr(fmt.Sprintf("uin=%d", this.Uin)))
	xormCond = xormCond.And(builder.Expr(fmt.Sprintf("protype_id=%d", protypeId)))

	records := make([]*table.TblPropInfo, 0)
	err := mysql.FindRecordsByMultiConds(engine, tableName, &xormCond, 0, 0, &records)
	if err != nil {
		return nil, err
	}

	if len(records) > 0 {
		return records[0], nil
	}

	return nil, nil
}
