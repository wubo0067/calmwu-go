/*
 * @Author: calmwu
 * @Date: 2017-11-30 10:19:27
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-11-30 11:03:49
 */

package utils

import (
	"bytes"
	"compress/zlib"
	"io"
)

func ZlibCompress(data []byte, compressLevel int) ([]byte, error) {
	if len(data) == 0 {
		return []byte{}, nil
	}

	compressBuf := bytes.NewBuffer(nil)
	compress, err := zlib.NewWriterLevel(compressBuf, compressLevel)
	if err != nil {
		ZLog.Errorf("zlib.NewWriterLevel failed! reason[%s]", err.Error())
		return nil, err
	}

	_, err = compress.Write(data)
	if err != nil {
		ZLog.Errorf("zlib.Writer failed! reason[%s]", err.Error())
		return nil, err
	}

	err = compress.Close()
	if err != nil {
		ZLog.Errorf("zlib.Close failed! reason[%s]", err.Error())
		return nil, err
	}

	return compressBuf.Bytes(), nil
}

func ZlibDCompress(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return []byte{}, nil
	}

	dcompressBuf := bytes.NewBuffer(nil)
	dcompress, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		ZLog.Errorf("zlib.NewWriterLevel failed! reason[%s]", err.Error())
		return nil, err
	}

	_, err = io.Copy(dcompressBuf, dcompress)
	if err != nil {
		return nil, err
	}

	err = dcompress.Close()
	if err != nil {
		return nil, err
	}

	return dcompressBuf.Bytes(), nil
}
