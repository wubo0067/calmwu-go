/*
 * @Author: CALM.WU
 * @Date: 2023-01-01 11:57:33
 * @Last Modified by: CALM.WU
 * @Last Modified time: 2023-01-01 12:44:02
 */

package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

const (
	_kallsyms = "/proc/kallsyms"
)

var (
	_lock       sync.RWMutex
	_ksym_cache map[string]string
)

// It reads the /proc/kallsyms file and stores the symbol name and address in a map
func LoadKallSyms() error {
	_lock.Lock()
	defer _lock.Unlock()

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

		_ksym_cache[ar[0]] = ar[2]
	}
	return nil
}

// > FindKsym() returns the kernel symbol name for a given address
func FindKsym(addr string) (string, error) {
	if len(_ksym_cache) == 0 {
		return "", fmt.Errorf("ksym cache is empty")
	}

	_lock.RLock()
	defer _lock.RUnlock()

	if sym, ok := _ksym_cache[addr]; ok {
		return sym, nil
	}
	return "", fmt.Errorf("kernel not found ksym for addr:%s", addr)
}
