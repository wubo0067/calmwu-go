/*
 * @Author: calmwu
 * @Date: 2017-10-08 19:57:03
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-10-08 20:03:40
 */

package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"context"
)

func Cdd(ctx context.Context) int {
	fmt.Println(ctx.Value("NLJB"))
	select {
	// 结束时候做点什么 ...
	case <-ctx.Done():
		return -3
	}
}

func Bdd(ctx context.Context) int {
	fmt.Println(ctx.Value("HELLO"))
	fmt.Println(ctx.Value("WROLD"))
	ctx = context.WithValue(ctx, "NLJB", "NULIJIABEI")
	go fmt.Println(Cdd(ctx))
	select {
	// 结束时候做点什么 ...
	case <-ctx.Done():
		return -2
	}
}

func Add(ctx context.Context) int {
	ctx = context.WithValue(ctx, "HELLO", "WROLD")
	ctx = context.WithValue(ctx, "WROLD", "HELLO")
	go fmt.Println(Bdd(ctx))
	select {
	// 结束时候做点什么 ...
	case <-ctx.Done():
		return -1
	}
}

func DeadlineTest(logger *log.Logger) {
	ctx := context.Background()

	_, ok := ctx.Deadline()
	if !ok {
		logger.Println("ctx not set deadline time")
		ctx1, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		_, ok = ctx1.Deadline()
		if ok {
			logger.Println("ctx1 set deadline time")
			<-ctx1.Done()
			logger.Printf("ctx1 timeout, %\n", ctx1.Err())
		}
	} else {
		logger.Println("ctx set deadline time")
	}

}

func main() {
	logger := log.New(os.Stdout, "", log.Ldate|log.Lmicroseconds|log.Lshortfile)
	// 自动取消(定时取消)

	DeadlineTest(logger)

	timeout := 3 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	logger.Println("start context")
	logger.Println(Add(ctx))

	// 手动取消
	//  {
	//      ctx, cancel := context.WithCancel(context.Background())
	//      go func() {
	//          time.Sleep(2 * time.Second)
	//          cancel() // 在调用处主动取消
	//      }()
	//      fmt.Println(Add(ctx))
	//  }
	logger.Println("Exit ", time.Now().String())

	ctxVal := context.WithValue(context.Background(), "floor-1", 1)
	ctxVal = context.WithValue(ctxVal, "floor-2", 2)
	ctxVal = context.WithValue(ctxVal, "floor-3", 3)

	fmt.Println(ctxVal.Value("floor-3").(int))

	parentCtx, parentCancel := context.WithCancel(context.Background())
	go func(ctx context.Context) {
		childCtx, _ := context.WithCancel(ctx)
		secTicker := time.NewTicker(time.Second)
		defer secTicker.Stop()
	L:
		for {
			select {
			case <-childCtx.Done():
				logger.Println("child Context Done")
				break L
			case <-secTicker.C:
				logger.Println("child Context second ticker")
			}
		}
	}(parentCtx)

	time.Sleep(5 * time.Second)
	parentCancel()
	time.Sleep(2 * time.Second)
	fmt.Println("context test completed")
}
