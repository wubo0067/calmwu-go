package main

import (
	"bytes"
	"fmt"
	"reflect"

	"strconv"

	"github.com/emirpasic/gods/sets/hashset"
)

// http://blog.csdn.net/vipally/article/details/40952817
// https://gist.github.com/drewolson/4771479
// http://stackoverflow.com/questions/10858787/what-are-the-uses-for-tags-in-go

// type OperationalItf interface {
// 	CheckLogin(c int) string
// 	QueryBulletin(c int) string
// }

/*
if t.Kind() == reflect.Ptr {
			// 如果是指针，则获取其所指向的元素
			t = t.Elem()
*/

type Operational struct {
	Name string
	Age  int
	Sex  int
}

type TblUserInfoS struct {
	Uin            int     `mapstructure:"uin"`
	CreateTime     string  `mapstructure:"createtime"`
	NickName       string  `mapstructure:"nickname"`
	RegisterRegion string  `mapstructure:"registerregion"`
	TotalCost      float32 `mapstructure:"totalcost"`
	LoginTime      string  `mapstructure:"logintime"`
	LogoutTime     string  `mapstructure:"logouttime"`
	TradeUnionID   int     `mapstructure:"tradeunionid"`
	VersionID      int     `mapstructure:"versionid"`
	ShipLst        []int
	ShipMap        map[string]string
	NameLst        []string
	AgeMap         map[string]int
	LineUpMap      map[int]int
}

func (o *Operational) CheckLogin(c int) string {
	return fmt.Sprintf("CheckLogin %d", c)
}

func (o *Operational) QueryBulletin(c int, d int) string {
	return fmt.Sprintf("QueryBulletin %d %d", c, d)
}

func makeUpdateCql(tbName string, object interface{}, keys *hashset.Set) {
	updateCql := "UPDATE " + tbName + " SET "
	whereCond := " WHERE "

	v_t := reflect.TypeOf(object).Elem()
	v_v := reflect.ValueOf(object).Elem()

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
					for index, _ := range fieldValMapKeys {
						fieldValMapKey := fieldValMapKeys[index]
						if fieldValMapKey.Kind() == reflect.String {
							updateCql += "'" + fieldValMapKeys[index].String() + "' : "
						} else {
							// TODO，这里特别说明，只支持string和整型
							updateCql += strconv.Itoa(int(fieldValMapKeys[index].Int())) + " : "
						}
						fmt.Printf("key[%v] [%s]\n", fieldValMapKey.Interface(), fieldValMapKey.String())
						//updateCql += "'" + fieldValMapKeys[index].String() + "' : "
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
			whereCond += fieldName + "=" + fmt.Sprintf("%v", fieldVal) + " AND "
		}
		fieldIndex++
	}

	updateCql = updateCql[:len(updateCql)-2]
	whereCond = whereCond[:len(whereCond)-5]
	updateCql += whereCond

	fmt.Println(updateCql)
}

func updateTblUserInfo(tblUserInfo *TblUserInfoS) {
	v_t := reflect.TypeOf(tblUserInfo).Elem()
	v_v := reflect.ValueOf(tblUserInfo).Elem()
	// if v_t.Kind() == reflect.Ptr {
	// 	v_t = v_t.Elem()
	// }

	var fieldIndex = 0
	for fieldIndex < v_t.NumField() {
		field := v_t.Field(fieldIndex)

		fieldName := field.Name
		fieldVal := v_v.FieldByName(fieldName).Interface()

		if field.Type.Kind() == reflect.String {
			fmt.Printf("val is string: [%s]\n", v_v.FieldByName(fieldName).String())
		}

		fmt.Printf("field name[%s] value[%v] type[%s]\n", fieldName, fieldVal, field.Type.Kind())
		fieldIndex++
	}

	keys := hashset.New()
	keys.Add("Uin")
	makeUpdateCql("tbl_userinfo", tblUserInfo, keys)
}

func Join(a []interface{}, sep string) string {
	if len(a) == 0 {
		return ""
	}
	if len(a) == 1 {
		return fmt.Sprintf("%v", a[0])
	}

	buffer := &bytes.Buffer{}
	buffer.WriteString(fmt.Sprintf("%v", a[0]))
	for i := 1; i < len(a); i++ {
		buffer.WriteString(sep)
		buffer.WriteString(fmt.Sprintf("%v", a[i]))
	}
	return buffer.String()
}

func test_userinfo() {
	var userInfo TblUserInfoS
	userInfo.Uin = 1
	userInfo.CreateTime = "2006-01-02 15:04:05+0800"
	userInfo.LoginTime = "2006-01-02 15:04:05+0800"
	userInfo.VersionID = 1
	userInfo.ShipLst = make([]int, 0)
	userInfo.ShipLst = append(userInfo.ShipLst, 1)
	userInfo.ShipLst = append(userInfo.ShipLst, 2)
	userInfo.ShipLst = append(userInfo.ShipLst, 3)

	userInfo.NameLst = make([]string, 0)
	userInfo.NameLst = append(userInfo.NameLst, "a")
	userInfo.NameLst = append(userInfo.NameLst, "b")
	userInfo.NameLst = append(userInfo.NameLst, "c")

	userInfo.ShipMap = make(map[string]string)
	userInfo.ShipMap["a"] = "b"
	userInfo.ShipMap["c"] = "d"

	userInfo.AgeMap = make(map[string]int)
	userInfo.AgeMap["e"] = 999
	userInfo.AgeMap["f"] = 888

	userInfo.LineUpMap = make(map[int]int)
	userInfo.LineUpMap[1] = 1
	userInfo.LineUpMap[10] = 1999

	updateTblUserInfo(&userInfo)
}

func main() {
	test_userinfo()

	var t interface{} = 1
	fmt.Printf("t type[%s], string[%s]\n", reflect.TypeOf(t).Name(),
		reflect.TypeOf(t).String())

	//var f float64
	if _, ok := t.(float64); !ok {
		fmt.Println("t is not float64")
	}

	//var ci interface{} = new(Operational)
	var ci interface{} = Operational{Name: "calmwu"}
	//fmt.Println(ci.CheckLogin(10))

	ci_t := reflect.TypeOf(ci)
	ci_t_string := ci_t.String()
	// 变量的具体类型 *main.Operational
	fmt.Println("ci_t_string: ", ci_t_string)
	// 变量类型的方法数量 2
	fmt.Println("ci_t num_method: ", ci_t.NumMethod())
	// 类型成员数量，如果是指针，这个方法就会panic
	fmt.Println("ci_t num_field: ", ci_t.NumField())

	ci_v := reflect.ValueOf(ci)
	// ptr
	fmt.Println(ci_v.Kind().String())
	fmt.Println(ci_v.NumField())

	//------------------------------------------------------------------------
	ci_1 := new(Operational)
	ci_1.Name = "Test!!!!!!!!!!!!!!!!"
	var ci_p interface{} = ci_1
	// 值的处理
	ci_p_v := reflect.ValueOf(ci_p)
	//&{}
	fmt.Println(ci_p_v)
	ci_p_v_e := reflect.ValueOf(ci_p).Elem()
	//{}
	fmt.Println(ci_p_v_e)
	// 通过值拿到值的类型
	fmt.Println(ci_p_v.Type().String())

	name_v := ci_p_v_e.FieldByName("Name")
	fmt.Println(name_v.String())

	// 可以拿到具体的类型
	ci_p_t := reflect.TypeOf(ci_p)

	// ci_p_t type:  *main.Operational
	// ci_p_t Elem type:  main.Operational
	fmt.Println("ci_p_t type: ", ci_p_t.String())
	// 如果是指针，拿到引用的具体类型
	fmt.Println("ci_p_t Elem type: ", ci_p_t.Elem().String())

	var i int = 0
	//var vals = []reflect.Value{reflect.ValueOf(1)}

	for i < ci_p_v.NumMethod() {
		ci_p_m := ci_p_v.Method(i)
		// [<func(int) string Value>]----[func(int) string]
		fmt.Printf("[%s]----[%s], [%d]\n", ci_p_m.String(), ci_p_m.Type().String(), ci_p_m.Type().NumIn())
		// 通过类型拿到方法的名字
		fmt.Println("Method ", ci_p_t.Method(i).Name, ci_p_t.Method(i).PkgPath, ci_p_t.Method(i).Index)
		method_params := make([]reflect.Value, ci_p_m.Type().NumIn())
		var n int = 0
		for n < ci_p_m.Type().NumIn() {
			method_params[n] = reflect.ValueOf(n + 100)
			n++
		}
		data := ci_p_m.Call(method_params)
		fmt.Println(data)
		i++
	}

	sliceInts := make([]Operational, 3, 3)
	sliceInts[0].Name = "1"
	sliceInts[1].Name = "2"
	sliceInts[2].Name = "3"
	fmt.Println("-------------------------")
	// 去掉了指针
	sliceV := reflect.Indirect(reflect.ValueOf(&sliceInts))
	// 获取数据
	fmt.Println(sliceV)
	// 这是个[]slice []main.Operational
	fmt.Println(sliceV.Type())
	// 这样可以获取slice元素的类型,这个是结构类型 main.Operational
	fmt.Println(sliceV.Type().Elem())
	// struct
	fmt.Println(sliceV.Type().Elem().Kind())
	// 动态创建结构指针
	newSP := reflect.New(sliceV.Type().Elem())
	fmt.Println(newSP)
	// 这样可以拿到具体的类型*main.Operational
	if newSP.Kind() == reflect.Ptr {
		fmt.Println("newS is point")
	}
	// 这个是真实类型，如果用reflect.TypeOf(newS)，这是获取声明类型
	fmt.Println(newSP.Type())
	// 输出动态创建对象的每个field
	newST := newSP.Type().Elem()
	fmt.Println(newST)
	i = 0
	for i < newST.NumField() {
		fieldType := newST.Field(i)
		fmt.Println("field ", i, "Name is", fieldType.Name, "Type is", fieldType.Type)
		switch fieldType.Type.Kind() {
		case reflect.String:
			newSP.Elem().Field(i).SetString("2323")
		case reflect.Int:
			newSP.Elem().Field(i).SetInt(999)
		}
		i++
	}
	fmt.Println(newSP)

	i = int(10)
	testRefect(&i)

	testStructMapField()
}

func testRefect(i interface{}) {
	fmt.Println(reflect.TypeOf(i))
	fmt.Println(reflect.ValueOf(i))
	fmt.Println(reflect.ValueOf(i).Type())
}

func testStructMapField() {
	type InterfaceDecS struct {
		InterfaceName2Path map[string]string
	}

	interfaceDec := new(InterfaceDecS)
	interfaceDec.InterfaceName2Path = make(map[string]string)
	interfaceDec.InterfaceName2Path["adduser"] = "/api/v1/AddUser"
	interfaceDec.InterfaceName2Path["deluser"] = "/api/v1/DeleteUser"
	interfaceDec.InterfaceName2Path["updateuser"] = "/api/v1/UpdateUser"

	getMapFields(interfaceDec)
}

func getMapFields(obj interface{}) {
	v := reflect.Indirect(reflect.ValueOf(obj))
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldName := field.Name
		fieldV := v.Field(i)
		fmt.Printf("fieldName[%s] type:%v kind:%v\n", fieldName, fieldV.Type(), field.Type.Kind())
		if field.Type.Kind() == reflect.Map {
			keys := fieldV.MapKeys()
			for index, _ := range keys {
				key := keys[index].String()
				value := fieldV.MapIndex(reflect.ValueOf(key))
				fmt.Printf("key[%s] value[%s]\n", key, value.String())
			}
		}
	}
}
