/*
 * @Author: calmwu
 * @Date: 2017-09-23 11:33:04
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-09-23 11:34:45
 * @Comment:
 */

package main

import "sailcraft/base"
import "fmt"

func main() {
	key := "Captain66093"

	hashVal1 := base.HashStr2Uint32(key)
	hashVal2 := base.HashStr2Uint32(key)

	fmt.Printf("hashVal1:%d hashVal2:%d\n", hashVal1, hashVal2)

	s := []string{"1", "2", "3"}
	s = append(s[:2], s[3:]...)
	fmt.Printf("s:%v\n", s)
}
