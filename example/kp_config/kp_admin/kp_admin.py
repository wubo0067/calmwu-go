#!/usr/bin/env python
#coding=utf-8

'''
@version: ??
@author: charlieyou
@contact: youcongabc@qq.com
@software: PyCharm
@file: tools.py
@time: 2016/10/9 9:56
'''

import getopt
import os
import sys
import traceback


keepalived_conf_tpl = '''
global_defs {
   router_id LVS_DEVEL
}

local_address_group laddr_g1 {
    local_ip
}

virtual_server_group shanks1 {
    local_ip port
}

vrrp_instance LVS_Cluster{

    state MASTER
    interface eth0
    virtual_router_id r_id
    priority 100
    nopreempt FALSE
    advert_int 1
    authentication {
        auth_type PASS
        auth_pass 08856CD8
    }
    virtual_ipaddress {
        local_ip
    }
}


virtual_server local_ip port {
    delay_loop 6
    lb_algo rr
    lb_kind FNAT
    protocol TCP
    laddr_group_name laddr_g1
    real_servers
}
'''

real_server_tpl = '''
    real_server server_ip port {
        weight 1
        TCP_CHECK {
            connect_timeout 3
            nb_get_retry 3
            delay_before_retry 3
            connect_port port
        }
    }
'''

def show_usage():
    print """usage: kp_admin.py --cmd=start/reload --id=xxx --local_ip=xx.xx.xx.xx --port=xxx --real_svrs=x.x.x.x:x.x.x.x:...
    """
if __name__ == '__main__':
    if not os.path.exists("/etc/rc.d/init.d/keepalived"):
        print "keepalived not existed!"
        sys.exit(-1)

    try:
        (opts, args) = getopt.getopt(sys.argv[1:], 'c:i:p:r:', ["cmd=", "id=", "local_ip=", "port=", "real_svrs="])
    except Exception as exc:
        show_usage()
        sys.exit(-1)

    if len(opts) == 0:
        show_usage()
        sys.exit(-1)

    cmd = ""
    id = None
    local_ip = ""
    port = 0
    real_server_lst = []
    for opt, arg in opts:
        if opt in ("-c", "--cmd"):
            if arg not in ("start", "reload"):
                print "illegal cmd, only start and reload are supported"
                sys.exit(-1)
            else:
                cmd = arg
        elif opt in ("-i", "--id"):
            id = arg
        elif opt in ("-l", "--local_ip"):
            local_ip = arg
        elif opt in ("-p", "--port"):
            port = arg
        elif opt in ("-r", "--real_svrs"):
            real_server_lst = arg.split(":")
        else:
            print "unsupported param {0}".format(opt)

    if "start" == cmd:
        real_server_conf = ""
        for real_svr in real_server_lst:
            real_server_conf += real_server_tpl.replace("server_ip", real_svr).replace("port", port)

        conf = keepalived_conf_tpl.replace("local_ip", local_ip).replace("port", port).replace("r_id", id).replace("real_servers", real_server_conf)

        if not os.path.exists("/etc/keepalived/"):
            os.makedirs("/etc/keepalived/")
        try:
            f = open("/etc/keepalived/keepalived.conf", "w")
            f.write(conf)
            f.close()
        except Exception as exc:
            print traceback.format_exc()
            sys.exit(-1)

        cmd = "/etc/rc.d/init.d/keepalived start"
        ret = os.popen(cmd).read()
        if ret.find("OK") != -1:
            sys.exit(0)
        elif ret.find("FAILED") != -1:
            print "keepalived start failed!"
            sys.exit(-1)
        else:
            print "keepalived already started!"
            sys.exit(-1)
    else:
        try:
            f = open("/etc/keepalived/keepalived.conf", "r")
            data = f.readall()
            f.close()
        except Exception as exc:
            print traceback.format_exc()
            sys.exit(-1)

        pos = data.find("real_server")
        if pos == -1:
            pos = data.rfind("}")
            if pos == -1:
                print "invalid conf"
                sys.exit(-1)

        conf = data[:pos]
        for real_svr in real_server_lst:
            conf += real_server_tpl.format(real_svr, port)
        conf += "}"
        try:
            f = open("/etc/keepalived/keepalived.conf", "w")
            f.write(conf)
            f.close()
        except Exception as exc:
            print traceback.format_exc()
            sys.exit(-1)

        cmd = "/etc/rc.d/init.d/keepalived start"
        ret = os.popen(cmd).read()
        if ret.find("OK") != -1:
            sys.exit(0)
        else:
            print "keepalived reload failed!".format(cmd)
            sys.exit(-1)