/*
 * @Author: CALM.WU
 * @Date: 2021-02-05 21:56:22
 * @Last Modified by:   CALM.WU
 * @Last Modified time: 2021-02-05 21:56:22
 */

package utils

import (
	"sync"
	"testing"

	"github.com/pkg/errors"
)

type (
	kvT struct {
		key string
		val int
	}
)

var (
	_cache = map[string]int{}
)

func TestBarrierCalls_1(t *testing.T) {
	fn := func(args interface{}) (interface{}, error) {
		if kv, ok := args.(*kvT); ok {
			_cache[kv.key] = kv.val
			return kv.val, nil
		}
		return nil, errors.New("args type is not kvT")
	}

	barrierCalls := NewBarrierCalls()
	iv, fresh, err := barrierCalls.Do("Hello", &kvT{
		key: "Hello",
		val: 99,
	}, fn)

	t.Logf("iv: %v, fresh: %v, err: %v", iv, fresh, err)

	iv, fresh, err = barrierCalls.Do("Hello", &kvT{
		key: "Hello",
		val: 888,
	}, fn)

	t.Logf("iv: %v, fresh: %v, err: %v", iv, fresh, err)
}

func TestBarrierCalls_2(t *testing.T) {
	fn := func(args interface{}) (interface{}, error) {
		if kv, ok := args.(*kvT); ok {
			_cache[kv.key] = kv.val
			//time.Sleep(time.Microsecond * 5)
			return kv.val, nil
		}
		return nil, errors.New("args type is not kvT")
	}

	barrierCalls := NewBarrierCalls()

	var wg sync.WaitGroup
	wg.Add(10)

	for i := 0; i < 10; i++ {
		go func(i int) {
			defer wg.Done()
			iv, fresh, err := barrierCalls.Do("Hello", &kvT{
				key: "Hello",
				val: 99 + i,
			}, fn)

			t.Logf("i: %d, iv: %v, fresh: %v, err: %v", i, iv, fresh, err)

			// 当fresh为false的时候，说明没有执行函数，而是直接使用的缓存值
		}(i)
	}

	wg.Wait()
}
