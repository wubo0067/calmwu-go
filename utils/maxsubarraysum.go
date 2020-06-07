/*
 * @Author: calmwu
 * @Date: 2020-06-07 21:31:04
 * @Last Modified by: calmwu
 * @Last Modified time: 2020-06-07 21:31:55
 */

package utils

func Max(x int, y int) int {
	if x < y {
		return y
	}
	return x
}

// MaxSubarraySum 输出连续子数组最大合
func MaxSubarraySum(array []int) int {
	var currentMax int = 0
	var maxTillNow int = 0
	for _, v := range array {
		currentMax = Max(v, currentMax+v)
		maxTillNow = Max(maxTillNow, currentMax)
	}
	return maxTillNow
}
