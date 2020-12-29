/*
 * @Author: calmwu
 * @Date: 2018-01-11 15:21:24
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-01-27 15:50:37
 * @Comment:
 */

package common

import (
	"net"
	"sailcraft/base"
	"strings"
	"time"
)

func QueryGeoInfo(remoteIp string) (ISOCountry, CountryName string) {
	ISOCountry = "UNKNOWN"
	CountryName = "UNKNOWN"

	ip := net.ParseIP(remoteIp)
	record, err := GConfig.geoDB.City(ip)
	if err != nil || record == nil {
		base.GLog.Error("GeoIP Query city by ip[%s] failed", remoteIp)
		return
	}
	ISOCountry = record.Country.IsoCode
	name, ok := record.Country.Names["zh-CN"]
	if ok {
		CountryName = name
	}

	if len(ISOCountry) == 0 && 0 == strings.Compare(CountryName, "Unknown") {
		ISOCountry = remoteIp
	}
	return
}

func GetCassandraMillionSeconds() int64 {
	//return time.Now().Format("2006-01-02 15:04:05+0800")
	return time.Now().UnixNano() / 1000000
}
