/*
 * @Author: calmwu
 * @Date: 2017-10-27 15:37:01
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-10-27 17:29:11
 * @Comment:
 */
package redistool

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	c_TIME_DEFAULT time.Time
	TimeType       = reflect.TypeOf(c_TIME_DEFAULT)
)

func ConvertSliceToRedisList(sliceObj interface{}) ([]string, error) {
	sliceValue := reflect.Indirect(reflect.ValueOf(sliceObj))
	if sliceValue.Kind() == reflect.Slice {
		if sliceValue.Len() <= 0 {
			return nil, errors.New("could not convert a empty slice")
		}

		// 获取slice的元素类型
		//fmt.Printf("slice elem type:%v", sliceValue.Type().Elem().Kind())

		redisList := make([]string, sliceValue.Len())
		var i = 0

		switch sliceValue.Type().Elem().Kind() {
		case reflect.String:
			return nil, errors.New("string array does not require conversion")
		case reflect.Int:
			for i < sliceValue.Len() {
				redisList[i] = strconv.FormatInt(sliceValue.Index(i).Int(), 10)
				i++
			}
		case reflect.Int32:
			for i < sliceValue.Len() {
				redisList[i] = strconv.FormatInt(sliceValue.Index(i).Int(), 10)
				i++
			}
		case reflect.Int64:
			for i < sliceValue.Len() {
				redisList[i] = strconv.FormatInt(sliceValue.Index(i).Int(), 10)
				i++
			}
		case reflect.Uint:
			for i < sliceValue.Len() {
				redisList[i] = strconv.FormatUint(sliceValue.Index(i).Uint(), 10)
				i++
			}
		case reflect.Uint32:
			for i < sliceValue.Len() {
				redisList[i] = strconv.FormatUint(sliceValue.Index(i).Uint(), 10)
				i++
			}
		case reflect.Uint64:
			for i < sliceValue.Len() {
				redisList[i] = strconv.FormatUint(sliceValue.Index(i).Uint(), 10)
				i++
			}
		case reflect.Float32:
			for i < sliceValue.Len() {
				redisList[i] = strconv.FormatFloat(sliceValue.Index(i).Float(), 'e', 3, 32)
				i++
			}
		case reflect.Float64:
			for i < sliceValue.Len() {
				redisList[i] = strconv.FormatFloat(sliceValue.Index(i).Float(), 'e', 3, 64)
				i++
			}
		case reflect.Bool:
			for i < sliceValue.Len() {
				redisList[i] = strconv.FormatBool(sliceValue.Index(i).Bool())
				i++
			}
		}

		return redisList, nil
	}
	return nil, errors.New("sliceObj kind is not slice")
}

func ConvertRedisListToSlice(redisL []string, sliceObj interface{}) error {
	sliceValue := reflect.Indirect(reflect.ValueOf(sliceObj))
	if sliceValue.Kind() == reflect.Slice {
		if len(redisL) != sliceValue.Len() {
			return errors.New("sliceObj len is not equal to redisL")
		}

		sliceValueElemKind := sliceValue.Type().Elem().Kind()
		if reflect.String == sliceValueElemKind {
			return errors.New("string array does not require conversion")
		}

		//fmt.Println(sliceValueElemKind)
		var i = 0

		switch sliceValueElemKind {
		case reflect.Int:
			for i < len(redisL) {
				redisV := redisL[i]
				if len(redisV) == 0 {
					redisV = "0"
				}
				num, err := strconv.ParseInt(redisV, 10, 64)
				if err != nil {
					return err
				}
				sliceValue.Index(i).SetInt(num)
				i++
			}
		case reflect.Int32:
			for i < len(redisL) {
				redisV := redisL[i]
				if len(redisV) == 0 {
					redisV = "0"
				}
				num, err := strconv.ParseInt(redisV, 10, 32)
				if err != nil {
					return err
				}
				sliceValue.Index(i).SetInt(num)
				i++
			}
		case reflect.Int64:
			for i < len(redisL) {
				redisV := redisL[i]
				if len(redisV) == 0 {
					redisV = "0"
				}
				num, err := strconv.ParseInt(redisV, 10, 64)
				if err != nil {
					return err
				}
				sliceValue.Index(i).SetInt(num)
				i++
			}
		case reflect.Uint:
			for i < len(redisL) {
				redisV := redisL[i]
				if len(redisV) == 0 {
					redisV = "0"
				}
				num, err := strconv.ParseUint(redisV, 10, 64)
				if err != nil {
					return err
				}
				sliceValue.Index(i).SetUint(num)
				i++
			}
		case reflect.Uint32:
			for i < len(redisL) {
				redisV := redisL[i]
				if len(redisV) == 0 {
					redisV = "0"
				}
				num, err := strconv.ParseUint(redisV, 10, 32)
				if err != nil {
					return err
				}
				sliceValue.Index(i).SetUint(num)
				i++
			}
		case reflect.Uint64:
			for i < len(redisL) {
				redisV := redisL[i]
				if len(redisV) == 0 {
					redisV = "0"
				}
				num, err := strconv.ParseUint(redisV, 10, 64)
				if err != nil {
					return err
				}
				sliceValue.Index(i).SetUint(num)
				i++
			}
		case reflect.Float32:
			for i < len(redisL) {
				redisV := redisL[i]
				if len(redisV) == 0 {
					redisV = "0.000"
				}
				num, err := strconv.ParseFloat(redisV, 32)
				if err != nil {
					return err
				}
				sliceValue.Index(i).SetFloat(num)
				i++
			}
		case reflect.Float64:
			for i < len(redisL) {
				redisV := redisL[i]
				if len(redisV) == 0 {
					redisV = "0.000"
				}
				num, err := strconv.ParseFloat(redisV, 64)
				if err != nil {
					return err
				}
				sliceValue.Index(i).SetFloat(num)
				i++
			}
		case reflect.Bool:
			for i < len(redisL) {
				redisV := redisL[i]
				if len(redisV) == 0 {
					redisV = "false"
				}
				b, err := strconv.ParseBool(redisV)
				if err != nil {
					return err
				}
				sliceValue.Index(i).SetBool(b)
				i++
			}
		}
		return nil
	}
	return errors.New("sliceObj kind is not slice")
}

func ConvertObjToRedisHash(obj interface{}) (map[string]string, error) {
	redisHash := make(map[string]string)

	v := reflect.Indirect(reflect.ValueOf(obj))
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldName := field.Name
		fieldValue := v.FieldByName(fieldName)

		fieldTagStr := field.Tag.Get("xorm")
		if fieldTagStr != "" {
			//fmt.Printf("fieldTagStr:%s name:[%s]\n", fieldTagStr, getNameFromTag(fieldTagStr))
			tagName := getNameFromTag(fieldTagStr)
			if tagName != "" {
				fieldName = tagName
			}
		}

		fmt.Printf("fileName[%s] type:%v value:%v valueType:%v kind:%v\n",
			fieldName, field.Type, fieldValue, fieldValue.Type(), field.Type.Kind())
		switch field.Type.Kind() {
		case reflect.String:
			redisHash[fieldName] = fieldValue.String()
		case reflect.Int:
			redisHash[fieldName] = strconv.FormatInt(fieldValue.Int(), 10)
		case reflect.Int32:
			redisHash[fieldName] = strconv.FormatInt(fieldValue.Int(), 10)
		case reflect.Int64:
			redisHash[fieldName] = strconv.FormatInt(fieldValue.Int(), 10)
		case reflect.Uint:
			redisHash[fieldName] = strconv.FormatUint(fieldValue.Uint(), 10)
		case reflect.Uint32:
			redisHash[fieldName] = strconv.FormatUint(fieldValue.Uint(), 10)
		case reflect.Uint64:
			redisHash[fieldName] = strconv.FormatUint(fieldValue.Uint(), 10)
		case reflect.Bool:
			redisHash[fieldName] = strconv.FormatBool(fieldValue.Bool())
		case reflect.Float32:
			// FormatFloat 将浮点数 f 转换为字符串形式
			// f：要转换的浮点数
			// fmt：格式标记（b、e、E、f、g、G）
			// prec：精度（数字部分的长度，不包括指数部分）
			// bitSize：指定浮点类型（32:float32、64:float64），结果会据此进行舍入。
			//
			// 格式标记：
			// 'b' (-ddddp±ddd，二进制指数)
			// 'e' (-d.dddde±dd，十进制指数)
			// 'E' (-d.ddddE±dd，十进制指数)
			// 'f' (-ddd.dddd，没有指数)
			// 'g' ('e':大指数，'f':其它情况)
			// 'G' ('E':大指数，'f':其它情况)
			//
			// 如果格式标记为 'e'，'E'和'f'，则 prec 表示小数点后的数字位数
			// 如果格式标记为 'g'，'G'，则 prec 表示总的数字位数（整数部分+小数部分）
			// 参考格式化输入输出中的旗标和精度说明
			redisHash[fieldName] = strconv.FormatFloat(fieldValue.Float(), 'e', 3, 32)
		case reflect.Float64:
			redisHash[fieldName] = strconv.FormatFloat(fieldValue.Float(), 'e', 8, 64)
		case reflect.Struct:
			if field.Type.String() == "time.Time" && field.Type.ConvertibleTo(TimeType) {
				t := fieldValue.Convert(TimeType).Interface().(time.Time)
				redisHash[fieldName] = t.Format("2006-01-02 15:04:05") //strconv.FormatInt(t.Unix(), 10)
			}
		default:
			return nil, fmt.Errorf("fieldName[%s] type:%v is not support!", fieldName, field.Type)
		}
	}

	return redisHash, nil
}

func ConvertRedisHashToObj(hashV map[string]string, objP interface{}) error {
	v := reflect.ValueOf(objP)
	if v.Kind() == reflect.Ptr {
		// 排除指针，指向原始类型
		t := reflect.TypeOf(objP).Elem()
		//--------- redis.Server *redis.Server
		//fmt.Println("---------", t, v.Type())
		i := 0
		for i < t.NumField() {
			field := t.Field(i)
			fieldName := field.Name

			fieldTagStr := field.Tag.Get("xorm")
			if fieldTagStr != "" {
				//fmt.Printf("fieldTagStr:%s name:[%s]\n", fieldTagStr, getNameFromTag(fieldTagStr))
				tagName := getNameFromTag(fieldTagStr)
				if tagName != "" {
					fieldName = tagName
				}
			}

			//fmt.Println("field ", i, "Name is", fieldName, "Type is", field.Type)

			if redisV, ok := hashV[fieldName]; !ok {
				fmt.Printf("field[%s] value is not exists!", fieldName)
				i++
				continue
			} else {
				if v.Elem().Field(i).CanSet() {
					switch field.Type.Kind() {
					case reflect.String:
						v.Elem().Field(i).SetString(redisV)
					case reflect.Int:
						num, err := strconv.ParseInt(redisV, 10, 64)
						if err != nil {
							return err
						}
						v.Elem().Field(i).SetInt(num)
					case reflect.Int32:
						num, err := strconv.ParseInt(redisV, 10, 32)
						if err != nil {
							return err
						}
						v.Elem().Field(i).SetInt(num)
					case reflect.Int64:
						num, err := strconv.ParseInt(redisV, 10, 64)
						if err != nil {
							return err
						}
						v.Elem().Field(i).SetInt(num)
					case reflect.Uint:
						num, err := strconv.ParseUint(redisV, 10, 64)
						if err != nil {
							return err
						}
						v.Elem().Field(i).SetUint(num)
					case reflect.Uint32:
						num, err := strconv.ParseUint(redisV, 10, 32)
						if err != nil {
							return err
						}
						v.Elem().Field(i).SetUint(num)
					case reflect.Uint64:
						num, err := strconv.ParseUint(redisV, 10, 64)
						if err != nil {
							return err
						}
						v.Elem().Field(i).SetUint(num)
					case reflect.Float32:
						// FormatFloat 将浮点数 f 转换为字符串形式
						// f：要转换的浮点数
						// fmt：格式标记（b、e、E、f、g、G）
						// prec：精度（数字部分的长度，不包括指数部分）
						// bitSize：指定浮点类型（32:float32、64:float64），结果会据此进行舍入。
						//
						// 格式标记：
						// 'b' (-ddddp±ddd，二进制指数)
						// 'e' (-d.dddde±dd，十进制指数)
						// 'E' (-d.ddddE±dd，十进制指数)
						// 'f' (-ddd.dddd，没有指数)
						// 'g' ('e':大指数，'f':其它情况)
						// 'G' ('E':大指数，'f':其它情况)
						//
						// 如果格式标记为 'e'，'E'和'f'，则 prec 表示小数点后的数字位数
						// 如果格式标记为 'g'，'G'，则 prec 表示总的数字位数（整数部分+小数部分）
						// 参考格式化输入输出中的旗标和精度说明
						fNum, err := strconv.ParseFloat(redisV, 32)
						if err != nil {
							return err
						}
						v.Elem().Field(i).SetFloat(fNum)
					case reflect.Float64:
						fNum, err := strconv.ParseFloat(redisV, 64)
						if err != nil {
							return err
						}
						v.Elem().Field(i).SetFloat(fNum)
					case reflect.Bool:
						b, err := strconv.ParseBool(redisV)
						if err != nil {
							return err
						}
						v.Elem().Field(i).SetBool(b)
					case reflect.Struct:
						fieldV := v.Elem().Field(i)
						if field.Type.String() == "time.Time" {
							timeV, err := time.Parse("2006-01-02 15:04:05", redisV)
							if err != nil {
								return err
							}
							fieldV.Set(reflect.ValueOf(timeV))
						}
					}
				} else {
					return fmt.Errorf("field[%s] unexport!", field.Name)
				}
			}
			i++
		}
		return nil
	} else {
		return fmt.Errorf("objP type.kind is not reflect.Ptr")
	}
}

func getNameFromTag(tagStr string) string {
	tagStr = strings.TrimSpace(tagStr)
	tagStrSize := len(tagStr)
	hashQuote := false
	lastIndex := 0

	for i := tagStrSize - 1; i >= 0; i-- {
		if tagStr[i] == '\'' {
			if !hashQuote {
				hashQuote = true
				lastIndex = i
			} else {
				return strings.TrimSpace(tagStr[i+1 : lastIndex])
			}
		}
	}
	return ""
}
