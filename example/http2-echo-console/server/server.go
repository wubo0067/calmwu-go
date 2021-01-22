/*
 * @Author: CALM.WU
 * @Date: 2021-01-19 16:34:15
 * @Last Modified by: CALM.WU
 * @Last Modified time: 2021-01-22 11:02:37
 */

package main

import (
	"io"
	"log"
	"net"
	"net/http"

	httpRouter "github.com/julienschmidt/httprouter"
)

type Server struct {
	router *httpRouter.Router
}

func (s *Server) Initialize() error {
	s.router = httpRouter.New()
	s.router.POST("/echo-console", s.echoConsoleHandler)

	// create the http server
	httpSrv := &http.Server{
		Handler: s.router,
	}

	httpSrvListener, err := net.Listen("tcp", ":30009")
	if err != nil {
		log.Fatalf("net listent :30009 failed. err:%s", err.Error())
	}

	log.Println("Http Server is listening on :30009")

	return httpSrv.ServeTLS(httpSrvListener, "../server.crt", "../server.key")
}

func (s *Server) echoConsoleHandler(w http.ResponseWriter, req *http.Request, _ httpRouter.Params) {
	// 只支持http2协议
	if req.ProtoMajor != 2 {
		log.Println("Not a HTTP/2 request, rejected!")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer req.Body.Close()

	rdBuf := make([]byte, 4*1024)

	for {
		rdLen, err := req.Body.Read(rdBuf)
		log.Printf("read from body %d bytes", rdLen)
		if rdLen > 0 {
			w.Write(rdBuf[:rdLen])

			if f, ok := w.(http.Flusher); ok {
				log.Println("flush to client")
				f.Flush()
			}
		}

		if err != nil {
			log.Printf("receive err:%s", err.Error())
			if err == io.EOF {
				w.Header().Set("Status", "200 OK, Read Completed")
				log.Println("client closed")
			}
			break
		}
	}
}

func main() {
	server := &Server{}
	server.Initialize()
}
