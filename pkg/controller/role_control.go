package controller

import (
	"fmt"
	"github.com/golang/glog"

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
const redisRoleName = "redis-node"

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
	exist, err = c.IsRolePresent(redisCluster)
	if err != nil {
		return false, err
	}
	if !exist {
		return false, nil
	}
	exist, err = c.IsRoleBindingPresent(redisCluster)
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

// IsRolePresent use to know if the ClusterRole exist
func (c *RoleControl) IsRolePresent(redisCluster *rapi.RedisCluster) (bool, error) {
	_, err := c.KubeClient.RbacV1beta1().Roles(redisCluster.Namespace).Get(redisRoleName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// IsRoleBindingPresent use to know if the ClusterRole exist
func (c *RoleControl) IsRoleBindingPresent(redisCluster *rapi.RedisCluster) (bool, error) {
	_, err := c.KubeClient.RbacV1beta1().RoleBindings(redisCluster.Namespace).Get(fmt.Sprintf("%s-%s", redisRoleName, redisCluster.Namespace), metav1.GetOptions{})
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
			glog.Errorf("unable to create a Service account: %s/%s, err:%v", redisCluster.Namespace, redisNodeServiceAccountName, err)
			return err
		}
	}

	exist, err = c.IsRolePresent(redisCluster)
	if err != nil {
		return err
	}
	if !exist {
		if _, err = c.KubeClient.RbacV1beta1().Roles(redisCluster.Namespace).Create(newRole(redisRoleName)); err != nil {
			glog.Errorf("unable to create a Roles: %s/%s, err:%v", redisCluster.Namespace, redisRoleName, err)
			return err
		}
	}

	exist, err = c.IsRoleBindingPresent(redisCluster)
	if err != nil {
		return err
	}
	if !exist {
		if _, err := c.KubeClient.RbacV1beta1().RoleBindings(redisCluster.Namespace).Create(newRoleBinding(redisNodeServiceAccountName, redisCluster.Namespace)); err != nil {
			glog.Errorf("unable to create a RoleBindings: %s/%s, err:%v", redisCluster.Namespace, redisNodeServiceAccountName, err)
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

func newRoleBinding(name, ns string) *krbacv1.RoleBinding {
	return &krbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s", name, ns),
		},
		Subjects: []krbacv1.Subject{
			{Kind: "ServiceAccount", Name: name, Namespace: ns},
		},
		RoleRef: krbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     name,
		},
	}
}

func newRole(name string) *krbacv1.Role {
	return &krbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Rules: []krbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"services", "endpoints", "pods"},
				Verbs:     []string{"list", "get"},
			},
		},
	}
}
