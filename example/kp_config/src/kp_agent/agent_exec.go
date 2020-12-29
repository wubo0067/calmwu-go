package main

import (
	"fmt"
	"kp_proto"
	"os/exec"
	"strings"
)

const KPADMIN_CMD_TEMPLATE = "/usr/local/bin/kp_admin.py --cmd=%s --id=%d --local_ip=%s --port=%d --real_svrs=%s"

func exec_kpadmin(cmd_type string, virtual_router_id int32, host_ip string, port int32, relserverlst []string) (int32, string) {

	// kp_shellcmd := fmt.Sprintf(KPADMIN_CMD_TEMPLATE, cmd_type,
	//     virtual_router_id,
	//     host_ip,
	//     port,
	//     strings.Join(relserverlst, ":"))

	var kp_shellcmd []string

	kp_shellcmd = append(kp_shellcmd, "/usr/local/bin/kp_admin.py")
	kp_shellcmd = append(kp_shellcmd, fmt.Sprintf("--cmd=%s", cmd_type))
	kp_shellcmd = append(kp_shellcmd, fmt.Sprintf("--id=%d", virtual_router_id))
	kp_shellcmd = append(kp_shellcmd, fmt.Sprintf("--local_ip=%s", host_ip))
	kp_shellcmd = append(kp_shellcmd, fmt.Sprintf("--port=%d", port))
	kp_shellcmd = append(kp_shellcmd, fmt.Sprintf("--real_svrs=%s", strings.Join(relserverlst, ":")))

	g_log.Debug("exec_cmd[%v]", kp_shellcmd)

	cmd := exec.Command("/usr/local/bin/python", kp_shellcmd...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		g_log.Error("exec kp_admin failed! reason[%s] output[%s]", err.Error(), string(out))
		return -1, string(out)
	}
	return 0, string(out)
}

func exec_kpreload() (kp_proto.KPErrorCode, string) {
	cmd := exec.Command("/etc/init.d/keepalived", "reload")
	out, err := cmd.CombinedOutput()
	if err != nil {
		g_log.Error("exec [/etc/init.d/keepalived reload] failed! reason[%s] output[%s]", err.Error(), string(out))
		return kp_proto.KPErrorCode_E_KPERR_FAILED_RELOAD, string(out)
	}
	return kp_proto.KPErrorCode_E_KPERR_OK, string(out)
}

func exec_close_fullnat_toa_entry() int {
	cmd_info := "echo 0 > /proc/sys/net/ipv4/vs/fullnat_toa_entry"
	cmd := exec.Command("/bin/sh", "-c", cmd_info)
	out, err := cmd.CombinedOutput()
	if err != nil {
		g_log.Error("exec [echo 0 > /proc/sys/net/ipv4/vs/fullnat_toa_entry] failed! reason[%s] output[%s]", err.Error(), string(out))
		return -1
	}
	g_log.Debug("Execute command[echo 0 > /proc/sys/net/ipv4/vs/fullnat_toa_entry] successed")
	return 0
}
