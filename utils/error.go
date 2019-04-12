/*
 * @Author: calmwu
 * @Date: 2018-01-27 16:59:37
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-05-17 13:45:49
 * @Comment:
 */

package utils

import (
	"errors"
	"fmt"

	"github.com/mozhata/merr"
)

func NewError(args ...interface{}) error {
	var err error
	var rawData []interface{}
	for _, arg := range args {
		switch arg.(type) {
		case error:
			err = arg.(error)
			ZLog.Errorf("error", err)
			continue
		default:
			rawData = append(rawData, arg)
		}
	}
	if err == nil {
		err = errors.New(fmt.Sprintf("%v", rawData))
	}
	return errors.New(fmt.Sprintf("%v [error => %s]", rawData, err.Error()))
}

/*
use example

func GetLocationFor(u *User) (*Location,error){
  respMsg, err := grpcClient.GetLocationFor(u.name)
  if err != nil{
    // here we directly send the error
    return nil, errors.New("while getting location from grpc client in GetLocationFor", err)
  }
  // process the respMsg and move on
}
*/

func StrError(err error) string {
	e := merr.WrapErr(err)
	return fmt.Sprintf("err: %s\nreason: %s\ncall stack: %s\n", e.Error(), e.RawErr(), e.CallStack())
}
