/*
 * @Author: CALM.WU
 * @Date: 2023-02-28 11:10:47
 * @Last Modified by: CALM.WU
 * @Last Modified time: 2023-02-28 11:15:34
 */

package utils

import (
	"encoding/binary"
	"unsafe"
)

var byteOrder binary.ByteOrder

func init() {
	byteOrder = initHostByteOrder()
}

func initHostByteOrder() binary.ByteOrder {
	buf := [2]byte{}
	*(*uint16)(unsafe.Pointer(&buf[0])) = uint16(0x0102)

	// 高位在低地址
	if buf[0] == 0x01 {
		return binary.BigEndian
	}
	return binary.LittleEndian
}

// getHostByteOrder returns the host's native binary.ByteOrder.
func GetHostByteOrder() binary.ByteOrder {
	return byteOrder
}
