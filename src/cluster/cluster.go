package cluster

import (
	"../client"
	"../types"
	"../utils"
)

var state types.ClusterState = types.NewClusterState()

// GetState returns the most recently updated swarm cluster state
func GetState() types.ClusterState {
	return state
}

// GetRunningActiveNodes returns all the running and active nodes in the swarm cluster
func GetRunningActiveNodes() []types.Node {
	nodes, err := client.GetRunningActiveNodes()

	if err != nil {
		return nil
	}

	return nodes
}

// GetContainerStats uses the client package to retreive a container's usage statistics
func GetContainerStats(containerID string) *types.ContainerStats {
	stats, err := client.GetContainerStats(containerID)

	if err != nil {
		return nil
	}

	return &types.ContainerStats{
		Raw:   stats,
		Usage: utils.ExtractContainerResourceUsage(stats),
	}
}

// UpdateState updates the cluster state snapshot kept in memory
func UpdateState() {
	tasks, err := client.GetRunningTasks()
	utils.PanicOnError(err)

	for _, t := range tasks {
		state.RunningTasks[t.ID] = t
	}

	services, err := client.GetServices()
	utils.PanicOnError(err)

	for _, s := range services {
		state.Services[s.ID] = s
	}

	nodes, err := client.GetRunningActiveNodes()
	utils.PanicOnError(err)

	for _, n := range nodes {
		state.RunningActiveNodes[n.ID] = n
	}
}

// AddLabelToNode adds a label to a swarm cluster node
func AddLabelToNode(nodeID string, label string, value string) {
	client.AddLabelToNode(nodeID, label, value)
}

// RemoveLabelFromNode removes a label from a swarm cluster node
func RemoveLabelFromNode(nodeID string, label string) {
	client.RemoveLabelFromNode(nodeID, label)
}
