/*
 * @Author: calmwu
 * @Date: 2018-01-04 18:58:55
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-01-04 19:01:59
 * @Comment:
 */

package main

import (
	"encoding/hex"
	"fmt"

	"github.com/wubo0067/calmwu-go/utils"
)

func main() {
	dhKey, _ := utils.GenerateDHKey()

	publicKey := dhKey.Bytes()
	fmt.Printf("publicKey len: %d\n", len(publicKey))
	fmt.Printf("publicKey: %s\n", hex.EncodeToString(publicKey[:]))

}
