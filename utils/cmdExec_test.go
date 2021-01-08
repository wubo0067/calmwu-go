// +build linux

/*
 * @Author: calm.wu
 * @Date: 2020-09-22 14:33:54
 * @Last Modified by: CALM.WU
 * @Last Modified time: 2021-01-08 12:03:23
 */

package utils

import "testing"

func TestCmdExec(t *testing.T) {
	cmds := []string{"/bin/bash", "-c", "sleep 20 && echo end"}

	outRes, _, err := CmdExec(cmds...)
	if err != nil {
		t.Fatal(err.Error())
	} else {
		t.Log(outRes)
	}
}
