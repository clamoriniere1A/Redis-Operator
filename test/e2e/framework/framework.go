package framework

import (
	"fmt"

	scclientset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"

	rclient "github.com/amadeusitgroup/redis-operator/pkg/client/clientset/versioned"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/amadeusitgroup/redis-operator/pkg/client"
)

// Framework stores necessary info to run e2e
type Framework struct {
	KubeConfig *rest.Config
}

type frameworkContextType struct {
	KubeConfigPath string
	ImageTag       string
}

// FrameworkContext stores globally the framework context
var FrameworkContext frameworkContextType

// NewFramework creates and initializes the a Framework struct
func NewFramework() (*Framework, error) {
	Logf("KubeconfigPath-> %q", FrameworkContext.KubeConfigPath)
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", FrameworkContext.KubeConfigPath)
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve kubeConfig:%v", err)
	}
	return &Framework{
		KubeConfig: kubeConfig,
	}, nil
}

func (f *Framework) kubeClient() (clientset.Interface, error) {
	return clientset.NewForConfig(f.KubeConfig)
}

func (f *Framework) redisOperatorClient() (rclient.Interface, error) {
	c, err := client.NewClient(f.KubeConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create rediscluster client:%v", err)
	}
	return c, err
}

func (f *Framework) serviceCatalogClient() (scclientset.Interface, error) {
	c, err := scclientset.NewForConfig(f.KubeConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create servicecatalog client:%v", err)
	}
	return c, err
}
