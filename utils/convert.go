package utils

import (
	"reflect"
	"strconv"

	"github.com/mitchellh/mapstructure"
)

var i64Type reflect.Type
var f64Type reflect.Type
var strType reflect.Type

func init() {
	i64Type = reflect.TypeOf(int64(0))
	f64Type = reflect.TypeOf(float64(0))
	strType = reflect.TypeOf("")
}

func ConvertToInt(i interface{}, defaultValue int) int {
	return int(ConvertToInt64(i, int64(defaultValue)))
}

func ConvertToInt64(i interface{}, defaultValue int64) (ret int64) {
	defer func() {
		if r := recover(); r != nil {
			// 转换出错，设置默认值
			ZLog.Errorf("Convert %v to %T failed[%v]", i, ret, r)
			ret = defaultValue
		}
	}()

	v := reflect.ValueOf(i)

	// 字符串类型
	if v.Kind() == reflect.String {
		strValue := v.String()
		ret, err := strconv.ParseInt(strValue, 10, 64)
		if err != nil {
			ZLog.Errorf("Convert \"%s\" to %T failed[%v]", strValue, ret, err)
			ret = defaultValue
		}

		return ret
	}

	i64Value := v.Convert(i64Type)
	ret = i64Value.Int()
	return ret
}

func ConvertToFloat64(i interface{}, defaultValue float64) (ret float64) {
	defer func() {
		if r := recover(); r != nil {
			// 转换出错，设置默认值
			ZLog.Errorf("Convert %v to %T failed[%v]", i, ret, r)
			ret = defaultValue
		}
	}()

	v := reflect.ValueOf(i)

	// 字符串类型
	if v.Kind() == reflect.String {
		strValue := v.String()
		ret, err := strconv.ParseFloat(strValue, 64)
		if err != nil {
			ZLog.Errorf("Convert \"%s\" to %T failed[%v]", strValue, ret, err)
			ret = defaultValue
		}

		return ret
	}

	f64Value := v.Convert(f64Type)
	ret = f64Value.Float()
	return ret
}

func ConvertToString(i interface{}, defaultValue string) (ret string) {
	defer func() {
		if r := recover(); r != nil {
			// 转换出错，设置默认值
			ZLog.Errorf("Convert %v to %T failed[%v]", i, ret, r)
			ret = defaultValue
		}
	}()

	v := reflect.ValueOf(i)
	strValue := v.Convert(strType)
	ret = strValue.String()
	return ret
}

func ConvertHashToObj(m interface{}, rawVal interface{}, tagName string) error {
	config := &mapstructure.DecoderConfig{
		TagName:  tagName,
		Metadata: nil,
		Result:   rawVal,
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(m)
}
