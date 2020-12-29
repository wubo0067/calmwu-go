package main

import (
	"fmt"
	//corev1 "k8s.io/api/core/v1"
	// "math"
)

func add_num(num_slice *[]int) {
	// 改变切片也需要传入指针，这是修改切片对象
	for i := 0; i < 10; i++ {
		*num_slice = append(*num_slice, i)
	}
}

func print_nums(num_slice []int) {
	fmt.Println("num_slice:", num_slice)
}

func change_nums(num_slice []int) {
	// 这是修改切片内容
	for i, elem := range num_slice {
		num_slice[i] = elem * 10
	}
}

func reverse(data []byte) {
	// 这里做个反转
	half_length := len(data) / 2
	length := len(data)
	for i := 0; i < half_length; i++ {
		j := length - i - 1
		data[i], data[j] = data[j], data[i]
	}
}

func main() {
	data := []byte("1234567890")
	fmt.Printf("data[%s]\n", data)
	reverse(data)
	fmt.Printf("reverse data[%s]\n", data)

	data_1 := data[:3]
	data_1[0] = 'z'
	fmt.Printf("data[%s]\n", data)
	fmt.Printf("data_1[%s]\n", data_1)

	data_2 := make([]byte, len(data))
	copy(data_2, data)
	data_2[0] = 'j'
	fmt.Printf("data[%s]\n", data)
	fmt.Printf("data_2[%s]\n", data_2)

	num_array := [...]int{1, 2}

	var temp_array interface{} = num_array
	_, ok := temp_array.(int)
	if !ok {
		fmt.Println("num_array type is not int")
	}

	_, ok = temp_array.([0]int)
	if ok {
		fmt.Println("num_array int array")
	}

	// 数组必须长度完全一致，类型才能匹配，第一个返回的是值
	resultoftype, ok := temp_array.([2]int)
	if ok {
		fmt.Println("num_array int[2] array, resultoftype:", resultoftype)
	}

	fmt.Println("len(num_array)", len(num_array))
	fmt.Println("cap(num_array)", cap(num_array))

	//num_slice := num_array[:]
	var num_slice []int
	fmt.Printf("num_slice is nil=%v length:%d\n", num_slice == nil, len(num_slice))
	// 新增一批数据
	add_num(&num_slice)
	// 因为空间不够，这里会返回新的切片
	num_slice_1 := append(num_slice, 999)
	fmt.Println("len(num_slice) ", len(num_slice))
	fmt.Println("len(num_slice_1)", len(num_slice_1))
	print_nums(num_slice)
	print_nums(num_slice_1)

	num_slice_2 := make([]int, 5, 10)
	num_slice_2 = append(num_slice_2, 1)
	num_slice_2[1] = 12
	print_nums(num_slice_2)
	change_nums(num_slice_2)
	print_nums(num_slice_2)

	// 这里前8个元素都是0
	//var a1 []int = make([]int, 8)
	var a1 []int
	//a1 = append(a1, 1, 2, 3, 4, 5, 6, 7, 8)
	a1 = append(a1, 3)
	fmt.Printf("a1 len:[%d] %v\n", len(a1), a1)

	var index int = 0
	for index, _ = range a1 {
		if a1[index] == 3 {
			break
		}
	}

	fmt.Printf("index:[%d]\n", index)

	// 删除这个元素
	a1 = append(a1[:index], a1[index+1:]...)
	fmt.Printf("a1 len:[%d] %v\n", len(a1), a1)

	// 测试append
	fmt.Println("---------test append")
	buf := make([]byte, 0, 100)
	msg := []byte{'1', '2', '3'}
	// ...把数组打散
	buf = append(buf, msg...)
	fmt.Println(string(buf))

	s := make([]byte, 16)
	fmt.Printf("s len:%d cap:%d\n", len(s), cap(s))
	copy(s, []byte{'a', 'b'})
	fmt.Printf("s:%v len:%d cap:%d\n", s, len(s), cap(s))
	s = s[2:8]
	fmt.Printf("s len:%d cap:%d\n", len(s), cap(s))
	s = s[0:14]
	fmt.Printf("s:%v len:%d cap:%d\n", s, len(s), cap(s))
}
