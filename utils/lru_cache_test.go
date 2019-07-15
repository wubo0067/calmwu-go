/*
 * @Author: calm.wu
 * @Date: 2019-07-15 10:59:05
 * @Last Modified by: calm.wu
 * @Last Modified time: 2019-07-15 11:08:12
 */

package utils

import (
	"fmt"
	"testing"
)

func TestLRUCache(t *testing.T) {
	lruCache, err := NewLRUCache(10)
	if err != nil {
		t.Fatal(err.Error())
	}

	lruCache.Set("0", "a")
	lruCache.Set("1", "b")
	lruCache.Set("2", "c")
	lruCache.Set("3", "d")
	fmt.Println("Cache:", lruCache)

	fmt.Println(lruCache.Get("0"))

	lruCache.Set("0", "calmwu")
	fmt.Println(lruCache.Get("0"))

	lruCache.Clear()
	fmt.Println("Cache:", lruCache)
}
