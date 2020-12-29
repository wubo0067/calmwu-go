/*
 * @Author: calm.wu
 * @Date: 2019-09-23 10:13:35
 * @Last Modified by: calm.wu
 * @Last Modified time: 2019-09-23 10:14:11
 */

package main

// #include <stdio.h>
// #include <stdlib.h>
//
// static void myprint(char* s) {
//   printf("%s\n", s);
// }
import "C"
import "unsafe"

func main() {
	cs := C.CString("Hello from stdio")
	C.myprint(cs)
	C.free(unsafe.Pointer(cs)) // yoko注，去除这行将发生内存泄漏
}
