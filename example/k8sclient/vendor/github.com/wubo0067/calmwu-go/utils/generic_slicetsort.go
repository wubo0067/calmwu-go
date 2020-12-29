/*
 * @Author: calmwu
 * @Date: 2020-03-07 15:55:24
 * @Last Modified by: calmwu
 * @Last Modified time: 2020-03-07 16:35:23
 */

package utils

import (
	"sort"

	"github.com/cheekybits/genny/generic"
)

// SortedSliceObjType 被排序的对象类型
type SortedSliceObjType generic.Type

// SortedSliceName 被排序的slice名
type SortedSliceName generic.Type

// SortedSliceWrap
type SortedSliceNameWrap struct {
	Ssot        []*SortedSliceObjType
	CompareFunc func(left, right *SortedSliceObjType) bool
}

// Len 长度
func (ssw SortedSliceNameWrap) Len() int { return len(ssw.Ssot) }

// Swap 交换
func (ssw SortedSliceNameWrap) Swap(i, j int) { ssw.Ssot[i], ssw.Ssot[j] = ssw.Ssot[j], ssw.Ssot[i] }

// Less 比较
func (ssw SortedSliceNameWrap) Less(i, j int) bool {
	return ssw.CompareFunc(ssw.Ssot[i], ssw.Ssot[j])
}

// Sort 排序
func (ssw SortedSliceNameWrap) Sort() {
	sort.Sort(ssw)
}
