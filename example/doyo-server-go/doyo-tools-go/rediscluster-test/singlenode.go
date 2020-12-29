/*
 * @Author: calmwu
 * @Date: 2018-10-27 20:07:49
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-10-27 22:25:54
 */

package main

import (
	base "doyo-server-go/doyo-base-go"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"
)

var (
	c_TIME_DEFAULT time.Time
	TimeType       = reflect.TypeOf(c_TIME_DEFAULT)
)

type ProtoDoyoRecUserLogin struct {
	UserID       string `json:"UserID"`
	UserLanguage string `json:"UserLanguage"` // http://www.lingoes.cn/zh/translator/langcode.htm
	UserCountry  string `json:"UserCountry"`  // https://zh.wikipedia.org/wiki/ISO_3166-1
	UserGender   int    `json:"UserGender"`   // 性别
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

func ConvertObjToStringMap(obj interface{}) (map[string]interface{}, error) {
	redisHash := make(map[string]interface{})

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

func ConvertStringMapToObj(hashV map[string]string, objP interface{}) error {
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

func main() {
	logger := base.NewSimpleLog(nil)

	redisdb := redis.NewClient(&redis.Options{
		Addr:         "123.59.40.19:6379",
		DialTimeout:  time.Second,
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
		PoolSize:     3,
	})

	res, err := redisdb.Set("calmwu", "ryzen", 60*time.Second).Result()
	if err != nil {
		logger.Println(err.Error())
	} else {
		logger.Printf("set res:%s", res)
	}

	user := new(ProtoDoyoRecUserLogin)
	user.UserID = "doyo123456"
	user.UserLanguage = "zh"
	user.UserCountry = "cn"
	user.UserGender = 1

	redisUser, err := ConvertObjToStringMap(user)
	if err != nil {
		logger.Println(err.Error())
	} else {
		cmdStatus := redisdb.HMSet("doyo123456", redisUser)
		logger.Printf("hmset %s", cmdStatus.String())
	}
}
