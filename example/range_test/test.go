package main

import (
	"flag"
	"fmt"
	"time"
)

/*
https://tour.golang.org/concurrency/4
*/

var (
	VersionTag = ""
	BuildTime  = ""
)

type Foo struct {
	bar string
}

func test_range_chan(num_chan <-chan int) {
	var index int = 0
	// The loop for i := range c receives values from the channel repeatedly until it is closed.
	for num := range num_chan {
		fmt.Printf("[%d] --- [%d]\n", index, num)
		index++
		/*
		   if index == 10 {
		       fmt.Printf("%s\n", "goroutine test_range_chan break!")
		       break
		   }*/
	}
	fmt.Println("test_range_chan exit!")
}

func maybeForeverLoop() {
	v := []int{1, 2, 3}
	for i := range v {
		v = append(v, i)
		fmt.Printf("v len:%d\n", len(v))
	}
	fmt.Printf("%v\n", v)
}

func main() {
	version := flag.Bool("v", false, "version")
	flag.Parse()

	if *version {
		fmt.Println("Version Tag: " + VersionTag)
		fmt.Println("Build Time: " + BuildTime)
	}

	foo_lst := []Foo{
		{"AAA"},
		{"BBB"},
		{"CCC"},
	}

	foo2_lst := make([]*Foo, len(foo_lst))
	for i, value := range foo_lst {
		foo2_lst[i] = &value
	}
	fmt.Println(foo_lst[0], foo_lst[1], foo_lst[2])
	// 这里都是临时变量的值
	fmt.Println(foo2_lst[0], foo2_lst[1], foo2_lst[2])

	// 需要这样来赋值
	foo3_lst := make([]*Foo, len(foo_lst))
	for i, _ := range foo_lst {
		foo3_lst[i] = &foo_lst[i]
	}
	fmt.Println(foo3_lst[0], foo3_lst[1], foo3_lst[2])

	num_chan := make(chan int, 10)
	go test_range_chan(num_chan)

	for index := 0; index < 10; index++ {
		num_chan <- index
		time.Sleep(time.Second)
	}

	close(num_chan)
	// 关闭 range会退出
	fmt.Println("close num_chan")

	//time.Sleep(time.Second)
	maybeForeverLoop()

	fmt.Println("main exit")
}
