/*
 * @Author: calmwu
 * @Date: 2017-10-17 16:56:47
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-05-19 10:37:25
 * @Comment:
 */
package mysql

import (
	"fmt"
	"reflect"
	"sailcraft/base"
	"strings"

	"github.com/go-xorm/builder"
	"github.com/go-xorm/core"
)

func InsertRecord(engine *DBEngineInfoS, tableName string, object interface{}, omitColumns ...string) (int64, error) {
	if engine != nil && len(tableName) > 0 {
		err := engine.RealEngine.Ping()
		if err == nil {
			affected, err := engine.RealEngine.Table(tableName).Omit(omitColumns...).Insert(object)
			if err != nil {
				base.GLog.Error("InsertMultiRecords Table[%s] failed! reason[%s]", tableName,
					err.Error())
				return -1, err
			}
			return affected, nil
		} else {
			return -1, err
		}
	}
	return -1, fmt.Errorf("engine is nil")
}

func InsertSliceRecordsToSameTable(engine *DBEngineInfoS, tableName string, sliceObjs interface{}) (int64, error) {
	return InsertRecord(engine, tableName, sliceObjs)
}

func UpdateRecord(engine *DBEngineInfoS, tableName string, primaryKeys *core.PK, objPtr interface{}) (affectedRows int64, err error) {
	affectedRows = 0
	err = nil

	if engine != nil && len(tableName) > 0 {
		err = engine.RealEngine.Ping()
		if err == nil {
			affectedRows, err = engine.RealEngine.Table(tableName).Id(primaryKeys).AllCols().Update(objPtr)
			return
		}
	}
	return 0, fmt.Errorf("engine is nil or tableName is empty")
}

func UpdateRecordCols(engine *DBEngineInfoS, tableName string, primaryKeys *core.PK, objPtr interface{}, updateCols ...string) (affectedRows int64, err error) {
	affectedRows = 0
	err = nil

	if engine != nil && len(tableName) > 0 {
		err = engine.RealEngine.Ping()
		if err == nil {
			session := engine.RealEngine.Table(tableName)
			session = session.Id(primaryKeys)

			if len(updateCols) > 0 {
				tableInfo := engine.RealEngine.TableInfo(objPtr)
				omitCols := make([]string, 0)
				for _, col := range tableInfo.Columns() {
					if col.IsPrimaryKey {
						continue
					}

					update := false
					for _, colName := range updateCols {
						lowerColName := strings.ToLower(colName)
						if strings.ToLower(col.Name) == lowerColName {
							update = true
							break
						}
					}

					if update {
						continue
					}

					omitCols = append(omitCols, col.Name)
				}
				session = session.Omit(omitCols...)
			}

			affectedRows, err = session.AllCols().Update(objPtr)

			return
		}
	}
	return 0, fmt.Errorf("engine is nil or tableName is empty")
}

// func UpdateRecordSpecifiedFields(engine *DBEngineInfoS, tableName string, primaryKeys *core.PK, fieldMap map[string]interface{}) (affectedRows int64, err error) {
// 	affectedRows = 0
// 	err = nil

// 	if engine != nil && len(tableName) > 0 {
// 		err = engine.RealEngine.Ping()
// 		if err == nil {
// 			// 这里是需要加入primarykey的
// 			session := engine.RealEngine.Table(tableName).Id(primaryKeys)
// 			affectedRows, err = engine.RealEngine.Table(tableName).Id(primaryKeys).Update(fieldMap)
// 			return
// 		}
// 	}
// 	return 0, fmt.Errorf("engine is nil or tableName is empty")
// }

func UpdateRecordSpecifiedFieldsByCond(engine *DBEngineInfoS, tableName string, cond string, fieldMap map[string]interface{}) (affectedRows int64, err error) {
	affectedRows = 0
	err = nil

	if engine != nil && len(tableName) > 0 && len(cond) > 0 {
		err = engine.RealEngine.Ping()
		if err == nil {
			affectedRows, err = engine.RealEngine.Table(tableName).Where(cond).Update(fieldMap)
			return
		}
	}
	return 0, fmt.Errorf("engine is nil or tableName is empty")
}

// 根据primary key 查询得到一条记录
func GetRecord(engine *DBEngineInfoS, tableName string, objPtr interface{}) (bool, error) {
	if engine != nil {
		objV := reflect.ValueOf(objPtr)
		if objV.Kind() == reflect.Ptr {
			err := engine.RealEngine.Ping()
			if err == nil {
				exist, err := engine.RealEngine.Table(tableName).Get(objPtr)
				if exist && err == nil {
					return true, nil
				} else if !exist && err == nil {
					return false, nil
				} else {
					return false, err
				}
			}
		} else {
			return false, fmt.Errorf("objPtr kind is not Ptr")
		}
	}
	return false, fmt.Errorf("engine is nil")
}

// 根据查询条件来查询一条记录
func GetRecordByCond(engine *DBEngineInfoS, tableName string, cond string, objPtr interface{}) (bool, error) {
	if engine != nil && len(cond) > 0 {
		objV := reflect.ValueOf(objPtr)
		if objV.Kind() == reflect.Ptr {
			err := engine.RealEngine.Ping()
			if err == nil {
				exist, err := engine.RealEngine.Table(tableName).Where(cond).Get(objPtr)
				if exist && err == nil {
					return true, nil
				} else if !exist && err == nil {
					return false, nil
				} else {
					return false, err
				}
			}
		} else {
			return false, fmt.Errorf("objPtr kind is not Ptr")
		}
	}
	return false, fmt.Errorf("engine is nil")
}

// cond：where条件，如果传入空字符串代表没有查询条件，如果查询条件为空，limit也是无效的
// 只支持单一条件
// tableName：查询的表名
// limit：限制大小，如果不适用填写0
// start：limit的偏移量
func FindRecordsBySimpleCond(engine *DBEngineInfoS, tableName string, cond string, limit int, start int, resultSlicePtr interface{}) error {
	if engine != nil {
		resultV := reflect.ValueOf(resultSlicePtr)
		if resultV.Kind() == reflect.Ptr {
			err := engine.RealEngine.Ping()
			if err == nil {
				if limit > 0 {
					err = engine.RealEngine.Table(tableName).Where(cond).Limit(limit, start).Find(resultSlicePtr)
				} else {
					err = engine.RealEngine.Table(tableName).Where(cond).Find(resultSlicePtr)
				}
				return err
			}
		} else {
			return fmt.Errorf("resultSlicePtr kind is not Ptr")
		}
	}
	return fmt.Errorf("engine is nil or cond is empty!")
}

func FindRecordsByMultiConds(engine *DBEngineInfoS, tableName string, cond *builder.Cond, limit int, start int, resultSlicePtr interface{}) error {
	if engine != nil {
		resultV := reflect.ValueOf(resultSlicePtr)
		if resultV.Kind() == reflect.Ptr {
			err := engine.RealEngine.Ping()
			if err == nil {
				session := engine.RealEngine.Table(tableName)
				if cond != nil {
					session = session.Where(*cond)
				}
				if limit > 0 {
					session = session.Limit(limit, start)
				}

				err = session.Find(resultSlicePtr)
				return err
			}
		} else {
			return fmt.Errorf("resultSlicePtr kind is not Ptr")
		}
	}
	return fmt.Errorf("engine is nil or cond is empty!")
}

func FindDistinctRecordsByMultiConds(engine *DBEngineInfoS, tableName string, distinctColumns []string, cond *builder.Cond, limit int, start int, resultSlicePtr interface{}) error {
	if engine != nil {
		resultV := reflect.ValueOf(resultSlicePtr)
		if resultV.Kind() == reflect.Ptr {
			err := engine.RealEngine.Ping()
			if err == nil {
				session := engine.RealEngine.Table(tableName)
				if len(distinctColumns) > 0 {
					session = session.Distinct(distinctColumns...)
				}

				if cond != nil {
					session = session.Where(*cond)
				}
				if limit > 0 {
					session = session.Limit(limit, start)
				}

				err = session.Find(resultSlicePtr)
				return err
			}
		} else {
			return fmt.Errorf("resultSlicePtr kind is not Ptr")
		}
	}
	return fmt.Errorf("engine is nil or cond is empty!")
}

// OrderBy("name desc")
func FindRecordsBySimpleCondWithOrderBy(engine *DBEngineInfoS, tableName string, cond string, limit int, start int, orderbyLst []string, resultSlicePtr interface{}) error {
	if engine != nil {
		resultV := reflect.ValueOf(resultSlicePtr)
		if resultV.Kind() == reflect.Ptr {
			err := engine.RealEngine.Ping()
			if err == nil {
				session := engine.RealEngine.Table(tableName).Where(cond)
				// 设置orderby
				if len(orderbyLst) > 0 {
					for _, orderbyContent := range orderbyLst {
						session = session.OrderBy(orderbyContent)
					}
				}
				if limit > 0 {
					session = session.Limit(limit, start)
				}
				err = session.Find(resultSlicePtr)
				return err
			}
		} else {
			return fmt.Errorf("resultSlicePtr kind is not Ptr")
		}
	}
	return fmt.Errorf("engine is nil or cond is empty!")
}

func DeleteRecordsByMultiConds(engine *DBEngineInfoS, tableName string, cond *builder.Cond, objPtr interface{}) (affectedRows int64, err error) {
	affectedRows = 0
	err = nil

	if engine != nil && len(tableName) > 0 {
		err = engine.RealEngine.Ping()
		if err == nil {
			affectedRows, err = engine.RealEngine.Table(tableName).Where(*cond).Delete(objPtr)
			return
		}
	}
	return 0, fmt.Errorf("engine is nil or tableName is empty")
}

func DeleteRecord(engine *DBEngineInfoS, tableName string, primaryKeys *core.PK, objPtr interface{}) (affectedRows int64, err error) {
	affectedRows = 0
	err = nil

	if engine != nil && len(tableName) > 0 {
		err = engine.RealEngine.Ping()
		if err == nil {
			affectedRows, err = engine.RealEngine.Table(tableName).Id(primaryKeys).Delete(objPtr)
			return
		}
	}
	return 0, fmt.Errorf("engine is nil or tableName is empty")
}

func SelectRecordsByCond(engine *DBEngineInfoS, tableName string, cond string, record interface{}) ([]interface{}, error) {
	if engine != nil && len(tableName) > 0 {
		recordT := reflect.Indirect(reflect.ValueOf(record)).Type()
		fmt.Println("recordT", recordT.String())
		// 这个不带package名字
		fmt.Println("recordT", recordT.Name())
		// // 动态创建对象
		// obj := reflect.New(t)
		// fmt.Println("-------", obj)
		// fmt.Println("-------", record)
		// fmt.Println("-------", obj.Type())
		rows, err := engine.RealEngine.Table(tableName).Where(cond).Rows(record)
		if err != nil {
			return nil, err
		} else {
			defer rows.Close()
			// 结果集
			result := make([]interface{}, 0)

			for rows.Next() {
				newRecord := reflect.New(recordT)
				err = rows.Scan(newRecord.Interface())
				if err != nil {
					return nil, err
				} else {
					result = append(result, newRecord.Interface())
				}
			}
			return result, nil
		}
	}
	return nil, fmt.Errorf("engine is nil or tableName is empty")
}

func SelectRecordsByCond2(engine *DBEngineInfoS, tableName string, cond string, tblStructName string) ([]interface{}, error) {
	if engine != nil && len(tableName) > 0 {
		record, err := NewTableObj(tblStructName)
		if err != nil {
			return nil, err
		}

		rows, err := engine.RealEngine.Table(tableName).Where(cond).Rows(record)
		if err != nil {
			return nil, err
		} else {
			defer rows.Close()
			// 结果集
			result := make([]interface{}, 0)

			for rows.Next() {
				newRecord, _ := NewTableObj(tblStructName)
				err = rows.Scan(newRecord)
				if err != nil {
					return nil, err
				} else {
					result = append(result, newRecord)
				}
			}
			return result, nil
		}
	}
	return nil, fmt.Errorf("engine is nil or tableName is empty")
}
