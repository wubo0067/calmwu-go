/*
 * @Author: CALM.WU
 * @Date: 2023-01-01 11:57:33
 * @Last Modified by: CALM.WU
 * @Last Modified time: 2023-01-05 16:06:14
 */

package utils

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

const (
	_kallsyms = "/proc/kallsyms"
)

type Ksym struct {
	address uint64
	name    string
}

var (
	_lock       sync.RWMutex
	_ksym_cache []*Ksym
)

// It reads the /proc/kallsyms file and stores the symbol name and address in a map
func LoadKallSyms() error {
	_lock.Lock()
	defer _lock.Unlock()

	if _ksym_cache != nil {
		return errors.Errorf("%s has been loaded", _kallsyms)
	}

	fd, err := os.Open(_kallsyms)
	if err != nil {
		return errors.Wrapf(err, "open %s failed", _kallsyms)
	}

	defer fd.Close()

	scaner := bufio.NewScanner(fd)
	for scaner.Scan() {
		line := scaner.Text()
		ar := strings.Split(line, " ")
		if len(ar) < 3 {
			continue
		}

		address, _ := strconv.ParseUint(ar[0], 16, 64)

		ksym := new(Ksym)
		ksym.address = address
		ksym.name = ar[2]

		_ksym_cache = append(_ksym_cache, ksym)
	}

	sort.Slice(_ksym_cache, func(i, j int) bool {
		return _ksym_cache[i].address < _ksym_cache[j].address
	})

	return nil
}

// It uses a binary search to find the symbol name for a given address
func FindKsym(addr uint64) (name string, offset uint32, err error) {
	if len(_ksym_cache) == 0 {
		err = fmt.Errorf("ksym cache is empty")
		return "", 0, err
	}

	_lock.RLock()
	defer _lock.RUnlock()

	//var result int64
	start := 0
	end := len(_ksym_cache)

	// fmt.Printf("+++start:%d, end:%d, count:%d\n", start, end, len(_ksym_cache))

	for start < end {
		mid := start + (end-start)/2
		//result = (int64)(addr - _ksym_cache[mid].address)

		// fmt.Printf("start:%d, mid:%d, end:%d, _ksym_cache[%d].address:%x\n",
		// 	start, mid, end, mid, _ksym_cache[mid].address)

		if addr < _ksym_cache[mid].address {
			end = mid
		} else if addr > _ksym_cache[mid].address {
			start = mid + 1
		} else {
			return _ksym_cache[mid].name, 0, nil
		}
	}

	// fmt.Printf("---start:%d, end:%d, count:%d\n", start, end, len(_ksym_cache))

	if start >= 1 && _ksym_cache[start-1].address < addr && addr < _ksym_cache[start].address {
		return _ksym_cache[start-1].name, (uint32)(addr - _ksym_cache[start-1].address), nil
	}

	err = fmt.Errorf("kernel not found ksym for addr:%x", addr)
	return "", 0, err
}
