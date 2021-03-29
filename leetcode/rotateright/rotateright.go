/*
 * @Author: CALM.WU
 * @Date: 2021-03-27 17:20:07
 * @Last Modified by: CALM.WU
 * @Last Modified time: 2021-03-27 17:40:48
 */

package main

// ListNode single list node
type ListNode struct {
	Val  int
	Next *ListNode
}

func searchTailAndPrevNode(l *ListNode) (tailNode, tailPrevNode *ListNode) {
	tailNode = nil
	tailPrevNode = nil
	return
}

func rotateRight(head *ListNode, k int) *ListNode {
	if k == 0 {
		return head
	}

	headNode := head

	for k > 0 {
		k--
	}

	return headNode
}

func main() {

}
