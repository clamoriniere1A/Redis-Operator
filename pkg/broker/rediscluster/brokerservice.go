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
								"properties": getProperties(),
							},
						},
					},
				},
			},
		},
	}
}

func getProperties() properties {
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
		numberOfMasterParameterKey: propertie{
			Title:   "Number of masters",
			Type:    "integer",
			Default: 3,
		},
		replicationFactorParameterKey: propertie{
			Title:   "Replication Factor",
			Type:    "integer",
			Default: 1,
		},
	}

	return props
}
