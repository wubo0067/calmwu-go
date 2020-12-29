package main

import (
	"fmt"
	"time"
)

func create_counter(start int) chan int {
	next := make(chan int)
	go func(i int) {
		for {
			curr_time := time.Now()
			fmt.Printf("[%s] +++PUT++ next-chan i:%d\n", curr_time.Format(time.UnixDate), i)
			next <- i
			i++
		}
	}(start)
	return next
}

func main() {
	chan_a := create_counter(1)
	chan_b := create_counter(100)

	fmt.Printf("chan_a cap:%d\n", cap(chan_a))

	for i := 0; i < 5; i++ {
		// 等待2秒后获取，因为这里等待，所以PUT会阻塞直到获取，如果chan有缓冲区，那么put会直接插入，直到chan满
		time.Sleep(2 * time.Second)
		a := <-chan_a
		b := <-chan_b
		curr_time := time.Now()
		fmt.Printf("[%s] ---GET--- chan_a -> %d, chan_b -> %d\n", curr_time.Format(time.UnixDate), a, b)
	}

	flag := false

	if flag == false {
		fmt.Printf("flag is false!!\n")
	}
}
