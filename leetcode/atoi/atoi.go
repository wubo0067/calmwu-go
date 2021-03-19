/*
 * @Author: CALM.WU
 * @Date: 2021-03-18 14:55:03
 * @Last Modified by: CALM.WU
 * @Last Modified time: 2021-03-18 18:23:38
 */

package main

import (
	"fmt"
	"math"
	"unsafe"
)

const (
	MaxUint64 = ^uint64(0)
	// 无符号32位数上线限
	MaxUint32 = ^uint32(0)
	MinUint32 = 0
	// 有符号32位数上下限
	MaxInt32 = int(MaxUint32 >> 1)
	MinInt32 = -MaxInt32 - 1
)

type ByteType int

const (
	Start     ByteType = iota + 1 // 起始
	Def                           // 普通字符
	Num                           // 数字
	Minus                         // -
	Plus                          // +
	Space                         // 空格
	FristZero                     //
)

var (
	_minNumASCII = 48
	_maxNumASCII = 57

	_powNum = [10]uint64{
		1,
		10,
		100,
		1000,
		10000,
		100000,
		1000000,
		10000000,
		100000000,
		1000000000,
	}
)

func String2Bytes(s string) []byte {
	sh := (*[2]uintptr)(unsafe.Pointer(&s))
	// bh := reflect.SliceHeader{
	// 	Data: sh.Data,
	// 	Len:  sh.Len,
	// 	Cap:  sh.Len,
	// }
	bh := [3]uintptr{sh[0], sh[1], sh[1]}
	return *(*[]byte)(unsafe.Pointer(&bh))
}

func myAtoi(s string) int {
	if len(s) == 0 {
		return 0
	}

	var preByteType ByteType = Start // 扫描的字符类型
	isNegative := false              //
	nums := make([]int, 0, 200)

	bs := String2Bytes(s)
	//bs = bytes.TrimSpace(bs)
	//bs = bytes.TrimPrefix(bs, []byte{'0'})

	for _, b := range bs {
		bNum := int(b)

		if preByteType == Start || preByteType == Space {
			if b == ' ' {
				preByteType = Space
			} else if b == '+' {
				preByteType = Plus
			} else if b == '-' {
				isNegative = true
				preByteType = Minus
			} else if bNum >= _minNumASCII && bNum <= _maxNumASCII {
				if b == '0' {
					//fmt.Printf("----------FirstZero\n")
					preByteType = FristZero
				} else {
					//fmt.Printf("----------add FirstZero bNum = %d\n", bNum)
					preByteType = Num
					nums = append(nums, bNum-_minNumASCII)
				}
			} else {
				preByteType = Def
				return 0
			}
		} else if preByteType == Plus || preByteType == Minus {
			if b == '+' || b == '-' {
				return 0
			} else if bNum >= _minNumASCII && bNum <= _maxNumASCII {
				preByteType = Num
				nums = append(nums, bNum-_minNumASCII)
			} else {
				preByteType = Def
				return 0
			}
		} else if preByteType == Num {
			if bNum >= _minNumASCII && bNum <= _maxNumASCII {
				preByteType = Num
				nums = append(nums, bNum-_minNumASCII)
			} else {
				break
			}
		} else if preByteType == FristZero {
			if b == '+' || b == '-' {
				return 0
			} else if bNum >= _minNumASCII && bNum <= _maxNumASCII {
				if b == '0' {
					//fmt.Printf("----------FirstZero\n")
					preByteType = FristZero
				} else {
					preByteType = Num
					nums = append(nums, bNum-_minNumASCII)
				}
			} else {
				preByteType = Def
				return 0
			}
		}
	}

	//fmt.Printf("nums = %v\n", nums)

	var result, temp uint64
	numWide := len(nums) - 1
	for _, n := range nums {
		temp = uint64(n)

		if MaxUint64-result < temp {
			result = uint64(MaxInt32 + 1)
			break
		}
		result = result + temp*uint64((math.Pow10(numWide)))
		//result = result + temp*_powNum[numWide]
		numWide--
	}

	//fmt.Printf("result = %d\n", result)

	if result > uint64(MaxInt32) {
		if isNegative {
			return MinInt32
		}
		return MaxInt32
	}

	if isNegative {
		return 0 - int(result)
	}

	return int(result)
}

func main() {
	fmt.Printf("MaxInt32 = %d MinInt32 = %d MaxUint64=%d\n", MaxInt32, MinInt32, uint64(math.MaxUint64))

	var bigNum uint64 = 18446744073709551610
	bigNum += uint64(7 * int(math.Pow10(0)))
	fmt.Printf("bigNum ===> {%d}\n\n", bigNum)

	content := "  -42"
	fmt.Printf("{%s} ===> {%d}\n\n", content, myAtoi(content))

	content = "4193 with words"
	fmt.Printf("{%s} ===> {%d}\n\n", content, myAtoi(content))

	content = "words and 987"
	fmt.Printf("{%s} ===> {%d}\n\n", content, myAtoi(content))

	content = "-91283472332"
	fmt.Printf("{%s} ===> {%d}\n\n", content, myAtoi(content))

	content = "+-12"
	fmt.Printf("{%s} ===> {%d}\n\n", content, myAtoi(content))

	content = "00000-42a1234"
	fmt.Printf("{%s} ===> {%d}\n\n", content, myAtoi(content))

	content = "  0000000000012345678"
	fmt.Printf("{%s} ===> {%d}\n\n", content, myAtoi(content))

	content = "18446744073709551617"
	fmt.Printf("{%s} ===> {%d}\n\n", content, myAtoi(content))
}
