/*
 * @Author: calmwu
 * @Date: 2018-11-14 15:16:18
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-11-15 15:29:39
 */

package main

import (
	"flag"
	"fmt"
	"time"

	doyobase "doyo-server-go/doyo-base-go"

	"github.com/seiflotfy/cuckoofilter"
)

/*
[calmwu@localhost test]$ ./test_cuckoofilter --capacity=10000000
insert 10000000 records, memory info: Alloc = 23 MiB	TotalAlloc = 245 MiB	Sys = 69 MiB	NumGC = 17
[calmwu@localhost test]$ ./test_cuckoofilter --capacity=10000000 --type=map
insert 10000000 records, memory info: Alloc = 1086 MiB	TotalAlloc = 1469 MiB	Sys = 1464 MiB	NumGC = 11
*/

const (
	fmtUserID = "doyo%08d"
)

var (
	cmdParamType     = flag.String("type", "ck", "ck/map")
	cmdParamCapacity = flag.Int("capacity", 10000, "")
	ckFilter         *cuckoofilter.CuckooFilter
	mapFilter        map[string]interface{}
	findList         = []string{"doyo00123456", "doyo01123456", "doyo11123456", "doyo14523456", "doyo14521256", "doyo56521256"}
)

func testMemUsage() {
	if *cmdParamType == "ck" {
		for i := 0; i < *cmdParamCapacity; i++ {
			userID := fmt.Sprintf(fmtUserID, i)
			ckFilter.InsertUnique([]byte(userID))
		}
	} else if *cmdParamType == "map" {
		for i := 0; i < *cmdParamCapacity; i++ {
			userID := fmt.Sprintf(fmtUserID, i)
			mapFilter[userID] = struct{}{}
		}
	}

	memUsage := doyobase.MemUsage()
	fmt.Printf("insert %d records, memory info: %s", *cmdParamCapacity, memUsage)
}

func testCKFindTimeTake() {
	now := time.Now()
	defer doyobase.TimeTaken(now, "testCKFindTimeTake")

	if *cmdParamType == "ck" {
		for index := range findList {
			target := findList[index]
			bExist := ckFilter.Lookup([]byte(target))
			if bExist {
				fmt.Printf("User:%s exist\n", target)
			} else {
				fmt.Printf("User:%s is not exist\n", target)
			}
		}
	} else if *cmdParamType == "map" {
		for index := range findList {
			target := findList[index]
			_, bExist := mapFilter[target]
			if bExist {
				fmt.Printf("User:%s exist\n", target)
			} else {
				fmt.Printf("User:%s is not exist\n", target)
			}
		}
	}

}

func main() {
	flag.Parse()

	if *cmdParamType == "ck" {
		ckFilter = cuckoofilter.NewCuckooFilter(uint(*cmdParamCapacity))
	} else if *cmdParamType == "map" {
		mapFilter = make(map[string]interface{})
	}
	testMemUsage()
	testCKFindTimeTake()
}
