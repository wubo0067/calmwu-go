/*
 * @Author: CALM.WU
 * @Date: 2021-03-27 17:20:07
 * @Last Modified by: CALM.WU
 * @Last Modified time: 2021-03-27 17:40:48
 */

package main

import "fmt"

// ListNode single list node
type ListNode struct {
	Val  int
	Next *ListNode
}

func listLength(l *ListNode) int {
	len := 1
	for l.Next != nil {
		len++
		l = l.Next
	}
	return len
}

// searchTailAndPrevNode return tailNode, tailPrevNode
func searchTailAndPrevNode(l *ListNode) (*ListNode, *ListNode) {
	tailPrevNode := l

	for {
		if tailPrevNode.Next.Next == nil {
			return tailPrevNode.Next, tailPrevNode
		}
		tailPrevNode = tailPrevNode.Next
	}
}

func rotateRight(head *ListNode, k int) *ListNode {
	if k == 0 || head == nil || head.Next == nil {
		return head
	}

	//headNode := head

	len := listLength(head)
	k = k % len
	//fmt.Printf("k: %d, len: %d\n", k, len)

	for k > 0 {
		tailNode, tailPrevNode := searchTailAndPrevNode(head)
		if tailPrevNode == nil {
			return tailNode
		}
		tailPrevNode.Next = nil
		tailNode.Next = head
		head = tailNode
		k--
	}

	return head
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

	res := rotateRight(l1, 1)
	printListNodes(res)

	l1 = &ListNode{
		2,
		&ListNode{
			4,
			&ListNode{
				3,
				nil,
			},
		},
	}

	res = rotateRight(l1, 4)
	printListNodes(res)
}
