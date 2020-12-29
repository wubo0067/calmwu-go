/*
 * @Author: calmwu
 * @Date: 2018-07-06 17:14:35
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-07-07 11:30:20
 * @Comment:
 */

package main

import (
	"fmt"
	"os"
	"sailcraft/base"
	"time"

	"github.com/emirpasic/gods/sets/hashset"
	"github.com/gocql/gocql"
	"github.com/urfave/cli"
)

var AreaSoutheastAsia = []string{"BN",
	"HK",
	"ID",
	"KH",
	"LA",
	"MM",
	"MY",
	"PH",
	"SG",
	"TH",
	"VN"}

var AreaEurope = []string{"AR",
	"AU",
	"BE",
	"BG",
	"BO",
	"BR",
	"BY",
	"BZ",
	"CA",
	"CH",
	"CL",
	"CY",
	"CZ",
	"DE",
	"ES",
	"FI",
	"FR",
	"GB",
	"GE",
	"GR",
	"HR",
	"HU",
	"IT",
	"LT",
	"LU",
	"NL",
	"NZ",
	"PL",
	"PT",
	"RO",
	"RS",
	"RU",
	"SE",
	"UA",
	"US"}

var (
	appFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "date, s",
			Value: "now",
			Usage: "指定计算的日期，例如20180528",
		},
	}
	// 计算的日期间隔
	dayIntervals   = []int{1, 3, 7, 15, 30}
	cassandraHosts = []string{"10.161.118.71", "10.161.118.83"}
)

const (
	// 查询当天注册的用户
	CqlSelTblUserOnlineFmt = "select uin, ISOCountryCode from tbl_useronline where createtime>='%s 00:00:00+0000' and createtime<='%s 23:59:59+0000' allow filtering"
	// 按日期查询活跃用户
	CqlSelDateActiveFmt = "select uin from tbl_dateactiveuin_%s"
	// 付费统计
	CqlSelDailyCountryRevenue = "select date, isocountrycode, totalrechargecount, totalrevenue from tbl_DailyCountryRevenue"
	// 分区注册统计
	CqlSelDailyCountryRegisterCount = "select date, isocountrycode, registercount from tbl_DailyCountryRegisterCount"
	// 统计每天充值人数
	CqlSelDailyRechargePlayerFmt = "select uin, isocountrycode from tbl_UserRechargeRecord where time>='%s 00:00:00+0000' and time<='%s 23:59:59+0000' allow filtering"
	// 按区域统计活跃
	CqlSelCountryDauFmt = "select isocountrycode, logincount from tbl_CountryDau where date='%s'"
)

func initCassandra() *gocql.Session {
	cluster := gocql.NewCluster(cassandraHosts[0:]...)
	cluster.Keyspace = "ks_statisticmodule"
	cluster.Consistency = gocql.Quorum

	base.GLog.Debug("Now start cassandra cluster create session!")
	session, err := cluster.CreateSession()
	if err != nil {
		base.GLog.Error("Cassandra cluster:%v create session failed! error:[%s]", cassandraHosts,
			err.Error())
		return nil
	} else {
		base.GLog.Info("Cassandra cluster:%v create session successed!", cassandraHosts)
	}
	return session
}

func changeDateFmt(date string) string {
	return fmt.Sprintf("%s-%s-%s", date[:4], date[4:6], date[6:])
}

func queryUins(session *gocql.Session, cqlContent string, europeCountrySet *hashset.Set, europeUins *hashset.Set,
	southAsisCountrySet *hashset.Set, southAsiaUins *hashset.Set) {
	var uin int
	var isoCountryCode string

	base.GLog.Debug("%s", cqlContent)

	iter := session.Query(cqlContent).PageSize(200).Iter()

	for iter.Scan(&uin, &isoCountryCode) {
		if europeCountrySet.Contains(isoCountryCode) {
			europeUins.Add(uin)
		} else if southAsisCountrySet.Contains(isoCountryCode) {
			southAsiaUins.Add(uin)
		}
	}

	iter.Close()
}

type RegionRevenue struct {
	southasiatotalrechargecount int
	southasiatotalrevenue       float32
	europetotalrechargecount    int
	europetotalrevenue          float32
}

type RegionRegister struct {
	southasiaregistercount int
	europeregistercount    int
}

func calcAreaDau(cSession *gocql.Session, startDate string, southAsisCountrySet *hashset.Set) {
	nowYearDay := time.Now().YearDay()

	iterDate, _ := time.Parse("20060102 00:00:00", fmt.Sprintf("%s 00:00:00", startDate))
	startYearDay := iterDate.YearDay()

	for ; startYearDay < nowYearDay; startYearDay++ {
		dateName := fmt.Sprintf("%d%02d%02d", iterDate.Year(), iterDate.Month(), iterDate.Day())

		var isoCountryCode string
		var loginCount int
		europeDau := 0
		southAsiaDau := 0

		cqlSelCountryDau := fmt.Sprintf(CqlSelCountryDauFmt, dateName)

		iter := cSession.Query(cqlSelCountryDau).PageSize(200).Iter()
		for iter.Scan(&isoCountryCode, &loginCount) {
			if southAsisCountrySet.Contains(isoCountryCode) {
				southAsiaDau += loginCount
			} else {
				europeDau += loginCount
			}
		}
		iter.Close()
		base.GLog.Debug("date:%s southAsiaDau:%d\teuropeDau:%d", dateName, southAsiaDau, europeDau)

		iterDate = iterDate.AddDate(0, 0, 1)
	}
}

func calcDailyRechargePlayerCount(cSession *gocql.Session, startDate string, southAsisCountrySet *hashset.Set) {
	nowYearDay := time.Now().YearDay()

	iterDate, _ := time.Parse("20060102 00:00:00", fmt.Sprintf("%s 00:00:00", startDate))
	startYearDay := iterDate.YearDay()

	for ; startYearDay < nowYearDay; startYearDay++ {
		base.GLog.Debug("startYearDay[%d] nowYearDay[%d]", startYearDay, nowYearDay)

		dateName := fmt.Sprintf("%d%02d%02d", iterDate.Year(), iterDate.Month(), iterDate.Day())

		var uin int
		var isoCountryCode string
		dailyUinSet := hashset.New()
		europeCount := 0
		southAsiaCount := 0

		cqlSelDailyRechargePlayer := fmt.Sprintf(CqlSelDailyRechargePlayerFmt, changeDateFmt(dateName), changeDateFmt(dateName))
		//base.GLog.Debug("%s", cqlSelDailyRechargePlayer)

		iter := cSession.Query(cqlSelDailyRechargePlayer).PageSize(200).Iter()
		for iter.Scan(&uin, &isoCountryCode) {
			if !dailyUinSet.Contains(uin) {
				dailyUinSet.Add(uin)

				if southAsisCountrySet.Contains(isoCountryCode) {
					southAsiaCount++
				} else {
					europeCount++
				}
			}
		}
		iter.Close()
		base.GLog.Debug("date:%s recharge player southAsiaCount:%d\teuropeCount:%d", dateName, southAsiaCount, europeCount)

		iterDate = iterDate.AddDate(0, 0, 1)
	}
}

func calcRegionRegister(cSession *gocql.Session, southAsisCountrySet *hashset.Set) {
	iter := cSession.Query(CqlSelDailyCountryRegisterCount).PageSize(200).Iter()

	var date string
	var isoCountryCode string
	var registercount int

	regionRegisterStatis := make(map[string]*RegionRegister)

	var isSouthAsia bool = true
	for iter.Scan(&date, &isoCountryCode, &registercount) {
		if southAsisCountrySet.Contains(isoCountryCode) {
			isSouthAsia = true
		} else {
			isSouthAsia = false
		}

		v, ok := regionRegisterStatis[date]
		if !ok {
			v = new(RegionRegister)
			regionRegisterStatis[date] = v
		}

		if isSouthAsia {
			v.southasiaregistercount += registercount
		} else {
			v.europeregistercount += registercount
		}
	}

	iter.Close()

	for k, r := range regionRegisterStatis {
		base.GLog.Debug("%s\t%d\t%d", k, r.europeregistercount, r.southasiaregistercount)
	}
}

func calcRegionRevenue(cSession *gocql.Session, southAsisCountrySet *hashset.Set) {
	iter := cSession.Query(CqlSelDailyCountryRevenue).PageSize(200).Iter()

	var date string
	var isoCountryCode string
	var totalrechargecount int
	var totalrevenue float32

	regionRevenueStatis := make(map[string]*RegionRevenue)

	var isSouthAsia bool = true
	for iter.Scan(&date, &isoCountryCode, &totalrechargecount, &totalrevenue) {
		if southAsisCountrySet.Contains(isoCountryCode) {
			isSouthAsia = true
		} else {
			isSouthAsia = false
		}

		v, ok := regionRevenueStatis[date]
		if !ok {
			v = new(RegionRevenue)
			regionRevenueStatis[date] = v
		}

		if isSouthAsia {
			base.GLog.Debug("southasia %s %d %f\n", date, totalrechargecount, totalrevenue)
			v.southasiatotalrechargecount += totalrechargecount
			v.southasiatotalrevenue += totalrevenue
		} else {
			base.GLog.Debug("europe %s %d %f", date, totalrechargecount, totalrevenue)
			v.europetotalrechargecount += totalrechargecount
			v.europetotalrevenue += totalrevenue
		}
	}

	iter.Close()

	for k, r := range regionRevenueStatis {
		base.GLog.Debug("%s\t%d\t%f\t%d\t%f", k, r.europetotalrechargecount, r.europetotalrevenue,
			r.southasiatotalrechargecount, r.southasiatotalrevenue)
	}
}

func areaCalcPreserve(date string) {
	cSession := initCassandra()
	if cSession == nil {
		base.GLog.Error("cSession is nil")
		return
	}

	// 初始化地区set
	southAsisCountrySet := hashset.New()
	europeCountrySet := hashset.New()
	for index := range AreaSoutheastAsia {
		southAsisCountrySet.Add(AreaSoutheastAsia[index])
	}
	for index := range AreaEurope {
		europeCountrySet.Add(AreaEurope[index])
	}

	dailyEuropeRegisterUins := hashset.New()
	dailySouthAsiaRegisterUins := hashset.New()
	// 在tblUserOnline中查询当日注册的用户
	cqlQueryUserOnline := fmt.Sprintf(CqlSelTblUserOnlineFmt, changeDateFmt(date), changeDateFmt(date))
	queryUins(cSession, cqlQueryUserOnline, europeCountrySet, dailyEuropeRegisterUins,
		southAsisCountrySet, dailySouthAsiaRegisterUins)

	base.GLog.Debug("europeUins:%d southAsiaUins:%d", dailyEuropeRegisterUins.Size(), dailySouthAsiaRegisterUins.Size())

	// 按日期查询活跃表
	for _, interval := range dayIntervals {

		activeDate, _ := time.Parse("20060102 00:00:00", fmt.Sprintf("%s 00:00:00", date))
		activeDate = activeDate.AddDate(0, 0, interval)
		activeDateName := fmt.Sprintf("%d%02d%02d", activeDate.Year(), activeDate.Month(), activeDate.Day())

		europeActiveCount := 0
		southAsiaActiveCount := 0
		var uin int

		cqlQueryActiveUser := fmt.Sprintf(CqlSelDateActiveFmt, activeDateName)
		base.GLog.Debug("%s", cqlQueryActiveUser)

		iter := cSession.Query(cqlQueryActiveUser).PageSize(200).Iter()

		for iter.Scan(&uin) {
			if dailyEuropeRegisterUins.Contains(uin) {
				europeActiveCount++
			} else if dailySouthAsiaRegisterUins.Contains(uin) {
				southAsiaActiveCount++
			}
		}

		iter.Close()

		base.GLog.Debug("%s europeActiveCount:%d southAsiaActiveCount:%d", activeDate.String(), europeActiveCount, southAsiaActiveCount)
		base.GLog.Debug("%s europe:%f southasia:%f", date,
			float32(europeActiveCount)/float32(dailyEuropeRegisterUins.Size()),
			float32(southAsiaActiveCount)/float32(dailySouthAsiaRegisterUins.Size()))
	}

	calcRegionRevenue(cSession, southAsisCountrySet)
	calcRegionRegister(cSession, southAsisCountrySet)
	calcDailyRechargePlayerCount(cSession, date, southAsisCountrySet)
	calcAreaDau(cSession, date, southAsisCountrySet)
}

func main() {
	base.InitLog("export.log")
	defer base.GLog.Close()

	app := cli.NewApp()
	app.Name = "area_calcpreserve"
	app.Usage = "area_calcpreserve --date=yyyymmdd"
	app.Flags = appFlags

	app.Action = func(c *cli.Context) error {
		date := c.String("date")
		areaCalcPreserve(date)
		return nil
	}

	app.Run(os.Args)
}
