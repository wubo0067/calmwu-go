/*
 * @Author: calmwu
 * @Date: 2020-06-07 21:23:36
 * @Last Modified by: calmwu
 * @Last Modified time: 2020-06-07 21:25:47
 */

package utils

import (
	crand "crypto/rand"
	"io"
	"math/rand"
)

func GeneratePassword(minLength int, maxLength int) string {
	var chars = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*()-_=+,.?/:;{}[]`~")

	var length = rand.Intn(maxLength-minLength) + minLength

	newPassword := make([]byte, length)
	randomData := make([]byte, length+(length/4))
	clen := byte(len(chars))
	maxrb := byte(256 - (256 % len(chars)))
	i := 0
	for {
		if _, err := io.ReadFull(crand.Reader, randomData); err != nil {
			panic(err)
		}
		for _, c := range randomData {
			if c >= maxrb {
				continue
			}
			newPassword[i] = chars[c%clen]
			i++
			if i == length {
				return Bytes2String(newPassword)
			}
		}
	}
}
