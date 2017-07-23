package types

// Service models a docker service
type Service struct {
	ID   string
	Name string
}

// RunningServiceInstance represents a running service instance on a particular node in a swarm cluster with the resources it consumers on the node
type RunningServiceInstance struct {
	Node           Node
	ContainerStats ContainerStats
}

// ServiceState represents the running state of a service in a swarm cluster
type ServiceState struct {
	Service                 Service
	RunningServiceInstances []RunningServiceInstance
}

// ServicesConfig represents the deserialized service configuration json passed to the program
type ServicesConfig struct {
	Services []ServiceConfig `json:"services"`
}

// ServiceConfig represents the configuration section for a single service in the ServicesConfig object
type ServiceConfig struct {
	Name        string                 `json:"name"`
	MinReplicas int                    `json:"min_replicas"`
	MaxReplicas int                    `json:"max_replicas"`
	ScaleOut    ServiceScaleConditions `json:"scale_out"`
	ScaleIn     ServiceScaleConditions `json:"scale_in"`
	NodeLabel   string                 `json:"node_label"`
}

// ServiceScaleConditions represents the resource usage that triggers a scale out/in for a service
type ServiceScaleConditions struct {
	CPU    float64 `json:"cpu"`
	Memory float64 `json:"memory"`
	Period string  `json:"period"`
}

// ServiceStagedScaling represents a scale out/in operation that has been staged to be completed
type ServiceStagedScaling struct {
	ServiceID       string
	StagedTimestamp int64
}
