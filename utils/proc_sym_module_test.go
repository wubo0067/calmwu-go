/*
 * @Author: CALM.WU
 * @Date: 2023-01-10 16:14:02
 * @Last Modified by: CALM.WU
 * @Last Modified time: 2023-09-12 17:47:28
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
	__executable = "/mnt/Program/pyroscope/pyroscope"
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
	f, tab := crack(__executable, t)
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

// env GO111MODULE=off go test -v -run=TestNewProcSyms
