package main

import (
    "os"
    "fmt"
    "text/template"
)

var kpconf_template string = `! Configuration File for keepalived" 

global_defs {
   notification_email {
   }
   router_id LVS_DEVEL
}

local_address_group laddr_g1 {
    {{.LocalAddr}}
}	

virtual_server_group shanks1 {
    {{with .VirtualSvrInfoLst}}{{range .}}{{.VirtualIP}} {{.Port}}
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
        {{range .VirtualIPLst}}{{.}}
        {{end}}
    }
}

{{with .VirtualSvrInfoLst}}{{range .}}virtual_server {{.VirtualIP}} {{.Port}} {
    delay_loop 10
    lb_algo {{.Scheduler}}
    lb_kind FNAT
    protocol {{.Protocol}}
    laddr_group_name laddr_g1
    {{$val_protocol := .Protocol}}
    {{$val_connecttimeout := .ConnectTimeout}}
    {{with .RealSvrInfoLst}}{{range.}}real_server {{.IP}} {{.Port}} {
        weight {{.Weight}}
        {{if eq $val_protocol "TCP"}}notify_up "/etc/keepalived/rs_up.sh {{.IP}} {{.Port}} tcp"
        notify_down "/etc/keepalived/rs_down.sh {{.IP}} {{.Port}} tcp"    
        TCP_CHECK {
            connect_timeout {{$val_connecttimeout}}
            nb_get_retry 3
            delay_before_retry 3
            connect_port {{.Port}} 
        }{{else}}UDP_CHECK {
        }{{end}}
    }
    {{end}}{{end}}
}

{{end}}{{end}}`

type RealSvrInfo struct {
    IP          string
    Port        int
    Weight      int
}

type VirtualSvrInfo struct {
    VirtualIP       string
    Port            int
    Scheduler       string
    Protocol        string
    ConnectTimeout  int32

    RealSvrInfoLst  []*RealSvrInfo
}

type KPConfRenderData struct {
    LocalAddr               string
    VirtualSvrInfoLst       []*VirtualSvrInfo
    VrrpRouterID            int
    VirtualIPLst            []string
    VrrpInterface           string
}

func main() {
    template_head_info := KPConfRenderData{
        LocalAddr : "10.12.16.178",
        VirtualSvrInfoLst : []*VirtualSvrInfo{
            &VirtualSvrInfo{
                VirtualIP : "12.23.34.45",
                Port : 12,
                Scheduler : "rr",
                Protocol : "TCP",
                ConnectTimeout : 9,
                RealSvrInfoLst : []*RealSvrInfo{
                    &RealSvrInfo{"1.1.1.1", 9991, 10},
                    &RealSvrInfo{"2.2.2.2", 9992, 11},
                    &RealSvrInfo{"3.3.3.3", 9993, 12},
                },
            },
            &VirtualSvrInfo{
                VirtualIP : "10.12.67.45",
                Port : 22,
                Scheduler : "wrr",
                Protocol : "UDP",
                ConnectTimeout : 9,
                RealSvrInfoLst : []*RealSvrInfo{
                    &RealSvrInfo{"4.4.4.4", 9994, 13},
                    &RealSvrInfo{"5.5.5.5", 9995, 14},
                    &RealSvrInfo{"6.6.6.6", 9996, 15},
                },                
            },            
        },
        VrrpRouterID : 100,
        VirtualIPLst : []string{"12.23.34.45", "10.12.67.45",},
        VrrpInterface : "eth0",
    }

    templ, err := template.New("kpconf").Parse(kpconf_template)
    if err != nil {
        fmt.Printf("%s\n", err.Error())
        return
    }

    // create keepalived.conf
    h_kpconfig_file, err := os.OpenFile("keepalived.conf", 
        os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0600)
    if err != nil {
        fmt.Println(err.Error())
        return
    }
    templ.Execute(h_kpconfig_file, template_head_info)
    h_kpconfig_file.Close();
    templ.Execute(os.Stdout, template_head_info)
}
