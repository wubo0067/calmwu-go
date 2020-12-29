/*
 * @Author: calmwu
 * @Date: 2017-11-16 11:08:49
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-11-16 11:15:16
 */

package main

import (
	"fmt"
	"net"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Hello from consul")
	addr, _ := net.InterfaceAddrs()
	fmt.Fprintf(w, "Hello from consul, %v", addr)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("health check!")
}

func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/health", healthHandler)
	http.ListenAndServe("10.186.40.75:8990", nil)
}
