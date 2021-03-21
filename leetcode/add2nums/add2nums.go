/*
 * @Author: CALM.WU
 * @Date: 2021-03-20 21:17:10
 * @Last Modified by: CALM.WU
 * @Last Modified time: 2021-03-20 23:22:52
 */

package main

import "fmt"

// ListNode single list node
type ListNode struct {
	Val  int
	Next *ListNode
}

func addTwoNumbers(l1 *ListNode, l2 *ListNode) *ListNode {
	res := &ListNode{}
	tailNode := &res.Next
	val := 0
	isNumCarry := false
	var partNode *ListNode = nil

	for l1 != nil && l2 != nil {
		val = l1.Val + l2.Val

		if isNumCarry {
			val++
		}

		*tailNode = &ListNode{}

		if val >= 10 {
			isNumCarry = true
			(*tailNode).Val = val - 10
		} else {
			isNumCarry = false
			(*tailNode).Val = val
		}

		// 上级指针的地址，下次给这个赋值就是修改了上级node的next
		tailNode = &(*tailNode).Next

		l1 = l1.Next
		l2 = l2.Next

		if l1 != nil && l2 != nil {
			continue
		} else if l1 == nil {
			partNode = l2
		} else {
			partNode = l1
		}

		for partNode != nil {
			*tailNode = &ListNode{}

			if isNumCarry {
				(*tailNode).Val = partNode.Val + 1
				if (*tailNode).Val >= 10 {
					(*tailNode).Val = (*tailNode).Val - 10
					isNumCarry = true
				} else {
					isNumCarry = false
				}
			} else {
				isNumCarry = false
				(*tailNode).Val = partNode.Val
			}

			partNode = partNode.Next
			tailNode = &(*tailNode).Next
		}
	}

	if isNumCarry {
		*tailNode = &ListNode{1, nil}
	}

	return res.Next
}

func printListNodes(l *ListNode) {
	i := 0

	for l != nil {
		fmt.Printf("res[%d] = %d\n", i, l.Val)
		l = l.Next
		i++
	}

	fmt.Println("------------------------------")
}

func main() {
	l1 := &ListNode{
		2,
		&ListNode{
			4,
			&ListNode{
				3,
				nil,
			},
		},
	}

	l2 := &ListNode{
		5,
		&ListNode{
			6,
			&ListNode{
				4,
				nil,
			},
		},
	}

	res := addTwoNumbers(l1, l2)
	printListNodes(res)

	l1 = &ListNode{
		0,
		nil,
	}

	l2 = &ListNode{
		0,
		nil,
	}

	res = addTwoNumbers(l1, l2)
	printListNodes(res)
}
