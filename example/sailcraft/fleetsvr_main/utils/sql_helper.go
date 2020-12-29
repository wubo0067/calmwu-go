package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sailcraft/base"
	"strings"

	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
)

func UpdateMultiRecords(engine *xorm.Engine, tableName string, sliceObjPtr interface{}, omitColumns ...string) (int, error) {
	if engine == nil {
		return 0, fmt.Errorf("engine is nil")
	}

	if tableName == "" {
		return 0, fmt.Errorf("table name is empty")
	}

	sliceValue := reflect.Indirect(reflect.ValueOf(sliceObjPtr))
	if sliceValue.Kind() != reflect.Slice {
		return 0, fmt.Errorf("needs a pointer to a slice")
	}

	tableInfo := engine.TableInfo(sliceValue.Index(0).Interface())
	if len(tableInfo.PrimaryKeys) != 1 {
		return 0, fmt.Errorf("only support one primary key")
	}

	if sliceValue.Len() <= 0 {
		return 0, fmt.Errorf("could not insert a empty slice")
	}

	pkName := tableInfo.PrimaryKeys[0]
	pkValues := make([]string, 0)
	pkFieldName := tableInfo.PKColumns()[0].FieldName

	columns := tableInfo.Columns()

	updateColumns := make([]*core.Column, 0)
	for _, col := range columns {
		if col.Name == pkName {
			continue
		}

		isOmit := false
		for _, omit := range omitColumns {
			if col.Name == omit {
				isOmit = true
				break
			}
		}

		if isOmit {
			continue
		}

		updateColumns = append(updateColumns, col)
	}

	if len(updateColumns) <= 0 {
		return 0, fmt.Errorf("without columns to update")
	}

	whenCaseParam := make(map[string](map[string]string))

	for i := 0; i < sliceValue.Len(); i++ {
		v := reflect.Indirect(sliceValue.Index(i))
		pkValue := interfaceToSqlStr(v.FieldByName(pkFieldName).Interface())
		pkValues = append(pkValues, pkValue)

		for _, col := range updateColumns {
			if col.Name != pkName {
				if _, ok := whenCaseParam[col.Name]; !ok {
					whenCaseParam[col.Name] = make(map[string]string)
				}

				whenCaseParam[col.Name][pkValue] = interfaceToSqlStr(v.FieldByName(col.FieldName).Interface())
			}
		}
	}

	caseWhenSlice := make([]string, 0)
	for colName, valueCase := range whenCaseParam {
		singleCaseWhen := fmt.Sprintf("%s=CASE %s", colName, pkName)
		for pkValue, colValue := range valueCase {
			singleCaseWhen = fmt.Sprintf("%s WHEN %s THEN %s", singleCaseWhen, pkValue, colValue)
		}

		caseWhenSlice = append(caseWhenSlice, singleCaseWhen)
	}

	sqlStr := fmt.Sprintf("UPDATE %s SET %s END WHERE %v IN (%s)", tableName, strings.Join(caseWhenSlice, " END, "), pkName, strings.Join(pkValues, ","))

	err := engine.Ping()
	if err != nil {
		return 0, err
	}

	result, err := engine.Exec(sqlStr)
	if err != nil {
		return 0, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(affected), nil
}

func interfaceToSqlStr(i interface{}) string {
	v := reflect.Indirect(reflect.ValueOf(i))
	switch v.Kind() {
	case reflect.String:
		return fmt.Sprintf("'%s'", v.String())
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		return fmt.Sprintf("%d", v.Int())
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint, reflect.Uintptr:
		return fmt.Sprintf("%d", v.Uint())
	case reflect.Struct:
		data, err := json.Marshal(i)
		if err != nil {
			base.GLog.Error(err)
			return ""
		}

		return string(data)
	}

	return ""
}
