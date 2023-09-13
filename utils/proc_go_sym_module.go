/*
 * @Author: CALM.WU
 * @Date: 2023-09-12 14:18:45
 * @Last Modified by: CALM.WU
 * @Last Modified time: 2023-09-12 18:35:10
 */

/*
 .gopclntab（Go Program Counter Look-Up Table）是 Go 语言中的一个特殊的 ELF 节（section），它在可执行文件中存储了有关程序计数器（Program Counter，PC）和源代码行号之间映射关系的信息。
 这个节的作用非常重要，它主要用于以下几个目的：
调试信息： .gopclntab 节包含了程序计数器（PC）到源代码文件和行号的映射信息。这使得调试器能够在运行时将机器码指令映射回源代码行，从而帮助开发人员进行代码调试。
堆栈跟踪： 当发生运行时错误或异常时，.gopclntab 节的信息可用于生成堆栈跟踪，以便在错误发生时能够识别哪个源代码行导致了问题。
性能分析： 这个节中的信息还用于性能分析工具，如 go tool pprof，以帮助分析程序的性能瓶颈和函数调用图。
反射： .gopclntab 节的信息对于 Go 的反射功能也是重要的，因为它允许反射库了解函数的签名和参数类型等信息。
.gopclntab 节的存在使得 Go 语言在运行时能够动态地获取有关源代码位置的信息，这对于调试和性能分析是非常有价值的。它是 Go 编译器和运行时系统的重要组成部分，用于提供丰富的运行时信息。
*/

package utils

import (
	"debug/elf"
	"debug/gosym"
	"os"

	"github.com/pkg/errors"
)

type GoSymTable struct {
	symIndex *gosym.Table
}

var (
	ErrGSTTextSectionEmpty  = errors.New("empty .text section")
	ErrGSTGoPCLNTabNotExist = errors.New("no .gopclntab section")
	ErrGSTGoSymTabNotExist  = errors.New("no .gosymtab section")
	ErrGSTGoTooOld          = errors.New("gosymtab: go sym tab too old")
	ErrGSTGoParseFailed     = errors.New("gosymtab: go sym tab parse failed")
	ErrGSTGoFailed          = errors.New("gosymtab: go sym tab failed")
	ErrGSTGoOOB             = errors.New("go table oob")
	ErrGSTGoSymbolsNotFound = errors.New("gosymtab: no go symbols found")
)

func (psm *ProcSymsModule) loadProcGoModule(pid int) error {
	var (
		f    *os.File
		elfF *elf.File
		err  error
	)
	if f, elfF, err = psm.open(pid); err != nil {
		return errors.Wrapf(err, "psm open:'/proc/%d/root%s'.", pid, psm.Pathname)
	}
	defer f.Close()

	switch elfF.Type {
	case elf.ET_EXEC:
		psm.Type = EXEC
	case elf.ET_DYN:
		psm.Type = SO
	default:
		return ErrProcModuleNotSupport
	}

	s := elfF.Section(".gosymtab")
	if s == nil {
		return ErrGSTGoSymTabNotExist
	}

	symdat, err := s.Data()
	if err != nil {
		return errors.Wrapf(err, "read %s gosymtab", psm.Pathname)
	}
	pclndat, err := elfF.Section(".gopclntab").Data()
	if err != nil {
		return errors.Wrapf(err, "read %s gopclntab", psm.Pathname)
	}

	pcln := gosym.NewLineTable(pclndat, elfF.Section(".text").Addr)
	tab, err := gosym.NewTable(symdat, pcln)
	if err != nil {
		return errors.Wrapf(err, "parsing %s gosymtab", psm.Pathname)
	}

	psm.goSymTable = &GoSymTable{
		symIndex: tab,
	}
	//fmt.Printf("loadProcGoModule:'%s' success.\n", psm.Pathname)
	return nil
}

func (gst *GoSymTable) __resolveGoPC(pc uint64) (string, uint32, error) {
	symFunc := gst.symIndex.PCToFunc(pc)
	if symFunc != nil {
		return symFunc.Name, uint32(symFunc.End - pc), nil
	}
	return "", 0, ErrGSTGoSymbolsNotFound
}
