package rediscluster

import "github.com/pmorie/go-open-service-broker-client/v2"

// GenerateBrokerService Used to generate the open service brocker service definition for the redis-cluster broker
func GenerateBrokerService() *v2.Service {
	boolFalse := false
	boolTrue := true
	return &v2.Service{
		Name:          "redis-broker-service",
		ID:            "redis-broker-serviceclasse",
		Description:   "Services from the redis-cluster broker!",
		Bindable:      true,
		PlanUpdatable: &boolFalse,
		Metadata: map[string]interface{}{
			"displayName": "redis-broker-service",
			"imageUrl":    "https://cdn4.iconfinder.com/data/icons/redis-2/1451/Untitled-2-256.png",
		},
		Plans: []v2.Plan{
			{
				Name:        "default",
				ID:          "redis-cluster-default-plan",
				Description: "The default plan for the redis broker service",
				Free:        &boolTrue,
				ParameterSchemas: &v2.ParameterSchemas{
					ServiceInstances: &v2.ServiceInstanceSchema{
						Create: &v2.InputParameters{
							Parameters: map[string]interface{}{
								"$schema":    "http://json-schema.org/draft-04/schema",
								"type":       "object",
								"title":      "Parameters",
								"properties": getPropertiesForStaticConf(),
							},
						},
					},
				},
			},
			{
				Name:        "auto-scale",
				ID:          "redis-cluster-autoscale-plan",
				Description: "`Redis-Cluster scale up and down is done automaticaly by the redis-operator thanks to prometheus metrics",
				Free:        &boolTrue,
				ParameterSchemas: &v2.ParameterSchemas{
					ServiceInstances: &v2.ServiceInstanceSchema{
						Create: &v2.InputParameters{
							Parameters: map[string]interface{}{
								"$schema":    "http://json-schema.org/draft-04/schema",
								"type":       "object",
								"title":      "Parameters",
								"properties": getDefaultProperties(),
							},
						},
					},
				},
			},
		},
	}
}

func getPropertiesForStaticConf() properties {
	props := getDefaultProperties()

	props[numberOfMasterParameterKey] = propertie{
		Title:   "Number of masters",
		Type:    "integer",
		Default: 3,
	}
	props[replicationFactorParameterKey] = propertie{
		Title:   "Replication Factor",
		Type:    "integer",
		Default: 1,
	}

	return props
}

func getDefaultProperties() properties {
	props := properties{
		clusterNameParameterKey: propertie{
			Title:     "Redis cluster name",
			Type:      "string",
			MaxLength: 40,
			Default:   "redis-cluster",
		},
		provisioningParameterKey: propertie{
			Title:   "provisioning mode",
			Type:    "string",
			Default: "self-hosted",
			Enum:    []interface{}{provisioningSelfHosted, provisioningUserNamespace},
		},
		memoryByNodeParameterKey: propertie{
			Title:     "Memory by Redis Node",
			Type:      "string",
			MaxLength: 10,
			MinLength: 0,
			Default:   "1Gi",
		},
		cpuByNodeParameterKey: propertie{
			Title:     "CPU by Redis Node",
			Type:      "string",
			MaxLength: 10,
			MinLength: 0,
			Default:   "1",
		},
	}

	return props
}
