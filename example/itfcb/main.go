/*
 * @Author: calm.wu
 * @Date: 2019-07-16 09:50:02
 * @Last Modified by: calm.wu
 * @Last Modified time: 2019-07-16 10:00:34
 */

package main

import (
	"fmt"
	"os"
)

// 状态监视器
type StatusObserver interface {
	OnChange(newInfo string)
}

type FileStatusObserver struct {
}

func (sfo *FileStatusObserver) OnChange(newInfo string) {
	fmt.Printf("FileStatusObserver %s\n", newInfo)
}

func SimulationCreateFile(so StatusObserver) {
	hFile, err := os.OpenFile("sim.txt", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		so.OnChange(fmt.Sprintf("create file failed! reason:%s", err.Error()))
	} else {
		so.OnChange("create file successed")
	}
	hFile.Close()
}

func main() {
	so := new(FileStatusObserver)
	SimulationCreateFile(so)
}
