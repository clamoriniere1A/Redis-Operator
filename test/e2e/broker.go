package e2e

import (
	// for test lisibility
	. "github.com/onsi/ginkgo"
	// for test lisibility
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	rapi "github.com/amadeusitgroup/redis-operator/pkg/api/redis/v1"

	"github.com/amadeusitgroup/redis-operator/test/e2e/framework"
)

const (
	serviceInstanceName = "redis-instance"
	serviceBindingName  = "redis-binding"
	servicePlanDefault  = "default"
	clusterNameBroker   = "clusterbroker"
)

var redisclusterBroker *rapi.RedisCluster

func deleteRedisClusterServiceInstance(instanceName string) {
	framework.HODeleteServiceInstance(serviceCatalogClient, clusterNs, instanceName)()
}

func deleteRedisClusterServiceBinding(bindingName string) {
	framework.HODeleteServiceBinding(serviceCatalogClient, clusterNs, bindingName)()
}

var _ = Describe("Redis broker", func() {

	Describe("Create a ServiceInstance", func() {
		It("should create a RedisCluster service instance", func() {
			params := framework.NewServiceInstanceParameters(clusterNameBroker, clusterNs, nil, nil)
			serviceIntance := framework.NewRedisClusterServiceInstance(serviceInstanceName, clusterNs, servicePlanDefault, params)

			Eventually(framework.HOCreateServiceInstance(serviceCatalogClient, clusterNs, serviceIntance), "5s", "1s").ShouldNot(HaveOccurred())

			Eventually(framework.HOIsServiceInstanceCreated(serviceCatalogClient, clusterNs, serviceIntance, false), "2m", "5s").ShouldNot(HaveOccurred())

			Eventually(framework.HOGetRedisCluster(redisClient, redisclusterBroker, clusterNameBroker, clusterNs), "5m", "10s").ShouldNot(HaveOccurred())
			var err error
			redisclusterBroker, err = redisClient.RedisoperatorV1().RedisClusters(clusterNs).Get(clusterNameBroker, metav1.GetOptions{})
			Ω(err).Should(BeNil())

			Eventually(framework.HOIsRedisClusterStarted(redisClient, redisclusterBroker, clusterNs), "3m", "5s").ShouldNot(HaveOccurred())

			Eventually(framework.HOIsServiceInstanceCreated(serviceCatalogClient, clusterNs, serviceIntance, true), "30s", "5s").ShouldNot(HaveOccurred())

			Eventually(framework.HOIsServiceInstanceSucceed(serviceCatalogClient, clusterNs, serviceIntance, true), "30s", "5s").ShouldNot(HaveOccurred())

			serviceBinding := framework.NewRedisClusterServiceBinding(serviceBindingName, clusterNs, serviceInstanceName)
			framework.Logf("serviceBinding created localy")
			Eventually(framework.HOCreateServiceBinding(serviceCatalogClient, clusterNs, serviceBinding), "5s", "1s").ShouldNot(HaveOccurred())
			framework.Logf("serviceBinding created in kube api server")
			Eventually(framework.HOIsServiceBindingCreated(serviceCatalogClient, clusterNs, serviceBinding), "2m", "5s").ShouldNot(HaveOccurred())
			framework.Logf("serviceBinding created checked ok")
		})
	})

	AfterEach(func() {
		deleteRedisClusterServiceBinding(serviceBindingName)
		deleteRedisClusterServiceInstance(serviceInstanceName)
		deleteRedisCluster(redisClient, redisclusterBroker)
	})
})
