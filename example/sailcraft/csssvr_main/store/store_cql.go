/*
 * @Author: calmwu
 * @Date: 2018-01-11 15:22:50
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-01-11 15:25:19
 * @Comment:
 */

package store

import (
	"fmt"
	"reflect"
	"sailcraft/base"
	"strconv"
	"time"

	"github.com/emirpasic/gods/sets/hashset"
	"github.com/gocql/gocql"
	"github.com/mitchellh/mapstructure"
)

func queryUserOnlineTable(session *gocql.Session, uin int) *TblUserOnineS {
	cqlSelectUserOnline := fmt.Sprintf("SELECT * FROM tbl_UserOnline WHERE uin=%d", uin)
	userOnlineRes := QueryRecords(session, cqlSelectUserOnline)
	if userOnlineRes != nil && len(userOnlineRes) > 0 {
		var userOnline TblUserOnineS
		err := decodeTblUserOnlineRecord(userOnlineRes[0], &userOnline)
		if err != nil {
			base.GLog.Error("Decode TblUserOnineS object failed! reason[%s]", err.Error())
		} else {
			return &userOnline
		}
	}
	return nil
}

func QueryRecords(session *gocql.Session, cqlContent string) []map[string]interface{} {
	if session == nil {
		return nil
	}

	iter := session.Query(cqlContent).Iter()
	defer iter.Close()

	result, err := iter.SliceMap()
	if err != nil {
		base.GLog.Error("Query[%s] failed! error[%s]\n", cqlContent, err.Error())
	} else {
		base.GLog.Debug("Query[%s] successed!", cqlContent)
		if len(result) == 0 {
			base.GLog.Warn("Query[%s] result is empty", cqlContent)
		}
		return result
	}

	return nil
}

func execCql(session *gocql.Session, cqlContent string) {
	if session != nil {
		query := session.Query(cqlContent)

		if err := query.Exec(); err != nil {
			base.GLog.Error("exec[%s] failed! reason[%s]", cqlContent, err.Error())
		} else {
			base.GLog.Debug("exec[%s] successed!", cqlContent)
		}
	} else {
		base.GLog.Error("session is invalid!")
	}
	return
}

func genUpdateCql(tbName string, tbObjPtr interface{}, keys *hashset.Set) (string, error) {
	if keys.Empty() {
		base.GLog.Error("keys is empty!")
		return "", fmt.Errorf("%s", "keys is empty!")
	}

	updateCql := "UPDATE " + tbName + " SET "
	whereCond := " WHERE "

	v_t := reflect.TypeOf(tbObjPtr).Elem()
	v_v := reflect.ValueOf(tbObjPtr).Elem()

	var fieldIndex = 0
	for fieldIndex < v_t.NumField() {
		field := v_t.Field(fieldIndex)

		fieldName := field.Name
		fieldVal := v_v.FieldByName(fieldName)

		if !keys.Contains(fieldName) {
			updateCql += fieldName + "="
			if field.Type.Kind() == reflect.String {
				updateCql += "'" + fmt.Sprintf("%v", fieldVal.Interface()) + "', "
			} else if field.Type.Kind() == reflect.Array ||
				field.Type.Kind() == reflect.Slice {
				if fieldVal.Len() > 0 {
					updateCql += "["
					var index = 0
					for index < fieldVal.Len() {
						// 数组元素
						fieldValAryUnitVal := fieldVal.Index(index)
						if fieldValAryUnitVal.Kind() == reflect.String {
							updateCql += fmt.Sprintf("'%v'", fieldValAryUnitVal.Interface()) + ", "
						} else {
							updateCql += fmt.Sprintf("%v", fieldValAryUnitVal.Interface()) + ", "
						}
						index++
					}
					updateCql = updateCql[:len(updateCql)-2]
					updateCql += "], "
					//fmt.Println(updateCql)
				}
			} else if field.Type.Kind() == reflect.Map {
				if fieldVal.Len() > 0 {
					updateCql += "{"
					fieldValMapKeys := fieldVal.MapKeys()
					for index := range fieldValMapKeys {
						fieldValMapKey := fieldValMapKeys[index]
						if fieldValMapKey.Kind() == reflect.String {
							updateCql += "'" + fieldValMapKeys[index].String() + "' : "
						} else {
							// TODO，这里特别说明，只支持string和整型
							updateCql += strconv.Itoa(int(fieldValMapKeys[index].Int())) + " : "
						}
						fieldValMapVal := fieldVal.MapIndex(fieldValMapKey)
						if fieldValMapVal.Kind() == reflect.String {
							updateCql += fmt.Sprintf("'%v'", fieldValMapVal.String()) + ", "
						} else {
							updateCql += fmt.Sprintf("%v", fieldValMapVal.Interface()) + ", "
						}
					}
					updateCql = updateCql[:len(updateCql)-2]
					updateCql += "}, "
				}
			} else {
				updateCql += fmt.Sprintf("%v", fieldVal.Interface()) + ", "
			}
		} else {
			if field.Type.Kind() == reflect.String {
				whereCond += fieldName + "=" + fmt.Sprintf("'%s'", fieldVal) + " AND "
			} else {
				whereCond += fieldName + "=" + fmt.Sprintf("%v", fieldVal) + " AND "
			}

		}
		fieldIndex++
	}

	updateCql = updateCql[:len(updateCql)-2]
	whereCond = whereCond[:len(whereCond)-5]
	updateCql += whereCond

	return updateCql, nil
}

func decodeCATimeField(from reflect.Type, to reflect.Type, v interface{}) (interface{}, error) {
	if from.String() == "time.Time" {
		//base.GLog.Debug("tblUserStatisticsInfo convert time field!")
		now := v.(time.Time)
		return int64(now.UnixNano() / int64(time.Millisecond)), nil
	}
	return v, nil
}

func customDecodeCARecord(record map[string]interface{}, config *mapstructure.DecoderConfig) error {
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		base.GLog.Error("NewDecoder failed! reason[%s]", err.Error())
		return err
	}

	err = decoder.Decode(record)
	if err != nil {
		base.GLog.Error("Decode TblUserInfo record failed! reason[%s]", err.Error())
		return err
	}
	return nil
}

func decodeTblUserOnlineRecord(record map[string]interface{}, userOnline *TblUserOnineS) error {

	// decodeHook := func(from reflect.Type, to reflect.Type, v interface{}) (interface{}, error) {
	// 	if from.String() == "time.Time" {
	// 		//base.GLog.Debug("tblUserStatisticsInfo convert time field!")
	// 		now := v.(time.Time)
	// 		return int64(now.UnixNano() / 1000000), nil
	// 	}
	// 	return v, nil
	// }

	config := &mapstructure.DecoderConfig{
		DecodeHook: decodeCATimeField,
		Result:     userOnline,
	}

	return customDecodeCARecord(record, config)

	// decoder, err := mapstructure.NewDecoder(config)
	// if err != nil {
	// 	base.GLog.Error("NewDecoder failed! reason[%s]", err.Error())
	// 	return err
	// }

	// err = decoder.Decode(record)
	// if err != nil {
	// 	base.GLog.Error("Decode TblUserInfo record failed! reason[%s]", err.Error())
	// 	return err
	// }
	// return nil
}

// func decodeTblCDKeyInfoRecord(record map[string]interface{}, cdKeyInfo *TblCDKeyInfoS) error {
// 	config := &mapstructure.DecoderConfig{
// 		DecodeHook: decodeCATimeField,
// 		Result:     cdKeyInfo,
// 	}

// 	return customDecodeCARecord(record, config)
// }

// func decodeTblCDKeyBatchInfoRecord(record map[string]interface{}, cdKeyBatchInfo *TblCDKeyBatchInfoS) error {
// 	config := &mapstructure.DecoderConfig{
// 		DecodeHook: decodeCATimeField,
// 		Result:     cdKeyBatchInfo,
// 	}

// 	return customDecodeCARecord(record, config)
// }
