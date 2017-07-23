package types

// RunningTask represents a task in the docker swarm cluster
type RunningTask struct {
	ID          string
	NodeID      string
	ServiceID   string
	ContainerID string
}
