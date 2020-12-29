/*
 * @Author: calm.wu
 * @Date: 2019-08-05 16:16:26
 * @Last Modified by: calm.wu
 * @Last Modified time: 2019-08-05 16:44:22
 */

package main

import "fmt"

func a() {
	i := 0
	defer fmt.Println(i) // 0
	i++
	return
}

func b() {
	for i := 0; i < 4; i++ {
		defer fmt.Println(i)
	}
}

func c() (i int) {
	defer func() { i++ }()
	return 1
}

func willPanic() {
	panic("will panic")
}

func testPanic() (i int) {
	i = 9

	defer func() {
		recover()
		//i = 11
	}()

	//willPanic()

	return 10
}

func main() {
	a()
	b()
	c := c()
	fmt.Println(c)
	d := testPanic()
	fmt.Print(d)
}
