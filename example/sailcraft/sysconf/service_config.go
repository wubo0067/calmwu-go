/*
 * @Author: calmwu
 * @Date: 2018-01-10 11:59:42
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-01-10 12:13:41
 */

// 所有注册服务的域名、服务端口
package sysconf

type ServiceInfoS struct {
	ServiceTag        string `json:"ServiceTag"`
	ServiceDomainName string `json:"DomainName"`
	ServicePort       int    `json:"Port"`
}

type ServiceConfigS struct {
	Services []ServiceConfigS `json:"Services"`
}
