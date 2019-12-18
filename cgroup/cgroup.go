/*
 * @Author: calmwu
 * @Date: 2019-12-14 10:44:22
 * @Last Modified by: calmwu
 * @Last Modified time: 2019-12-14 10:47:41
 */

// Package cgroup wrapper cgroup operation
package cgroup

import (
	"bufio"
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
	calm_utils "github.com/wubo0067/calmwu-go/utils"
)

// FindCgroupMountPoint 找出挂载了某个subsystem的hierarchy cgroup根节点所在的目录 FindCgroupMountPoint(“memory”)
func FindCgroupMountPoint(subsystem string) string {
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		calm_utils.Errorf("Open /proc/self/mountinfo failed. err:%s", err.Error())
		return ""
	}
	defer f.Close()

	// 读取文件
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		txt := scanner.Text()
		fields := strings.Split(txt, " ")
		for _, opt := range strings.Split(fields[len(fields)-1], ",") {
			if opt == subsystem {
				return fields[4]
			}
		}
	}

	if err := scanner.Err(); err != nil {
		calm_utils.Errorf("Scanner /proc/self/mountinfo failed. err:%s", err.Error())
		return ""
	}

	return ""
}

// GetCgroupPath 得到cgroup在文件系统中的绝对路径
func GetCgroupPath(subsystem string, cgroupPath string, autoCreate bool) (string, error) {
	cgroupRoot := FindCgroupMountPoint(subsystem)
	if _, err := os.Stat(path.Join(cgroupRoot, cgroupPath)); err == nil ||
		(autoCreate && os.IsNotExist(err)) {
		if os.IsNotExist(err) {
			if err = os.Mkdir(path.Join(cgroupRoot, cgroupPath), 0755); err != nil {
				return "", errors.Errorf("create cgroup dir failed. err:%s", err.Error())
			}
		}
		return path.Join(cgroupRoot, cgroupPath), nil
	} else {
		return "", errors.Errorf("cgroup path error %s", err.Error())
	}
}
