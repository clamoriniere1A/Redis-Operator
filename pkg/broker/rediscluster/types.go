package rediscluster

// Instance rediscluster instance information
type Instance struct {
	Name      string
	Namespace string

	ServiceName string
	RequirePass string
}

// ServiceInstanceParameter rediscluster instance parameters
type ServiceInstanceParameter struct {
	ClusterName       string  `json:"cluster-name,omitempty"`
	ClusterNamespace  string  `json:"cluster-namespace,omitempty"`
	NumberOfMasters   *uint32 `json:"number-of-masters,omitempty"`
	ReplicationFactor *uint32 `json:"replication-factor,omitempty"`
	MemoryByNode      *uint32 `json:"memory-by-node,omitempty"`
	Proxy             bool    `json:"proxy,omitempty"`
	Monitoring        bool    `json:"monitoring,omitempty"`
	Requirepass       string  `json:"requirepass,omitempty"`
}

type properties map[parameterKey]propertie

type propertie struct {
	Title     string        `json:"title"`
	Type      string        `json:"type"`
	Default   interface{}   `json:"default,omitempty"`
	MaxLength uint32        `json:"maxLength,omitempty"`
	Enum      []interface{} `json:"enum,omitempty"`
}

type provisioningType string

const (
	provisioningSelfHosted    provisioningType = "self-hosted"
	provisioningUserNamespace provisioningType = "user-namespace"
)

type parameterKey string

const (
	clusterNameParameterKey       parameterKey = "cluster-name"
	provisioningParameterKey      parameterKey = "provisioning"
	numberOfMasterParameterKey    parameterKey = "number-of-masters"
	replicationFactorParameterKey parameterKey = "replication-factor"
)
