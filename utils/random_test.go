/*
 * @Author: calm.wu
 * @Date: 2019-04-22 15:17:31
 * @Last Modified by: calm.wu
 * @Last Modified time: 2019-04-22 15:18:29
 */

package utils

import "testing"

func TestGenerateRandomID(t *testing.T) {
	id := GenerateRandomID()
	t.Log(id)
}
