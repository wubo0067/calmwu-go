/*
 * @Author: calmwu
 * @Date: 2017-09-18 10:36:33
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-09-18 10:37:04
 * @Comment:
 */

package base

import (
	"bytes"
	"fmt"
	"strings"
)

func Join(a []interface{}, sep string) string {
	if len(a) == 0 {
		return ""
	}
	if len(a) == 1 {
		return fmt.Sprintf("%v", a[0])
	}

	buffer := &bytes.Buffer{}
	buffer.WriteString(fmt.Sprintf("%v", a[0]))
	for i := 1; i < len(a); i++ {
		buffer.WriteString(sep)
		buffer.WriteString(fmt.Sprintf("%v", a[i]))
	}
	return buffer.String()
}

func ArrayToString(a []int, delim string) string {
	return strings.Trim(strings.Replace(fmt.Sprint(a), " ", delim, -1), "[]")
	//return strings.Trim(strings.Join(strings.Split(fmt.Sprint(a), " "), delim), "[]")
	//return strings.Trim(strings.Join(strings.Fields(fmt.Sprint(a)), delim), "[]")
}
