/*
 * @Author: calmwu
 * @Date: 2018-01-11 10:46:42
 * @Last Modified by: calmwu
 * @Last Modified time: 2018-01-11 16:23:53
 */

package store

import (
	"sailcraft/base"
	"sailcraft/csssvr_main/common"
	"sailcraft/csssvr_main/proto"
	"sync"
	"time"

	"github.com/gocql/gocql"
)

type CassandraStoreMgr struct {
	cassandraSessionMap map[string]*gocql.Session
	initFlag            bool
	workerMgr           *CassandraWorkerMgr
}

var (
	CasMgr   *CassandraStoreMgr = nil
	initOnce sync.Once
)

func init() {
	if CasMgr == nil {
		CasMgr = new(CassandraStoreMgr)
		CasMgr.cassandraSessionMap = make(map[string]*gocql.Session)
		CasMgr.workerMgr = new(CassandraWorkerMgr)
	}
}

func (casMgr *CassandraStoreMgr) InitCassandraSessions(cassandraConf *common.CassandraConfS) error {
	var err error
	var session *gocql.Session

	initOnce.Do(func() {
		for _, keyspace := range cassandraConf.KeySpaces {
			cluster := gocql.NewCluster(cassandraConf.ClusterHosts...)
			cluster.Keyspace = keyspace
			cluster.Consistency = gocql.Consistency(cassandraConf.Consistency)
			if cassandraConf.DisableInitialHostLookup == 1 {
				// 用指定的ip，不然就是内网ip了
				cluster.DisableInitialHostLookup = true
				cluster.IgnorePeerAddr = true
			}
			cluster.NumConns = cassandraConf.WorkerRoutingCount
			cluster.ReconnectInterval = 1 * time.Second
			cluster.RetryPolicy = &gocql.SimpleRetryPolicy{NumRetries: 3}
			cluster.Timeout = time.Duration(10 * time.Second)
			cluster.SocketKeepalive = time.Duration(30 * time.Second)
			cluster.PoolConfig.HostSelectionPolicy = gocql.TokenAwareHostPolicy(gocql.RoundRobinHostPolicy())

			base.GLog.Debug("Now start cassandra cluster create session!")
			session, err = cluster.CreateSession()
			if err != nil {
				base.GLog.Error("Cassandra cluster:%v create session failed! error:[%s]", cassandraConf.ClusterHosts,
					err.Error())
			} else {
				base.GLog.Info("Cassandra cluster:%v create session successed!", cassandraConf.ClusterHosts)
			}
			casMgr.cassandraSessionMap[keyspace] = session
		}
		// 启动处理协程
		casMgr.workerMgr.Start(cassandraConf.WorkerRoutingCount)
	})
	return err
}

func (casMgr *CassandraStoreMgr) FiniCassandraSessions() {
	for k, _ := range casMgr.cassandraSessionMap {
		casMgr.cassandraSessionMap[k].Close()
	}
	casMgr.workerMgr.Stop()
}

func (casMgr *CassandraStoreMgr) SubmitRequest(cpd *proto.CassandraProcDataS) {
	casMgr.workerMgr.submitRequest(cpd)
}

func (casMgr *CassandraStoreMgr) QueryResult(req *base.ProtoRequestS, remoteIP string) interface{} {
	reply := make(chan *proto.CassandraProcResultS)
	var cpd *proto.CassandraProcDataS = new(proto.CassandraProcDataS)

	cpd.RemoteIP = remoteIP
	cpd.ReqData = req
	cpd.ResultChan = reply
	casMgr.workerMgr.submitRequest(cpd)

	select {
	case caProcRes, ok := <-reply:
		if ok {
			return caProcRes.Result
		}
	case <-time.After(3 * time.Second):
		base.GLog.Error("Cassandra Query result timeout!")
	}
	return nil
}

func (casMgr *CassandraStoreMgr) GetSessionByKeyspace(keyspace string) *gocql.Session {
	if session, ok := casMgr.cassandraSessionMap[keyspace]; ok {
		return session
	}

	base.GLog.Error("Cassandra keyspace[%s] is invalid!", keyspace)
	return nil
}
