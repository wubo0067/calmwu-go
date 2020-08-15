/*
 * @Author: calmwu
 * @Date: 2017-09-19 16:27:11
 * @Last Modified by: calmwu
 * @Last Modified time: 2020-08-15 16:17:29
 * @Comment:
 */

package utils

import (
	"github.com/spaolacci/murmur3"
)

const (
	HashSeed = 0x2a
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

/*From the paper "A Fast, Minimal Memory, Consistent Hash Algorithm" by John Lamping, Eric Veach (2014).
http://arxiv.org/abs/1406.2294
*/

// JumpHash Hash consistently chooses a hash bucket number in the range [0, numBuckets) for the given key. numBuckets must >= 1.
func JumpHash(key uint64, numBuckets int) int32 {
	var b int64 = -1
	var j int64

	if numBuckets <= 0 {
		numBuckets = 1
	}

	for j < int64(numBuckets) {
		b = j
		key = key*2862933555777941757 + 1
		j = int64(float64(b+1) * (float64(int64(1)<<31) / float64((key>>33)+1)))
	}

	return int32(b)
}
