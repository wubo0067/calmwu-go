/*
 * @Author: calmwu
 * @Date: 2019-12-01 09:59:11
 * @Last Modified by: calmwu
 * @Last Modified time: 2019-12-01 10:04:47
 */

package main

import (
	"fmt"
	"sync"
)

// 对这个代码的深入理解：https://imil.net/blog/2019/11/27/Understanding-golang-channel-range-again/

// 这段代码的问题在哪，因为range没有退出，生产者没有close
func rangeChannel() {
	var wg sync.WaitGroup

	c := make(chan string)
	for _, t := range []string{"a", "b", "c"} {
		wg.Add(1)
		go func(s string) {
			c <- s
			wg.Done()
		}(t)
	}

	// 这里需要加上close，加上了，运行程序还是没有输出，这是为什么呢，因为go函数都没有跑起来
	// 现在加上了waitgroup还是报错。因为这里不是缓冲的channel，而wait又阻碍了读取。
	// 所以只有让wait协程化，怎么协程化比较合适呢
	go func() {
		wg.Wait()
		close(c)
	}()

	for s := range c {
		fmt.Println(s)
	}
}

func main() {
	rangeChannel()
}
