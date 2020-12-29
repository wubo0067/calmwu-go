/*
 * @Author: calmwu
 * @Date: 2018-03-16 12:23:15
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-03-16 15:14:13
 * @Comment:
 */

package proto

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

// go test -v -run TestVIPInit
func TestVIPInit(t *testing.T) {
	var vip TBPlayerVIPInfoS
	now := time.Now()
	vip.Init(now)
	fmt.Printf("vip:%+v\n", vip)

	jsonData, _ := json.Marshal(&vip)
	fmt.Printf("jsonData:%s\n", string(jsonData))

	vip1 := new(TBPlayerVIPInfoS)
	json.Unmarshal(jsonData, vip1)
	fmt.Printf("vip1:%+v\n", vip1)
}

func TestVIP(t *testing.T) {
	var vip UserVIPType
	vip |= E_USER_VIP_NORMALMONTHLY
	fmt.Println("vip:", vip.String())

	vip |= E_USER_VIP_LUXURYMONTHLY
	fmt.Println("vip:", vip.String())

	if vip&E_USER_VIP_ALL != 0 {
		fmt.Println("super vip")
	}

	if vip&E_USER_VIP_LUXURYMONTHLY != 0 {
		fmt.Println("monthly vip")
	}

	vip ^= E_USER_VIP_NORMALMONTHLY
	fmt.Println("vip:", vip.String())

	vip ^= E_USER_VIP_LUXURYMONTHLY
	fmt.Println("vip:", vip.String())
}
