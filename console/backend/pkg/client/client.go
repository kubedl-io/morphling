package client

import (
	"context"

	morphlingv1alpha1 "github.com/alibaba/morphling/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	setupLog = ctrl.Log.WithName("setup")
	cmgr     = &ClientMgr{}
	scheme   = runtime.NewScheme()
)

type ClientMgr struct {
	config     *rest.Config
	scheme     *runtime.Scheme
	ctrlCache  cache.Cache
	ctrlClient client.Client
	//kubeClient clientset.Interface
}

func Init() *ClientMgr {
	cmgr.config = ctrl.GetConfigOrDie()
	cmgr.scheme = scheme
	_ = clientgoscheme.AddToScheme(cmgr.scheme)
	_ = morphlingv1alpha1.AddToScheme(cmgr.scheme)

	ctrlCache, err := cache.New(cmgr.config, cache.Options{Scheme: cmgr.scheme})
	if err != nil {
		klog.Fatal(err)
	}
	cmgr.ctrlCache = ctrlCache

	c, err := client.New(cmgr.config, client.Options{Scheme: cmgr.scheme})
	if err != nil {
		klog.Fatal(err)
	}

	cmgr.ctrlClient = &client.DelegatingClient{
		Reader: &client.DelegatingReader{
			CacheReader:  ctrlCache,
			ClientReader: c,
		},
		Writer:       c,
		StatusClient: c,
	}
	return cmgr
}

func Start() {
	go func() {
		stopChan := make(chan struct{})
		cmgr.ctrlCache.Start(stopChan)
	}()
}

func (c *ClientMgr) GetCtrlClient() client.Client {
	return c.ctrlClient
}

// IndexField is Used for filtering Pods from PodList
func (c *ClientMgr) IndexField(obj runtime.Object, field string, extractValue client.IndexerFunc) error {
	return c.ctrlCache.IndexField(context.Background(), obj, field, extractValue)
}
