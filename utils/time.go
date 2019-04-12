/*
 * @Author: calmwu
 * @Date: 2017-09-18 10:36:09
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-11-30 17:27:06
 * @Comment:
 */

package utils

import (
	"fmt"
	"strconv"
	"time"
)

func GetDate() string {
	now := time.Now()
	year, month, day := now.Date()
	return fmt.Sprintf("%d%02d%02d", year, month, day)
}

func GetDateByLocation(location *time.Location) string {
	now := time.Now().In(location)
	year, month, day := now.Date()
	return fmt.Sprintf("%d%02d%02d", year, month, day)
}

func GetTimeByTz(tz string) (*time.Time, error) {
	localtion, err := time.LoadLocation(tz)
	if err != nil {
		return nil, err
	}
	now := time.Now().In(localtion)
	return &now, nil
}

func GetDateNum(location *time.Location) int {
	var dateName string
	if location == nil {
		dateName = GetDate()
	} else {
		dateName = GetDateByLocation(location)
	}
	dateNum, _ := strconv.ParseInt(dateName, 10, 32)
	return int(dateNum)
}

func GetDateNum2(now *time.Time) int {
	year, month, day := now.Date()
	dateName := fmt.Sprintf("%d%02d%02d", year, month, day)
	dateNum, _ := strconv.ParseInt(dateName, 10, 32)
	return int(dateNum)
}

func GetDateHour() string {
	now := time.Now()
	year, month, day := now.Date()
	return fmt.Sprintf("%d%02d%02d%02d", year, month, day, now.Hour())
}

func GetTimeStampMs() string {
	return time.Now().Format("2006-01-02 15:04:05.000")
}

func GetTimeStampSec() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// const (
//     ANSIC       = "Mon Jan _2 15:04:05 2006"
//     UnixDate    = "Mon Jan _2 15:04:05 MST 2006"
//     RubyDate    = "Mon Jan 02 15:04:05 -0700 2006"
//     RFC822      = "02 Jan 06 15:04 MST"

//  // RFC822 with numeric zone RFC850      = "Monday, 02-Jan-06 15:04:05 MST"
//     RFC822Z     = "02 Jan 06 15:04 -0700"
//     RFC1123     = "Mon, 02 Jan 2006 15:04:05 MST"

//  // RFC1123 with numeric zone RFC3339     = "2006-01-02T15:04:05Z07:00"
//     RFC1123Z    = "Mon, 02 Jan 2006 15:04:05 -0700"
//     RFC3339Nano = "2006-01-02T15:04:05.999999999Z07:00"

//  // Handy time stamps. Stamp      = "Jan _2 15:04:05"
//     Kitchen     = "3:04PM"
//     StampMilli = "Jan _2 15:04:05.000"
//     StampMicro = "Jan _2 15:04:05.000000"
//     StampNano  = "Jan _2 15:04:05.000000000"
// )

func TimeName(t time.Time) string {
	return t.Format("2006-01-02 15:04:05 -0700")
}

func PrintPerformTimeConsuming(funcName string, startTime time.Time) {
	timeConumeSeconds := time.Now().Sub(startTime).Seconds()
	ZLog.Debugf("function[%s] using [%f] seconds", funcName, timeConumeSeconds)
}

func GetDayEndTimeLocal() time.Time {
	year, month, day := time.Now().Date()
	endTime := time.Date(year, month, day, 23, 59, 59, 0, time.Local)
	return endTime
}

func GetDayEndTimeUtc() time.Time {
	year, month, day := time.Now().UTC().Date()
	endTime := time.Date(year, month, day, 23, 59, 59, 0, time.UTC)
	return endTime
}

func GetNextDayStartTimeLocal() time.Time {
	year, month, day := time.Now().Date()
	startTime := time.Date(year, month, day+1, 0, 0, 0, 0, time.Local)
	return startTime
}

func GetNextDayStartTimeUtc() time.Time {
	year, month, day := time.Now().UTC().Date()
	startTime := time.Date(year, month, day+1, 0, 0, 0, 0, time.UTC)
	return startTime
}

func GetNextDayStartTimeByLocation(location *time.Location) time.Time {
	year, month, day := time.Now().In(location).Date()
	startTime := time.Date(year, month, day+1, 0, 0, 0, 0, location)
	return startTime
}

func GetWeekName(location *time.Location) int {
	if location == nil {
		location = time.Local
	}
	year, week := time.Now().In(location).ISOWeek()
	weekName, _ := strconv.ParseInt(fmt.Sprintf("%d%02d", year, week), 10, 32)
	return int(weekName)
}

func GetWeekDay(location *time.Location) int32 {
	if location == nil {
		location = time.Local
	}
	return int32(time.Now().In(location).Weekday())
}

func GetMonthName(location *time.Location) int {
	if location == nil {
		location = time.Local
	}
	year, month, _ := time.Now().In(location).Date()
	monthName, _ := strconv.ParseInt(fmt.Sprintf("%d%02d", year, month), 10, 32)
	return int(monthName)
}

// 计算某年某月的天数
func GetMonthlyDayCount(year int, month int) int {
	var days int
	if month != 2 {
		if month == 4 || month == 6 || month == 9 || month == 11 {
			days = 30
		} else {
			days = 31
		}
	} else {
		if ((year%4) == 0 && (year%100) != 0) || (year%400) == 0 {
			days = 29
		} else {
			days = 28
		}
	}
	return days
}
