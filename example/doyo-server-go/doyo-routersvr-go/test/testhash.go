/*
 * @Author: calmwu
 * @Date: 2018-09-28 11:33:30
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-10-12 16:35:50
 */

package main

import (
	base "doyo-server-go/doyo-base-go"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func main() {
	secs := time.Now().Unix()
	fmt.Println(secs)

	n := base.HashStr2Uint32("")
	fmt.Println(n)

	n = base.HashStr2Uint32("doyo-server-go/doyo-base-go")
	fmt.Println(n)

	servTypeLst := []string{"123", "234", "5555"}
	fmt.Println(servTypeLst[len(servTypeLst)-1])
	for index := range servTypeLst {
		fmt.Println(index)
	}

	a := "/data/offline_media/1/11.mp4"
	as := strings.Split(a, "/")
	lNumName := strings.Split(as[len(as)-1], ".")[0]
	lNum, _ := strconv.Atoi(strings.Split(lNumName, ".")[0])
	fmt.Println(lNum)

}
