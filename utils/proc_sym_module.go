/*
 * @Author: CALM.WU
 * @Date: 2023-01-10 14:20:15
 * @Last Modified by: CALM.WU
 * @Last Modified time: 2023-01-10 16:17:22
 */

package utils

import (
	"bufio"
	"debug/elf"
	"fmt"
	"os"
	"sort"
	"strings"
	"syscall"

	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
)

const (
	__miniProcMapsEntryDefaultFieldCount = 6
)

type ProcSymModuleType int

var (
	__errProcModuleNotSupport       = errors.New("proc module not support")
	__errProcModuleNotSymbolSection = errors.New("proc module not symbol section")
)

const (
	UNKNOWN ProcSymModuleType = iota
	EXEC
	SO
	VDSO
)

type ProcSym struct {
	address uint64
	name    string
}

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

type ProcSymsModule struct {
	// StartAddr is the starting address of current mapping.
	StartAddr uint64
	// EndAddr is the ending address of current mapping.
	EndAddr uint64
	// Perm is the permission of current mapping.
	Perms ProcMapPermissions
	// Offset is the offset of current mapping.
	Offset uint64
	// Dev is the device of current mapping.
	Dev uint64
	// Inode is the inode of current mapping. find / -inum 101417806 or lsof -n -i 'inode=174919'
	Inode uint64
	//
	Pathname string
	//
	Type ProcSymModuleType
	//
	procSymTable []*ProcSym
}

// It reads the contents of /proc/pid/maps, parses each line, and returns a slice of ProcMap entries.
func (psm *ProcSymsModule) __loadProcModule(pid int) error {
	//
	nsRelativePath := fmt.Sprintf("/proc/%d/root%s", pid, psm.Pathname)

	// 打开elf文件
	elfFile, err := elf.Open(nsRelativePath)
	if err != nil {
		return errors.Wrapf(err, "open elfFile '%s' failed.", nsRelativePath)
	}
	defer elfFile.Close()

	// 获取module类型
	switch elfFile.Type {
	case elf.ET_EXEC:
		psm.Type = EXEC
	case elf.ET_DYN:
		psm.Type = SO
	default:
		return __errProcModuleNotSupport
	}

	// from .text section read symbol and address
	symbols, err := elfFile.Symbols()
	if err != nil {
		return __errProcModuleNotSymbolSection
	}

	for _, symbol := range symbols {
		if int(symbol.Section) < len(elfFile.Sections) &&
			elfFile.Sections[symbol.Section].Name == ".text" &&
			len(symbol.Name) > 0 {
			ps := new(ProcSym)
			ps.name = symbol.Name
			ps.address = symbol.Value
			psm.procSymTable = append(psm.procSymTable, ps)
		}
	}

	// 排序
	sort.Slice(psm.procSymTable, func(i, j int) bool {
		return psm.procSymTable[i].address < psm.procSymTable[j].address
	})

	return nil
}

// A method of the ProcSymsModule struct. It is used to print the ProcSymsModule struct.
func (psm *ProcSymsModule) String() string {
	return fmt.Sprintf("%x-%x %#v %x %x %d %s, symbols:%d",
		psm.StartAddr, psm.EndAddr, psm.Perms, psm.Offset, psm.Dev, psm.Inode, psm.Pathname, len(psm.procSymTable))
}

func (psm *ProcSymsModule) __resolveAddr(addr uint64) (string, uint32, string, error) {
	// 二分查找
	index := sort.Search(len(psm.procSymTable), func(i int) bool {
		return psm.procSymTable[i].address > addr
	})

	// addr小于所有symbol的最小地址
	if index == 0 {
		return "", 0, "", errors.Errorf("can't find symbol in module:'%s'", psm.Pathname)
	}

	// 找到了
	ps := psm.procSymTable[index-1]
	return ps.name, uint32(addr - ps.address), psm.Pathname, nil
}

type ProcSyms struct {
	// pid
	Pid int
	// ProcSymsModule slice
	Modules []*ProcSymsModule
	// inode, Determine whether to refresh
	InodeID uint64
}

// It parses a line from the /proc/<pid>/maps file and returns a ProcSymsModule struct
func __parseProcMapEntry(line string, pss *ProcSyms) error {
	// 7ff8be1a5000-7ff8be1c0000 r-xp 00000000 fd:00 570150                     /usr/lib64/libpthread-2.28.so
	fields := strings.Fields(line)
	field_count := len(fields)
	if field_count != __miniProcMapsEntryDefaultFieldCount {
		return nil
	}

	var perms string
	var devMajor, devMinor uint64

	psm := new(ProcSymsModule)
	psm.Type = UNKNOWN

	fmt.Sscanf(line, "%x-%x %s %x %x:%x %d %s", &psm.StartAddr, &psm.EndAddr, &perms,
		&psm.Offset, &devMajor, &devMinor, &psm.Inode, &psm.Pathname)

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

	if err := psm.__loadProcModule(pss.Pid); err != nil {
		if errors.Is(err, __errProcModuleNotSupport) ||
			errors.Is(err, __errProcModuleNotSymbolSection) {
			return nil
		}
		return errors.Wrapf(err, "parse module '%s' failed.", psm.Pathname)
	}

	pss.Modules = append(pss.Modules, psm)

	return nil
}

// It reads the contents of /proc/pid/maps, parses each line, and returns a slice of ProcMap entries.
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
		err := __parseProcMapEntry(scanner.Text(), pss)
		if err != nil {
			return nil, errors.Wrap(err, "NewProcMap __parseProcMapEntry failed")
		}
	}

	return pss, nil
}

// Used to find the symbol of the specified address.
func (pss *ProcSyms) FindPsym(addr uint64) (string, uint32, string, error) {
	if len(pss.Modules) == 0 {
		return "", 0, "", errors.New("ProcSyms is not initialized")
	}

	for _, psm := range pss.Modules {
		if addr >= psm.StartAddr && addr <= psm.EndAddr {
			return psm.__resolveAddr(addr - psm.StartAddr)
		}
	}
	return "", 0, "", errors.Errorf("addr:%x not in /proc/%d/maps", addr, pss.Pid)
}
