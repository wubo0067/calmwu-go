/*
 * @Author: calm.wu
 * @Date: 2020-01-16 11:15:48
 * @Last Modified by: calm.wu
 * @Last Modified time: 2020-01-16 11:19:15
 */

 // Package utils for calmwu golang tools
package utils

// MergeMap map合并
func MergeMap(a, b map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = mergeMaps(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}

