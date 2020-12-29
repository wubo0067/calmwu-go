/*
 * @Author: calmwu
 * @Date: 2017-09-19 16:27:11
 * @Last Modified by: calmwu
 * @Last Modified time: 2017-09-23 11:45:05
 * @Comment:
 */
package base

import "github.com/spaolacci/murmur3"

const (
	HashSeed = 0
)

func HashStr2Uint32(s string) uint32 {
	hMurmur32 := murmur3.New32WithSeed(HashSeed)
	hMurmur32.Write([]byte(s))
	return hMurmur32.Sum32()
}

func HashStr2Uint64(s string) uint64 {
	hMurmur64 := murmur3.New64WithSeed(HashSeed)
	hMurmur64.Write([]byte(s))
	return hMurmur64.Sum64()
}
