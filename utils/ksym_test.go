/*
 * @Author: CALM.WU
 * @Date: 2023-01-05 14:26:19
 * @Last Modified by: CALM.WU
 * @Last Modified time: 2023-01-05 16:05:49
 */

package utils

import (
	"testing"
)

// GO111MODULE=off go test -v -run=TestFindKsym
func TestFindKsym(t *testing.T) {
	err := LoadKallSyms()
	if err != nil {
		t.Fatal(err.Error())
	}

	count := len(__ksym_cache)
	t.Logf("Ksym count: %d\n", count)

	t.Logf("first ksym: %#v", __ksym_cache[0])
	t.Logf("last ksym: %#v", __ksym_cache[count-1])

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

	// ffffffffc03b4480 d fuse_acl_xattr_handlers	[fuse]
	addr = uint64(0xffffffffc03b4480)
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

	//ffffffffc01482c0 t sata_spd_string	[libata]
	addr = uint64(0xffffffffc01482c0)
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
