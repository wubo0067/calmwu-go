/*
 * @Author: calmwu
 * @Date: 2017-11-30 11:15:08
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-12-08 16:54:33
 */

package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"fmt"
	"math/big"

	"github.com/monnand/dhkx"
)

// https://www.cnblogs.com/si812cn/archive/2009/11/26/1611272.html
// g是p的原根Diffie-Hellman密钥交换算法的有效性依赖于计算离散对数的难度。简言之，可以如下定义离散对数：首先定义一个素数p的原根，为其各次幂产生从1 到p-1的所有整数根，也就是说，如果a是素数p的一个原根，那么数值
//a mod p, a2 mod p, ..., ap-1 mod p 是各不相同的整数，并且以某种排列方式组成了从1到p-1的所有整数。
const (
	KeyBytesSize = 32
	pValue       = "8b79f180cbd3f282de92e8b8f2d092674ffda61f01ed961f8ef04a1b7a3709ff748c2abf6226cf0c4538e48838193da456e92ee530ef7aa703e741585e475b26cd64fa97819181cef27de2449cd385c49c9b030f89873b5b7eaf063a788f00db3cb670c73846bc4f76af062d672bde8f29806b81548411ab48b99aebfd9c2d09"
	gValue       = "029843c81d0ea285c41a49b1a2f8e11a56a4b39040dfbc5ec040150c16f72f874152f9c44c659d86f7717b2425b62597e9a453b13da327a31cde2cced600915252d30262d1e54f4f864ace0e484f98abdbb37ebb0ba4106af5f0935b744677fa2f7f3826dcef3a1586956105ebea805d871f34c46c25bc30fc66b2db26cb0a93"
)

var (
	dhGroup *dhkx.DHGroup = nil
)

// GenerateDHKey 算出来的ka发送给对方
func GenerateDHKey() (*dhkx.DHKey, error) {
	once.Do(func() {
		p, _ := new(big.Int).SetString(pValue, 16)
		g, _ := new(big.Int).SetString(gValue, 16)
		dhGroup = dhkx.CreateGroup(p, g)
	})

	return dhGroup.GeneratePrivateKey(nil)
}

// GenerateEncryptionKey 根据对方返回的kb计算加密密钥
func GenerateEncryptionKey(pub []byte, privateKey *dhkx.DHKey) ([]byte, error) {
	//s := []byte(*kb)
	peerPubKey := dhkx.NewPublicKey(pub)
	// 计算密钥
	encryptionKey, err := dhGroup.ComputeKey(peerPubKey, privateKey)
	if err != nil {
		return nil, err
	}

	return encryptionKey.Bytes(), nil
}

func NewCipherBlock(encryptionKey []byte) (cipher.Block, error) {
	cipherBlock, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("New cipherBlock failed. err[%s]", err.Error())
	}
	return cipherBlock, nil
}

// EncryptPlainText 传入明文，返回密文
func EncryptPlainText(cipherBlock cipher.Block, plainText []byte) ([]byte, error) {
	if cipherBlock == nil {
		return nil, errors.New("cipher block object is nil")
	}

	plainTextLen := len(plainText)
	plainTextBuff := plainText

	// 明文数据必须是blocksize的倍数，aes.blocksize=16
	if plainTextLen%aes.BlockSize != 0 {
		// 重新生成
		plainTextLen = (((plainTextLen) + ((aes.BlockSize) - 1)) & ^((aes.BlockSize) - 1))
		plainTextBuff = make([]byte, plainTextLen)
		copy(plainTextBuff, plainText)
	}

	// 生成加密缓冲区
	cipherTextLen := aes.BlockSize + plainTextLen
	cipherTextBuff := make([]byte, cipherTextLen)

	//GLog.Debug("plainTextLen[%d] cipherTextLen[%d]", plainTextLen, cipherTextLen)

	iv := cipherTextBuff[:aes.BlockSize]
	mode := cipher.NewCBCEncrypter(cipherBlock, iv)
	mode.CryptBlocks(cipherTextBuff[aes.BlockSize:], plainTextBuff)
	return cipherTextBuff, nil
}

// DecryptCipherText in-place模式
func DecryptCipherText(cipherBlock cipher.Block, cipherText []byte) ([]byte, error) {
	if cipherBlock == nil {
		return nil, errors.New("cipher block object is nil!")
	}

	cipherTextLen := len(cipherText)

	if cipherTextLen < aes.BlockSize {
		return nil, fmt.Errorf("cipherText too short! cipherTextLen[%d]", cipherTextLen)
	}

	iv := cipherText[:aes.BlockSize]
	//g_log.Debug("ciphertext len[%d]", len(ciphertext))
	cipherText = cipherText[aes.BlockSize:]
	//g_log.Debug("ciphertext len[%d]", len(ciphertext))
	mode := cipher.NewCBCDecrypter(cipherBlock, iv)
	mode.CryptBlocks(cipherText, cipherText)
	//g_log.Debug("ciphertext len[%d]", len(ciphertext))
	return cipherText, nil
}
