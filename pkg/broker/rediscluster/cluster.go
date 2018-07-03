package rediscluster

import (
	"fmt"

	"k8s.io/api/core/v1"
	kresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/amadeusitgroup/redis-operator/pkg/api/redis"
	rapi "github.com/amadeusitgroup/redis-operator/pkg/api/redis/v1"
)

// NewRedisCluster builds and returns a new RedisCluster instance
func NewRedisCluster(name, namespace, instanceID, tag string, nbMaster, replication int32, resourceConf ClusterNodeResourcesConf) (*rapi.RedisCluster, error) {
	resourcesRequirements, err := GetResourcesRequirements(resourceConf)
	if err != nil {
		return nil, err
	}
	rc := &rapi.RedisCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       rapi.ResourceKind,
			APIVersion: redisk8soperatorio.GroupName + "/" + rapi.ResourceVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: map[string]string{"cluster-instance-id": instanceID},
		},
		Spec: rapi.RedisClusterSpec{
			ServiceName:       fmt.Sprintf("%s-service", name),
			NumberOfMaster:    &nbMaster,
			ReplicationFactor: &replication,
			PodTemplate: &v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": name,
					},
					Annotations: map[string]string{
						"cluster-instance-id": instanceID,
					},
				},
				Spec: v1.PodSpec{
					ServiceAccountName: "redis-node",
					Volumes: []v1.Volume{
						{Name: "data"},
					},
					Containers: []v1.Container{
						{
							Name:            "redis-node",
							Image:           fmt.Sprintf("redisoperator/redisnode:%s", tag),
							ImagePullPolicy: v1.PullIfNotPresent,
							Args: []string{
								"--v=6",
								"--logtostderr=true",
								"--alsologtostderr=true",
								fmt.Sprintf("--rs=%s-service", name),
								"--t=10s",
								"--d=10s",
								"--ns=$(POD_NAMESPACE)",
								"--cluster-node-timeout=2000",
							},
							Ports: []v1.ContainerPort{
								{Name: "redis", ContainerPort: 6379},
								{Name: "cluster", ContainerPort: 16379},
							},
							VolumeMounts: []v1.VolumeMount{
								{Name: "data", MountPath: "/redis-data"},
							},
							Env: []v1.EnvVar{
								{Name: "POD_NAMESPACE", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{FieldPath: "metadata.namespace"}}},
							},
							LivenessProbe: &v1.Probe{
								Handler: v1.Handler{
									HTTPGet: &v1.HTTPGetAction{
										Path: "/live",
										Port: intstr.FromInt(8080),
									},
								},
								InitialDelaySeconds: 10,
								TimeoutSeconds:      5,
								PeriodSeconds:       10,
								SuccessThreshold:    1,
								FailureThreshold:    30,
							},
							ReadinessProbe: &v1.Probe{
								Handler: v1.Handler{
									HTTPGet: &v1.HTTPGetAction{
										Path: "/ready",
										Port: intstr.FromInt(8080),
									},
								},
								TimeoutSeconds:   5,
								PeriodSeconds:    10,
								SuccessThreshold: 1,
								FailureThreshold: 3,
							},
							Resources: resourcesRequirements,
						},
					},
				},
			},
		},
	}

	// Set Redis configuration if memory limit is provided
	if resourceConf.MemoryRequest != "" {
		for id, container := range rc.Spec.PodTemplate.Spec.Containers {
			if container.Name == "redis-node" {
				// Set the limits
				limitEnvVar := v1.EnvVar{
					Name:      "MEMORY_REQUEST",
					ValueFrom: &v1.EnvVarSource{ResourceFieldRef: &v1.ResourceFieldSelector{ContainerName: "redis-node", Resource: "requests.memory"}},
				}
				rc.Spec.PodTemplate.Spec.Containers[id].Env = append(rc.Spec.PodTemplate.Spec.Containers[id].Env, limitEnvVar)

				// Set the argument
				rc.Spec.PodTemplate.Spec.Containers[id].Args = append(rc.Spec.PodTemplate.Spec.Containers[id].Args, "--max-memory=$(MEMORY_REQUEST)")
			}
		}
	}

	return rc, nil
}

// GetResourcesRequirements return the v1.ResourceRequirements struct depending ot the arguments
func GetResourcesRequirements(resourceConf ClusterNodeResourcesConf) (v1.ResourceRequirements, error) {
	resourceRQ := v1.ResourceRequirements{}
	var errs []error
	var err error
	resourceRQ.Requests, err = setResourceRequirements(v1.ResourceCPU, resourceConf.CPURequest, resourceRQ.Requests)
	if err != nil {
		errs = append(errs, err)
	}
	resourceRQ.Requests, err = setResourceRequirements(v1.ResourceMemory, resourceConf.MemoryRequest, resourceRQ.Requests)
	if err != nil {
		errs = append(errs, err)
	}
	resourceRQ.Limits, err = setResourceRequirements(v1.ResourceCPU, resourceConf.CPULimit, resourceRQ.Limits)
	if err != nil {
		errs = append(errs, err)
	}
	resourceRQ.Limits, err = setResourceRequirements(v1.ResourceMemory, resourceConf.MemoryLimit, resourceRQ.Limits)
	if err != nil {
		errs = append(errs, err)
	}

	// default limits if not provided by requested
	if resourceConf.CPURequest != "" && resourceConf.CPULimit == "" {
		resourceRQ.Limits, err = setResourceRequirements(v1.ResourceCPU, resourceConf.CPURequest, resourceRQ.Limits)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if resourceConf.MemoryRequest != "" && resourceConf.MemoryLimit == "" {
		resourceRQ.Limits, err = setResourceRequirements(v1.ResourceMemory, resourceConf.MemoryRequest, resourceRQ.Limits)
		if err != nil {
			errs = append(errs, err)
		}

	}

	return resourceRQ, errors.NewAggregate(errs)
}

func setResourceRequirements(resourceType v1.ResourceName, value string, out v1.ResourceList) (v1.ResourceList, error) {
	if value != "" {
		if out == nil {
			out = v1.ResourceList{}
		}
		q, err := kresource.ParseQuantity(value)
		if err != nil {
			return out, err
		}
		out[resourceType] = q
	}
	return out, nil
}

// ClusterNodeResourcesConf configuration for cluster node resources
type ClusterNodeResourcesConf struct {
	CPURequest    string
	CPULimit      string
	MemoryRequest string
	MemoryLimit   string
}
