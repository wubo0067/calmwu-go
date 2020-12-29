/*
 * @Author: calmwu
 * @Date: 2017-06-26 09:52:55
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-07-06 17:36:05
 */
package main

import (
	"fmt"
	"os"
	"time"

	"bufio"

	"github.com/emirpasic/gods/sets/hashset"
	"github.com/urfave/cli"
)

var (
	version   = "0.0.1"
	buildtime = ""

	appFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "date, s",
			Value: "now",
			Usage: "start date",
		},
		cli.StringFlag{
			Name:  "dumppath, p",
			Value: "/srv/sandmonk/PreserveFiles",
			Usage: "dump file path",
		},
	}

	dayIntervals = []int{1, 3, 7, 15, 30}
)

func checkExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		// 检查是否是路径不存在错误
		return false, nil
	}
	return true, err
}

func calcActiveUinCount(date string, dumppath string, uinSet *hashset.Set) int {
	dateActiveUinCsv := fmt.Sprintf("%s/dateactiveuin_%s.csv", dumppath, date)
	dateActiveUinCsvF, err := os.Open(dateActiveUinCsv)
	if err != nil {
		fmt.Printf("Open CSV[%s] failed! reason[%s]\n", dateActiveUinCsv, err.Error())
		return -1
	}
	defer dateActiveUinCsvF.Close()

	var dateActiveUinCount = 0
	dateActiveUinCsvReader := bufio.NewReader(dateActiveUinCsvF)
	for {
		uin, _, err := dateActiveUinCsvReader.ReadLine()
		if err != nil {
			break
		}
		uinName := string(uin)
		if uinSet.Contains(uinName) {
			dateActiveUinCount++
		}
	}

	return dateActiveUinCount
}

func calcPreserve(date string, dumppath string) {
	// 判断导出目录是否存在
	ret, err := checkExists(dumppath)
	if err != nil || !ret {
		fmt.Printf("DumpPath[%s] is invalid!\n", dumppath)
		os.Exit(-1)
	}

	// 读取dateregisteruin_xxxxxxxx.csv文件
	dateRegisterUinCsv := fmt.Sprintf("%s/dateregisteruin_%s.csv", dumppath, date)
	dateRegisterUinCsvF, err := os.Open(dateRegisterUinCsv)
	if err != nil {
		fmt.Printf("Open CSV[%s] failed! reason[%s]\n", dateRegisterUinCsv, err.Error())
		os.Exit(-1)
	}
	defer dateRegisterUinCsvF.Close()

	var dateRegisterUinCount = 0
	dateRegisterUinSet := hashset.New()
	dateRegisterUinCsvReader := bufio.NewReader(dateRegisterUinCsvF)
	for {
		uin, _, err := dateRegisterUinCsvReader.ReadLine()
		if err != nil {
			break
		}
		dateRegisterUinSet.Add(string(uin))
		dateRegisterUinCount++
	}
	fmt.Printf("%s register uin count %d\n", date, dateRegisterUinCount)

	if dateRegisterUinCount == 0 {
		fmt.Printf("%s register player is zero!!!\n", date)
		os.Exit(0)
	}

	for _, interval := range dayIntervals {

		startTime, err := time.Parse("20060102 15:04:05", fmt.Sprintf("%s 00:00:00", date))
		if err != nil {
			fmt.Printf("Parse startdate failed! reason[%s]\n", err.Error())
			os.Exit(-1)
		}
		future := startTime.AddDate(0, 0, interval)
		year, month, day := future.Date()
		activeDate := fmt.Sprintf("%d%02d%02d", year, month, day)

		activeCount := calcActiveUinCount(activeDate, dumppath, dateRegisterUinSet)
		if activeCount >= 0 {
			preserve := float32(activeCount) / float32(dateRegisterUinCount)
			fmt.Printf("%s: %d days activeCount %d preserve %f\n", date, interval, activeCount, preserve)
		}
	}
}

func main() {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("dumpPreserveTable version=%s buildtime=%s\n", version, buildtime)
	}

	app := cli.NewApp()
	app.Name = "calcPreserve"
	app.Usage = "calc preserve"
	app.Flags = appFlags

	app.Action = func(c *cli.Context) error {
		date := c.String("date")
		dumppath := c.String("dumppath")
		calcPreserve(date, dumppath)

		return nil
	}

	app.Run(os.Args)

	fmt.Fprintf(os.Stderr, "[%s] %s exit!\n", app.Name, time.Now().String())

}
