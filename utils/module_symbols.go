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
	ErrGSTTextSectionEmpty  = errors.New("empty .text section")
	ErrGSTGoPCLNTabNotExist = errors.New("no .gopclntab section")
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

	// addr小于所有symbol的最小地址
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

func (gomst *GoModuleSymbolTbl) GenerateTbl(goSymTabSection *elf.Section, elfF *elf.File) error {
	goSymTabData, err := goSymTabSection.Data()
	if err != nil {
		return errors.Wrapf(err, "read %s gosymtab section data.", gomst.moduleName)
	}

	pclndat, err := elfF.Section(".gopclntab").Data()
	if err != nil {
		return errors.Wrapf(err, "read %s gopclntab section data.", gomst.moduleName)
	}

	pcln := gosym.NewLineTable(pclndat, elfF.Section(".text").Addr)
	tab, err := gosym.NewTable(goSymTabData, pcln)
	if err != nil {
		return errors.Wrapf(err, "parsing %s gosymtab.", gomst.moduleName)
	}

	gomst.symIndex = tab
	return nil
}

func (gomst *GoModuleSymbolTbl) Count() int {
	return len(gomst.symIndex.Funcs)
}

func (gomst *GoModuleSymbolTbl) Symbols() []*ModuleSymbol {
	return nil
}

type ModuleSymbolTblMgr struct {
	lc   *lru.Cache[string, SymbolTable] // 管理所有module的符号表
	lock sync.Mutex
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
			glog.Warningf("evicted module symbol table:'%s', buildID:'%s'", v.ModuleName(), k)
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

		__singleModuleSymbolTblMgr.lock.Lock()
		defer __singleModuleSymbolTblMgr.lock.Unlock()

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

	// 判断是否是golang module
	if goSymTab := elfF.Section(".gosymtab"); goSymTab == nil {
		// is native module
		nmst := new(NativeModuleSymbolTbl)
		nmst.buildID = buildID
		nmst.moduleName = moduleName
		// 通过buildID查找对应的debug文件
		debugFilePath := findDebugFile(buildID, appRootFS, moduleName, elfF)
		if debugFilePath != "" {
			// 如果debug文件存在，打开
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
			glog.Infof("native module:'%s' buildID:'%s' create symbol table ok.", moduleName, buildID)
		}
	} else {
		// is golang module
		gomst := new(GoModuleSymbolTbl)
		gomst.buildID = buildID
		gomst.moduleName = moduleName
		if err = gomst.GenerateTbl(goSymTab, elfF); err == nil {
			st = gomst
			__singleModuleSymbolTblMgr.lc.Add(buildID, st)
			glog.Infof("go module:'%s' buildID:'%s' create symbol table ok.", moduleName, buildID)
		}
	}

	return st, err
}

// deleteModuleSymbolTbl deletes the module symbol table for the given build ID.
func deleteModuleSymbolTbl(buildID string) {
	if __singleModuleSymbolTblMgr != nil {

		__singleModuleSymbolTblMgr.lock.Lock()
		defer __singleModuleSymbolTblMgr.lock.Unlock()

		__singleModuleSymbolTblMgr.lc.Remove(buildID)
	}
}
