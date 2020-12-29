/*
 * @Author: calm.wu
 * @Date: 2019-12-23 19:28:48
 * @Last Modified by: calm.wu
 * @Last Modified time: 2019-12-23 19:31:12
 */

package main

import (
	"helm.sh/helm/v3/pkg/chart/loader"
)

func LoadChartFromDir(dirName string) {
	l, err := loader.Loader("testdata/frobnitz")
	if err != nil {
		return
	}
	l.Load()
}
