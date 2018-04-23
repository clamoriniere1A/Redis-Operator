package e2e

import (
	api "k8s.io/api/core/v1"
	clientset "k8s.io/client-go/kubernetes"

	scclientset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"

	// for test lisibility
	. "github.com/onsi/ginkgo"

	rapi "github.com/amadeusitgroup/redis-operator/pkg/api/redis/v1"
	"github.com/amadeusitgroup/redis-operator/pkg/client/clientset/versioned"
	"github.com/amadeusitgroup/redis-operator/test/e2e/framework"
)

var redisClient versioned.Interface
var kubeClient clientset.Interface
var serviceCatalogClient scclientset.Interface

var rediscluster *rapi.RedisCluster

const clusterName = "cluster1"
const clusterNs = api.NamespaceDefault

var _ = BeforeSuite(func() {
	redisClient, kubeClient, serviceCatalogClient = framework.BuildAndSetClients()
})

var _ = AfterSuite(func() {
	deleteRedisCluster(redisClient, rediscluster)
	deleteRedisClusterServiceInstance(serviceInstanceName)
})
