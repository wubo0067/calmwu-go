/*
 * @Author: calmwu
 * @Date: 2019-02-23 20:25:17
 * @Last Modified by: calmwu
 * @Last Modified time: 2019-02-23 20:25:56
 */

package main

import "fmt"

func JumpHash(key uint64, buckets int) int {
	var b, j int64
	if buckets <= 0 {
		buckets = 1
	}
	for j < int64(buckets) {
		b = j
		key = key*2862933555777941757 + 1
		j = int64(float64(b+1) * (float64(int64(1)<<31) / float64((key>>33)+1)))
	}
	return int(b)
}

func MainJumpHash() {
	buckets := make(map[int]int, 10)
	count := 10
	for i := uint64(0); i < 120000; i++ {
		b := JumpHash(i, count)
		buckets[b] = buckets[b] + 1
	}
	fmt.Printf("buckets: %v\n", buckets)
	//add two buckets
	count = 12
	for i := uint64(0); i < 120000; i++ {
		oldBucket := JumpHash(i, count-2)
		newBucket := JumpHash(i, count)
		//如果对象需要移动到新的bucket中,则首先从原来的bucket删除，再移动
		if oldBucket != newBucket {
			buckets[oldBucket] = buckets[oldBucket] - 1
			buckets[newBucket] = buckets[newBucket] + 1
		}
	}
	fmt.Printf("buckets after add two servers: %v\n", buckets)
}

func repeatStrings(s string, count int) string {
	b := make([]byte, len(s)*count)
	bp := copy(b, s)
	fmt.Printf("bp:%d\n", bp)
	for bp < len(b) {
		x := copy(b[bp:], b[:bp])
		fmt.Printf("+++bp:%d len(b):%d x:%d\n", bp, len(b), x)
		bp *= 2
		fmt.Printf("---bp:%d\n", bp)
	}
	return string(b)
}
