/*
 * @Author: CALM.WU
 * @Date: 2021-06-22 10:38:23
 * @Last Modified by: CALM.WU
 * @Last Modified time: 2021-06-22 10:39:22
 */

package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"syscall"
)

// WritePidFile write current process id to file
func WritePidFile(path string) error {
	_, err := os.Stat(path)

	if err == nil { // file already exists
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return fmt.Errorf("Could not read %s: %v", path, err)
		}

		pid, err := strconv.Atoi(string(data))
		if err != nil {
			return fmt.Errorf("Could not parse pid file %s contents '%s': %v", path, string(data), err)
		}

		if process, err := os.FindProcess(pid); err == nil {
			if err := process.Signal(syscall.Signal(0)); err == nil {
				return fmt.Errorf("process with pid %d is still running", pid)
			}
		}
	}

	return ioutil.WriteFile(path, []byte(fmt.Sprintf("%d", os.Getpid())), 0664)

}
