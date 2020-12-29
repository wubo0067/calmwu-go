/*
 * @Author: calmwu
 * @Date: 2019-12-29 17:09:57
 * @Last Modified by: calmwu
 * @Last Modified time: 2019-12-29 17:13:56
 */

//https://www.golangprograms.com/example-of-interface-with-type-embedding-and-method-overriding-in-go-language.html
//https://travix.io/type-embedding-in-go-ba40dd4264df
//https://hackthology.com/object-oriented-inheritance-in-go.html
//https://medium.com/random-go-tips/method-overriding-680cfd51ce40

package main

import (
	"fmt"

	"github.com/micro/go-micro/v2"
)

type Bouncer interface {
	Bounce()
}

type Ball struct {
	Radius   int
	Material string
}

func (b Ball) Bounce() {
	fmt.Printf("Bouncing ball %+v\n", b)
}

type Football struct {
	Bouncer
}

// football可以自己实现接口方法，这样屏蔽了成员实现的方法
func (fb Football) Bounce() {
	fmt.Printf("Bouncing football %+v\n", fb)
	fb.Bouncer.Bounce()
}

func main() {
	var b Bouncer = &Football{Ball{Radius: 10, Material: "leather"}}
	b.Bounce()

	service := micro.NewService()
	service.Init()

	testWrapCall()
}
