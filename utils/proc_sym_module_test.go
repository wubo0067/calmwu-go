/*
 * @Author: CALM.WU
 * @Date: 2023-01-10 16:14:02
 * @Last Modified by: CALM.WU
 * @Last Modified time: 2023-09-18 17:34:54
 */

package utils

import (
	"debug/elf"
	"debug/gosym"
	"encoding/binary"
	"fmt"
	"testing"
)

const (
	__pyroscope = "/mnt/Program/pyroscope/pyroscope"
	__fio       = "/usr/bin/fio"
)

func readUint64(data []byte) uint64 {
	return binary.LittleEndian.Uint64(data)
}

func readUint32(data []byte) uint32 {
	return binary.LittleEndian.Uint32(data)
}

func crack(file string, t *testing.T) (*elf.File, *gosym.Table) {
	// Open self
	f, err := elf.Open(file)
	if err != nil {
		t.Fatal(err)
	}
	return parse(file, f, t)
}

func parse(file string, f *elf.File, t *testing.T) (*elf.File, *gosym.Table) {
	s := f.Section(".gosymtab")
	if s == nil {
		t.Skip("no .gosymtab section")
	}
	symdat, err := s.Data()
	if err != nil {
		f.Close()
		t.Fatalf("reading %s gosymtab: %v", file, err)
	}
	pclndat, err := f.Section(".gopclntab").Data()
	if err != nil {
		f.Close()
		t.Fatalf("reading %s gopclntab: %v", file, err)
	}

	pcln := gosym.NewLineTable(pclndat, f.Section(".text").Addr)
	tab, err := gosym.NewTable(symdat, pcln)
	if err != nil {
		f.Close()
		t.Fatalf("parsing %s gosymtab: %v", file, err)
	}

	return f, tab
}

// GO111MODULE=off go test -v -run=TestGOSymTab
func TestGOSymTab(t *testing.T) {
	f, tab := crack(__pyroscope, t)
	defer f.Close()

	pc := uint64(0x1004f32)
	symFunc := tab.PCToFunc(pc)
	if symFunc != nil {
		t.Logf("pc:%d ===> func:'%s + %d', entry:%d, end:%d\n",
			pc, symFunc.Name, symFunc.End-pc, symFunc.Entry, symFunc.End)
	} else {
		t.Fatal("pc to func failed")
	}

	pc = uint64(0x7db549)
	symFunc = tab.PCToFunc(pc)
	if symFunc != nil {
		t.Logf("pc:%d ===> func:'%s + %d', entry:%d, end:%d\n",
			pc, symFunc.BaseName(), symFunc.End-pc, symFunc.Entry, symFunc.End)
	} else {
		t.Fatal("pc to func failed")
	}
}

// GO111MODULE=off go test -v -run=TestResolveGO
func TestResolveGO(t *testing.T) {
	pid := 4607

	pss, err := NewProcSyms(pid)
	if err != nil {
		t.Fatal(err.Error())
	}

	pcList := []uint64{
		0x1004f32,
		0x47f067,
		0x1a8e40d,
		0x1a8b052,
		0x1a991d8,
	}

	for _, pc := range pcList {
		name, offset, moduleName, err := pss.ResolvePC(pc)
		if err == nil {
			t.Logf("addr:0x%x %s+0x%02x [%s]\n\n", pc, name, offset, moduleName)
		} else {
			t.Fatal(err.Error())
		}
	}
}

// GO111MODULE=off go test -v -run=TestResolveCApp
func TestResolveCApp(t *testing.T) {
	pid := 4607

	pss, err := NewProcSyms(pid)
	if err != nil {
		t.Fatal(err.Error())
	}

	addr := uint64(0x00007fa9984962a6)
	name, offset, moduleName, err := pss.ResolvePC(addr)
	if err != nil {
		t.Fatal(err.Error())
	} else {
		fmt.Printf("addr:0x%x %s+0x%02x [%s]\n\n", addr, name, offset, moduleName)
	}

	addr = uint64(0x00007fa998495b47)
	name, offset, moduleName, err = pss.ResolvePC(addr)
	if err != nil {
		t.Fatal(err.Error())
	} else {
		fmt.Printf("addr:0x%x %s+0x%02x [%s]\n\n", addr, name, offset, moduleName)
	}

	addr = uint64(0x00007fa99848c1df)
	name, offset, moduleName, err = pss.ResolvePC(addr)
	if err != nil {
		t.Fatal(err.Error())
	} else {
		fmt.Printf("addr:0x%x %s+0x%02x [%s]\n\n", addr, name, offset, moduleName)
	}

	addr = uint64(0x41f3b3)
	name, offset, moduleName, err = pss.ResolvePC(addr)
	if err != nil {
		t.Fatal(err.Error())
	} else {
		fmt.Printf("addr:0x%x %s+0x%x [%s]\n\n", addr, name, offset, moduleName)
	}
}

// GO111MODULE=off go test -v -run=TestBuildID
func TestBuildID(t *testing.T) {
	fFIO, err := elf.Open(__fio)
	if err != nil {
		t.Fatal(err)
	}
	defer fFIO.Close()

	buildID, err := GetBuildID(fFIO)
	if err != nil {
		t.Errorf("get %s buildid failed.", err.Error())
	} else {
		// = readelf -n /usr/bin/fio
		t.Logf("%s buildid:'%s', type:%d", __fio, buildID.ID, buildID.Type)
	}
	fPyro, err := elf.Open(__fio)
	if err != nil {
		t.Fatal(err)
	}
	defer fFIO.Close()

	buildID, err = GetBuildID(fPyro)
	if err != nil {
		t.Errorf("get %s buildid failed.", err.Error())
	} else {
		// = go tool buildid /mnt/Program/pyroscope/pyroscope
		/*
					readelf -n /mnt/Program/pyroscope/pyroscope

			Displaying notes found in: .note.go.buildid
			  Owner                 Data size       Description
			  Go                   0x00000053       Unknown note type: (0x00000004)
			  description data: 55 45 70 49 34 69 44 37 53 47 67 45 47 66 77 59 4c 55 58 5a 2f 43 58 48 52 57 5f 4c 37 38 76 44 47 58 5a 69 50 6a 74 6f 71 2f 45 65 79 68 57 32 56 6f 33 73 75 65 77 7a 4e 48 56 45 36 6f 2f 53 53 64 59 38 34 77 32 45 44 68 4b 53 75 47 70 5f 6d 39 56
		*/
		t.Logf("%s buildid:'%s', type:%d", __pyroscope, buildID.ID, buildID.Type)
	}
}

// GO111MODULE=off go test -v -run=TestFindDebugFile
func TestFindDebugFile(t *testing.T) {
	fFIO, err := elf.Open(__fio)
	if err != nil {
		t.Fatal(err)
	}
	defer fFIO.Close()

	buildID, err := GetBuildID(fFIO)
	if err != nil {
		t.Errorf("get %s buildid failed.", err.Error())
	} else {
		// = readelf -n /usr/bin/fio
		debugFile := findDebugFile(buildID.ID, "/proc/1/root", __fio, fFIO)
		t.Logf("%s buildid:'%s', type:%d, debugfile:'%s'", __fio, buildID.ID, buildID.Type, debugFile)
	}
}

// dnf remove fio-debuginfo.x86_64
// dnf -y install fio-debuginfo.x86_64

// GO111MODULE=off go test -v -run=TestLoadSymbols
func TestLoadSymbols(t *testing.T) {
	fFIO, err := elf.Open(__fio)
	if err != nil {
		t.Fatal(err)
	}
	defer fFIO.Close()

	psm := new(ProcSymsModule)
	psm.buildSymTable(fFIO)

	t.Logf("%s symbol count:%d", __fio, len(psm.procSymTable))

	for _, sym := range psm.procSymTable {
		t.Logf("name:'%s', val:'%x'", sym.name, sym.pc)

	}

	// dyn name:'blktrace_lookup_device', val:'8e400', section:'15==>.text'
}
