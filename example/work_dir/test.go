/*
 * @Author: calmwu
 * @Date: 2018-12-26 19:20:01
 * @Last Modified by: calm.wu
 * @Last Modified time: 2019-03-15 17:36:23
 */

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"unsafe"

	"github.com/pquerna/ffjson/ffjson"

	"github.com/mitchellh/mapstructure"
)

type Basic struct {
	Vstring     string      `json:"Vstring"`
	Vint        int         `json:"Vint"`
	Vuint       uint        `json:"Vuint"`
	Vbool       bool        `json:"Vbool"`
	Vfloat      float64     `json:"Vfloat"`
	Vextra      string      `json:"Vextra"`
	Vsilent     bool        `json:"vsilent"`
	Vdata       interface{} `json:"Vdata"`
	VjsonInt    int         `json:"VjsonInt"`
	VjsonFloat  float64     `json:"VjsonFloat"`
	VjsonNumber json.Number `json:"VjsonNumber"`
}

func TestBasicTypeDecode() {
	input := map[string]interface{}{
		"Vstring":     "foo",
		"Vint":        42,
		"Vuint":       42,
		"Vbool":       true,
		"Vfloat":      42.42,
		"Vsilent":     true,
		"Vdata":       42,
		"VjsonInt":    json.Number("1234"),
		"VjsonFloat":  json.Number("1234.5"),
		"VjsonNumber": json.Number("1234.5"),
	}

	for k, v := range input {
		fmt.Printf("k:%s, v:%v\n", k, v)

		if k == "Vstring" {
			delete(input, k)
		}
	}

	var result Basic
	err := mapstructure.Decode(input, &result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "got an err: %s", err.Error())
	} else {
		fmt.Printf("result:%+v\n", result)
	}
}

func unsafePointerTest() {
	i := 10

	var fp *float64 = (*float64)(unsafe.Pointer(&i))
	*fp = *fp * 2.399

	fmt.Println("unsafePointer i:", i)

	basic := new(Basic)
	pVstring := (*string)(unsafe.Pointer(basic))
	*pVstring = "Hello"

	pVjsonFloat := (*float64)(unsafe.Pointer(uintptr(unsafe.Pointer(basic)) + unsafe.Offsetof(basic.VjsonFloat)))
	*pVjsonFloat = 9.323
	fmt.Printf("basic:%+v\n", basic)
}

func ffjsonTest() {
	b := new(Basic)
	b.Vbool = false
	b.Vstring = "Hello"
	b.VjsonNumber = "123456"
	serialData, _ := json.Marshal(b)

	decode := ffjson.NewDecoder()

	b2 := new(Basic)
	decode.Decode(serialData, b2)

	fmt.Printf("b2:%+v\n", b2)
}

func testSlice() {
	s := []int{1, 2, 3, 4, 5}

	// range的s是s的副本，所以其没有变化，这里会循环5次
	for i, n := range s {
		if i == 0 {
			s = s[:3]
			s[2] = n + 100
		}
		fmt.Println(i, n)
	}
	fmt.Println(s)
}

func deferTest() {
	x, y := 10, 100

	defer func(i int) {
		fmt.Printf("x, y := %d, %d\n", i, y)
	}(x)

	defer func(i *int) {
		fmt.Printf("x, y := %d, %d\n", *i, y)
	}(&x)

	x += 10
	y += 100

	fmt.Printf("x, y := %d, %d\n", x, y)
}

func interfaceNil(t interface{}) {
	// 对于interface的转换结果也要判断是否为nil
	v1, ok := t.(*int)
	if ok {
		if v1 != nil {
			fmt.Printf("t int is %d\n", *v1)
		} else {
			fmt.Printf("t *int is nil\n")
		}
	}

	switch v := t.(type) {
	case *int:
		if v != nil {
			fmt.Printf("t value:%d\n", *v)
		}
	default:
		fmt.Printf("t type is %s\n", reflect.TypeOf(t).String())
	}
}

func arrayTest() {
	nums := [5]int{}

	// 数组是拷贝
	func(a [5]int) {
		for i := 0; i < len(a); i++ {
			a[i] = i
		}
	}(nums)

	fmt.Printf("nums:%+v\n", nums)

	// slice传递的是指针
	func(b []int) {
		for i := 0; i < len(b); i++ {
			b[i] = i
		}
	}(nums[2:])

	fmt.Printf("nums:%+v\n", nums)
}

func jsonMarshal() {
	type Account struct {
		Name     string
		Password string `json:"-"`
		Balance  float64
	}
	joe := Account{Name: "Joe", Password: "123456", Balance: 102.4}
	s, _ := json.Marshal(joe)
	fmt.Println(string(s))
}

func jsonStreamDecode() {
	const s = `
	[
	  {"almonds": false},
	  {"cashews": true},
	  {"walnuts": false}
	]`
	dec := json.NewDecoder(strings.NewReader(s))

	t, err := dec.Token()
	if err != nil {
		panic(err)
	}
	if t != json.Delim('[') {
		panic("Expected '[' delimiter")
	}

	for dec.More() {
		var m map[string]bool
		err := dec.Decode(&m)
		if err != nil {
			panic(err)
		}

		fmt.Println("decoded", m)
	}

	t, err = dec.Token()
	if err != nil {
		panic(err)
	}
	if t != json.Delim(']') {
		panic("Expected ']' delimiter")
	}
}

func jsonEncodingStream() {
	for i := 0; i < 10; i++ {
		defer func(n int) {
			fmt.Println(n)
		}(i)
	}

	var outBuffer bytes.Buffer
	base64writer := base64.NewEncoder(base64.StdEncoding, &outBuffer)

	// Create a new JSON encoder, hooking up its output to base64writer.
	je := json.NewEncoder(base64writer)
	je.Encode(map[string]float64{"foo": 4.12, "pi": 3.14159})

	// Flush any partially encoded blocks left in the base64 encoder.
	base64writer.Close()

	fmt.Println(outBuffer.String())

	var m map[string]float64
	base64reader := base64.NewDecoder(base64.StdEncoding, &outBuffer)
	dec := json.NewDecoder(base64reader)

	// t, err := dec.Token()
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(t)

	dec.Decode(&m)
	fmt.Println(m)

}

func maybeUnlimitedLoops() {
	v := []int{1, 2, 3}
	for i := range v {
		v = append(v, i)
	}
	fmt.Printf("%v\n", v)
}

func main() {
	frameNums := make([]int, 0, 2)
	frameNums = append(frameNums, 1, 2)
	fmt.Printf("frameNums:%v len:%d\n", frameNums, len(frameNums))

	nums := []int{1, 2, 3, 4, 5}
	fmt.Printf("nums size:%d\n", len(nums))
	nums = nums[len(nums):]
	fmt.Printf("nums:%v size:%d\n", (nums == nil), len(nums))

	TestBasicTypeDecode()

	unsafePointerTest()

	ffjsonTest()

	testSlice()

	deferTest()

	arrayTest()

	var f1 *float64 = new(float64)
	interfaceNil(f1)
	var t1 *int
	interfaceNil(t1)

	jsonMarshal()

	jsonStreamDecode()

	jsonEncodingStream()

	maybeUnlimitedLoops()

	MainJumpHash()

	repeatStrings("123", 3)

	//testSizeof()

	var i int = func() int {
		return 19
	}()
	fmt.Print("i:%d\n", i)
}
