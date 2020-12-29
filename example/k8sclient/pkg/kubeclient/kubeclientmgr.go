/*
 * @Author: calmwu
 * @Date: 2020-05-24 11:28:20
 * @Last Modified by: calmwu
 * @Last Modified time: 2020-05-24 14:38:48
 */

// package kubeclient for manager the k8s clusters
package kubeclient

import (
	"sync"
	"time"

	"github.com/pkg/errors"
	calm_utils "github.com/wubo0067/calmwu-go/utils"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ClusterID = string

type ClusterConfigMap = map[ClusterID]*rest.Config

type IKubeClientMgr interface {
	// GetClient 根据集群id获取集群访问对象
	GetClient(clusterID ClusterID) (client.Client, error)

	// RegisterScheme
	RegisterScheme(pfunSchemeAdd func(*runtime.Scheme) error) error

	// Start
	Start() error

	// Stop
	Stop()
}

var (
	// KubesMgr
	KubesMgr IKubeClientMgr
	_once    sync.Once
)

// KubeClientMgr 集群访问对象
type KubeClientMgr struct {
	// kubeClusters is the cache of IKubeClientMgr keyed by ClusterID
	kubeClusters map[ClusterID]IKubeClient

	// stop is the stop channel to stop kubecluster
	stopCh chan struct{}

	// mu guards access to the map
	mu sync.RWMutex
}

// InitKubeClientMgr 初始化多集群管理对象
func InitKubeClientMgr(clusterCfgMap ClusterConfigMap) IKubeClientMgr {
	_once.Do(func() {

		mgr := &KubeClientMgr{
			kubeClusters: make(map[ClusterID]IKubeClient, len(clusterCfgMap)),
			stopCh:       make(chan struct{}),
		}

		for clusterID, config := range clusterCfgMap {
			kubeClient, err := NewKubeClient(config, "", clusterID)
			if err != nil {
				err = errors.Wrap(err, "failed to NewKubeClient.")
				calm_utils.Error(err.Error())
				return
			}

			mgr.kubeClusters[clusterID] = kubeClient
		}

		KubesMgr = mgr
	})
	return KubesMgr
}

// GetClient 获得访问k8s的client通过集群id
func (kcm *KubeClientMgr) GetClient(clusterID ClusterID) (client.Client, error) {
	kcm.mu.RLock()
	defer kcm.mu.RUnlock()

	if kubeClient, exist := kcm.kubeClusters[clusterID]; exist {
		return kubeClient.GetClient(), nil
	}
	return nil, errors.Errorf("clusterID:%s does not exist", clusterID)
}

// RegisterScheme 将crd的scheme加入
func (kcm *KubeClientMgr) RegisterScheme(pfunSchemeAdd func(*runtime.Scheme) error) error {
	kcm.mu.Lock()
	defer kcm.mu.Unlock()

	for clusterID, kubeClient := range kcm.kubeClusters {
		err := pfunSchemeAdd(kubeClient.GetScheme())
		if err != nil {
			err = errors.Wrapf(err, "clusterID:%s failed to AddToScheme.", clusterID)
			calm_utils.Error(err.Error())
			return err
		}
		calm_utils.Debugf("clusterID:%s scuccessed to AddToScheme", clusterID)
	}
	return nil
}

// Start 启动所有的kubeclient
func (kcm *KubeClientMgr) Start() error {
	kcm.mu.Lock()
	defer kcm.mu.Unlock()

	for clusterID, kubeClient := range kcm.kubeClusters {
		err := kubeClient.Start(kcm.stopCh)
		if err != nil {
			err = errors.Wrapf(err, "cluster:%s failed to start kubeClient.", clusterID)
			calm_utils.Error(err.Error())
			return err
		}
	}
	return nil
}

// Stop 停止所有的kubeclient
func (kcm *KubeClientMgr) Stop() {
	close(kcm.stopCh)
	time.Sleep(time.Second)
	calm_utils.Debug("KubeClientMgr closed")
}
