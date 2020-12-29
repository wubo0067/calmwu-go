package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	msg, err := ioutil.ReadFile("/var/log/message")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Read /var/log/message failed! reason[%s]", err.Error())
		return
	}

}
