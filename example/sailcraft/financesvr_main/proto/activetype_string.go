// Code generated by "stringer -type=ActiveType"; DO NOT EDIT.

package proto

import "fmt"

const _ActiveType_name = "E_ACTIVETYPE_SUPERGIFTPACKAGEE_ACTIVETYPE_MISSIONE_ACTIVETYPE_EXCHANGEE_ACTIVETYPE_CDKEYEXCHANGEE_ACTIVETYPE_NEWPLAYERBENEFITE_ACTIVETYPE_FIRSTRECHARGE"

var _ActiveType_index = [...]uint8{0, 29, 49, 70, 96, 125, 151}

func (i ActiveType) String() string {
	if i < 0 || i >= ActiveType(len(_ActiveType_index)-1) {
		return fmt.Sprintf("ActiveType(%d)", i)
	}
	return _ActiveType_name[_ActiveType_index[i]:_ActiveType_index[i+1]]
}
