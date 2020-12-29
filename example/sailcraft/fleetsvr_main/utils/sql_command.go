package utils

import (
	"errors"
	"strconv"
)

const (
	SQL_INSERT = "insert"
	SQL_DELETE = "delete"
	SQL_SELECT = "select"
	SQL_UPDATE = "update"
)

type QueryCondition struct {
	Attr  string
	Op    string
	Value interface{}
}

type SqlCommands struct {
	Action    string
	TableName string
	Fields    []string
	Where     []*QueryCondition
	Data      map[string]interface{}
	Limit     int
}

func QuoteString(value string) string {
	return "'" + value + "'"
}

func GetInterfaceValue(value interface{}) string {
	retValue := ""
	switch value.(type) {
	case int8:
		data := value.(int8)
		retValue = strconv.Itoa(int(data))
	case int16:
		data := value.(int16)
		retValue = strconv.Itoa(int(data))
	case int32:
		data := value.(int32)
		retValue = strconv.Itoa(int(data))
	case int64:
		data := value.(int64)
		retValue = strconv.FormatInt(data, 10)
	case int:
		data := value.(int)
		retValue = strconv.Itoa(data)
	case float32:
		data := value.(float32)
		retValue = strconv.FormatFloat(float64(data), 'E', -1, 64)
	case float64:
		data := value.(float64)
		retValue = strconv.FormatFloat(data, 'E', -1, 64)
	case string:
		data := value.(string)
		retValue = QuoteString(data)
	default:
		// 其他类型默认按照字符串处理
		data := value.(string)
		retValue = QuoteString(data)
	}

	return retValue
}

// 默认where全走and操作
func GenWhereClause(conds []*QueryCondition) string {
	where := ""
	for index, queryCondition := range conds {
		if index > 0 {
			where = where + " and "
		} else {
			where = " where "
		}
		where = where + queryCondition.Attr + queryCondition.Op + GetInterfaceValue(queryCondition.Value)
	}

	return where
}

func GenLimitClause(limit int) string {
	strLimit := ""
	if limit > 0 {
		strLimit = strLimit + " limit " + strconv.Itoa(limit)
	}
	return strLimit
}

func GenFromClause(tableName string) string {
	return " from " + tableName
}

func GenSelectClause(fields []string) string {
	return "select " + JoinString(fields, ",")
}

func (sqlCommands *SqlCommands) GenSelectSql() string {
	strSql := GenSelectClause(sqlCommands.Fields)
	strSql = strSql + GenFromClause(sqlCommands.TableName)

	strWhere := GenWhereClause(sqlCommands.Where)
	if strWhere != "" {
		strSql = strSql + strWhere
	}

	strLimit := GenLimitClause(sqlCommands.Limit)
	if strLimit != "" {
		strSql = strSql + strLimit
	}

	return strSql
}

func (sqlCommands *SqlCommands) GenInsertSql() string {
	strSql := "insert into " + sqlCommands.TableName

	strFields := ""
	strValues := ""
	for field, value := range sqlCommands.Data {
		if strFields == "" {
			strFields = field
			strValues = GetInterfaceValue(value)
		} else {
			strFields = strFields + "," + field
			strValues = strValues + "," + GetInterfaceValue(value)
		}
	}

	strSql = strSql + "(" + strFields + ")" + " values" + "(" + strValues + ")"

	return strSql
}

func (sqlCommands *SqlCommands) GenUpdateSql() string {
	strSql := "update " + sqlCommands.TableName

	strSetValues := " set "
	for field, value := range sqlCommands.Data {
		if strSetValues == "" {
			strSetValues = field + "=" + GetInterfaceValue(value)
		} else {
			strSetValues = strSetValues + "," + field + "=" + GetInterfaceValue(value)
		}
	}

	strSql = strSql + strSetValues

	strWhere := GenWhereClause(sqlCommands.Where)
	if strWhere != "" {
		strSql = strSql + strWhere
	}

	return strSql
}

func (sqlCommands *SqlCommands) GenDeleteSql() string {
	strSql := "delete from " + sqlCommands.TableName

	strWhere := GenWhereClause(sqlCommands.Where)
	if strWhere != "" {
		strSql = strSql + strWhere
	}

	return strSql
}

func (sqlCommands *SqlCommands) GenSql() (string, error) {
	strSql := ""

	switch sqlCommands.Action {
	case SQL_SELECT:
		strSql = sqlCommands.GenSelectSql()
	case SQL_INSERT:
		strSql = sqlCommands.GenInsertSql()
	case SQL_UPDATE:
		strSql = sqlCommands.GenUpdateSql()
	case SQL_DELETE:
		strSql = sqlCommands.GenDeleteSql()
	default:
		return strSql, errors.New("sql only support insert delete update select")
	}

	if strSql == "" {
		return strSql, errors.New("strSql is empty")
	}

	result := SqlQuote(strSql)
	if result {
		return strSql, nil
	}

	return strSql, errors.New("strSql is wrong")
}
