package client

import (
	"bufio"
	"context"
	"encoding/json"

	"../types"
	"../utils"
	dockerTypes "github.com/docker/docker/api/types"
	dockerClient "github.com/docker/docker/client"
)

var ctx context.Context
var cli *dockerClient.Client

func init() {
	ctx = context.Background()

	cliTmp, err := dockerClient.NewEnvClient()

	if cliTmp == nil {
		panic("could not get a new env client for docker")
	}

	utils.PanicOnError(err)

	cli = cliTmp
}

// GetServices gets a list of running services in the docker swarm cluster
func GetServices() ([]types.Service, error) {
	dockerServices, err := cli.ServiceList(ctx, dockerTypes.ServiceListOptions{})

	if err != nil {
		return nil, err
	}

	services := make([]types.Service, len(dockerServices))
	for i := 0; i < len(dockerServices); i++ {
		s := dockerServices[i]
		services[i] = types.Service{
			ID:   s.ID,
			Name: s.Spec.Name,
		}
	}

	return services, nil
}

// GetRunningActiveNodes gets a list of nodes in the docker swarm cluster
func GetRunningActiveNodes() ([]types.Node, error) {
	dockerNodes, err := cli.NodeList(ctx, dockerTypes.NodeListOptions{})

	if err != nil {
		return nil, err
	}

	nodes := make([]types.Node, len(dockerNodes))
	for i := 0; i < len(dockerNodes); i++ {
		n := dockerNodes[i]

		if n.Status.State != "ready" || n.Spec.Availability != "active" {
			continue
		}

		nodes[i] = types.Node{
			ID:       n.ID,
			IP:       n.Status.Addr,
			Hostname: n.Description.Hostname,
			Role:     string(n.Spec.Role),
			Leader:   n.ManagerStatus != nil && n.ManagerStatus.Leader,
		}
	}

	return nodes, nil
}

// GetRunningTasks gets a list of tasks in the docker swarm cluster
func GetRunningTasks() ([]types.RunningTask, error) {
	dockerTasks, err := cli.TaskList(ctx, dockerTypes.TaskListOptions{})

	if err != nil {
		return nil, err
	}

	tasks := make([]types.RunningTask, len(dockerTasks))
	cnt := len(tasks)
	for i := 0; i < len(dockerTasks); i++ {
		t := dockerTasks[i]
		if t.Status.State != "running" {
			continue
		}
		tasks[cnt-1] = types.RunningTask{
			ID:          t.ID,
			NodeID:      t.NodeID,
			ServiceID:   t.ServiceID,
			ContainerID: t.Status.ContainerStatus.ContainerID,
		}
		cnt--
	}

	tasks = tasks[cnt:]

	return tasks, nil
}

// GetContainerStats retrieves usages statistics for a particular node in a swarm cluster
func GetContainerStats(containerID string) (types.ContainerStatsRaw, error) {
	var result types.ContainerStatsRaw

	stats, err := cli.ContainerStats(ctx, containerID, false)

	if err != nil {
		return result, err
	}

	defer stats.Body.Close()

	s := bufio.NewScanner(stats.Body)

	s.Scan()

	jerr := json.Unmarshal(s.Bytes(), &result)

	if jerr != nil {
		return result, jerr
	}

	return result, nil
}

// AddLabelToNode adds a label to a swarm cluster node
func AddLabelToNode(nodeID string, label string, value string) {
	s, err := cli.SwarmInspect(ctx)

	if err != nil {
		return
	}

	n, _, err := cli.NodeInspectWithRaw(ctx, nodeID)

	if err != nil {
		return
	}

	if n.Spec.Labels == nil {
		n.Spec.Labels = map[string]string{}
	}

	n.Spec.Labels[label] = value

	cli.NodeUpdate(ctx, nodeID, s.Version, n.Spec)
}

// RemoveLabelFromNode removes a label from a swarm cluster node
func RemoveLabelFromNode(nodeID string, label string) {
	s, err := cli.SwarmInspect(ctx)

	if err != nil {
		return
	}

	n, _, err := cli.NodeInspectWithRaw(ctx, nodeID)

	if err != nil || n.Spec.Labels == nil {
		return
	}

	delete(n.Spec.Labels, label)

	cli.NodeUpdate(ctx, nodeID, s.Version, n.Spec)
}
