/*
 * @Author: calmwu
 * @Date: 2018-02-24 18:12:55
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-02-27 15:14:44
 * @Comment:
 */

package handler

import "testing"

func TestQueryFinanceUser(t *testing.T) {
	var Uin uint64 = 1
	userFinance, err := QueryFinanceUser(Uin)
	if err != nil {
		t.Error(err.Error())
		return
	}
	t.Logf("%v", userFinance)
}
