package main

import (
	"fmt"
	"log"

	"github.com/gocql/gocql"
)

// 自定义tracer的用法 https://github.com/relops/gocql-tracing-example/blob/master/main.go，就是将执行过程记录下来

var (
	cassandra_hosts = []string{"139.199.62.77", "118.89.34.64"}
)

func insert_dau(session *gocql.Session) {

	for i := 1000; i < 2000; i++ {
		sql_content := fmt.Sprintf("UPDATE tbl_dau set login_count=login_count+%d where date='%d'",
			i, i)
		fmt.Println(sql_content)
		err := session.Query(sql_content).Exec()
		if err != nil {
			fmt.Printf("failed: %s\n", err.Error())
		}
	}
}

func select_dau(session *gocql.Session) {
	//buf := &bytes.Buffer{}
	//trace := gocql.NewTraceWriter(session, buf)

	iter := session.Query("SELECT * from tbldau").PageSize(20).Iter()

	//session.SetTrace(trace)
	var date string
	var login_count int
	var index int = 0
	for iter.Scan(&date, &login_count) {
		fmt.Printf("[%d] tbl_dau: date[%s] login_count[%d]\n", index, date, login_count)
		index++
	}

	fmt.Println("record count: ", index)
	// 这里会
	//fmt.Println("trace info: ", buf.String())

	if err := iter.Close(); err != nil {
		fmt.Println(err.Error())
	}
}

func createTable(s *gocql.Session, table string) error {
	if err := s.Query(table).RetryPolicy(nil).Exec(); err != nil {
		log.Printf("error creating table table=%q err=%v\n", table, err)
		return err
	}
	return nil
}

func test_cas(session *gocql.Session) {

	err := createTable(session, `CREATE TABLE IF NOT EXISTS ks_statisticmodule.cas_user (
			name         varchar,
			age   	     int,
			address 	 varchar,
			PRIMARY KEY (name))`)
	if err != nil {
		fmt.Printf("create table failed! error[%s]\n", err.Error())
		return
	}

	name := "calmwu"
	age := 39
	address := "shenzhen"

	var nameCAS string
	var ageCAS int
	var addressCAS string
	var applied bool

	// 如果不存在应该插入，返回的是之前的值，这里应该是空
	if err := session.Query(`INSERT INTO cas_user (name, age, address)
		VALUES (?, ?, ?) IF NOT EXISTS`,
		name, age, address).Exec(); err != nil {
		fmt.Println("insert:", err)
	}

	// 不存在才插入，已经存在返回存在的数据，applied确认该操作是否执行
	address = "wuhan"
	if applied, err := session.Query(`INSERT INTO cas_user (name, age, address)
		VALUES (?, ?, ?) IF NOT EXISTS`,
		name, age, address).ScanCAS(&nameCAS, &addressCAS, &ageCAS); err != nil {
		fmt.Println("CAS insert:", err)
	} else if applied {
		fmt.Println("insert should not have been applied")
	}

	fmt.Printf("applied[%v] nameCAS[%s] ageCAS[%d] addressCAS[%s]\n", applied, nameCAS, ageCAS, addressCAS)
}

func test_batch(session *gocql.Session) {
	batch := session.NewBatch(gocql.UnloggedBatch)
	batch.Query("INSERT INTO cas_user (name, age, address) VALUES(?, ?, ?)", "jenny", 1, "shenzhen")
	batch.Query("INSERT INTO cas_user (name, age, address) VALUES(?, ?, ?)", "vivi", 2, "shenzhen-luohu")
	err := session.ExecuteBatch(batch)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("Batch insert successed!")
}

func FindRecord(session *gocql.Session, cqlContent string) (result []map[string]interface{}) {
	result = nil

	iter := session.Query(cqlContent).Iter()
	defer iter.Close()

	result, err := iter.SliceMap()
	if err != nil {
		fmt.Printf("SliceMap failed! error[%s]\n", err.Error())
		return
	}
	fmt.Printf("%v\n", result)
	return
}

func main() {
	cluster := gocql.NewCluster(cassandra_hosts[0:]...)
	cluster.Keyspace = "ks_statisticmodule"
	cluster.Consistency = gocql.Quorum
	session, _ := cluster.CreateSession()

	// insert_dau(session)
	select_dau(session)
	//test_cas(session)
	// test_batch(session)

	FindRecord(session, "select * from tbluserinfo where uin=1")
}
