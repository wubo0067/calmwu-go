/*
 * @Author: calm.wu
 * @Date: 2020-09-16 10:33:12
 * @Last Modified by: calm.wu
 * @Last Modified time: 2020-09-16 12:06:23
 */

package main

import (
	"fmt"
	"os"
	"testing"
)

func indexArray() {
	a := [...]int{1, 2, 3, 4, 5, 6, 7, 8}

	for i := range a {
		a[3] = 100
		if i == 3 {
			fmt.Fprintf(os.Stdout, "indexArray, %d %d", i, a[i])
		}
	}
}

func indexValArray() {
	a := [...]int{1, 2, 3, 4, 5, 6, 7, 8}

	for i, v := range a {
		a[3] = 100
		if i == 3 {
			fmt.Fprintf(os.Stdout, "indexValArray, %d %d", i, v)
		}
	}
}

func TestIndexRange(t *testing.T) {
	indexArray()
}

func TestValRange(t *testing.T) {
	indexValArray()
}

func BenchmarkIndexRange(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		indexArray()
	}
}

func BenchmarkValRange(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		indexValArray()
	}
}
