/*
 * @Author: CALM.WU
 * @Date: 2023-10-26 14:32:43
 * @Last Modified by: CALM.WU
 * @Last Modified time: 2023-10-26 14:46:33
 */

package utils

import (
	"debug/elf"
	"debug/gosym"
	"sort"
	"sync"

	"github.com/golang/glog"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/pkg/errors"
)

var (
	ErrGSTTextSectionEmpty  = errors.New(".text section is empty")
	ErrGSTGoPCLNTabNotExist = errors.New("no .gopclntab section")
	ErrGSTGoPCLNTabEmpty    = errors.New(".gopclntab section is empty")
	ErrGSTGoSymTabNotExist  = errors.New("no .gosymtab section")
	ErrGSTGoTooOld          = errors.New("gosymtab: go sym tab too old")
	ErrGSTGoParseFailed     = errors.New("gosymtab: go sym tab parse failed")
	ErrGSTGoFailed          = errors.New("gosymtab: go sym tab failed")
	ErrGSTGoOOB             = errors.New("go table oob")
	ErrGSTGoSymbolsNotFound = errors.New("gosymtab: no go symbols found")
)

type ModuleSymbol struct {
	Name    string
	Address uint64 // readelf -s ./xxx second column
}

type ResolveSymbol struct {
	Name   string
	Offset uint32
}

type SymbolTable interface {
	Resolve(addr uint64) (*ResolveSymbol, error) // 解析地址
	ModuleName() string                          //
	BuildID() string                             //
	Count() int                                  //
	Symbols() []*ModuleSymbol                    //
}

type ModuleSymbolTbl struct {
	moduleName string
	buildID    string // proc maps module buildid from readelf -n
}

func (mst *ModuleSymbolTbl) ModuleName() string {
	return mst.moduleName
}

func (mst *ModuleSymbolTbl) BuildID() string {
	return mst.buildID
}

type NativeModuleSymbolTbl struct {
	ModuleSymbolTbl
	symbolTable []*ModuleSymbol // 排序的符号列表
	symbolCount int             //
}

func (nmst *NativeModuleSymbolTbl) Resolve(addr uint64) (*ResolveSymbol, error) {
	// 二分查找
	index := sort.Search(nmst.symbolCount, func(i int) bool {
		return nmst.symbolTable[i].Address > addr
	})

	// addr 小于所有 symbol 的最小地址
	if index == 0 {
		return nil, errors.Errorf("addr:0x%x not in module:'%s', buildID:'%s' symbol table{0x%x---0x%x}",
			addr, nmst.moduleName, nmst.buildID, nmst.symbolTable[0].Address, nmst.symbolTable[nmst.symbolCount-1].Address)
	}

	// 找到了
	ms := nmst.symbolTable[index-1]
	return &ResolveSymbol{Name: ms.Name, Offset: uint32(addr - ms.Address)}, nil
}

func (nmst *NativeModuleSymbolTbl) GenerateTbl(elfF *elf.File) error {
	// from .text section read symbol and pc
	symbols, err := elfF.Symbols()
	if err != nil && !errors.Is(err, elf.ErrNoSymbols) {
		return errors.Wrapf(err, "read module:'%s' SYMTAB.", nmst.moduleName)
	}

	dynSymbols, err := elfF.DynamicSymbols()
	if err != nil && !errors.Is(err, elf.ErrNoSymbols) {
		return errors.Wrapf(err, "read module:'%s' DYNSYM.", nmst.moduleName)
	}

	pfnAddSymbol := func(m *NativeModuleSymbolTbl, syms []elf.Symbol) {
		for _, sym := range syms {
			if sym.Value != 0 && sym.Info&0xf == byte(elf.STT_FUNC) {
				ps := new(ModuleSymbol)
				ps.Name = sym.Name
				ps.Address = sym.Value
				m.symbolTable = append(m.symbolTable, ps)
				m.symbolCount += 1
			}
		}
	}

	pfnAddSymbol(nmst, symbols)
	pfnAddSymbol(nmst, dynSymbols)

	symbols = nil
	dynSymbols = nil

	if nmst.symbolCount == 0 {
		return ErrProcModuleHasNoSymbols
	}

	//fmt.Printf("-------------module:'%s' SymCount:%d.\n", psm.Pathname, psm.SymCount)

	// 按地址排序，地址相同按名字排序
	sort.Slice(nmst.symbolTable, func(i, j int) bool {
		if nmst.symbolTable[i].Address == nmst.symbolTable[j].Address {
			return nmst.symbolTable[i].Name < nmst.symbolTable[j].Name
		}
		return nmst.symbolTable[i].Address < nmst.symbolTable[j].Address
	})

	return nil
}

func (nmst *NativeModuleSymbolTbl) Count() int {
	return len(nmst.symbolTable)
}

func (nmst *NativeModuleSymbolTbl) Symbols() []*ModuleSymbol {
	return nmst.symbolTable
}

type GoModuleSymbolTbl struct {
	ModuleSymbolTbl
	symIndex *gosym.Table
}

func (gomst *GoModuleSymbolTbl) Resolve(addr uint64) (*ResolveSymbol, error) {
	symFunc := gomst.symIndex.PCToFunc(addr)
	if symFunc != nil {
		return &ResolveSymbol{Name: symFunc.Name, Offset: uint32(symFunc.End - addr)}, nil
	}
	return nil, ErrGSTGoSymbolsNotFound
}

func (gomst *GoModuleSymbolTbl) GenerateTbl(goSymTabSec *elf.Section, elfF *elf.File) error {
	var (
		err           error
		gosymtabData  []byte
		gopclntabData []byte
	)

	if sec := elfF.Section(".gopclntab"); sec != nil {
		if sec.Type == elf.SHT_NOBITS {
			// 如果没有 meta 数据，返回
			return errors.Wrapf(err, ".gopclntab section has no bits", gomst.moduleName)
		}

		gopclntabData, err = sec.Data()
		if err != nil {
			return errors.Wrapf(err, "read %s gopclntab section.", gomst.moduleName)
		}
	}
	if len(gopclntabData) <= 0 {
		return ErrGSTGoPCLNTabEmpty
	}

	gosymtabData, _ = goSymTabSec.Data()
	// if err != nil {
	// 	return errors.Wrapf(err, "read %s gosymtab section.", gomst.moduleName)
	// }

	lineTab := gosym.NewLineTable(gopclntabData, elfF.Section(".text").Addr)
	tab, err := gosym.NewTable(gosymtabData, lineTab)
	if err != nil {
		return errors.Wrapf(err, "build symtab or pclinetab for %s.", gomst.moduleName)
	}

	gomst.symIndex = tab
	gosymtabData = nil
	return nil
}

func (gomst *GoModuleSymbolTbl) Count() int {
	return len(gomst.symIndex.Funcs)
}

func (gomst *GoModuleSymbolTbl) Symbols() []*ModuleSymbol {
	return nil
}

type ModuleSymbolTblMgr struct {
	lc *lru.Cache[string, SymbolTable] // 管理所有 module 的符号表
}

var (
	__singleModuleSymbolTblMgr   *ModuleSymbolTblMgr
	__moduleSymbolTblMgrInitOnce sync.Once
	__nativeModuleSymbolTbl      SymbolTable = &NativeModuleSymbolTbl{}
	__goModuleSymbolTbl          SymbolTable = &GoModuleSymbolTbl{}
)

// InitModuleSymbolTblMgr initializes the module symbol table manager with the given capacity.
// It creates a new LRU cache with the specified capacity and sets up the module symbol table manager
// to use it. If the initialization has already been performed, this function does nothing.
// Returns an error if there was a problem creating the LRU cache.
func InitModuleSymbolTblMgr(capacity int) error {
	var err error

	__moduleSymbolTblMgrInitOnce.Do(func() {
		__singleModuleSymbolTblMgr = &ModuleSymbolTblMgr{}

		__singleModuleSymbolTblMgr.lc, err = lru.NewWithEvict[string, SymbolTable](capacity, func(k string, v SymbolTable) {
			// 做个类型转换，释放内存
			switch t := v.(type) {
			case *GoModuleSymbolTbl:
				if t != nil {
					glog.Warningf("GoModule symbol table:'%s', buildID:'%s' is evicted.",
						v.ModuleName(), k)
					t.symIndex = nil
					t = nil
				}
			case *NativeModuleSymbolTbl:
				glog.Warningf("evicted NativeModule symbol table:'%s', buildID:'%s'", v.ModuleName(), k)
				t.symbolTable = nil
				t = nil
			}
		})

		if err != nil {
			err = errors.Wrap(err, "new module symbol table lru cache failed.")
		}
	})
	return err
}

// getModuleSymbolTbl returns the symbol table for a given build ID.
// If the symbol table is found, it is returned along with a nil error.
// If the symbol table is not found, a nil table and an error are returned.
func getModuleSymbolTbl(buildID string) (SymbolTable, error) {
	var (
		st SymbolTable
		ok bool
	)

	if __singleModuleSymbolTblMgr != nil {
		st, ok = __singleModuleSymbolTblMgr.lc.Get(buildID)
		if ok && st != nil {
			return st, nil
		}
	}
	return nil, errors.Errorf("symbol table not found by buildID:'%s'", buildID)
}

// createModuleSymbolTbl creates a symbol table for a given module.
// It takes in the buildID, moduleName, appRootFS, and elfF as parameters.
// It returns a SymbolTable and an error.
func createModuleSymbolTbl(buildID string, moduleName string, appRootFS string, elfF *elf.File) (SymbolTable, error) {
	var (
		st        SymbolTable
		err       error
		elfDebugF *elf.File
	)

	// 判断是否是 golang module
	if sec := elfF.Section(".gosymtab"); sec == nil {
		// is native module
		nmst := new(NativeModuleSymbolTbl)
		nmst.buildID = buildID
		nmst.moduleName = moduleName
		// 通过 buildID 查找对应的 debug 文件
		debugFilePath := findDebugFile(buildID, appRootFS, moduleName, elfF)
		if debugFilePath != "" {
			// 如果 debug 文件存在，打开
			glog.Infof("found debug file:'%s' for module:'%s'", debugFilePath, moduleName)
			elfDebugF, err = elf.Open(debugFilePath)
			if err == nil {
				defer elfDebugF.Close()
				err = nmst.GenerateTbl(elfDebugF)
			}
		} else {
			err = nmst.GenerateTbl(elfF)
		}
		if err == nil {
			st = nmst
			__singleModuleSymbolTblMgr.lc.Add(buildID, st)
			glog.Infof("native module:'%s' buildID:'%s' create symbol table ok. current have %d modules in LRUCache",
				moduleName, buildID, __singleModuleSymbolTblMgr.lc.Len())
		}
	} else {
		// is golang module
		gomst := new(GoModuleSymbolTbl)
		gomst.buildID = buildID
		gomst.moduleName = moduleName
		if err = gomst.GenerateTbl(sec, elfF); err == nil {
			st = gomst
			__singleModuleSymbolTblMgr.lc.Add(buildID, st)
			glog.Infof("go module:'%s' buildID:'%s' create symbol table ok. current have %d modules in LRUCache",
				moduleName, buildID, __singleModuleSymbolTblMgr.lc.Len())
		}
	}
	if err != nil {
		glog.Errorf("create module:'%s' buildID:'%s' symbol table failed. err:%s", moduleName, buildID, err.Error())
	}
	return st, err
}

// DeleteModuleSymbolTbl deletes the module symbol table for the given build ID.
func DeleteModuleSymbolTbl(buildID string) {
	if __singleModuleSymbolTblMgr != nil {
		// Remove is thread safe
		if __singleModuleSymbolTblMgr.lc.Remove(buildID) {
			glog.Infof("delete module symbol table by buildID:'%s'. current have %d modules in LRUCache",
				buildID, __singleModuleSymbolTblMgr.lc.Len())
		}
	}
}
