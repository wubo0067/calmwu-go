/*
 * @Author: CALM.WU
 * @Date: 2023-01-10 14:20:15
 * @Last Modified by: CALM.WU
 * @Last Modified time: 2023-10-26 14:41:55
 */

package utils

import (
	"bufio"
	"debug/elf"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
	"syscall"

	"github.com/parca-dev/parca-agent/pkg/buildid"
	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
)

const (
	__miniProcMapsEntryDefaultFieldCount = 6
	__debugLinkSection                   = ".gnu_debuglink"
)

type ProcModuleType int

var (
	ErrProcModuleNotSupport       = errors.New("proc module not support")
	ErrProcModuleNotSymbolSection = errors.New("proc module not symbol section")
	ErrProcModuleHasNoSymbols     = errors.New("proc module has no symbols")
)

const (
	UNKNOWN ProcModuleType = iota
	EXEC
	SO
	VDSO
)

type ModuleSym struct {
	pc   uint64
	name string
}

type ProcModulePermissions struct {
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

type ProcModule struct {
	// StartAddr is the starting pc of current mapping.
	StartAddr uint64
	// EndAddr is the ending pc of current mapping.
	EndAddr uint64
	// Perm is the permission of current mapping.
	Perms ProcModulePermissions
	// Offset is the offset of current mapping.
	Offset uint64
	// Dev is the device of current mapping.
	Dev uint64
	// Inode is the inode of current mapping. find / -inum 101417806 or lsof -n -i 'inode=174919'
	Inode uint64
	// 内存段所属的文件的路径名
	Pathname string
	//
	Type ProcModuleType
	//
	procSymTable []*ModuleSym
	SymCount     int
	//
	goSymTable *GoSymTable
	BuildID    string
}

func (psm *ProcModule) open(appRootFS string) (*elf.File, error) {
	// rootfs: /proc/%d/root
	var (
		elfF *elf.File
		err  error
	)
	modulePath := fmt.Sprintf("%s%s", appRootFS, psm.Pathname)

	elfF, err = elf.Open(modulePath)
	if err != nil {
		return nil, errors.Wrapf(err, "open ELFfile:'%s'.", modulePath)
	}

	return elfF, nil
}

func cString(bs []byte) string {
	i := 0
	for ; i < len(bs); i++ {
		if bs[i] == 0 {
			break
		}
	}
	return Bytes2String(bs[:i])
}

func findDebugFile(buildID, appRootFS, pathName string, elfF *elf.File) string {
	// 首先在/usr/lib/debug/.build-id目录下根据buildid查找debug文件
	debugFile := fmt.Sprintf("/usr/lib/debug/.build-id/%s/%s.debug", buildID[:2], buildID[2:])
	fsDebugFile := fmt.Sprintf("%s%s", appRootFS, debugFile)

	//fmt.Printf("debugFile:'%s', fsDebugFile:'%s'\n", debugFile, fsDebugFile)
	_, err := os.Stat(fsDebugFile)
	if err == nil {
		return debugFile
	}

	// 读取.gnu_debuglink
	debugLinkSection := elfF.Section(__debugLinkSection)
	if debugLinkSection != nil {
		debugLinkData, err := debugLinkSection.Data()
		if err == nil {
			if len(debugLinkData) >= 6 {
				//_ = debugLinkData[len(debugLinkData)-4:]
				debugLink := cString(debugLinkData)
				//fmt.Printf("debugLink:'%s'\n", debugLink) //hex.EncodeToString(crc))

				// /usr/bin/ls.debug
				fsDebugFile := path.Join(appRootFS, path.Dir(pathName), debugLink)
				//fmt.Printf("fsDebugFile:'%s'\n", fsDebugFile)
				_, err = os.Stat(fsDebugFile)
				if err == nil {
					return fsDebugFile
				}
				// /usr/bin/.debug/ls.debug
				fsDebugFile = path.Join(appRootFS, path.Dir(pathName), ".debug", debugLink)
				//fmt.Printf("fsDebugFile:'%s'\n", fsDebugFile)
				_, err = os.Stat(fsDebugFile)
				if err == nil {
					return fsDebugFile
				}
				// /usr/lib/debug/usr/bin/ls.debug.
				fsDebugFile = path.Join(appRootFS, "/usr/lib/debug", path.Dir(pathName), debugLink)
				//fmt.Printf("fsDebugFile:'%s'\n", fsDebugFile)
				_, err = os.Stat(fsDebugFile)
				if err == nil {
					return fsDebugFile
				}
			}
		}
	}

	return ""
}

func (psm *ProcModule) buildSymTable(elfF *elf.File) error {
	// from .text section read symbol and pc
	symbols, err := elfF.Symbols()
	if err != nil && !errors.Is(err, elf.ErrNoSymbols) {
		return errors.Wrapf(err, "read module:'%s' SYMTAB.", psm.Pathname)
	}

	dynSymbols, err := elfF.DynamicSymbols()
	if err != nil && !errors.Is(err, elf.ErrNoSymbols) {
		return errors.Wrapf(err, "read module:'%s' DYNSYM.", psm.Pathname)
	}

	pfnAddSymbol := func(m *ProcModule, syms []elf.Symbol) {
		for _, sym := range syms {
			if sym.Value != 0 && sym.Info&0xf == byte(elf.STT_FUNC) {
				ps := new(ModuleSym)
				ps.name = sym.Name
				ps.pc = sym.Value
				m.procSymTable = append(m.procSymTable, ps)
				m.SymCount = m.SymCount + 1
			}
		}
	}

	pfnAddSymbol(psm, symbols)
	pfnAddSymbol(psm, dynSymbols)

	if psm.SymCount == 0 {
		return ErrProcModuleHasNoSymbols
	}

	//fmt.Printf("-------------module:'%s' SymCount:%d.\n", psm.Pathname, psm.SymCount)

	// 按地址排序，地址相同按名字排序
	sort.Slice(psm.procSymTable, func(i, j int) bool {
		if psm.procSymTable[i].pc == psm.procSymTable[j].pc {
			return psm.procSymTable[i].name < psm.procSymTable[j].name
		}
		return psm.procSymTable[i].pc < psm.procSymTable[j].pc
	})

	return nil
}

// It reads the contents of /proc/pid/maps, parses each line, and returns a slice of ProcMap entries.
func (psm *ProcModule) loadProcModule(appRootFS string) error {
	var (
		elfF      *elf.File
		elfDebugF *elf.File
		err       error
	)

	// 打开elf文件
	if elfF, err = psm.open(appRootFS); err != nil {
		return errors.Wrapf(err, "psm open:'%s%s'.", appRootFS, psm.Pathname)
	}
	defer elfF.Close()

	// 获取module类型，编译使用了-fPIE生成位置无关的执行程序，Type会是ET_DYN，否则就是ET_EXEC
	// 	?  bin git:(feature-xm-ebpf-collector) ? readelf -h ./x-monitor|grep 'Type:'
	//   Type:                              EXEC (Executable file)
	// ?  bin git:(feature-xm-ebpf-collector) ? readelf -h /bin/fio|grep 'Type:'
	//   Type:                              DYN (Shared object file)
	// ?  bin git:(feature-xm-ebpf-collector) ? ps -ef|grep ssh
	switch elfF.Type {
	case elf.ET_EXEC:
		psm.Type = EXEC
	case elf.ET_DYN:
		psm.Type = SO
	default:
		return ErrProcModuleNotSupport
	}

	psm.BuildID, err = buildid.FromELF(elfF)
	if err != nil {
		return errors.Wrapf(err, "failed to get build ID for %s", psm.Pathname)
	}

	// 查找对应debug文件
	debugFilePath := findDebugFile(psm.BuildID, appRootFS, psm.Pathname, elfF)
	if debugFilePath != "" {
		// 直接加载debug文件
		elfDebugF, err = elf.Open(debugFilePath)
		if err == nil {
			defer elfDebugF.Close()
			err = psm.buildSymTable(elfDebugF)
		}
	} else {
		// 直接从elf文件中加载symbol
		err = psm.buildSymTable(elfF)
	}

	return err
}

// A method of the ProcModule struct. It is used to print the ProcModule struct.
func (psm *ProcModule) String() string {
	return fmt.Sprintf("%x-%x %#v %x %x %d %s, symbols:%d",
		psm.StartAddr, psm.EndAddr, psm.Perms, psm.Offset, psm.Dev, psm.Inode, psm.Pathname, len(psm.procSymTable))
}

func (psm *ProcModule) resolvePC(pc uint64) (string, uint32, string, error) {
	//size := len(psm.procSymTable)
	// 二分查找
	index := sort.Search(psm.SymCount, func(i int) bool {
		return psm.procSymTable[i].pc > pc
	})

	// addr小于所有symbol的最小地址
	if index == 0 {
		return "", 0, "", errors.Errorf("pc:0x%x not in symtab{0x%x---0x%0x} with module:'%s'",
			pc, psm.procSymTable[0].pc, psm.procSymTable[psm.SymCount-1].pc, psm.Pathname)
	}

	// 找到了
	ps := psm.procSymTable[index-1]
	return ps.name, uint32(pc - ps.pc), psm.Pathname, nil
}

type ProcSyms struct {
	// pid
	Pid int
	// ProcModule slice
	Modules []*ProcModule
	// inode, Determine whether to refresh
	InodeID uint64
}

// It parses a line from the /proc/<pid>/maps file and returns a ProcModule struct
func parseProcMapEntry(line string, pss *ProcSyms) error {
	// 7ff8be1a5000-7ff8be1c0000 r-xp 00000000 fd:00 570150                     /usr/lib64/libpthread-2.28.so
	var (
		err                error
		perms              string
		devMajor, devMinor uint64
		appRootFS          = fmt.Sprintf("/proc/%d/root", pss.Pid)
	)

	fields := strings.Fields(line)
	field_count := len(fields)
	if field_count != __miniProcMapsEntryDefaultFieldCount {
		return nil
	}

	psm := new(ProcModule)
	psm.Type = UNKNOWN

	fmt.Sscanf(line, "%x-%x %s %x %x:%x %d %s", &psm.StartAddr, &psm.EndAddr, &perms,
		&psm.Offset, &devMajor, &devMinor, &psm.Inode, &psm.Pathname)

	//fmt.Printf("parse line:'%s'\n", line)

	if len(psm.Pathname) == 0 ||
		strings.Contains(psm.Pathname, "[vdso]") ||
		strings.Contains(psm.Pathname, "[vsyscall]") {
		return nil
	}

	permBytes := String2Bytes(perms)
	if permBytes[2] != 'x' {
		return nil
	}

	for _, ch := range perms {
		switch ch {
		case 'r':
			psm.Perms.Readable = true
		case 'w':
			psm.Perms.Writable = true
		case 'x':
			psm.Perms.Executable = true
		case 's':
			psm.Perms.Shared = true
		case 'p':
			psm.Perms.Private = true
		}
	}

	psm.Dev = unix.Mkdev(uint32(devMajor), uint32(devMinor))

	// 测试golang程序的load
	if err = psm.loadProcGoModule(appRootFS); err != nil {
		if err = psm.loadProcModule(appRootFS); err != nil {
			if errors.Is(err, ErrProcModuleNotSupport) || errors.Is(err, ErrProcModuleHasNoSymbols) {
				// 不加入，忽略，继续
				psm = nil
				return nil
			}
			return errors.Wrapf(err, "load module:'%s' failed.", psm.Pathname)
		}
	}

	pss.Modules = append(pss.Modules, psm)

	return nil
}

func NewProcSyms(pid int) (*ProcSyms, error) {
	procMapsFile, err := os.Open(fmt.Sprintf("/proc/%d/maps", pid))
	if err != nil {
		return nil, errors.Wrap(err, "NewProcMap open failed")
	}
	defer procMapsFile.Close()

	fileExe, err := os.Stat(fmt.Sprintf("/proc/%d/exe", pid))
	if err != nil {
		return nil, errors.Wrap(err, "stat execute file failed.")
	}
	stat := fileExe.Sys().(*syscall.Stat_t)

	pss := new(ProcSyms)
	pss.Pid = pid
	pss.InodeID = stat.Ino

	scanner := bufio.NewScanner(procMapsFile)

	for scanner.Scan() {
		// maps每一行的信息
		text := scanner.Text()
		err := parseProcMapEntry(text, pss)
		if err != nil {
			return nil, errors.Wrapf(err, "parse text:'%s' failed", text)
		}
	}

	return pss, nil
}

// ResolvePC 根据程序计数器(PC)解析符号信息
// 如果 ProcSyms 中的模块为空，则返回错误
// 如果 PC 在模块的地址范围内，则返回符号名称、偏移量和路径名
// 如果模块类型为 SO，则返回符号名称、偏移量和路径名
// 如果模块类型为 EXEC，则返回符号名称、偏移量和路径名
// 如果在解析过程中出现错误，则返回错误
func (pss *ProcSyms) ResolvePC(pc uint64) (string, uint32, string, error) {
	if len(pss.Modules) == 0 {
		return "", 0, "", errors.New("proc modules is empty")
	}

	for _, psm := range pss.Modules {
		if pc >= psm.StartAddr && pc <= psm.EndAddr {
			if psm.Type == SO {
				return psm.resolvePC(pc - psm.StartAddr)
			} else if psm.Type == EXEC {
				if psm.goSymTable != nil {
					symName, offset, err := psm.goSymTable.__resolveGoPC(pc)
					if err == nil {
						return symName, offset, psm.Pathname, nil
					}
				} else {
					return psm.resolvePC(pc)
				}
			}
		}
	}
	return "", 0, "", errors.Errorf("pc:0x%x is outside the valid ranges in /proc/%d/maps", pc, pss.Pid)
}

// GetModules 返回进程符号表中的所有模块。
func (pss *ProcSyms) GetModules() []*ProcModule {
	return pss.Modules
}
