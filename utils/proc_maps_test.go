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
	"os"
	"testing"

	"github.com/parca-dev/parca-agent/pkg/stack/unwind"
)

const (
	__pyroscope = "/mnt/Program/pyroscope/pyroscope"
	__fio       = "/usr/bin/fio"
	__stack_bin = "/home/pingan/Program/x-monitor/bin/stack_unwind_cli"
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

// GO111MODULE=off go test -v -run=TestSymbolTblReuse -v -logtostderr
func TestSymbolTblReuse(t *testing.T) {
	InitModuleSymbolTblMgr(128)

	pid := 2533093
	xm_pss, _ := NewProcSyms(pid)
	t.Logf("pid:%d module count:%d", pid, len(xm_pss.Modules()))

	pid = 1
	first_pss, _ := NewProcSyms(pid)
	t.Logf("pid:%d module count:%d", pid, len(first_pss.Modules()))

	t.Logf("module symbol table cache count:%d", __singleModuleSymbolTblMgr.lc.Len())
}

// GO111MODULE=off go test -v -run=TestResolvePCXMonitor -v -logtostderr
func TestResolvePCXMonitor(t *testing.T) {
	// 容量为 4，会导致 lru 淘汰前面的
	InitModuleSymbolTblMgr(4)
	pid := 2533093 // x-monitor

	pss, err := NewProcSyms(pid)
	if err != nil {
		t.Fatal(err.Error())
	}

	ms := pss.Modules()
	for i, m := range ms {
		t.Logf("%d: %s", i, m.String())
	}

	// addr := uint64(0x00007fa9984962a6)
	// name, offset, moduleName, err := pss.ResolvePC(addr)
	// if err != nil {
	// 	t.Fatal(err.Error())
	// } else {
	// 	fmt.Printf("addr:0x%x %s+0x%02x [%s]\n\n", addr, name, offset, moduleName)
	// }

	// addr = uint64(0x00007fa998495b47)
	// name, offset, moduleName, err = pss.ResolvePC(addr)
	// if err != nil {
	// 	t.Fatal(err.Error())
	// } else {
	// 	fmt.Printf("addr:0x%x %s+0x%02x [%s]\n\n", addr, name, offset, moduleName)
	// }

	// addr = uint64(0x00007fa99848c1df)
	// name, offset, moduleName, err = pss.ResolvePC(addr)
	// if err != nil {
	// 	t.Fatal(err.Error())
	// } else {
	// 	fmt.Printf("addr:0x%x %s+0x%02x [%s]\n\n", addr, name, offset, moduleName)
	// }

	pcList := []uint64{
		0x424f8b,
		0x4f3340,
		0x4b4690,
		0x449a2f,
		0x527b40,
	}

	// 会重新加载 x-monitor
	for _, addr := range pcList {
		name, offset, moduleName, err := pss.ResolvePC(addr)
		if err != nil {
			t.Error(err.Error())
		} else {
			t.Logf("addr:0x%x %s+0x%x [%s]", addr, name, offset, moduleName)
		}
	}

	t.Logf("module symbol table cache count:%d", __singleModuleSymbolTblMgr.lc.Len())

	/*
		proc_maps_test.go:166: addr:0x424f8b fini_collector_proc_schedstat+0x0 [/mnt/Program/x-monitor/bin/x-monitor]
		proc_maps_test.go:166: addr:0x4f3340 memdup+0x0 [/mnt/Program/x-monitor/bin/x-monitor]
		proc_maps_test.go:166: addr:0x4b4690 cc_array_iter_index+0x0 [/mnt/Program/x-monitor/bin/x-monitor]
		proc_maps_test.go:166: addr:0x449a2f __get_all_childpids+0x0 [/mnt/Program/x-monitor/bin/x-monitor]
		proc_maps_test.go:166: addr:0x527b40 ZSTDv07_loadEntropy+0x0 [/mnt/Program/x-monitor/bin/x-monitor]
	*/
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

// GO111MODULE=off go test -v -run=TestPrintGOAppSymbols -v -logtostderr
func TestPrintGOAppSymbols(t *testing.T) {
	InitModuleSymbolTblMgr(128)

	pid := 4607 // pyroscope

	pss, err := NewProcSyms(pid)
	if err != nil {
		t.Fatal(err.Error())
	}

	ms := pss.Modules()
	for _, m := range ms {
		st, err := getModuleSymbolTbl(m.BuildID)
		if st != nil {
			funcs := st.(*GoModuleSymbolTbl).symIndex.Funcs
			for _, f := range funcs {
				t.Logf("name:'%s', entry:0x%x, end:0x%x", f.Name, f.Entry, f.End)
			}
		} else {
			t.Log(err.Error())
		}
	}
}

// dnf remove fio-debuginfo.x86_64
// dnf -y install fio-debuginfo.x86_64
// rpm -ql fio-debuginfo-3.19-3.el8.x86_64

// GO111MODULE=off go test -v -run=TestFioDebugSymbols -v -logtostderr
func TestFioDebugSymbols(t *testing.T) {
	InitModuleSymbolTblMgr(128)
	pmm := new(ProcMapsModule)
	pmm.Pathname = __fio
	pmm.RootFS = "/proc/1/root"
	err := pmm.loadProcModule()
	if err != nil {
		t.Fatal(err)
	} else {
		if st, err := getModuleSymbolTbl(pmm.BuildID); err == nil {
			t.Logf("%s have %d symbols", __fio, st.Count())
			for _, sym := range st.Symbols() {
				t.Logf("name:'%-40s', addr:'%#x'", sym.Name, sym.Address)
			}
		}
	}
}

// GO111MODULE=off go test -v -run=TestTextSection
func TestTextSection(t *testing.T) {
	fFIO, err := elf.Open(__fio)
	if err != nil {
		t.Fatal(err)
	}
	defer fFIO.Close()

	/*
	   proc_sym_module_test.go:246: .text sectionHeader elf.SectionHeader{Name:".text", Type:elf.SHT_PROGBITS, Flags:elf.SHF_ALLOC+elf.SHF_EXECINSTR, Addr:0x1bfc0, Offset:0x1bfc0, Size:0x73d12, Link:0x0, Info:0x0, Addralign:0x10, Entsize:0x0, FileSize:0x73d12}
	   proc_sym_module_test.go:261: debug .text sectionHeader elf.SectionHeader{Name:".text", Type:elf.SHT_NOBITS, Flags:elf.SHF_ALLOC+elf.SHF_EXECINSTR, Addr:0x1bfc0, Offset:0x360, Size:0x73d12, Link:0x0, Info:0x0, Addralign:0x10, Entsize:0x0, FileSize:0x73d12}
	*/
	txtSec := fFIO.Section(".text")
	t.Logf(".text sectionHeader %#v", txtSec.SectionHeader)

	buildID, err := GetBuildID(fFIO)
	if err != nil {
		t.Fatal(err)
	}

	debugFIO := findDebugFile(buildID.ID, "/proc/1/root", __fio, fFIO)
	if debugFIO != "" {
		fDebugFIO, err := elf.Open(debugFIO)
		if err != nil {
			t.Fatal(err)
		}

		debugTxtSec := fDebugFIO.Section(".text")
		t.Logf("debug .text sectionHeader %#v", debugTxtSec.SectionHeader)

		syms, err := fDebugFIO.Symbols()
		if err != nil {
			t.Fatal(err)
		}

		for _, sym := range syms {
			if sym.Value != 0 && sym.Info&0xf == byte(elf.STT_FUNC) {
				t.Logf("sym:'%s' val:0x%x, secionIndex:%d, section:'%s'", sym.Name, sym.Value, sym.Section, fDebugFIO.Sections[sym.Section].Name)
			}
		}
		// sym:'eta_time_within_slack' val:0x41350, secionIndex:15, section:'.text'
	}
}

// GO111MODULE=off go test -v -run=TestUnwindTable
func TestUnwindTable(t *testing.T) {
	tb, machine, err := unwind.GenerateCompactUnwindTable(__stack_bin, "stack_unwind_cli")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Logf("machine: %s", machine.GoString())
		for i, _ := range tb {
			row := &tb[i]
			t.Logf("%d: %#v", i, row)
		}
	}
}

// GO111MODULE=off go test -v -run=TestCheckInterpreterBin
func TestCheckInterpreterBin(t *testing.T) {
	InterpreterBinList := []string{
		"/usr/libexec/platform-python",
		"/usr/bin/python3.6",
		"/usr/lib/jvm/java-1.8.0-openjdk-1.8.0.362.b09-4.el9.x86_64/jre/bin/java",
		"/usr/lib/jvm/java-1.8.0-openjdk-1.8.0.312.b07-2.el8_5.x86_64/jre/bin/java",
	}

	for _, interpreter := range InterpreterBinList {
		_, err := os.Stat(interpreter)
		if err == nil {
			f, err := elf.Open(interpreter)
			if err == nil {
				t.Logf("===>check '%s'", interpreter)
				symbols, err := f.DynamicSymbols()
				if err == nil {
					for _, sym := range symbols {
						t.Logf("sym:'%s'", sym.Name)
						if v, ok := interpreterTags[sym.Name]; ok {
							t.Logf("'%s' type is '%s'<===", interpreter, func() string {
								switch v {
								case PythonLangType:
									return "python"
								case JavaLangType:
									return "java"
								}
								return "unknown"
							}())
							break
						}
					}
				}
			}
		}
	}
}
