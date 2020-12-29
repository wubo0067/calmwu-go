// stacker
package main

// 坑爹，不用就会编译报错，要死啊！
import (
	"errors"
	"fmt"
)

type Stack []interface{}

//Stack 的方法
func (stack Stack) Len() int {
	return len(stack)
}

// 该方法没有返回值
func (stack *Stack) Push(x interface{}) {
	*stack = append(*stack, x)
}

func (stack Stack) Top() (interface{}, error) {
	if len(stack) == 0 {
		return nil, errors.New("Can't Top en empty stack")
	}
	return stack[len(stack)-1], nil
}

func (stack *Stack) Pop() (interface{}, error) {
	var stack_len int = len(*stack)
	if stack_len == 0 {
		return nil, errors.New("Can't Pop en empty stack")
	}
	// 得到最后一个元素
	//temp_stack := *stack
	x := (*stack)[stack_len-1]
	*stack = (*stack)[:stack_len-1]
	return x, nil
}

func main() {
	var m_stack Stack
	m_stack.Push(1)
	m_stack.Push("Hi")
	m_stack.Push([]string{"ping", "ls", "mpstat", "top"})
	m_stack.Push(3.14)

	fmt.Println("m_stack size =", m_stack.Len())
	for {
		item, err := m_stack.Pop()
		if err != nil {
			break
		}
		fmt.Println(item)
	}
}
