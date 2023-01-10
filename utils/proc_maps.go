/*
 * @Author: CALM.WU
 * @Date: 2023-01-10 14:20:15
 * @Last Modified by: CALM.WU
 * @Last Modified time: 2023-01-10 16:17:22
 */

package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
)

const (
	__defaultProcMapsEntryDefaultFieldCount = 6
)

type ProcMapPermissions struct {
	// Readable is true if the mapping is readable.
	Readable bool
	// Writable is true if the mapping is writable.
	Writable bool
	// Executable is true if the mapping is executable.
	Executable bool
	// Shared is true if the mapping is shared.
	Shared bool
	// Private is true if the mapping is private.
	Private bool
}

type ProcMap struct {
	// StartAddr is the starting address of current mapping.
	StartAddr uint64
	// EndAddr is the ending address of current mapping.
	EndAddr uint64
	// Perm is the permission of current mapping.
	Perm *ProcMapPermissions
	// Offset is the offset of current mapping.
	Offset uint64
	// Dev is the device of current mapping.
	Dev uint64
	// Inode is the inode of current mapping. find / -inum 101417806 or lsof -n -i 'inode=174919'
	Inode uint64
	//
	Pathname string
}

func __parseProcMap(line string) (*ProcMap, error) {
	fields := strings.Fields(line)
	field_count := len(fields)
	if field_count < 5 {
		return nil, errors.Errorf("truncated  procmap entry")
	}

	procMap := new(ProcMap)
	if field_count == __defaultProcMapsEntryDefaultFieldCount {
		fmt.Sscan(line, "%x-%x %s %x %x:%x %d", &procMap.StartAddr, &procMap.EndAddr, &procMap.Perm,
			&procMap.Offset, &procMap.Dev, &procMap.Inode)
		procMap.Pathname = ""
	} else if field_count > __defaultProcMapsEntryDefaultFieldCount {
		fmt.Sscan(line, "%x-%x %s %x %x:%x %d %s", &procMap.StartAddr, &procMap.EndAddr, &procMap.Perm,
			&procMap.Offset, &procMap.Dev, &procMap.Inode, &procMap.Pathname)
	}
	return procMap, nil
}

// It reads the contents of /proc/pid/maps, parses each line, and returns a slice of ProcMap structs
func NewProcMaps(pidProcMapsFile string) ([]*ProcMap, error) {
	procMapsFile, err := os.Open(pidProcMapsFile)
	if err != nil {
		return nil, errors.Wrap(err, "NewProcMap open failed")
	}
	defer procMapsFile.Close()

	// use nil slice not empty slice
	var maps []*ProcMap
	scanner := bufio.NewScanner(procMapsFile)

	for scanner.Scan() {
		m, err := __parseProcMap(scanner.Text())
		if err != nil {
			return nil, errors.Wrap(err, "NewProcMap __parseProcMap failed")
		}
		maps = append(maps, m)
	}

	return maps, nil
}
