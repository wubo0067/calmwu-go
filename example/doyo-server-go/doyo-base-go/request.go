/*
 * @Author: calmwu
 * @Date: 2017-09-20 17:12:00
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-11-23 15:00:04
 * @Comment:
 */

package base

import (
	"bytes"
	"compress/zlib"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
)

func UnpackRequest(c *gin.Context) *ProtoRequestS {
	bodyData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		ZLog.Errorf("Read request body failed! reason[%s]", err.Error())
		return nil
	}

	ZLog.Debugf("Request Data:\n%s", bodyData)

	var req ProtoRequestS
	err = json.Unmarshal(bodyData, &req)
	if err != nil {
		ZLog.Errorf("decode body failed! reason[%s]", err.Error())
		return nil
	}

	return &req
}

func SendResponse(c *gin.Context, res *ProtoResponseS) {
	response, err := json.Marshal(res)
	if err == nil {
		ZLog.Debugf("send respone to %s\nResponse Data:\n%s", c.Request.RemoteAddr, response)
		c.Data(http.StatusOK, "text/plain; charset=utf-8", response)
	} else {
		ZLog.Errorf("Json Marshal ProtoResponseS failed! reason[%s]", err.Error())
	}
}

func GetClientAddrFromGin(c *gin.Context) string {
	var remoteAddr string
	remoteAddrLst, ok := c.Request.Header["X-Real-Ip"]
	if !ok {
		remoteAddr = "Unknown"
	} else {
		remoteAddr = remoteAddrLst[0]
	}
	return remoteAddr
}

func UnpackClientRequest(c *gin.Context) (*ProtoRequestS, error) {
	var req ProtoRequestS
	dcompressR, _ := zlib.NewReader(c.Request.Body)
	err := json.NewDecoder(dcompressR).Decode(&req)
	return &req, err
}

func SendResponseToClient(c *gin.Context, res *ProtoResponseS) {
	var compressBuf bytes.Buffer
	compressW := zlib.NewWriter(&compressBuf)
	json.NewEncoder(compressW).Encode(res)
	compressW.Close()
	//ZLog.Debugf("compress size[%d]", compressBuf.Len())
	c.Data(http.StatusOK, "text/plain; charset=utf-8", compressBuf.Bytes())
}

func PostRequest(url string, req *ProtoRequestS) (*ProtoResponseS, error) {
	serialData, err := json.Marshal(req)
	if err != nil {
		ZLog.Errorf("PostRequest to url[%s] Marshal failed! reason[%s]",
			url, err.Error())
		return nil, err
	}

	res, err := http.Post(url, "text/plain; charset=utf-8", strings.NewReader(string(serialData)))
	if err != nil {
		ZLog.Errorf("PostRequest to url[%s] Post failed! reason[%s]",
			url, err.Error())
		return nil, err
	}

	if res != nil {
		defer res.Body.Close()
	}

	bodyData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		ZLog.Errorf("Read body failed! reason[%s]", err.Error())
		return nil, err
	}

	ZLog.Debugf("Response Data:\n%s", bodyData)

	var protoRes ProtoResponseS
	err = json.Unmarshal(bodyData, &protoRes)
	if err != nil {
		ZLog.Errorf("decode ProtoResponseS failed! reason[%s]", err.Error())
		return nil, err
	}
	return &protoRes, nil
}

func MapstructUnPackByJsonTag(m interface{}, rawVal interface{}) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:  "json",
		Metadata: nil,
		Result:   rawVal,
	})
	if err != nil {
		ZLog.Errorf("mapstructure.NewDecoder failed! reason[%s]", err.Error())
		return err
	}

	err = decoder.Decode(m)
	if err != nil {
		ZLog.Errorf("Decode %s failed! reason[%s]", reflect.TypeOf(m).String(), err.Error())
		return err
	}
	return nil
}

type WebItfResData struct {
	Param   interface{}
	RetCode int
}

type webItfResponseFunc func()

func RequestPretreatment(c *gin.Context, interfaceName string, realReqPtr interface{}) (*WebItfResData, webItfResponseFunc, error) {
	var err error
	req := UnpackRequest(c)
	if req == nil {
		err = fmt.Errorf("unpack interface[%s] request failed!", interfaceName)
		ZLog.Errorf(err.Error())
		return nil, nil, err
	}

	err = MapstructUnPackByJsonTag(req.ReqData.Params, realReqPtr)
	if err != nil {
		err = fmt.Errorf("Uin[%d] Decode %s failed! reason[%s]",
			req.Uin, reflect.Indirect(reflect.ValueOf(realReqPtr)).Type().String(), err.Error())
		ZLog.Errorf(err.Error())
		return nil, nil, err
	}

	webItfResData := new(WebItfResData)

	return webItfResData, func() {
		if req != nil {
			var res ProtoResponseS
			res.Version = req.Version
			res.EventId = req.EventId
			res.ReturnCode = ProtoReturnCode(webItfResData.RetCode)
			res.TimeStamp = time.Now().UTC().Unix()
			res.ResData.InterfaceName = req.ReqData.InterfaceName
			if err != nil {
				res.ResData.Params = err.Error()
			} else {
				res.ResData.Params = webItfResData.Param
			}
			SendResponse(c, &res)
		}
	}, nil
}

// reference: https://gist.github.com/dmichael/5710968
func timeoutDialer(connectTimeout time.Duration, readWritetimeout time.Duration) func(network, addr string) (c net.Conn, err error) {
	return func(network, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(network, addr, connectTimeout)
		if err != nil {
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(readWritetimeout))
		return conn, nil
	}
}

func NewTimeoutHttpClient(connectTimeout time.Duration, readWritetimeout time.Duration) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Dial: timeoutDialer(connectTimeout, readWritetimeout),
		},
	}
}

// golang http 长连接优化 https://blog.csdn.net/kdpujie/article/details/73177179
// https://www.tuicool.com/articles/2YrmQjV
// MaxIdleConn MaxIdleConnsPerHost=2

// https://colobu.com/2016/07/01/the-complete-guide-to-golang-net-http-timeouts/
// https://stackoverflow.com/questions/36773837/best-way-to-use-http-client-in-a-concurrent-application
// Clients are safe for concurrent use by multiple goroutines.
func NewBaseHttpClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			DisableKeepAlives:     false,
			TLSHandshakeTimeout:   10 * time.Second, // 限制TLS握手使用的时间
			ResponseHeaderTimeout: 10 * time.Second, // 限制读取response header的时间
			IdleConnTimeout:       90 * time.Second, // 连接最大空闲时间，超过这个时间就会被关闭。
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   10,
			ExpectContinueTimeout: 1 * time.Second, // 限制client在发送包含 Expect: 100-continue的header到收到继续发送body的response之间的时间等待。注意在1.6中设置这个值会禁用HTTP/2
		},
	}
}

// https://blog.csdn.net/wangshubo1989/article/details/77508738
// https://colobu.com/2016/06/07/simple-golang-tls-examples/ InsecureSkipVerify: true,
func GenerateTLSConfig() *tls.Config {
	// 先生成key
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	// X.509是一种非常通用的证书格式，生成证书
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	// 生成pem和key，写入内存对象
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{Certificates: []tls.Certificate{tlsCert}}
}
