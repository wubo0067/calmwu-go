/*
 * @Author: calmwu
 * @Date: 2017-11-29 19:24:05
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-11-30 16:42:09
 * @Comment:
 */

package protocol

import (
	"compress/zlib"
	"crypto/cipher"
	"encoding/binary"
	"fmt"
	"hash/crc32"

	"github.com/wubo0067/calmwu-go/utils"
)

// 压缩--->CRC校验--->加密
func FlodPayloadData(rawData []byte, cipherBlock cipher.Block) ([]byte, error) {
	if len(rawData) == 0 {
		return []byte{}, nil
	}

	// 压缩
	cData, err := utils.ZlibCompress(rawData, zlib.BestCompression)
	if err != nil {
		return nil, err
	}
	cDataLen := uint32(len(cData))
	//fmt.Printf("cData len:%d\n", cDataLen)

	// crc32
	crcBuf := make([]byte, 8, 8+cDataLen)
	// 对压缩数据计算crc值
	crc := crc32.ChecksumIEEE(cData)
	//fmt.Printf("FlodTrasnferData crc:%x\n", crc)
	// 将crc值写入前4个字节
	binary.BigEndian.PutUint32(crcBuf[:4], crc)
	crcBuf = append(crcBuf, cData...)
	// 写入压缩数据长度，解密时需要
	binary.BigEndian.PutUint32(crcBuf[4:8], uint32(len(crcBuf)))
	//fmt.Printf("crcBuf len:%d\n", len(crcBuf))

	// 加密
	cipherBuf, err := utils.EncryptPlainText(cipherBlock, crcBuf)
	if err != nil {
		return nil, err
	}
	//fmt.Printf("cipherBuf len:%d\n", len(cipherBuf))
	return cipherBuf, nil
}

// 解密--->crc校验--->解压缩
func UnFlodPayloadData(cipherData []byte, cipherBlock cipher.Block) ([]byte, error) {
	if len(cipherData) == 0 {
		return []byte{}, nil
	}

	//fmt.Printf("cipherData len:%d\n", len(cipherData))
	// 解密
	crcBuf, err := utils.DecryptCipherText(cipherBlock, cipherData)
	if err != nil {
		return nil, err
	}
	//fmt.Printf("crcBuf len:%d\n", len(crcBuf))

	// crc校验
	expectedCrc := binary.BigEndian.Uint32(crcBuf[:4])
	crcBufLen := binary.BigEndian.Uint32(crcBuf[4:8])
	//fmt.Printf("crcBufLen:%d\n", crcBufLen)
	crc := crc32.ChecksumIEEE(crcBuf[8:crcBufLen])
	if crc != expectedCrc {
		return nil, fmt.Errorf("Got invalid checksum for payload: %x, %x", crc, expectedCrc)
	}

	// 解压
	rawText, err := utils.ZlibDCompress(crcBuf[8:crcBufLen])
	if err != nil {
		return nil, err
	}
	return rawText, nil
}
