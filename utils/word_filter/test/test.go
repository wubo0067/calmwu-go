/*
 * @Author: calmwu
 * @Date: 2018-04-28 11:12:02
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-04-28 11:52:56
 * @Comment:
 */

package main

import (
	"doyo-server-go/doyo-routersvr-go/base/word_filter"
	"fmt"
	"time"
)

var (
	sensitiveWordFile = "../../../sysconf/SensitiveWordPrecision.txt"
	dicFiles          = []string{sensitiveWordFile}
)

func main() {
	word_filter.LoadDicFiles(dicFiles)
	time.Sleep(1000 * time.Millisecond) //asyn,sleep until trie is built

	str := []rune("学习大大 and 李大钊 and System hello filter かんりにん")
	fmt.Println(string(str))
	word_filter.FilterText(sensitiveWordFile, str, []rune{}, '*')
	fmt.Println(string(str))
	fmt.Printf("\n------------------------------------------------\n")
}
