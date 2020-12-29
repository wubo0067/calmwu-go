/**
 * @author LiZhengde
 * @email zhengde.li@gridice.com
 * @create date 2018-04-14 04:52:09
 * @modify date 2018-04-14 04:52:09
 * @desc 自定义错误
 */

package custom_errors

import (
	"errors"
	"fmt"
	"runtime/debug"
)

func New(format string, a ...interface{}) error {
	return errors.New(fmt.Sprintf("errros: %s\nStackTrace:\n%s", fmt.Sprintf(format, a), debug.Stack()))
}

func InvalidUin() error {
	return New("uin is invalid")
}

func NullPoint() error {
	return New("null point")
}
