/*
 * @Author: calmwu
 * @Date: 2019-08-24 21:20:54
 * @Last Modified by: calmwu
 * @Last Modified time: 2019-08-24 21:33:50
 */

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

var image []byte

func init() {
	var err error
	image, err = ioutil.ReadFile("calm.jpg")
	if err != nil {
		log.Panic(err.Error())
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Got connection: %s", r.Proto)
	log.Printf("r.URL.Path:%s\n", r.URL.Path)

	if r.URL.Path == "/image" {
		log.Println("Handling image")
		//w.Write([]byte("Hello Again!"))
		w.Header().Set("Content-Type", "image/png")
		w.Write(image)
		return
	}

	log.Println("Handling 1st")

	pusher, ok := w.(http.Pusher)
	if !ok {
		log.Println("Can't push to client")
	} else {
		// push后，客户端会自动的调用该接口
		err := pusher.Push("/image", nil)
		if err != nil {
			log.Printf("Failed push: %v", err)
		}
	}

	//w.Write([]byte("Hello"))
	w.Header().Add("Content-Type", "text/html")
	fmt.Fprintf(w, `<html><body><img src="/image"></body></html>`)
}

func main() {
	srv := &http.Server{
		Addr:    "calm.org:8000",
		Handler: http.HandlerFunc(handler),
	}

	log.Printf("Serving on https://calm.org:8000")
	log.Fatal(srv.ListenAndServeTLS("./server.crt", "./server.key"))
}
