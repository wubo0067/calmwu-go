/*
 * @Author: calmwu
 * @Date: 2018-01-30 15:28:57
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-01-30 16:39:51
 * @Comment:
 */

package utils

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

type Event struct {
	ReferenceCounter // 这是个referencecountable对象
	Name             string
	ID               uint32
}

// 对象创建方法
func NewEvent(rc ReferenceCounter) ReferenceCountable {
	fmt.Printf("New Event\n")
	event := new(Event)
	event.ReferenceCounter = rc
	return event
}

func ResetEvent(i interface{}) error {
	ev, ok := i.(*Event)
	if !ok {
		return fmt.Errorf("illegal object[%s] sent to ResetEvent", reflect.TypeOf(i).String())
	}
	fmt.Printf("Reset Event ID=%d\n", ev.ID)
	ev.Name = ""
	ev.ID = 0
	return nil
}

var eventPool = NewReferenceCountedPool(NewEvent, ResetEvent)

func AcquireEvent() *Event {
	return eventPool.Get().(*Event)
}

func TestEventPool(t *testing.T) {
	ev := AcquireEvent()
	fmt.Printf("%+v\n", eventPool.Stats())

	ev.DecrementReferenceCount()
	fmt.Printf("%+v\n", eventPool.Stats())

	ev = AcquireEvent()
	fmt.Printf("%+v\n", eventPool.Stats())

	ev.DecrementReferenceCount()
	fmt.Printf("%+v\n", eventPool.Stats())
}

func TestRoutineEventPool(t *testing.T) {
	eventChan := make(chan *Event, 100)

	go func() {
		for event := range eventChan {
			fmt.Printf("%+v\n", event)
			defer event.DecrementReferenceCount()
			//event.DecrementReferenceCount()
		}
		fmt.Println("event process routine exit!")
	}()

	i := 0
	for i < 10 {
		i++
		ev := AcquireEvent()
		defer ev.DecrementReferenceCount()
		ev.Name = fmt.Sprintf("Name_%d", i)
		ev.ID = uint32(i)
		// 在放入routine的时候计数递增
		ev.IncrementReferenceCount()
		eventChan <- ev
		//ev.DecrementReferenceCount()
	}

	//time.Sleep(time.Second)
	close(eventChan)
	time.Sleep(time.Second)
	fmt.Printf("%+v\n", eventPool.Stats())
}
