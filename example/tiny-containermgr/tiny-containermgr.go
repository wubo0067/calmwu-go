/*
 * @Author: CALM.WU
 * @Date: 2021-03-31 15:27:40
 * @Last Modified by: CALM.WU
 * @Last Modified time: 2021-03-31 16:40:23
 */

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	// We want to have a full control on the container's
	// stdout, so we are creating a pipe to redirect it.
	rd, wr, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	defer rd.Close()
	defer wr.Close()

	// Start runc in detached mode.
	fmt.Println("Launching runc")

	cmd := exec.Command("runc", "run", "--detach", os.Args[1])
	cmd.Stdin = nil  // i.e. /dev/null
	cmd.Stderr = nil // i.e. /dev/null
	// 容器的标准输出是管道，如果这个关闭了，容器就会收到sigpipe消息，然后退出
	cmd.Stdout = wr
	if err := cmd.Run(); err != nil {
		panic(err)
	}

	// Read some data from the container's stdout.
	buf := make([]byte, 1024)
	for i := 0; i < 10; i++ {
		n, err := rd.Read(buf)
		if err != nil {
			panic(err)
		}
		output := strings.TrimSuffix(string(buf[:n]), "\n")
		fmt.Printf("Container produced: [%s]\n", output)

	}

	// Get bored quickly, give up and exit.
	fmt.Println("We are done, exiting...")
}
