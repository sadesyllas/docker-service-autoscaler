package types

// ClusterState represents all the objects currently in the swarm cluster
type ClusterState struct {
	RunningTasks       map[string]RunningTask
	Services           map[string]Service
	RunningActiveNodes map[string]Node
}

// NewClusterState creates a new ClusterState object
func NewClusterState() ClusterState {
	return ClusterState{
		RunningTasks:       map[string]RunningTask{},
		Services:           map[string]Service{},
		RunningActiveNodes: map[string]Node{},
	}
}
