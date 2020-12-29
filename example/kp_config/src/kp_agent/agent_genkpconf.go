package main

import (
	"os"
	//"fmt"
	"kp_proto"
	"text/template"
)

var kpconf_template string = `! Configuration File for keepalived" 

global_defs {
   notification_email {
   }
   router_id LVS_DEVEL
}

local_address_group laddr_g1 {
    {{.LocalIP}}
}	
{{$val_virtualip := .VirtualIP}}
{{$val_linknodetype := .LinkNodeType}}
virtual_server_group shanks1 {
    {{with .VirtualSvrInfoLst}}{{range .}}{{$val_virtualip}} {{.Port}}
    {{end}}{{end}}
}

vrrp_instance LVS_Cluster{

    state MASTER   
    interface {{.VrrpInterface}}  
    virtual_router_id {{.VrrpRouterID}}
    priority 100 
    nopreempt FALSE 
    advert_int 1 
    authentication {
        auth_type PASS  
        auth_pass 08856CD8
    }
    virtual_ipaddress {
        {{$val_virtualip}}
    }
}

{{with .VirtualSvrInfoLst}}{{range .}}virtual_server {{$val_virtualip}} {{.Port}} {
    delay_loop {{.DelayLoop}}
    lb_algo {{.Scheduler}}
    lb_kind FNAT
    protocol {{.NetProtoName}}
    laddr_group_name laddr_g1
    {{if eq $val_linknodetype 1}}alpha
    {{end}}
    {{$val_scheduler := .Scheduler}}
    {{$val_protocol := .NetProtoName}}
    {{with .RealSvrInfoLst}}{{range.}}real_server {{.RsIP}} {{.RsPort}} {
        {{if eq $val_scheduler "wrr"}}weight {{.Weight}}{{end}}
        {{if eq $val_linknodetype 1}}
        {{if eq $val_protocol "TCP"}}notify_up "/usr/local/bin/kp_agent --mode=notify --rs_ip={{.RsIP}} --rs_port={{.RsPort}} --proto_name=TCP --status=up --notify_port=10001"
        notify_down "/usr/local/bin/kp_agent --mode=notify --rs_ip={{.RsIP}} --rs_port={{.RsPort}} --proto_name=TCP --status=down --notify_port=10001"     
        TCP_CHECK {
            connect_timeout 10
            connect_port {{.RsPort}}
        }{{else}}UDP_CHECK {
        }{{end}}
        {{end}}
    }
    {{end}}{{end}}
}

{{end}}{{end}}`

func generate_kpconfigfile() (kp_proto.KPErrorCode, string) {
	// 生成keepalived.conf文件
	f_kpconfig, err := os.OpenFile("/etc/keepalived/keepalived.conf",
		os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0600)
	if err != nil {
		g_log.Error("open file /etc/keepalived/keepalived.conf failed, reason[%s]", err.Error())
		return kp_proto.KPErrorCode_E_KPERR_FAILED_RELOAD, err.Error()
	}

	defer f_kpconfig.Close()

	templ, err := template.New("kpconf").Parse(kpconf_template)
	if err != nil {
		g_log.Error("Parse kpconf_template failed! reason[%s]", err.Error())
		return kp_proto.KPErrorCode_E_KPERR_FAILED_RELOAD, err.Error()
	}

	err = templ.Execute(f_kpconfig, g_kpconfinfo)
	if err != nil {
		g_log.Error("template execute failed! reason[%s]", err.Error())
		return kp_proto.KPErrorCode_E_KPERR_FAILED_RELOAD, err.Error()
	}

	return kp_proto.KPErrorCode_E_KPERR_OK, ""
}
