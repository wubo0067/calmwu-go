package utils

import (
	crand "crypto/rand"
	"encoding/hex"
	"io"
	"math"
	"math/big"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"
)

const letterBytes = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ~!@#$%^&*<>?"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var (
	once sync.Once
	// SeededSecurely 变量
	SeededSecurely bool
)

// RandomInt 生成随机整数
func RandomInt(n int) int {
	rand.Seed(time.Now().Unix())
	rnd := rand.Intn(n)
	return rnd
}

// RandomBytes 随机生成字节数组
func RandomBytes(size int) ([]byte, error) {
	bytes := make([]byte, size)
	_, err := crand.Read(bytes)
	return bytes, err
}

// RandomRangeIn  trying to generate 8 digit numbers, the range would be (10000000, 99999999)
func RandomRangeIn(low, hi int) int {
	return low + rand.Intn(hi-low)
}

// SeedMathRand 设置随机种子
func SeedMathRand() {
	once.Do(func() {
		n, err := crand.Int(crand.Reader, big.NewInt(math.MaxInt64))
		if err != nil {
			rand.Seed(time.Now().UTC().UnixNano())
			return
		}
		rand.Seed(n.Int64())
		SeededSecurely = true
	})
}

// RandStringBytesMaskImpr 根据掩码生成随机字符串
func RandStringBytesMaskImpr(n int) string {
	b := make([]byte, n)
	// A rand.Int63() generates 63 random bits, enough for letterIdxMax letters!
	for i, cache, remain := n-1, rand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func generateID(r io.Reader) string {
	b := make([]byte, 32)
	for {
		if _, err := io.ReadFull(r, b); err != nil {
			panic(err) // This shouldn't happen
		}
		id := hex.EncodeToString(b)
		// if we try to parse the truncated for as an int and we don't have
		// an error then the value is all numeric and causes issues when
		// used as a hostname. ref #3869
		if _, err := strconv.ParseInt(id, 10, 64); err == nil {
			continue
		}
		return id
	}
}

// GenerateRandomID 生成随机uuid
func GenerateRandomID() string {
	return generateID(crand.Reader)
}

// GenerateRandomPrivateMacAddr 生成mac地址
func GenerateRandomPrivateMacAddr() (string, error) {
	buf := make([]byte, 6)
	_, err := crand.Read(buf)
	if err != nil {
		return "", err
	}

	// Set the local bit for local addresses
	// Addresses in this range are local mac addresses:
	// x2-xx-xx-xx-xx-xx , x6-xx-xx-xx-xx-xx , xA-xx-xx-xx-xx-xx , xE-xx-xx-xx-xx-xx
	buf[0] = (buf[0] | 2) & 0xfe

	hardAddr := net.HardwareAddr(buf)
	return hardAddr.String(), nil
}
