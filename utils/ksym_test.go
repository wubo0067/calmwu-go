/*
 * @Author: CALM.WU
 * @Date: 2023-01-05 14:26:19
 * @Last Modified by: CALM.WU
 * @Last Modified time: 2023-01-05 16:05:49
 */

package utils

import (
	"testing"

	"github.com/emirpasic/gods/sets/hashset"
)

// GO111MODULE=off go test -v -run=TestFindKsym
func TestFindKsym(t *testing.T) {
	err := LoadKallSyms()
	if err != nil {
		t.Fatal(err.Error())
	}

	count := len(ksymCache)
	t.Logf("Ksym count: %d\n", count)

	t.Logf("first ksym: %#v", ksymCache[0])
	t.Logf("last ksym: %#v", ksymCache[count-1])

	// 0000000000000000 A fixed_percpu_data
	addr := uint64(0x0000000000000000)
	name, err := FindKsym(addr)
	if err != nil {
		t.Fatal(err.Error())
	} else {
		t.Logf("addr:0x%x name:%s", addr, name)
	}

	// ffffffffa1b3c5d0 t complete_walk
	addr = uint64(0xffffffffa1b3c5d0)
	name, err = FindKsym(addr)
	if err != nil {
		t.Fatal(err.Error())
	} else {
		t.Logf("addr:0x%x name:%s", addr, name)
	}

	// ffffffffc0a45970 t nfs_file_open	[nfs]
	addr = uint64(0xffffffffc0a45970)
	name, err = FindKsym(addr)
	if err != nil {
		t.Fatal(err.Error())
	} else {
		t.Logf("addr:0x%x name:%s", addr, name)
	}

	// ffffffffc0124330 t crc_pcl      [crc32c_intel]
	addr = uint64(0xffffffffc0124330)
	name, err = FindKsym(addr)
	if err != nil {
		t.Fatal(err.Error())
	} else {
		t.Logf("addr:0x%x name:%s", addr, name)
	}

	//fffffffc02f67f0 t xfs_btree_overlapped_query_range	[xfs]
	addr = uint64(0xfffffffc02f67f0)
	name, err = FindKsym(addr)
	if err != nil {
		t.Fatal(err.Error())
	} else {
		t.Logf("addr:0x%x name:%s", addr, name)
	}

	//ffffffffc0146c00 t ata_platform_remove_one	[libata]
	addr = uint64(0xffffffffc0146c00)
	name, err = FindKsym(addr)
	if err != nil {
		t.Fatal(err.Error())
	} else {
		t.Logf("addr:0x%x name:%s", addr, name)
	}

	//ffffffffc0146cd0 t ata_pci_device_do_suspend	[libata]
	addr = uint64(0xffffffffc0146cd0)
	name, err = FindKsym(addr)
	if err != nil {
		t.Fatal(err.Error())
	} else {
		t.Logf("addr:0x%x name:%s", addr, name)
	}
}

/*
15:27:50 342284  342288  x-monitor       __x64_sys_openat
        ffffffffbab2deb1 __x64_sys_openat+0x1 [kernel]
        ffffffffba8042bb do_syscall_64+0x5b [kernel]
        ffffffffbb2000ad entry_SYSCALL_64_after_hwframe+0x65 [kernel]
*/

// GO111MODULE=off go test -v -run=TestKernelSymbols
func TestKernelSymbols(t *testing.T) {
	err := LoadKallSyms()
	if err != nil {
		t.Fatal(err.Error())
	}

	set := hashset.New("nfs_file_open",
		"ext4_punch_hole",
		"__tracepoint_xfs_btree_overlapped_query_range",
		"xfs_btree_overlapped_query_range",
		"fuse_acl_xattr_handlers",
		"uncore_down_prepare",
		"__x64_sys_sendmsg",
		"__x64_sys_mmap",
		"__x64_sys_memfd_create",
		"__x64_sys_pwritev",
		"__x64_sys_madvise",
	)

	// ksymCache sort by address
	for i, ksym := range ksymCache {
		if set.Contains(ksym.name) {
			t.Logf("ksym: %#v, next ksym: %#v", ksym, ksymCache[i+1])

			if symbol, err := FindKsym(ksym.address); err == nil {
				t.Logf("addr:0x%x name:%s", ksym.address, symbol)
			} else {
				t.Fatalf("addr:0x%x, err:%s", ksym.address, err.Error())
			}
		}
	}

	t.Log("-----------------")
	// addr:0xffffffff9880a1c0 name:uncore_down_prepare
	address := uint64(0xffffffff9880a1c1)
	if symbol, err := FindKsym(address); err == nil {
		t.Logf("addr:0x%x name:%s", address, symbol)
	} else {
		t.Fatalf("addr:0x%x err:%s", address, err)
	}

	address = uint64(0xffffffffc0a45980)
	if symbol, err := FindKsym(address); err == nil {
		t.Logf("addr:0x%x name:%s", address, symbol)
	} else {
		t.Fatalf("addr:0x%x err:%s", address, err)
	}

	t.Log("-----------------")

	for _, ksymName := range set.Values() {
		if KsymNameExists(ksymName.(string)) {
			t.Logf("%s in /proc/kallsyms", ksymName)
		} else {
			t.Logf("%s not in /proc/kallsyms", ksymName)
		}
	}
}
