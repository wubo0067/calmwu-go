package main

import (
	"fmt"
	"reflect"
	"time"

	"github.com/wubo0067/calmwu-go/utils"
)

// https://stackoverflow.com/questions/48422482/how-to-convert-utc-to-india-local-time-using-golang
// https://stackoverflow.com/questions/37237223/access-current-system-time-zone

func greet() {
	defer utils.MeasureFunc()()
	fmt.Println("greet")
	time.Sleep(time.Second)
}

func foo() {
	greet()
}

func main() {
	now := time.Now()

	// 得到时区
	tzName, tzOffset := now.Zone()
	fmt.Printf("TZName:%s TZOffset:%d\n", tzName, tzOffset)

	location, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		fmt.Printf("LoadLocation failed! reason:%s\n", err.Error())
	} else {
		fmt.Println("---------Location:", location.String())
	}

	location = time.FixedZone(tzName, tzOffset)
	fmt.Println("---------Location:", location.String())

	//Time time.Time struct
	fmt.Printf("%s %s %s\n", reflect.TypeOf(now).Name(), reflect.TypeOf(now).String(), reflect.ValueOf(now).Kind().String())

	fmt.Println(now, now.Unix())
	fmt.Println(now.UTC(), now.UTC().Unix())

	// 计算当前时间在指定时区的时间
	location, _ = time.LoadLocation("UTC")
	fmt.Println(now.In(location))
	fmt.Println("--------------------------------")

	fmt.Println(now.UnixNano())
	fmt.Println(now.UnixNano()/int64(time.Millisecond), time.Millisecond, int64(time.Millisecond))

	fmt.Println("--------------------------------")
	fmt.Println(now.UTC().Format(time.RFC3339))
	fmt.Println(now)

	fmt.Println(time.Now().Format("2006-01-02T15:04:05-0700"))
	fmt.Println(time.Now().Format("2006-01-02 15:04:05+0800"))

	year, month, day := now.Date()
	date := fmt.Sprintf("%d%02d%02d%02d", year, month, day, now.Hour())
	fmt.Println(date)

	location, _ = time.LoadLocation("Local")
	fmt.Println("Location:", location.String())

	t := time.Date(year, month, day+20, 23, 59, 59, 0, location)
	fmt.Println(t.Local())

	t = time.Date(year, month, day+1, 0, 0, 0, 0, time.UTC)
	fmt.Println(t.Local())
	fmt.Println(t.UTC())

	var index = 1
	for index <= 7 {
		nTime := time.Now()
		nextDay := nTime.AddDate(0, 0, index)
		year, month, day := nextDay.Date()
		fmt.Printf("nextDay [%v] %d%02d%02d\n", nextDay, year, month, day)
		//fmt.Println(nextDay.Format("20060101"))
		index++
	}

	endTime := time.Now().Sub(now).Seconds()

	t = time.Unix(int64(endTime), 0)
	fmt.Println(t.String())
	fmt.Println(endTime)

	// now, _ = time.Parse("20060102 15:04:05", "20170703 00:00:00")
	// index = 0
	// for index < 30 {
	// 	future := now.AddDate(0, 0, index)
	// 	year, month, day := future.Date()
	// 	fmt.Printf("future [%v] %d%02d%02d\n", future, year, month, day)
	// 	index++
	// }

	nAry := make([]int, 0)
	nAry = append(nAry, 1)
	nAry = append(nAry, 2)
	fmt.Printf("%+v\n", nAry)

	//CreateTime:1500019209758 DeadTime:1500027061167
	//           1500033649

	foo()
}
