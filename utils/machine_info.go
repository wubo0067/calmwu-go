/*
 * @Author: CALM.WU
 * @Date: 2024-02-28 11:14:31
 * @Last Modified by: CALM.WU
 * @Last Modified time: 2024-02-28 14:50:29
 */

package utils

import (
	"bytes"
	"os"
	"path"
	"runtime"
	"strings"

	"golang.org/x/sys/unix"
)

var (
	__MachineIDFiles = []string{
		"sys/devices/virtual/dmi/id/product_uuid", "etc/machine-id", "var/lib/dbus/machine-id",
	}
)

// Summary: MachineID is a function that retrieves the machine ID by reading from a list of specified files.
// It iterates through the list of files and reads the content from the '/proc/1/root' directory.
// If a valid machine ID is found, it is processed to remove whitespace and dashes before being returned.
// Parameters:
//
//	None
//
// Returns:
//
//	string: The machine ID extracted from the files, or an empty string if no valid ID is found.
func MachineID() string {
	for _, f := range __MachineIDFiles {
		idStr, err := os.ReadFile(path.Join("/proc/1/root", f))
		if err != nil {
			continue
		}
		id := strings.TrimSpace(strings.Replace(Bytes2String(idStr), "-", "", -1))
		return id
	}
	return ""
}

// Summary: Uname is a function that retrieves the hostname and kernel version of the current system by switching to the system's UTS namespace and reading the Utsname struct.
// Parameters:
//
//	None
//
// Returns:
//
//	string: The hostname of the current system.
//	string: The kernel version of the current system.
//	error: An error, if any, encountered during the retrieval process.
func Uname() (string, string, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	f, err := os.Open("/proc/1/ns/uts")
	if err != nil {
		return "", "", err
	}
	defer f.Close()

	self, err := os.Open("/proc/self/ns/uts")
	if err != nil {
		return "", "", err
	}
	defer self.Close()

	// 回到自己的 uts namespace
	defer func() {
		unix.Setns(int(self.Fd()), unix.CLONE_NEWUTS)
	}()

	// 到系统的 uts namespace
	err = unix.Setns(int(f.Fd()), unix.CLONE_NEWUTS)
	if err != nil {
		return "", "", err
	}

	var uts unix.Utsname
	if err := unix.Uname(&uts); err != nil {
		return "", "", err
	}

	hostname := Bytes2String(bytes.Split(uts.Nodename[:], []byte{0})[0])
	kernelVersion := Bytes2String(bytes.Split(uts.Release[:], []byte{0})[0])
	return hostname, kernelVersion, nil
}
