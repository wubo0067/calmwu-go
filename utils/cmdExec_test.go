/*
 * @Author: calm.wu
 * @Date: 2020-09-22 14:33:54
 * @Last Modified by: calm.wu
 * @Last Modified time: 2020-09-22 14:38:52
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
