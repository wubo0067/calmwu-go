/*
 * @Author: calmwu
 * @Date: 2018-11-14 15:34:17
 * @Last Modified by: calmwu
 * @Last Modified time: 2019-02-22 18:50:02
 */

package utils

import (
	"bytes"
	"fmt"
	"runtime"
	"time"
)

func MemUsage() string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	var infoBuf bytes.Buffer

	infoBuf.WriteString(fmt.Sprintf("Alloc = %v MiB", bToMb(m.Alloc)))
	infoBuf.WriteString(fmt.Sprintf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc)))
	infoBuf.WriteString(fmt.Sprintf("\tSys = %v MiB", bToMb(m.Sys)))
	infoBuf.WriteString(fmt.Sprintf("\tNumGC = %v\n", m.NumGC))

	return infoBuf.String()
}

func bToMb(b uint64) uint64 {
	return b >> 20
}

func TimeTaken(t time.Time, name string) {
	elapsed := time.Since(t)
	if ZLog != nil {
		ZLog.Debugf("TIME: %s took %s\n", name, elapsed)
	} else {
		fmt.Printf("TIME: %s took %s\n", name, elapsed)
	}
}
