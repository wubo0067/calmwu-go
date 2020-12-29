/*
 * @Author: calm.wu
 * @Date: 2019-05-21 14:23:02
 * @Last Modified by: calm.wu
 * @Last Modified time: 2019-05-21 14:25:31
 */

package main

import "fmt"

func main() {
	s := []byte("")

	s1 := append(s, 'a')
	s2 := append(s, 'b')

	// 如果有此行，打印的结果是 a b，否则打印的结果是b b
	fmt.Println(s1, "===", s2)
	fmt.Println(string(s1), string(s2))
}
