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
	_lock        sync.RWMutex
	__ksym_cache []*Ksym
)

// It reads the /proc/kallsyms file and stores the symbol name and address in a map
func LoadKallSyms() error {
	_lock.Lock()
	defer _lock.Unlock()

	if __ksym_cache != nil {
		return errors.Errorf("%s has been loaded", _kallsyms)
	}

	fd, err := os.Open(_kallsyms)
	if err != nil {
		return errors.Wrapf(err, "open %s failed", _kallsyms)
	}

	defer fd.Close()

	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		line := scanner.Text()
		ar := strings.Split(line, " ")
		if len(ar) < 3 {
			continue
		}

		address, _ := strconv.ParseUint(ar[0], 16, 64)

		ksym := new(Ksym)
		ksym.address = address
		ksym.name = ar[2]

		__ksym_cache = append(__ksym_cache, ksym)
	}

	sort.Slice(__ksym_cache, func(i, j int) bool {
		return __ksym_cache[i].address < __ksym_cache[j].address
	})

	return nil
}

// It uses a binary search to find the symbol name for a given address
func FindKsym(addr uint64) (name string, offset uint32, err error) {
	if len(__ksym_cache) == 0 {
		err = fmt.Errorf("ksym cache is empty")
		return "", 0, err
	}

	_lock.RLock()
	defer _lock.RUnlock()

	// var result int64
	start := 0
	end := len(__ksym_cache)

	// fmt.Printf("+++start:%d, end:%d, count:%d\n", start, end, len(__ksym_cache))

	for start < end {
		mid := start + (end-start)/2
		// result = (int64)(addr - __ksym_cache[mid].address)

		// fmt.Printf("start:%d, mid:%d, end:%d, __ksym_cache[%d].address:%x\n",
		// 	start, mid, end, mid, __ksym_cache[mid].address)

		if addr < __ksym_cache[mid].address {
			end = mid
		} else if addr > __ksym_cache[mid].address {
			start = mid + 1
		} else {
			return __ksym_cache[mid].name, 0, nil
		}
	}

	// fmt.Printf("---start:%d, end:%d, count:%d\n", start, end, len(__ksym_cache))

	if start >= 1 && __ksym_cache[start-1].address < addr && addr < __ksym_cache[start].address {
		return __ksym_cache[start-1].name, (uint32)(addr - __ksym_cache[start-1].address), nil
	}

	err = fmt.Errorf("kernel not found ksym for addr:%x", addr)
	return "", 0, err
}
