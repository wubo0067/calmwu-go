/*
 * @Author: calmwu
 * @Date: 2019-01-28 16:19:32
 * @Last Modified by: calmwu
 * @Last Modified time: 2019-01-28 17:58:59
 */

// go test bench=. -benchmem -benchtime=3s

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/intel-go/bytebuf"
)

func ForSlice(s []string) {
	len := len(s)
	for i := 0; i < len; i++ {
		_, _ = i, s[i]
	}
}

func RangeForSlice(s []string) {
	for i, v := range s {
		_, _ = i, v
	}
}

const N = 1000

func initSlice() []string {
	s := make([]string, N)
	for i := 0; i < N; i++ {
		s[i] = "www.flysnow.org"
	}
	return s
}

func BenchmarkForSlice(b *testing.B) {
	s := initSlice()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ForSlice(s)
	}
}

func BenchmarkRangeForSlice(b *testing.B) {
	s := initSlice()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RangeForSlice(s)
	}
}

func RangeForMap1(m map[int]string) {
	for k, v := range m {
		_, _ = k, v
	}
}

func initMap() map[int]string {
	m := make(map[int]string, N)
	for i := 0; i < N; i++ {
		m[i] = fmt.Sprint("www.flysnow.org", i)
	}
	return m
}

func BenchmarkRangeForMap1(b *testing.B) {
	m := initMap()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RangeForMap1(m)
	}
}

func BenchmarkIntelByteBuf(b *testing.B) {
	b.ResetTimer()

	bbuf := bytebuf.NewPointer()
	base64writer := base64.NewEncoder(base64.StdEncoding, bbuf)
	for i := 0; i < b.N; i++ {

		je := json.NewEncoder(base64writer)
		je.Encode(map[string]float64{"foo": 4.12, "pi": 3.14159})

		// Flush any partially encoded blocks left in the base64 encoder.
		base64writer.Close()
	}
}

func BenchmarkStdByteBuf(b *testing.B) {
	b.ResetTimer()

	bbuf := new(bytes.Buffer)
	base64writer := base64.NewEncoder(base64.StdEncoding, bbuf)
	for i := 0; i < b.N; i++ {

		je := json.NewEncoder(base64writer)
		je.Encode(map[string]float64{"foo": 4.12, "pi": 3.14159})

		// Flush any partially encoded blocks left in the base64 encoder.
		base64writer.Close()
	}
}
