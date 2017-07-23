package types

// Node represents a node in the docker swarm cluster
type Node struct {
	ID       string
	IP       string
	Hostname string
	Role     string
	Leader   bool
}
