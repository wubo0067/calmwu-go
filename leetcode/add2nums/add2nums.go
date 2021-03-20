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
	courseNode := res
	isNumCarry := false
	var tailNode *ListNode = nil
	val := 0

	for l1 != nil && l2 != nil {
		val = l1.Val + l2.Val

		if isNumCarry {
			val++
		}

		if val >= 10 {
			isNumCarry = true
			courseNode.Val = val - 10
		} else {
			isNumCarry = false
			courseNode.Val = val
		}

		tailNode = courseNode
		tailNode.Next = &ListNode{}
		courseNode = tailNode.Next

		l1 = l1.Next
		l2 = l2.Next
	}

	var partNode *ListNode

	if l1 != nil && l2 == nil {
		partNode = l1
	} else {
		partNode = l2
	}

	for partNode != nil {
		if isNumCarry {
			courseNode.Val = partNode.Val + 1
			if courseNode.Val >= 10 {
				courseNode.Val = courseNode.Val - 10
				isNumCarry = true
			} else {
				isNumCarry = false
			}
		} else {
			isNumCarry = false
			courseNode.Val = partNode.Val
		}

		partNode = partNode.Next
		tailNode = courseNode
		tailNode.Next = &ListNode{}
		courseNode = tailNode.Next
	}

	if isNumCarry {
		courseNode.Val = 1
		courseNode.Next = nil
	} else {
		tailNode.Next = nil
	}

	return res
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
