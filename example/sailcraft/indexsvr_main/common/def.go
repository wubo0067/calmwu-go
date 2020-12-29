/*
 * @Author: calmwu
 * @Date: 2017-09-20 15:10:07
 * @Last Modified by:   calmwu
 * @Last Modified time: 2017-09-20 15:10:07
 * @Comment:
 */

package common

const (
	HTTP_METHOD_GET  = 1
	HTTP_METHOD_POST = 2
	HTTP_METHOD_PUT  = 4
)

const (
	QUERYTYPE_LIKE  = "like"
	QUERYTYPE_MATCH = "match"
)

var (
	GServName           string
	GServListenIP       string
	GServListenCtrlPort int
)
