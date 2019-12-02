/*
 * @Author: calmwu
 * @Date: 2019-12-02 19:22:16
 * @Last Modified by: calmwu
 * @Last Modified time: 2019-12-02 19:40:05
 */

package utils

import (
	"log"
	"testing"
	"time"
)

func TestTimer(t *testing.T) {
	index := 10

	myTimer := NewTimer()

L:
	for {
		myTimer.Reset(5 * time.Second)
		select {
		case <-myTimer.Chan():
			myTimer.SetRead()
			log.Printf("ticker %s", time.Now().String())
			index--
			if index == 0 {
				break L
			}
		}
	}

	// 这里必然返回false
	ret := myTimer.Stop()
	log.Printf("Stop ret:%v\n", ret)
}
