// Code generated by "stringer --type=UserVIPType"; DO NOT EDIT.

package proto

import "fmt"

const _UserVIPType_name = "E_USER_VIP_NOE_USER_VIP_NORMALMONTHLYE_USER_VIP_LUXURYMONTHLYE_USER_VIP_ALL"

var _UserVIPType_index = [...]uint8{0, 13, 37, 61, 75}

func (i UserVIPType) String() string {
	if i < 0 || i >= UserVIPType(len(_UserVIPType_index)-1) {
		return fmt.Sprintf("UserVIPType(%d)", i)
	}
	return _UserVIPType_name[_UserVIPType_index[i]:_UserVIPType_index[i+1]]
}
