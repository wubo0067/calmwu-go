/*
 * @Author: calmwu
 * @Date: 2017-10-17 11:09:24
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-09-26 11:05:07
 * @Comment:
 */

package mysql

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-xorm/builder"
	"github.com/go-xorm/core"
)

func CreateEngine() (*DBEngineInfoS, error) {
	driverName := "mysql"
	dbUser := "root"
	dbPwd := "root"
	dbAddr := "192.168.12.3:3306"
	dbName := "calmwu"

	return CreateDBEngnine(driverName, dbUser, dbPwd, dbAddr, dbName)
}

type TblUserS struct {
	Id      int    `xorm:"pk autoincr"` // 自增字段必须是pk
	Name    string `xorm:"varchar(25) notnull unique"`
	Age     int    `xorm:"default(678) notnull"`
	Address string `xorm:"varchar(128) null"`
}

func (tblUser TblUserS) TableName() string {
	//fmt.Println("tableName invoke userId", tblUser.Id)
	return fmt.Sprintf("TblUser_%d", tblUser.Id%3)
}

// go test -v
func TestCreateDBEngnine(t *testing.T) {

	engine, err := CreateEngine()

	if err != nil {
		t.Error(err.Error())
	} else {
		t.Log("create DBEngine successed!")
	}

	tables, err := GetDbMetas(engine)
	if err != nil {
		t.Error(err.Error())
	} else {
		for index, _ := range tables {
			table := tables[index]
			t.Log(table.Name, "\t", table.ColumnsSeq())
		}
	}
}

func TestPingDB(t *testing.T) {
	engine, err := CreateEngine()
	if err != nil {
		t.Error(err.Error())
	} else {
		t.Log("create DBEngine successed!")
	}

	DoDBKeepAlive(engine, time.Second*2)

	time.Sleep(time.Second * 6)
}

func TestCreateDBTable(t *testing.T) {
	engine, err := CreateEngine()
	if err != nil {
		t.Error(err.Error())
	} else {
		t.Log("create DBEngine successed!")
	}

	tblUser := new(TblUserS)

	for i := 0; i < 3; i++ {
		// mdgbz，函数对象不算实现了方法
		// tblUser.TableName = func() string {
		// 	return fmt.Sprintf("User_%d", i)
		// }
		tblUser.Id = i
		// t.Log("tblUser name:", tblUser.TableName())

		// // reflect.Indirect用这个后将ptr转成了对象，所以tablename方法不可调用了
		// if _, ok := reflect.ValueOf(tblUser).Interface().(xorm.TableName); ok {
		// 	t.Log("tblUser implement xorm.TableName interface")
		// } else {
		// 	t.Log("tblUser does not implement xorm.TableName interface")
		// }

		// err = engine.CreateTables(tblUser)
		// if err != nil {
		// 	t.Error(err.Error())
		// }
		err := CreateTable(engine, tblUser)
		if err != nil {
			t.Error("create table failed!", err.Error())
		}
	}
}

func TestSqlExec(t *testing.T) {
	engine, err := CreateEngine()
	if err != nil {
		t.Error(err.Error())
	} else {
		t.Log("create DBEngine successed!")
	}

	sqlContent := "TRUNCATE TABLE TblUser_0"
	affectedRows, _, err := SqlExec(engine, sqlContent)
	t.Log("affectRows: ", affectedRows)
	if err != nil {
		t.Error(err.Error())
	}

	sqlContent = "TRUNCATE TABLE TblUser_1"
	SqlExec(engine, sqlContent)
	sqlContent = "TRUNCATE TABLE TblUser_2"
	SqlExec(engine, sqlContent)
}
func TestInsertMultiRecords(t *testing.T) {
	engine, err := CreateEngine()
	if err != nil {
		t.Error(err.Error())
	} else {
		t.Log("create DBEngine successed!")
	}

	users := make([]*TblUserS, 10)
	for i := 0; i < 10; i++ {
		users[i] = new(TblUserS)
		users[i].Name = fmt.Sprintf("Name_%d", i)
		users[i].Age = i
		users[i].Address = fmt.Sprintf("Addr_%d", i)
	}

	//InsertMultiRecords(engine, users[0], users[1], users[2])
	// 作为一个slice和数组会忽略表名，批量插入到一张表中，这里用的是指针
	// 如果表不存在，程序会panic
	affected, err := InsertSliceRecordsToSameTable(engine, "TblUser_1", &users)
	if err != nil {
		t.Error(err.Error())
	} else {
		t.Log("affected:", affected)
	}

	userObj := new(TblUserS)
	userObj.Name = "Calmwu"
	userObj.Age = 98
	userObj.Address = "Shenzhen"
	affected, err = InsertRecord(engine, "TblUser_2", userObj)
	if err != nil {
		t.Error(err.Error())
	} else {
		t.Log("affected:", affected)
	}

	userObj1 := new(TblUserS)
	userObj1.Name = "Golang"
	//userObj1.Age = 17
	userObj1.Address = "google"
	// 这里用的是对象
	affected, err = InsertRecord(engine, "TblUser_2", userObj1, "age")
	if err != nil {
		t.Error(err.Error())
	} else {
		t.Log("affected:", affected)
	}
}

func TestUpdateRecord(t *testing.T) {
	engine, err := CreateEngine()
	if err != nil {
		t.Error(err.Error())
	} else {
		t.Log("create DBEngine successed!")
	}

	userObj := new(TblUserS)
	userObj.Id = 1
	userObj.Name = "Calmwu"
	userObj.Age = 0
	userObj.Address = "343sdsd"

	PK := core.NewPK(2)
	a, e := UpdateRecord(engine, "TblUser_2", PK, userObj)
	if e != nil {
		t.Error(e.Error())
	} else {
		t.Log("affect rows: ", a)
	}
}

func TestUpdateSpecifiedFields(t *testing.T) {
	engine, err := CreateEngine()
	if err != nil {
		t.Error(err.Error())
	} else {
		t.Log("create DBEngine successed!")
	}

	affectedRows, err := UpdateRecordSpecifiedFieldsByCond(engine, "TblUser_1", "Id=1", map[string]interface{}{"Name": "Jingdong", "Address": ""})
	if err != nil {
		t.Error(err.Error())
	} else {
		t.Log("UpdateRecordSpecifiedFieldsByCond successed! affectedRows:", affectedRows)
	}

	// Pk := core.NewPK(5)
	// affectedRows, err = UpdateRecordSpecifiedFields(engine, "TblUser_1", Pk, map[string]interface{}{"Name": "Jingdong"})
	// if err != nil {
	// 	t.Error(err.Error())
	// } else {
	// 	t.Log("UpdateRecordSpecifiedFields successed! affectedRows:", affectedRows)
	// }
}

func TestGetRecord(t *testing.T) {
	engine, err := CreateEngine()
	if err != nil {
		t.Error(err.Error())
	} else {
		t.Log("create DBEngine successed!")
	}
	userObj := new(TblUserS)
	userObj.Id = 100
	_, err = GetRecord(engine, "TblUser_0", userObj)
	if err != nil {
		t.Error(err.Error())
	} else {
		t.Log(userObj)
	}
}

func TestGetRecordByCond(t *testing.T) {
	engine, err := CreateEngine()
	if err != nil {
		t.Error(err.Error())
	} else {
		t.Log("create DBEngine successed!")
	}

	userObj := new(TblUserS)
	_, err = GetRecordByCond(engine, "TblUser_0", "Id=2", userObj)
	if err != nil {
		t.Error(err.Error())
	} else {
		t.Log(userObj)
	}

	userObj = new(TblUserS)
	exist, err := GetRecordByCond(engine, "TblUser_0", "Id=2000", userObj)
	if !exist {
		t.Log("record not found")
	}

	if err != nil {
		t.Error(err.Error())
	} else {
		t.Log(userObj)
	}
}

func TestFindRecords(t *testing.T) {
	engine, err := CreateEngine()
	if err != nil {
		t.Error(err.Error())
	} else {
		t.Log("create DBEngine successed!")
	}

	result := make([]TblUserS, 0)
	err = FindRecordsBySimpleCond(engine, "TblUser_1", "Id >= 1", 5, 1, &result)
	if err != nil {
		t.Error(err.Error())
	} else {
		t.Log(result)
	}

	result = make([]TblUserS, 0)
	err = FindRecordsBySimpleCond(engine, "TblUser_2", "", 0, 0, &result)
	if err != nil {
		t.Error(err.Error())
	} else {
		t.Log(result)
	}
}

func TestDeleteRecord(t *testing.T) {
	engine, err := CreateEngine()
	if err != nil {
		t.Error(err.Error())
	} else {
		t.Log("create DBEngine successed!")
	}

	userObj := new(TblUserS)

	PK := core.NewPK(7)
	a, e := DeleteRecord(engine, "TblUser_1", PK, userObj)
	if e != nil {
		t.Error(e.Error())
	} else {
		t.Log("affect rows: ", a)
	}

	PK = core.NewPK(9)
	a, e = DeleteRecord(engine, "TblUser_1", PK, userObj)
	if e != nil {
		t.Error(e.Error())
	} else {
		t.Log("affect rows: ", a)
	}
}

func TestDeleteMultiRecords(t *testing.T) {
	engine, err := CreateEngine()
	if err != nil {
		t.Error(err.Error())
	} else {
		t.Log("create DBEngine successed!")
	}

	userObj := new(TblUserS)

	xormCond := builder.In("Id", 5, 6, 7, 8, 9)
	a, e := DeleteRecordsByMultiConds(engine, "TblUser_1", &xormCond, userObj)
	if e != nil {
		t.Error(e.Error())
	} else {
		t.Log("affect rows: ", a)
	}
}

func TestSelectRecordsByCond(t *testing.T) {
	engine, err := CreateEngine()
	if err != nil {
		t.Error(err.Error())
	} else {
		t.Log("create DBEngine successed!")
	}

	record := new(TblUserS)
	// recordT := reflect.ValueOf(record).Elem().Type()
	// obj := reflect.New(recordT)

	// newST := obj.Type().Elem()
	// fmt.Println(newST)
	// var i = 0
	// for i < newST.NumField() {
	// 	fieldType := newST.Field(i)
	// 	fmt.Println("field ", i, "Name is", fieldType.Name, "Type is", fieldType.Type)
	// 	i++
	// }

	// fmt.Printf("record type[%s]\n", reflect.TypeOf(record).String())
	// // 这两个类型的区别，我是比较模糊的
	// // record type[*db.TblUserS]
	// // obj type[reflect.Value]
	// // obj type[*db.TblUserS]
	// fmt.Printf("obj type[%s]\n", reflect.TypeOf(obj).String())
	// fmt.Printf("obj type[%s]\n", obj.Type().String())

	results, err := SelectRecordsByCond(engine, "TblUser_1", "Id>4", record)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		for i, _ := range results {
			fmt.Println(results[i].(*TblUserS))
		}
	}
}

func TestSelectRecordsByCond2(t *testing.T) {
	engine, err := CreateEngine()
	if err != nil {
		t.Error(err.Error())
	} else {
		t.Log("create DBEngine successed!")
	}

	record := new(TblUserS)
	RegisterTableObj(record)

	results, err := SelectRecordsByCond2(engine, "TblUser_1", "Id>4", "TblUserS")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		for i, _ := range results {
			fmt.Println(results[i].(*TblUserS))
		}
	}

	results, err = SelectRecordsByCond2(engine, "TblUser_1", "", "db.TblUserS")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		for i, _ := range results {
			fmt.Println(results[i].(*TblUserS))
		}
	}
}

func TestFindRecordByMultiConds(t *testing.T) {
	engine, err := CreateEngine()
	if err != nil {
		t.Error(err.Error())
	} else {
		t.Log("create DBEngine successed!")
	}

	result := make([]TblUserS, 0)
	// conds := []*SqlWhereUnitS{
	// 	&SqlWhereUnitS{
	// 		ConditionContent: "Id = 5",
	// 		LogicOper:        E_LOGICOP_OR},
	// 	&SqlWhereUnitS{
	// 		ConditionContent: "Address = 'Addr_9'",
	// 		LogicOper:        E_LOGICOP_NONE},
	// }
	// xormCond, _ := MakeWhereCondition(conds)
	xormCond := builder.Expr("Id=5").And(builder.Expr("Address = 'Addr_9'"))
	err = FindRecordsByMultiConds(engine, "TblUser_1", &xormCond, 0, 0, &result)
	if err != nil {
		t.Error(err.Error())
	} else {
		fmt.Printf("result count: %d\n", len(result))
		for i, _ := range result {
			fmt.Println(result[i])
		}
	}

	result = make([]TblUserS, 0)
	// conds = []*SqlWhereUnitS{
	// 	&SqlWhereUnitS{
	// 		ConditionContent: "Id > 5",
	// 		LogicOper:        E_LOGICOP_AND},
	// 	&SqlWhereUnitS{
	// 		ConditionContent: "Age > 7",
	// 		LogicOper:        E_LOGICOP_NONE},
	// }
	// xormCond, _ = MakeWhereCondition(conds)

	xormCond = builder.Expr("Id > 5").Or(builder.Expr("Age > 7"))
	err = FindRecordsByMultiConds(engine, "TblUser_1", &xormCond, 0, 0, &result)
	if err != nil {
		t.Error(err.Error())
	} else {
		fmt.Printf("result count: %d\n", len(result))
		for i, _ := range result {
			fmt.Println(result[i])
		}
	}
}

func TestFindRecordOrderBy(t *testing.T) {
	engine, err := CreateEngine()
	if err != nil {
		t.Error(err.Error())
	} else {
		t.Log("create DBEngine successed!")
	}

	result := make([]TblUserS, 0)

	err = FindRecordsBySimpleCondWithOrderBy(engine, "TblUser_1", "Name = 'Name_0'", 0, 0, []string{"Age desc"}, &result)
	if err != nil {
		t.Error(err.Error())
	} else {
		fmt.Printf("result count: %d\n", len(result))
		for i, _ := range result {
			fmt.Println(result[i])
		}
	}
}
