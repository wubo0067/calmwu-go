/*
 * @Author: calmwu
 * @Date: 2018-11-15 16:43:04
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-11-15 16:50:00
 */

package main

import "fmt"

type BaseInfo struct {
	b1 string
	b2 string
}

type IntegrationInfo struct {
	BaseInfo
	i1 int
	i2 int
}

func main() {
	var b BaseInfo
	b.b1 = "b1----"
	b.b2 = "b2++++"

	var i IntegrationInfo
	i.BaseInfo = b
	fmt.Printf("i.b1[%s] i.b2[%s]\n", i.b1, i.b2)
	fmt.Printf("i.b1 addr[%x] i.b2 addr[%x]\n", &i.b1, &i.b2)
	fmt.Printf("i.BaseInfo.b1 addr[%x] i.BaseInfo.b2 addr[%x]\n", &i.BaseInfo.b1, &i.BaseInfo.b2)
}
