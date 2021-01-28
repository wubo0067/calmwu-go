/*
 * @Author: calmwu
 * @Date: 2021-01-28 23:01:14
 * @Last Modified by: calmwu
 * @Last Modified time: 2021-01-28 23:14:20
 */

package main

import (
	"errors"
	"fmt"

	werrors "github.com/pkg/errors"
)

var (
	//
	err1 = errors.New("this is first error")
	//
	err2 = errors.New("this is second error")
)

func useErrorIs() {
	wraperr1 := func() error {
		we := werrors.Wrap(err1, "wrapper first level")
		return we
	}()

	wraperr2 := func() error {
		we := werrors.Wrap(wraperr1, "wrapper first level")
		return we
	}()

	// 判断是不是同一个错误
	// true
	fmt.Printf("wraperr_2 is err1 = %v\n", errors.Is(wraperr2, err1))
	// false
	fmt.Printf("wraperr_2 is err2 = %v\n", errors.Is(wraperr2, err2))
}

type CalmErrorString struct {
	s string
}

func (e *CalmErrorString) Error() string {
	return e.s
}

func useErrorAs() {
	wraperr1 := func() error {
		we := werrors.Wrap(&CalmErrorString{
			s: "err is CalmErrorString",
		}, "wrapper first level")
		return we
	}()

	wraperr2 := func() error {
		we := werrors.Wrap(wraperr1, "wrapper first level")
		return we
	}()

	// 判断错误类型是否一直，这里切记As第二个参数是指针的指针
	var calmErrStr *CalmErrorString
	fmt.Printf("wraperr_2 type is CalmErrorString = %v\n", errors.As(wraperr2, &calmErrStr))
}

func main() {
	useErrorIs()

	useErrorAs()
}
