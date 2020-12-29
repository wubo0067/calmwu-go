package utils

import (
	"fmt"
	"reflect"
	"regexp"
	"sailcraft/base"
	"sailcraft/base/consul_api"
	financesvr_proto "sailcraft/financesvr_main/proto"
	"strconv"
)

// 一些通用的方法
func JoinString(attrs []string, op string) string {
	data := ""
	for index, value := range attrs {
		if index > 0 {
			data = data + op
		}
		data = data + value
	}

	return data
}

// sql防止注入判断
func SqlQuote(sql string) bool {
	str := `(?:')|(?:--)|(/\\*(?:.|[\\n\\r])*?\\*/)|(\b(select|update|and|or|delete|insert|trancate|char|chr|into|substr|ascii|declare|exec|count|master|into|drop|execute)\b)`
	re, err := regexp.Compile(str)
	if err != nil {
		return false
	}

	result := re.MatchString(sql)

	return result
}

// 这里支持int float bool string，暂时不支持其他类型
func ConvertMapInterfaceToMapString(attrMap map[string]interface{}) (map[string]string, error) {
	redisHash := make(map[string]string)

	for fieldName, value := range attrMap {
		fieldValue := reflect.ValueOf(value)
		t := fieldValue.Type()

		switch t.Kind() {
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
		default:
			return nil, fmt.Errorf("fieldName[%s] type:%v is not support!", fieldName, t)
		}
	}

	return redisHash, nil
}

func QueryPlayerTimeZone(uin int) string {
	// 查询玩家的时区从financesvr
	realReq := financesvr_proto.ProtoUserVIPTypeReq{
		Uin:    uint64(uin),
		ZoneID: -1,
	}
	financeSvrRes, err := consul_api.PostRequstByConsulDns(uint64(uin), "QueryUserVIPType", &realReq, ConsulClient, "FinanceSvr")
	if err != nil {
		base.GLog.Error("Uin[%d] Query[FinanceSvr:QueryUserVIPType] failed!", uin)
	} else {
		var realRes financesvr_proto.ProtoUserVIPTypeRes
		err = base.MapstructUnPackByJsonTag(financeSvrRes.ResData.Params, &realRes)
		if err == nil {
			base.GLog.Debug("Uin[%d] TimeZone[%s]", uin, realRes.TimeZone)
			return realRes.TimeZone
		}
	}
	return "Local"
}
