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

func TestTimerReset(t *testing.T) {
	nt := NewTimer()
	stopCh := make(chan struct{})

	go func() {
		for {
			select {
			case <-nt.C:
				log.Printf("time out")
			case <-stopCh:
				log.Printf("time loop exit")
				return
			}
		}
	}()

	nt.Reset(3 * time.Second)
	time.Sleep(4 * time.Second)
	sb := nt.Stop()
	nt.Logf("expired Stop return :v", sb)

	nt.Reset(5 * time.Second)
	time.Sleep(2 * time.Second)
	sb = nt.Stop()
	nt.Logf("no expired Stop return :v", sb)

	nt.Reset(5 * time.Second)
	time.Sleep(2 * time.Second)
	nt.Reset(3 * time.Second)
	time.Sleep(4 * time.Second)
	nt.Log("twice reset expire time 6 secs")

	close(stopCh)

	time.Sleep(1 * time.Second)
}
