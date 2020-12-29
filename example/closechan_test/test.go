package main

import (
	"crypto/rand"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"
)

type server struct {
	data_chan chan int
	exit      chan struct{}
	wg        sync.WaitGroup
}

func (s *server) start() {
	s.data_chan = make(chan int)
	s.exit = make(chan struct{})

	s.wg.Add(2)

	go s.start_sender()
	go s.start_recevier(1)
	go s.start_recevier(2)
}

func (s *server) stop() {
	// 一个关闭能保证所有的在该chan上的读取都能读到
	close(s.exit)
	s.wg.Wait()
	return
}

func (s *server) start_sender() {
	// 生成一个定时器，定时写入数据
	timer_tick := time.NewTicker(time.Second)
	count := 1
	for {
		select {
		case <-timer_tick.C:
			s.data_chan <- count
			count++
		case <-s.exit:
			fmt.Println("recevie close event!")
			s.wg.Done()
			return
		}
	}
}

func (s *server) start_recevier(id int) {
	for {
		select {
		case n := <-s.data_chan:
			fmt.Printf("recevier[%d] receive data[%d]\n", id, n)
		case <-s.exit:
			fmt.Printf("recevieer[%d] recevie close event!\n", id)
			s.wg.Done()
			return
		}
	}
}

func notifyExit(errChan chan<- error) {
	select {
	case <-time.After(5 * time.Second):
		// 这两个效果是一样的
		errChan <- nil
		close(errChan)
	}
}

func sortNotify() {

	values := make([]byte, 32*1024*1024)
	if _, err := rand.Read(values); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	done := make(chan struct{})
	go func() { // the sorting goroutine
		sort.Slice(values, func(i, j int) bool {
			return values[i] < values[j]
		})
		done <- struct{}{} // notify sorting is done
		close(done)
	}()

	// do some other things ...

	<-done // waiting here for notification
	fmt.Println(values[0], values[len(values)-1])

}

func testClose(numChan <-chan int) {

	for range numChan {
		fmt.Printf("close event received!")
		break
	}
}

func nullStructChannel() {
	ch := make(chan struct{})
	i := 0

	go func() {
		i = 1
		ch <- struct{}{}
	}()

	<-ch
	fmt.Println(i)
}

func testRangeClose() {
	numChannel := make(chan int, 10)

	for i := 0; i < 10; i++ {
		numChannel <- i
	}

	close(numChannel)

	for i := range numChannel {
		fmt.Println("---", i)
	}
}

func main() {
	// var server_i server
	// server_i.start()
	// // 这里其实应该等待信号
	// time.Sleep(10 * time.Second)
	// server_i.stop()

	// sortNotify()

	// errChan := make(chan error, 1)
	// go notifyExit(errChan)
	// err := <-errChan
	// fmt.Printf("err type:%T\n", err)

	nullStructChannel()

	testRangeClose()

	// numChan := make(chan int, 10)
	// go testClose(numChan)

	// numChan <- 1
	// numChan <- 2
	// numChan <- 3
	// numChan <- 4
	// close(numChan)
	// numChan <- 5

	time.Sleep(2 * time.Second)
}
