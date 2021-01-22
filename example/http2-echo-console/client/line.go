/*
 * @Author: CALM.WU
 * @Date: 2021-01-22 10:26:06
 * @Last Modified by: CALM.WU
 * @Last Modified time: 2021-01-22 10:35:33
 */

package main

import (
	"bufio"
	"io"
)

type LineIterator struct {
	reader *bufio.Reader
}

func NewLineIterator(rd io.Reader) *LineIterator {
	lineIter := new(LineIterator)
	lineIter.reader = bufio.NewReader(rd)
	return lineIter
}

func (ln *LineIterator) Next() ([]byte, error) {
	var lineBytes []byte
	for {
		line, isPrefix, err := ln.reader.ReadLine()
		if err != nil {
			return nil, err
		}

		lineBytes = append(lineBytes, line...)
		if !isPrefix {
			break
		}
	}
	return lineBytes, nil
}
