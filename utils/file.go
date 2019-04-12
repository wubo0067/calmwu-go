/*
 * @Author: calmwu
 * @Date: 2017-09-18 10:33:28
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-04-28 16:37:37
 * @Comment:
 */
package utils

import (
	"fmt"
	"os"
)

func CheckDir(dirPath string) error {
	fileInfo, err := os.Stat(dirPath)

	// 判断路径是否存在
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "dirPath[%s] is not exist!\n", dirPath)
		}
		fmt.Fprintf(os.Stderr, "error[%s]\n", err.Error())
		return err
	}

	// 检查是不是目录
	if !fileInfo.IsDir() {
		fmt.Fprintf(os.Stderr, "dirPath[%s] is not Directory!\n", dirPath)
		return err
	}
	return nil
}

func MkDir(dirPath string) error {
	err := CheckDir(dirPath)

	if err != nil {
		// 目录不存在创建
		err = os.MkdirAll(dirPath, 0777)
		if err != nil {
			fmt.Fprintf(os.Stderr, "path[%s] create failed! reason[%s]\n", dirPath, err.Error())
			return err
		}
	}
	return nil
}

func PathExist(path string) error {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return err
	}
	return nil
}
