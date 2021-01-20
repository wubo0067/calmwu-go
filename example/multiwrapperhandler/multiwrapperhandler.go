/*
 * @Author: calm.wu
 * @Date: 2019-07-10 11:32:02
 * @Last Modified by: calmwu
 * @Last Modified time: 2020-12-31 20:50:41
 */

package main

import (
	"context"
	"fmt"
	"strings"
)

type Person struct {
	name    string
	age     int
	address string
}

func defaultFilterHandler(ctx context.Context, p *Person) bool {
	fmt.Printf("---defaultFilterHandler---\n")
	return true
}

type PersonFilterHandler func(ctx context.Context, p *Person) bool

type PersonFilterHandlerWrapper func(PersonFilterHandler) PersonFilterHandler

func filterPersonName(ctx context.Context, p *Person) bool {
	fmt.Printf("----filterPersonName---\n")
	if strings.Contains(p.name, "wu") {
		return true
	}
	return false
}

func filterPersonAge(ctx context.Context, p *Person) bool {
	fmt.Printf("----filterPersonAge---\n")
	if p.age > 18 {
		return true
	}
	return false
}

func filterPersonAddress(ctx context.Context, p *Person) bool {
	fmt.Printf("----filterPersonAddress---\n")
	if p.address == "wuhan" {
		return true
	}
	return false
}

func testFilterPerson() {
	persons := []*Person{
		&Person{
			name:    "calmwu",
			age:     40,
			address: "wuhan",
		},
		&Person{
			name:    "jerrywu",
			age:     60,
			address: "chendu",
		},
		&Person{
			name:    "marry",
			age:     6,
			address: "chendu",
		},
	}

	filterHandlers := []PersonFilterHandler{
		filterPersonName, filterPersonAge, filterPersonAddress,
	}

	var res []*Person
	for _, person := range persons {
		isOk := true
		for _, handler := range filterHandlers {
			if !handler(context.TODO(), person) {
				isOk = false
				break
			}
		}
		if isOk {
			res = append(res, person)
		}
	}

	for _, person := range res {
		fmt.Printf("res:%v\n\n\n", person)
	}

	// wrapper1 := NewPersonFilterHandlerWrapper(1)
	// fn := wrapper1(filterPersonAddress)

	// wrapper2 := NewPersonFilterHandlerWrapper(2)
	// fn = wrapper2(fn)

	//fn(context.TODO(), res[0])

	// wrapper1 := NewPersonFilterHandlerWrapper(1, filterPersonAddress)

	// fn := wrapper1(defaultFilterHandler)
	// //b := fn(context.TODO(), persons[0])
	// //fmt.Printf("b: %v\n\n\n", b)

	// wrapper2 := NewPersonFilterHandlerWrapper(2, filterPersonAge)
	// fn = wrapper2(fn)
	// b := fn(context.TODO(), persons[1])
	// fmt.Printf("b: %v\n\n", b)

	// b = fn(context.TODO(), persons[0])
	// fmt.Printf("b: %v\n", b)

	fmt.Printf("--------------\n")
	fn := MakePersonFilterHandler(filterPersonName, filterPersonAddress, filterPersonAge)
	b := fn(context.TODO(), persons[0])
	fmt.Printf("b: %v\n\n", b)

	b = fn(context.TODO(), persons[1])
	fmt.Printf("b: %v\n\n", b)

	b = fn(context.TODO(), persons[2])
	fmt.Printf("b: %v\n\n", b)
}

func NewPersonFilterHandlerWrapper(i int, outHandler PersonFilterHandler) PersonFilterHandlerWrapper {
	return func(inHandler PersonFilterHandler) PersonFilterHandler {
		return func(ctx context.Context, p *Person) bool {
			fmt.Printf("level-%d\n", i)
			if outHandler(ctx, p) {
				if inHandler != nil {
					return inHandler(ctx, p)
				} else {
					return true
				}
			}
			return false
		}
	}
}

func MakePersonFilterHandler(handlers ...PersonFilterHandler) PersonFilterHandler {
	handlerWrappers := make([]PersonFilterHandlerWrapper, len(handlers))

	for i, handler := range handlers {
		handlerWrappers[i] = NewPersonFilterHandlerWrapper(i, handler)
	}

	var handler PersonFilterHandler

	size := len(handlers)
	for i := size - 1; i >= 0; i-- {
		handler = handlerWrappers[i](handler)
	}

	return handler
}

type ProcessHandler func(ctx context.Context) error

type ProcessHandlerWrapper func(ProcessHandler) ProcessHandler

func SayHello(ctx context.Context) error {
	levelNum := ctx.Value("levelNum").(int)
	var i int
	for i < levelNum {
		key := fmt.Sprintf("level-%d", i)
		val := ctx.Value(key).(int)
		fmt.Printf("%s:%d\n", key, val)
		i++
	}
	return nil
}

func NewWrapperSayHello(levelNum int) ProcessHandlerWrapper {
	// 传入的函数对象会被返回的函数对象调用
	return func(ph ProcessHandler) ProcessHandler {
		return func(ctx context.Context) error {
			key := fmt.Sprintf("level-%d", levelNum)
			ctx = context.WithValue(ctx, key, levelNum)
			fmt.Printf("wrapper %s:%d\n", key, levelNum)
			return ph(ctx)
		}
	}
}

func main() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "levelNum", 3)

	var wrapperHandlers []ProcessHandlerWrapper
	for i := 0; i < 3; i++ {
		wrapperHandlers = append(wrapperHandlers, NewWrapperSayHello(i))
	}

	fn := SayHello
	// wrapper wrapperhandler 最后的最后执行
	for i := 3; i > 0; i-- {
		fn = wrapperHandlers[i-1](fn)
	}

	// 然后执行wrapper
	fn(ctx)

	testFilterPerson()

	testCallContext()
}
