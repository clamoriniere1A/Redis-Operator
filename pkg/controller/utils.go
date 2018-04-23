package controller

import (
	"errors"
	"fmt"
	"net"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"

	rapi "github.com/amadeusitgroup/redis-operator/pkg/api/redis/v1"
	"github.com/amadeusitgroup/redis-operator/pkg/config"
	"github.com/amadeusitgroup/redis-operator/pkg/redis"
)

func getPodLabelsSet(rediscluster *rapi.RedisCluster) (labels.Set, error) {
	desiredLabels := labels.Set{}
	for k, v := range rediscluster.Spec.AdditionalLabels {
		desiredLabels[k] = v
	}
	for k, v := range rediscluster.Spec.PodTemplate.Labels {
		desiredLabels[k] = v
	}
	desiredLabels[rapi.ClusterNameLabelKey] = rediscluster.Name // add rediscluster name to the Pod labels
	return desiredLabels, nil
}

// createRedisClusterPodLabelSelector creates label selector to select the jobs related to a rediscluster, stepName
func createRedisClusterPodLabelSelector(rediscluster *rapi.RedisCluster) (labels.Selector, error) {
	set, err := getPodLabelsSet(rediscluster)
	if err != nil {
		return nil, err
	}
	return labels.SelectorFromSet(set), nil
}

// NewRedisAdmin builds and returns new redis.Admin from the list of pods
func NewRedisAdmin(pods []*apiv1.Pod, cfg *config.Redis) (redis.AdminInterface, error) {
	nodesAddrs := []string{}
	for _, pod := range pods {
		redisPort := redis.DefaultRedisPort
		for _, container := range pod.Spec.Containers {
			if container.Name == "redis-node" {
				for _, port := range container.Ports {
					if port.Name == "redis" {
						redisPort = fmt.Sprintf("%d", port.ContainerPort)
					}
				}
			}
		}
		nodesAddrs = append(nodesAddrs, net.JoinHostPort(pod.Status.PodIP, redisPort))
	}
	adminConfig := redis.AdminOptions{
		ConnectionTimeout:  time.Duration(cfg.DialTimeout) * time.Millisecond,
		RenameCommandsFile: cfg.GetRenameCommandsFile(),
	}

	return redis.NewAdmin(nodesAddrs, &adminConfig), nil
}

// IsPodReady check if pod is in ready condition, return the error message otherwise
func IsPodReady(pod *apiv1.Pod) (bool, error) {
	if pod == nil {
		return false, errors.New("No Pod")
	}

	// get ready condition
	var readycondition apiv1.PodCondition
	found := false
	for _, cond := range pod.Status.Conditions {
		if cond.Type == apiv1.PodReady {
			readycondition = cond
			found = true
			break
		}
	}

	if !found {
		return false, errors.New("Cound't find ready condition")
	}

	if readycondition.Status != apiv1.ConditionTrue {
		return false, errors.New(readycondition.Message)
	}

	return true, nil
}
