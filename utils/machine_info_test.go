/*
 * @Author: CALM.WU
 * @Date: 2024-02-28 11:20:23
 * @Last Modified by: CALM.WU
 * @Last Modified time: 2024-02-28 15:22:13
 */

package utils

import (
	"regexp"
	"testing"
)

var (
	kernelVersionRe = regexp.MustCompile(`^(\d+\.\d+\.\d+)`)
)

/*
   machine_info_test.go:15: MachineID: 0b2101b36c86914dabd67197969a7d04
   machine_info_test.go:19: Hostname: Redhat8-01, KernelVersion: 4.18.0-348.7.1.el8_5.x86_64
*/

// GO111MODULE=off go test -v -run=TestMachineInfo
func TestMachineInfo(t *testing.T) {
	id := MachineID()
	t.Log("MachineID:", id)

	hostname, kv, err := Uname()
	if err == nil {
		majorMinor := kernelVersionRe.FindString(kv)
		t.Logf("Hostname: %s, KernelVersion: %s, Major.Minor.Patch: %s", hostname, kv, majorMinor)
	} else {
		t.Error("Uname failed:", err)
	}
}
