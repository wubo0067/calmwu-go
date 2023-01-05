/*
 * @Author: CALM.WU
 * @Date: 2023-01-05 14:26:19
 * @Last Modified by: CALM.WU
 * @Last Modified time: 2023-01-05 16:05:49
 */

package utils

import (
	"fmt"
	"testing"
)

func TestFindKsym(t *testing.T) {
	err := LoadKallSyms()
	if err != nil {
		t.Fatal(err.Error())
	}

	fmt.Printf("Ksym count: %d\n", len(_ksym_cache))

	addr := uint64(0xffffffffbab2deb1)
	name, offset, err := FindKsym(addr)
	if err != nil {
		t.Fatal(err.Error())
	} else {
		fmt.Printf("addr:0x%x name:%s offset:0x%02x\n\n", addr, name, offset)
	}

	addr = uint64(0xffffffffba804260)
	name, offset, err = FindKsym(addr)
	if err != nil {
		t.Fatal(err.Error())
	} else {
		fmt.Printf("addr:0x%x name:%s offset:0x%02x\n\n", addr, name, offset)
	}

	addr = uint64(0xffffffffba8042bb)
	name, offset, err = FindKsym(addr)
	if err != nil {
		t.Fatal(err.Error())
	} else {
		fmt.Printf("addr:0x%x name:%s offset:0x%02x\n\n", addr, name, offset)
	}

	addr = uint64(0xffffffffbb2000ad)
	name, offset, err = FindKsym(addr)
	if err != nil {
		t.Fatal(err.Error())
	} else {
		fmt.Printf("addr:0x%x name:%s offset:0x%02x\n\n", addr, name, offset)
	}
}

/*
15:27:50 342284  342288  x-monitor       __x64_sys_openat
        ffffffffbab2deb1 __x64_sys_openat+0x1 [kernel]
        ffffffffba8042bb do_syscall_64+0x5b [kernel]
        ffffffffbb2000ad entry_SYSCALL_64_after_hwframe+0x65 [kernel]
*/
