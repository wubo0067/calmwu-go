/*
 * @Author: lizhengde
 * @Date: 2018-02-03 10:32:00
 * @Last Modified by: lizhengde
 * @Last Modified time: 2018-02-03 10:32:00
 * @Comment:
 */

package utils

import (
	"fmt"
	"time"
)

const (
	SecondsPerMinute = 60
	SecondsPerHour   = 60 * 60
	SecondsPerDay    = 24 * SecondsPerHour
	SecondsPerWeek   = 7 * SecondsPerDay
)

type LocalizedTime struct {
	loc *time.Location
}

var (
	GLocalizedTime *LocalizedTime = new(LocalizedTime)
)

func init() {
	GLocalizedTime.loc = time.UTC
}

/*
	设置本地区域，如"UTC"
*/
func (this *LocalizedTime) SetLocale(loc string) error {
	location, err := time.LoadLocation(loc)

	if err != nil {
		return err
	}

	this.loc = location

	return nil
}

/*
	获取本地区域
*/
func (this *LocalizedTime) GetLocale() *time.Location {
	if this.loc == nil {
		return time.UTC
	}

	return this.loc
}

/*
	获取当前时间
*/
func (this *LocalizedTime) Now() time.Time {
	return time.Now().In(this.GetLocale())
}

/*
	获取当前时间戳（秒）
*/
func (this *LocalizedTime) SecTimeStamp() int64 {
	return time.Now().Unix()
}

/*
	获取当前时间戳（纳秒）
*/
func (this *LocalizedTime) NSecTimeStamp() int64 {
	return time.Now().UnixNano()
}

/*
	根据给定的秒和纳秒返回time结构体（距离UTC时间1970年1月1日时间）
*/
func (this *LocalizedTime) Unix(sec int64, nsec int64) time.Time {
	return time.Unix(sec, nsec).In(this.GetLocale())
}

/*
	获取当前日期（年，月，日）
*/
func (this *LocalizedTime) NowDate() (year int, month time.Month, day int) {
	return this.Now().Date()
}

/*
	根据给定的秒和纳秒返回当前日期（年，月，日）
*/
func (this *LocalizedTime) UnixDate(sec int64, nsec int64) (year int, month time.Month, day int) {
	return this.Unix(sec, nsec).Date()
}

/*
	根据当天的时分秒返回time结构
*/
func (this *LocalizedTime) TodayClock(hour int, minute int, sec int) (time.Time, error) {
	clockStr := fmt.Sprintf("%s %02d:%02d:%02d", this.Now().Format("2006-01-02"), hour, minute, sec)

	clock, err := time.ParseInLocation("2006-01-02 15:04:05", clockStr, this.GetLocale())

	if err != nil {
		return time.Time{}, err
	}

	return clock, nil
}

/*
	根据年月日时分秒返回time结构
*/
func (this *LocalizedTime) Clock(year int, month int, day int, hour int, minute int, sec int) (time.Time, error) {
	clockStr := fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d", year, month, day, hour, minute, sec)

	clock, err := time.ParseInLocation("2006-01-02 15:04:05", clockStr, this.GetLocale())

	if err != nil {
		return time.Time{}, err
	}

	return clock, nil
}

/*
	传入一个时间戳（秒），判断是不是昨天
*/
func (this *LocalizedTime) IsYesterday(sec int64) bool {
	return this.IsToday(sec + SecondsPerDay)
}

/*
	传入一个时间戳（秒），判断是不是今天
*/
func (this *LocalizedTime) IsToday(sec int64) bool {
	y, m, d := this.UnixDate(sec, 0)
	nY, nM, nD := this.NowDate()

	return nY == y && nM == m && nD == d
}

func (this *LocalizedTime) IsCurrentWeek(sec int64) bool {
	return this.IsTheSameWeek(this.SecTimeStamp(), sec)
}

func (this *LocalizedTime) IsTheSameWeek(secX int64, secY int64) bool {
	t := this.Unix(secX, 0)
	weekDay := t.Weekday()
	minDay := secX/int64(SecondsPerDay) - int64(weekDay) + 1
	maxDay := minDay + int64(time.Saturday)

	dayY := secY/int64(SecondsPerDay) + 1
	if dayY >= minDay && dayY <= maxDay {
		return true
	}

	return false
}
