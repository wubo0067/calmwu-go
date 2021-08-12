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
			case <-nt.Chan():
				log.Printf("time out")
			case <-stopCh:
				log.Printf("time loop exit")
				return
			}
		}
	}()

	log.Print("start timer")
	nt.Reset(3 * time.Second)
	time.Sleep(4 * time.Second)
	sb := nt.Stop()
	log.Printf("expired 3 secs Stop return %v\n", sb)

	log.Print("start timer")
	nt.Reset(5 * time.Second)
	time.Sleep(2 * time.Second)
	sb = nt.Stop()
	log.Printf("no expired Stop return :%v\n", sb)

	log.Print("start timer-1")
	nt.Reset(5 * time.Second)
	time.Sleep(2 * time.Second)
	log.Print("start timer-2")
	nt.Reset(5 * time.Second)
	time.Sleep(6 * time.Second)
	log.Printf("twice reset expire time 6 secs")

	close(stopCh)

	time.Sleep(1 * time.Second)
}
