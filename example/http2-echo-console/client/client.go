/*
 * @Author: CALM.WU
 * @Date: 2021-01-19 16:58:07
 * @Last Modified by: CALM.WU
 * @Last Modified time: 2021-01-22 10:51:21
 */

package main

import (
	"bufio"
	"crypto/tls"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"golang.org/x/net/http2"
)

type Client struct {
	client *http.Client
}

func (c *Client) Initialize() {
	certs, err := tls.LoadX509KeyPair("../server.crt", "../server.key")
	if err != nil {
		log.Fatal(err.Error())
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			Certificates:       []tls.Certificate{certs},
			InsecureSkipVerify: true,
		},
	}

	http2.ConfigureTransport(tr)

	c.client = &http.Client{
		Transport: tr,
	}
}

func (c *Client) Console(rd io.ReadCloser) {
	req := &http.Request{
		Method: "POST",
		URL: &url.URL{
			Scheme: "https",
			Host:   "localhost:30009",
			Path:   "/echo-console",
		},
		Header: http.Header{},
		Body:   rd,
	}

	resp, err := c.client.Do(req)
	if err != nil {
		log.Fatalf("client.Do failed. %s", err.Error())
	}

	if resp.StatusCode == 500 {
		log.Fatalln("resp StatusCode is 500")
	}

	defer resp.Body.Close()

	bReader := bufio.NewReader(resp.Body)
	rBuf := make([]byte, 4*1024)

	totalBytesReceived := 0

	for {
		rLen, err := bReader.Read(rBuf)
		if rLen > 0 {
			totalBytesReceived += rLen
			log.Printf("totalBytesReceived: %d, content: %s", totalBytesReceived, string(rBuf[:rLen]))
		}

		if err != nil {
			if err == io.EOF {
				log.Println("End of interaction")
			}
			log.Println("---close---")
			break
		}
	}
}

func main() {
	rPipe, wPipe, err := os.Pipe()
	if err != nil {
		log.Fatalf("os.Pipe failed. %s", err.Error())
	}

	client := new(Client)
	client.Initialize()
	go client.Console(rPipe)

	lineBuf := make([]byte, 1024)
	for {
		log.Printf("> ")
		n, err := os.Stdin.Read(lineBuf)
		log.Printf("[%d %q %v]\n> ", n, lineBuf[:n], err)

		if err == nil {
			wPipe.Write(lineBuf[:n])
		} else {
			if err == io.EOF {
				log.Printf("Disconnect from the server")
				wPipe.Close()
			}
			break
		}
	}
	//io.Copy(wPipe, os.Stdin)

	time.Sleep(3 * time.Second)
	log.Println("---client exit---")
}
