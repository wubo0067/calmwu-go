/*
 * @Author: calmwu
 * @Date: 2020-05-24 11:28:07
 * @Last Modified by: calmwu
 * @Last Modified time: 2020-05-24 14:28:10
 */

// package kubeclient for manager the k8s clusters
package kubeclient

import (
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	calm_utils "github.com/wubo0067/calmwu-go/utils"
)

type IKubeClient interface {
	//
	GetScheme() *runtime.Scheme
	//
	Start(stopCh <-chan struct{}) error
	//
	GetClient() client.Client
}

// KubeClient 集群访问类型
type KubeClient struct {
	config    *rest.Config
	scheme    *runtime.Scheme
	cache     cache.Cache
	client    client.Client
	namespace string
	clusterID ClusterID
}

// NewKubeClient 构造一个集群访问对象
func NewKubeClient(config *rest.Config, namespace string, clusterID ClusterID) (IKubeClient, error) {
	mapper, err := apiutil.NewDynamicRESTMapper(config)
	if err != nil {
		err = errors.Wrapf(err, "cluster:%s failed to NewDynamicRESTMapper with config", clusterID)
		calm_utils.Error(err.Error())
		return nil, err
	}

	kubeClusterScheme := scheme.Scheme

	cacheOptions := cache.Options{
		Scheme:    kubeClusterScheme,
		Mapper:    mapper,
		Namespace: namespace,
	}

	// 创建cache
	kubeClusterCache, err := cache.New(config, cacheOptions)
	if err != nil {
		err = errors.Wrapf(err, "cluster:%s failed to NewDynamicRESTMapper with config", clusterID)
		calm_utils.Error(err.Error())
		return nil, err
	}

	// 创建client
	c, err := client.New(config, client.Options{Scheme: kubeClusterScheme, Mapper: mapper})
	if err != nil {
		err = errors.Wrapf(err, "cluster:%s failed to new controller-runtime client.", clusterID)
		calm_utils.Error(err.Error())
		return nil, err
	}

	kubeClusterClient := client.DelegatingClient{
		Reader: &client.DelegatingReader{
			CacheReader:  kubeClusterCache,
			ClientReader: c,
		},
		Writer:       c,
		StatusClient: c,
	}

	return &KubeClient{
		config:    config,
		scheme:    kubeClusterScheme,
		cache:     kubeClusterCache,
		client:    kubeClusterClient,
		namespace: namespace,
		clusterID: clusterID,
	}, nil
}

// GetScheme 返回runtime.Scheme
func (kc *KubeClient) GetScheme() *runtime.Scheme {
	return kc.scheme
}

// Start runs all the informers known to this cache until the given channel is closed.
func (kc *KubeClient) Start(stopCh <-chan struct{}) error {
	go func() {
		if err := kc.cache.Start(stopCh); err != nil {
			calm_utils.Fatalf("cluster:%s failed to start Cache. err:%s", kc.clusterID, err.Error())
		}
	}()

	if !kc.cache.WaitForCacheSync(stopCh) {
		err := errors.Errorf("cluster:%s wait for all the caches sync failed, beacuse it could not sync a cache", kc.clusterID)
		calm_utils.Error(err.Error())
		return err
	}
	calm_utils.Debugf("cluster:%s waits for all the caches to sync successed.", kc.clusterID)
	return nil
}

func (kc *KubeClient) GetClient() client.Client {
	return kc.client
}
