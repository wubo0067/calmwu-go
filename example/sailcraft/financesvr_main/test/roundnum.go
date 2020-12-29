/*
 * @Author: calmwu
 * @Date: 2018-02-23 11:11:18
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-03-29 17:39:42
 */

package main

import (
	"fmt"
)

func RoundUp(x, base int32) int32 {
	return (((x) + ((base) - 1)) & (^((base) - 1)))
}

func RoundDown(x, base int32) int32 {
	return ((x) & -(base))
}

func NormalizeRefreshIntervalHours(intervalHours int32) int32 {
	hours := [8]int32{1, 2, 3, 4, 6, 8, 12, 24}

	for _, val := range hours {
		if intervalHours <= val {
			return val
		}
	}
	return 24
}

func CalcRefreshStartHours(intervalHours, dateHours int32) int32 {
	if dateHours > 23 {
		dateHours = 23
	}
	return dateHours / intervalHours * intervalHours
}

func main() {
	a := NormalizeRefreshIntervalHours(3)
	fmt.Println("a=", a)
	a = NormalizeRefreshIntervalHours(7)
	fmt.Println("a=", a)
	a = NormalizeRefreshIntervalHours(15)
	fmt.Println("a=", a)
	a = NormalizeRefreshIntervalHours(25)
	fmt.Println("a=", a)

	a = CalcRefreshStartHours(3, 23)
	fmt.Println("a=", a)

	a = CalcRefreshStartHours(2, 23)
	fmt.Println("a=", a)

	a = CalcRefreshStartHours(24, 1)
	fmt.Println("a=", a)

	a = CalcRefreshStartHours(12, 3)
	fmt.Println("a=", a)

	a = CalcRefreshStartHours(8, 9)
	fmt.Println("a=", a)

	c := []int{1, 2, 3, 4, 5, 6}
	fmt.Printf("c[2:]=%v\n", c[3:])
	fmt.Printf("%v\n", c[:4])

	var b []int = c
	c[2] = 10
	// 应该是一样的
	fmt.Printf("b:%v\n", b)
	fmt.Printf("c:%v\n", c)

	i := 19
	j := 100
	k := float32(float32(i) / float32(j) * 100)
	fmt.Printf("%v\n", k)

	l := make([]int, 10)
	copy(l[3:], c)
	fmt.Printf("%v\n", l)

	cp := &c
	fmt.Printf("cp=%v\n", *cp)
	c1 := []int{9, 8, 7, 6}
	*cp = c1
	fmt.Printf("cp=%v\n", *cp)
	fmt.Printf("c=%v\n", c)

	a1 := 500
	a2 := 100
	a3 := int32(float32(a1) * float32(a2) / 100)
	fmt.Println(a3)

	numLst := []int{0, 9, 8, 7, 6}
	for _, num := range numLst {
		fmt.Println(num)
	}

	a4 := 410
	b4 := 3000
	c4 := float64(a4) / float64(b4)
	c5 := c4 + 0.005
	c6 := int(c5 * 100)
	fmt.Println("c4=", c4, "c5=", c5, "c6=", c6)

	slice := []int{1, 2, 3, 4, 5, 6}
	index := 0
	slice = append(slice[:index], slice[index+1:]...)
	fmt.Println(slice)
	index = 4
	slice = append(slice[:index], slice[index+1:]...)
	fmt.Println(slice)

	chance := 100
	var totalChance float32 = 25000.00
	fmt.Println(totalChance)
	f1 := int((float32(chance)/totalChance + 0.005) * 100)
	fmt.Println(f1)
	f2 := (float32(chance)/totalChance + 0.005) * 100.00
	fmt.Println(f2)
}
