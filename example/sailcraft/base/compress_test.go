/*
 * @Author: calmwu
 * @Date: 2017-11-30 10:46:01
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-11-30 10:56:41
 */

package base

import (
	"compress/zlib"
	"strings"
	"testing"
)

func TestZlibCompress(t *testing.T) {
	data := "Hello TestZlibCompress"

	compressData, err := ZlibCompress([]byte(data), zlib.BestCompression)
	if err != nil {
		t.Error(err.Error())
		return
	}

	dcompressData, err := ZlibDCompress(compressData)
	if err != nil {
		t.Error(err.Error())
		return
	}

	t.Logf("dcompressData:%s\n", string(dcompressData))

	if strings.Compare(data, string(dcompressData)) == 0 {
		t.Log("TestZlibCompress test successed")
	} else {
		t.Error("TestZlibCompress test failed")
	}
}
