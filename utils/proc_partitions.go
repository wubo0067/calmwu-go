//go:build linux
// +build linux

/*
 * @Author: CALM.WU
 * @Date: 2023-07-11 11:46:37
 * @Last Modified by: CALM.WU
 * @Last Modified time: 2023-07-11 12:04:05
 */

package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/sys/unix"
)

const (
	__defaultPartitionsFile = "/proc/partitions"
)

type ProcPartition struct {
	Major   uint32
	Minor   uint32
	Dev     uint64
	Blocks  uint64
	DevName string
}


// ProcPartitions reads the /proc/partitions file and returns a slice of ProcPartition structs
// representing the partitions listed in the file.
func ProcPartitions() ([]ProcPartition, error) {
	ppf, err := os.Open(__defaultPartitionsFile)
	if err != nil {
		return nil, err
	}
	defer ppf.Close()

	scanner := bufio.NewScanner(ppf)
	partitions := make([]ProcPartition, 0)

	for scanner.Scan() {
		var (
			major, minor uint32
			dev, blocks  uint64
			devName      string
		)

		line := scanner.Text()
		// skip header line
		if strings.HasPrefix(line, "major") {
			continue
		}
		// skip empty line
		if len(line) == 0 {
			continue
		}

		if _, err := fmt.Sscanf(line, "%d %d %d %s", &major, &minor, &blocks, &devName); err != nil {
			return nil, err
		}
		dev = dev = unix.Mkdev(major, minor)

		partitions = append(partitions, ProcPartition{
			Major:   major,
			Minor:   minor,
			Dev:     dev,
			Blocks:  blocks,
			DevName: devName,
		})
	}

	return partitions, nil
}
