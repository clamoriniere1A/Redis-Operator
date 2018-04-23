package utils

import (
	"k8s.io/api/core/v1"

	rapi "github.com/amadeusitgroup/redis-operator/pkg/api/redis/v1"
)

// GetClusterStatus used to return the status of a RedisCluster
func GetClusterStatus(rc *rapi.RedisCluster) rapi.RedisClusterConditionType {
	status := rapi.RedisClusterScaling

	if hasStatus(rc, rapi.RedisClusterRollingUpdate, v1.ConditionTrue) {
		status = rapi.RedisClusterRollingUpdate
	} else if hasStatus(rc, rapi.RedisClusterScaling, v1.ConditionTrue) {
		status = rapi.RedisClusterScaling
	} else if hasStatus(rc, rapi.RedisClusterRebalancing, v1.ConditionTrue) {
		status = rapi.RedisClusterRebalancing
	} else if hasStatus(rc, rapi.RedisClusterOK, v1.ConditionTrue) {
		if rc.Spec.NumberOfMaster != nil && rc.Spec.ReplicationFactor != nil &&
			rc.Status.Cluster.NumberOfMaster == *rc.Spec.NumberOfMaster &&
			rc.Status.Cluster.MaxReplicationFactor == rc.Status.Cluster.MinReplicationFactor &&
			rc.Status.Cluster.MaxReplicationFactor == *rc.Spec.ReplicationFactor {
			status = rapi.RedisClusterOK
		}
	}

	return status
}

func hasStatus(rc *rapi.RedisCluster, conditionType rapi.RedisClusterConditionType, status v1.ConditionStatus) bool {
	for _, cond := range rc.Status.Conditions {
		if cond.Type == conditionType && cond.Status == status {
			return true
		}
	}
	return false
}
