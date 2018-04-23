package controller

import (
	"fmt"

	kapiv1 "k8s.io/api/core/v1"
	krbacv1 "k8s.io/api/rbac/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"

	rapi "github.com/amadeusitgroup/redis-operator/pkg/api/redis/v1"
)

// RoleControlInterface use to create delete Role for the RedisCluster
type RoleControlInterface interface {
	IsSecurityRessourcePresent(redisCluster *rapi.RedisCluster) (bool, error)
	CreateSecurityResources(redisCluster *rapi.RedisCluster) error
}

// RoleControl contains all information for managing Kube accounts and roles
type RoleControl struct {
	KubeClient clientset.Interface
	Recorder   record.EventRecorder
}

var _ RoleControlInterface = &RoleControl{}

const redisNodeServiceAccountName = "redis-node"
const redisClusterRoleName = "redis-node"

// NewRoleControl builds and returns new NewRoleControl instance
func NewRoleControl(client clientset.Interface, rec record.EventRecorder) *RoleControl {
	ctrl := &RoleControl{
		KubeClient: client,
		Recorder:   rec,
	}

	return ctrl
}

// IsSecurityRessourcePresent returns true if all Security resources already exist
func (c *RoleControl) IsSecurityRessourcePresent(redisCluster *rapi.RedisCluster) (bool, error) {
	exist, err := c.IsRedisNodeAccountAndRolePresent(redisCluster)
	if err != nil {
		return false, err
	}
	if !exist {
		return false, nil
	}
	exist, err = c.IsClusterRolePresent(redisCluster)
	if err != nil {
		return false, err
	}
	if !exist {
		return false, nil
	}
	exist, err = c.IsClusterRoleBindingPresent(redisCluster)
	if err != nil {
		return false, err
	}
	if !exist {
		return false, nil
	}

	return true, nil

}

// IsRedisNodeAccountAndRolePresent use to know if the ServiceAccount exist
func (c *RoleControl) IsRedisNodeAccountAndRolePresent(redisCluster *rapi.RedisCluster) (bool, error) {
	_, err := c.KubeClient.Core().ServiceAccounts(redisCluster.Namespace).Get(redisNodeServiceAccountName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// IsClusterRolePresent use to know if the ClusterRole exist
func (c *RoleControl) IsClusterRolePresent(redisCluster *rapi.RedisCluster) (bool, error) {
	_, err := c.KubeClient.RbacV1beta1().ClusterRoles().Get(redisClusterRoleName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// IsClusterRoleBindingPresent use to know if the ClusterRole exist
func (c *RoleControl) IsClusterRoleBindingPresent(redisCluster *rapi.RedisCluster) (bool, error) {
	_, err := c.KubeClient.RbacV1beta1().ClusterRoleBindings().Get(fmt.Sprintf("%s-%s", redisClusterRoleName, redisCluster.Namespace), metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// CreateSecurityResources use to create the ServiceAccount and Role for the Redis-Node
func (c *RoleControl) CreateSecurityResources(redisCluster *rapi.RedisCluster) error {
	exist, err := c.IsRedisNodeAccountAndRolePresent(redisCluster)
	if err != nil {
		return err
	}
	if !exist {
		if _, err = c.KubeClient.CoreV1().ServiceAccounts(redisCluster.Namespace).Create(newServiceAccount(redisNodeServiceAccountName)); err != nil {
			return err
		}
	}

	exist, err = c.IsClusterRolePresent(redisCluster)
	if err != nil {
		return err
	}
	if !exist {
		if _, err = c.KubeClient.RbacV1beta1().ClusterRoles().Create(newClusterRole(redisClusterRoleName)); err != nil {
			return err
		}
	}

	exist, err = c.IsClusterRoleBindingPresent(redisCluster)
	if err != nil {
		return err
	}
	if !exist {
		if _, err := c.KubeClient.RbacV1beta1().ClusterRoleBindings().Create(newClusterRoleBinding(redisNodeServiceAccountName, redisCluster.Namespace)); err != nil {
			return err
		}
	}
	return nil
}

func newServiceAccount(name string) *kapiv1.ServiceAccount {
	return &kapiv1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func newClusterRoleBinding(name, ns string) *krbacv1.ClusterRoleBinding {
	return &krbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s", name, ns),
		},
		Subjects: []krbacv1.Subject{
			{Kind: "ServiceAccount", Name: name, Namespace: ns},
		},
		RoleRef: krbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     name,
		},
	}
}

func newClusterRole(name string) *krbacv1.ClusterRole {
	return &krbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Rules: []krbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"namespaces", "services", "endpoints", "pods"},
				Verbs:     []string{"list", "get"},
			},
		},
	}
}
