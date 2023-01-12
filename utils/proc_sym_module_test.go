/*
 * @Author: CALM.WU
 * @Date: 2023-01-10 16:14:02
 * @Last Modified by: CALM.WU
 * @Last Modified time: 2023-01-10 16:17:27
 */

package utils

import (
	"fmt"
	"testing"
)

func TestNewProcSyms(t *testing.T) {
	pid := 3638918

	pss, err := NewProcSyms(pid)
	if err != nil {
		t.Fatal(err.Error())
	}

	for i, psm := range pss.Modules {
		fmt.Printf("[%d] module: %s\n", i, psm.String())

		for j, ps := range psm.procSymTable {
			fmt.Printf("\t[%d] 0x%x\t%s\n", j, ps.address, ps.name)
		}
	}

	addr := uint64(0x00007fa9984962a6)
	name, offset, moduleName, err := pss.FindPsym(addr)
	if err != nil {
		t.Fatal(err.Error())
	} else {
		fmt.Printf("addr:0x%x %s+0x%02x [%s]\n\n", addr, name, offset, moduleName)
	}

	addr = uint64(0x00007fa998495b47)
	name, offset, moduleName, err = pss.FindPsym(addr)
	if err != nil {
		t.Fatal(err.Error())
	} else {
		fmt.Printf("addr:0x%x %s+0x%02x [%s]\n\n", addr, name, offset, moduleName)
	}

	addr = uint64(0x00007fa99848c1df)
	name, offset, moduleName, err = pss.FindPsym(addr)
	if err != nil {
		t.Fatal(err.Error())
	} else {
		fmt.Printf("addr:0x%x %s+0x%02x [%s]\n\n", addr, name, offset, moduleName)
	}

	addr = uint64(0x41f3b3)
	name, offset, moduleName, err = pss.FindPsym(addr)
	if err != nil {
		t.Fatal(err.Error())
	} else {
		fmt.Printf("addr:0x%x %s+0x%x [%s]\n\n", addr, name, offset, moduleName)
	}
}

// env GO111MODULE=off go test -v -run=TestNewProcSyms
