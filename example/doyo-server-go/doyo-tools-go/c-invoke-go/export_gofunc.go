/*
 * @Author: calmwu
 * @Date: 2019-02-21 14:51:29
 * @Last Modified by: calmwu
 * @Last Modified time: 2019-02-21 19:19:10
 */

// package name: export_gofunc
package main

import "C"
import "fmt"

/* struct Vertex { int X; int Y; }; */

//export SayHello
func SayHello(name string) {
	fmt.Printf("export_gfunc says: Hello, %s\n", name)
}

//export SayBye
func SayBye() {
	fmt.Println("Nautilus says: Bye!")
}

//export printStruct
func printStruct(p *C.struct_Vertex) {
	fmt.Printf("p:%+v\n", p)
}

func main() {
	// We need the main function to make possible
	// CGO compiler to compile the package as C shared library
}

// go build -buildmode=c-shared -o export_gofunc.so export_gofunc.go
// go build -buildmode=c-archive -o export_gofunc.a export_gofunc.go
// set GOARCH=arm
// set GOOS=linux
