/*
 * @Author: calm.wu
 * @Date: 2019-11-10 16:32:40
 * @Last Modified by: calm.wu
 * @Last Modified time: 2019-11-10 16:53:50
 */

package main

import (
	"os"
	"time"

	calm_utils "github.com/wubo0067/calmwu-go/utils"
)

func main() {
	logger := calm_utils.NewSimpleLog(nil)
	ticker := time.NewTicker(time.Duration(3 * time.Second))
	defer ticker.Stop()

	tickCount := 5
L:
	for {
		select {
		case <-ticker.C:
			logger.Printf("tickCount: %d\n", tickCount)
			tickCount--
			if tickCount == 0 {
				logger.Printf("childProcess exit!\n")
				break L
			}
		}
	}

	// exit code != 0，父进程cmd.Wait就会返回错误
	os.Exit(-1)
	return
}
