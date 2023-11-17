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

type ProcMapsModule struct {
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
	Inode    uint64
	Pathname string // 内存段所属的文件的路径名
	RootFS   string
	Type     ProcModuleType
	BuildID  string
}

func (pmm *ProcMapsModule) open() (*elf.File, error) {
	// rootfs: /proc/%d/root
	var (
		elfF *elf.File
		err  error
	)
	modulePath := fmt.Sprintf("%s%s", pmm.RootFS, pmm.Pathname)

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
	// 首先在/usr/lib/debug/.build-id 目录下根据 buildid 查找 debug 文件
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

// It reads the contents of /proc/pid/maps, parses each line, and returns a slice of ProcMap entries.
func (pmm *ProcMapsModule) loadProcModule() error {
	var (
		elfF *elf.File
		err  error
	)

	// 打开 elf 文件
	if elfF, err = pmm.open(); err != nil {
		return errors.Wrap(err, "pmm open.")
	}
	defer elfF.Close()

	// 获取 module 类型，编译使用了-fPIE 生成位置无关的执行程序，Type 会是 ET_DYN，否则就是 ET_EXEC
	// 	?  bin git:(feature-xm-ebpf-collector) ? readelf -h ./x-monitor|grep 'Type:'
	//   Type:                              EXEC (Executable file)
	// ?  bin git:(feature-xm-ebpf-collector) ? readelf -h /bin/fio|grep 'Type:'
	//   Type:                              DYN (Shared object file)
	// 获取文件类型，在计算 address 是要根据类型判断是否减去 start address
	switch elfF.Type {
	case elf.ET_EXEC:
		pmm.Type = EXEC
	case elf.ET_DYN:
		pmm.Type = SO
	default:
		return ErrProcModuleNotSupport
	}
	// 获取 buildID
	pmm.BuildID, err = buildid.FromELF(elfF)
	if err != nil {
		return errors.Wrapf(err, "failed to get build ID for %s", pmm.Pathname)
	}

	// 判断该 buildID 是否已经缓存
	if st, err := getModuleSymbolTbl(pmm.BuildID); st != nil && err == nil {
		// 已经 cache 了，不用继续解析了
		//glog.Infof("module:'%s' buildID:'%s' has cached.", pmm.Pathname, pmm.BuildID)
		return nil
	}
	// 生成符号表
	_, err = createModuleSymbolTbl(pmm.BuildID, pmm.Pathname, pmm.RootFS, elfF)

	return err
}

// A method of the ProcMapsModule struct. It is used to print the ProcMapsModule struct.
func (pmm *ProcMapsModule) String() string {
	return fmt.Sprintf("%x-%x %#v %x %x %d %s",
		pmm.StartAddr, pmm.EndAddr, pmm.Perms, pmm.Offset, pmm.Dev, pmm.Inode, pmm.Pathname)
}

type ProcMaps struct {
	// pid
	Pid int
	// ProcMapsModule slice
	ModuleList []*ProcMapsModule
	// inode, Determine whether to refresh
	InodeID uint64
}

// It parses a line from the /proc/<pid>/maps file and returns a ProcMapsModule struct
func parseProcMapsEntry(line string, pss *ProcMaps) error {
	// 7ff8be1a5000-7ff8be1c0000 r-xp 00000000 fd:00 570150                     /usr/lib64/libpthread-2.28.so
	var (
		err                error
		perms              string
		devMajor, devMinor uint64
	)

	fields := strings.Fields(line)
	field_count := len(fields)
	if field_count != __miniProcMapsEntryDefaultFieldCount {
		return nil
	}

	pmm := new(ProcMapsModule)
	pmm.Type = UNKNOWN
	pmm.RootFS = fmt.Sprintf("/proc/%d/root", pss.Pid)

	fmt.Sscanf(line, "%x-%x %s %x %x:%x %d %s", &pmm.StartAddr, &pmm.EndAddr, &perms,
		&pmm.Offset, &devMajor, &devMinor, &pmm.Inode, &pmm.Pathname)

	if len(pmm.Pathname) == 0 ||
		strings.Contains(pmm.Pathname, "[vdso]") ||
		strings.Contains(pmm.Pathname, "[vsyscall]") {
		return nil
	}

	permBytes := String2Bytes(perms)
	if permBytes[2] != 'x' {
		return nil
	}

	for _, ch := range perms {
		switch ch {
		case 'r':
			pmm.Perms.Readable = true
		case 'w':
			pmm.Perms.Writable = true
		case 'x':
			pmm.Perms.Executable = true
		case 's':
			pmm.Perms.Shared = true
		case 'p':
			pmm.Perms.Private = true
		}
	}

	pmm.Dev = unix.Mkdev(uint32(devMajor), uint32(devMinor))

	if err = pmm.loadProcModule(); err != nil {
		if errors.Is(err, ErrProcModuleNotSupport) || errors.Is(err, ErrProcModuleHasNoSymbols) {
			// 不加入，忽略，继续
			pmm = nil
			return nil
		}
		return errors.Wrapf(err, "load module:'%s' failed.", pmm.Pathname)
	}

	pss.ModuleList = append(pss.ModuleList, pmm)

	return nil
}

func NewProcSyms(pid int) (*ProcMaps, error) {
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

	pss := new(ProcMaps)
	pss.Pid = pid
	pss.InodeID = stat.Ino

	scanner := bufio.NewScanner(procMapsFile)

	for scanner.Scan() {
		// maps 每一行的信息
		text := scanner.Text()
		err := parseProcMapsEntry(text, pss)
		if err != nil {
			return nil, errors.Wrapf(err, "parse text:'%s' failed", text)
		}
	}

	return pss, nil
}

// ResolvePC 根据程序计数器 (PC) 解析符号信息
// 如果 ProcMaps 中的模块为空，则返回错误
// 如果 PC 在模块的地址范围内，则返回符号名称、偏移量和路径名
// 如果模块类型为 SO，则返回符号名称、偏移量和路径名
// 如果模块类型为 EXEC，则返回符号名称、偏移量和路径名
// 如果在解析过程中出现错误，则返回错误
func (pss *ProcMaps) ResolvePC(pc uint64) (string, uint32, string, error) {
	var (
		st   SymbolTable
		rs   *ResolveSymbol
		err  error
		addr uint64
		elfF *elf.File
	)
	if len(pss.ModuleList) == 0 {
		return "", 0, "", errors.New("proc modules is empty")
	}

	for _, pmm := range pss.ModuleList {
		if pc >= pmm.StartAddr && pc <= pmm.EndAddr {
			// 根据 module 类型计算地址
			if pmm.Type == SO {
				addr = pc - pmm.StartAddr
			} else if pmm.Type == EXEC {
				addr = pc
			}
			// 根据 buildID 找到 module 的 SymbolTable
			st, err = getModuleSymbolTbl(pmm.BuildID)
			if st == nil && err != nil {
				// 如果符号表不存在，创建符号表
				if elfF, err = pmm.open(); err == nil {
					st, err = createModuleSymbolTbl(pmm.BuildID, pmm.Pathname, pmm.RootFS, elfF)
					elfF.Close()
				}
			}

			if st != nil && err == nil {
				if rs, err = st.Resolve(addr); err == nil {
					// 解析 ok
					return rs.Name, rs.Offset, pmm.Pathname, nil
				} else {
					// 解析失败
					return "", 0, "", err
				}
			}
		}
	}
	return "", 0, "", errors.Errorf("pc:0x%x is outside the valid ranges in /proc/%d/maps", pc, pss.Pid)
}

// GetModules 返回进程符号表中的所有模块。
func (pss *ProcMaps) Modules() []*ProcMapsModule {
	return pss.ModuleList
}
