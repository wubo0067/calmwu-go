package main

import (
	"fmt"
	"time"
)

//import "sync"

func fibonacci(c, quit chan int) {
	x, y := 1, 1
	for {
		select {
		// 判断x通道是否可写
		case c <- x:
			x, y = y, x+y
		case <-quit:
			fmt.Println("fibonacci quit!")
			return
		case <-time.After(time.Second):
			fmt.Println("time out 1 second")
		default:
			// 会循环一直执行
			//fmt.Println("fibonacci block!")
		}
	}
}

func run_fibonacci() {
	c := make(chan int)
	quit := make(chan int)

	go func() {
		for i := 0; i < 10; i++ {
			// 读取
			fmt.Println(<-c)
		}
		time.Sleep(5 * time.Second)
		// 读取10个后，写入退出标志
		quit <- 0
	}()

	fibonacci(c, quit)
}

func sink(ch <-chan int) {
	for {
		i := <-ch
		fmt.Println("sink i", i)
	}
}

func source(ch chan<- int) {
	var i int = 0
	for {
		ch <- i
		i++
	}
}

// 测试自产自销
func run_fullduplex() {
	c := make(chan int)

	go source(c)
	go sink(c)
	return
}

func runBlock() chan<- int {
	zeroChan := make(chan int)

	// time.After不要放到select里面，否则计时是从select选择时候开始的，你怎么都不可能达到超时时间，要放到外面
	after := time.After(time.Second)

	go func() {
	L:
		for {
			select {
			case num := <-zeroChan:
				fmt.Println(num)
				//
			case <-after:
				fmt.Println("timeout")
				break L
			default:
				time.Sleep(time.Millisecond * 200)
			}
			fmt.Println("runLoop")
		}
	}()

	return zeroChan
}

func main() {
	//    run_fibonacci()

	//run_fullduplex()
	zeroChan := runBlock()
	var i int
	for i < 10 {
		zeroChan <- i
		i++
	}
	time.Sleep(3 * time.Second)
}
