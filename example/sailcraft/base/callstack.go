/*
 * @Author: calmwu
 * @Date: 2017-11-06 17:24:24
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-04-14 14:12:07
 * @Comment:
 */

package base

import (
	"fmt"
	"runtime"
	"strings"
)

func GetCallStack() string {
	pcs := make([]uintptr, 32)
	count := runtime.Callers(0, pcs)
	stackPC := pcs[:count]
	frames := runtime.CallersFrames(stackPC)
	var (
		f      runtime.Frame
		more   bool
		result string
		index  int
	)
	for {
		f, more = frames.Next()
		if index = strings.Index(f.File, "src"); index != -1 {
			// trim GOPATH or GOROOT prifix
			f.File = string(f.File[index+4:])
		}
		result = fmt.Sprintf("%s%s\n\t%s:%d\n", result, f.Function, f.File, f.Line)
		if !more {
			break
		}
	}
	return result
}

func DumpStacks() {
	buf := make([]byte, 262144)
	buf = buf[:runtime.Stack(buf, true)]

	fmt.Printf("=== BEGIN goroutine stack dump ===\n%s\n=== END goroutine stack dump ===", buf)
}
