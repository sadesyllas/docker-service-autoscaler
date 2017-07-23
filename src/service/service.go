package service

import (
	"math"
	"sort"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"../cluster"
	"../types"
)

var (
	config              types.ServicesConfig
	scaleOutStagingArea map[string]types.ServiceStagedScaling
	scaleInStagingArea  map[string]types.ServiceStagedScaling
)

// ScaleServices scales the autoscaled services based on the provided configuration
func ScaleServices() {
	wg := sync.WaitGroup{}
	for _, s := range config.Services {
		wg.Add(1)
		scaleService(s, &wg)
	}
	wg.Wait()
}

// UpdateConfig reads a json configuration file at configPath and caches the parsed ServicesConfig object
func UpdateConfig(configPath string) {
	config = types.ServicesConfig{
		Services: []types.ServiceConfig{
			types.ServiceConfig{
				Name:        "portainer",
				MinReplicas: 3,
				MaxReplicas: 5,
				ScaleOut: types.ServiceScaleConditions{
					CPU:    20,
					Memory: 50,
					Period: "1m",
				},
				ScaleIn: types.ServiceScaleConditions{
					CPU:    10,
					Memory: 25,
					Period: "1m",
				},
				NodeLabel: "portainer",
			},
		},
	}
}

// scaleService
func scaleService(serviceConfig types.ServiceConfig, wg *sync.WaitGroup) {
	defer wg.Done()

	serviceState := getServiceState(serviceConfig.Name)
	runningServiceInstancesCount := len(serviceState.RunningServiceInstances)
	serviceID := serviceState.Service.ID

	if runningServiceInstancesCount < serviceConfig.MinReplicas {
		newNodesNeeded := serviceConfig.MinReplicas - runningServiceInstancesCount
		newNodes := getNewNodesForService(serviceState, cluster.GetRunningActiveNodes(), newNodesNeeded)

		if len(newNodes) == 0 {
			log.Warnf("Needed to start %d new instances for service %s but no nodes are available",
				newNodesNeeded, serviceState.Service.Name)

			return
		}

		startServiceOnNodes(serviceConfig, newNodes)

		log.Infof("Started %d new instances for service %s", newNodesNeeded, serviceState.Service.Name)

		delete(scaleOutStagingArea, serviceID)
		delete(scaleInStagingArea, serviceID)

		return
	}

	healthyServiceNodes, sickServiceNodes := categorizeNodesForService(serviceConfig, serviceState)
	healthyServiceNodesCount, _ := len(healthyServiceNodes), len(sickServiceNodes)

	if healthyServiceNodesCount < serviceConfig.MinReplicas {
		if runningServiceInstancesCount == serviceConfig.MaxReplicas {
			log.Warnf("Scaling needed to match minimum healthy for service %s but already using max replicas",
				serviceState.Service.Name)

			return
		}

		scaleOutPeriod, _ := time.ParseDuration(serviceConfig.ScaleOut.Period)
		doScaleOut := false

		if scaleOutPeriod.Seconds() != 0.0 {
			if s, ok := scaleOutStagingArea[serviceID]; ok {
				if float64(time.Now().Unix()-s.StagedTimestamp) >= scaleOutPeriod.Seconds() {
					doScaleOut = true
				}
			} else {
				scaleOutStagingArea[serviceID] = types.ServiceStagedScaling{
					ServiceID:       serviceID,
					StagedTimestamp: time.Now().Unix(),
				}
			}
		} else {
			doScaleOut = true
		}

		if doScaleOut {
			newNodesNeededCount := serviceConfig.MinReplicas - healthyServiceNodesCount
			newNodesAllowedCount := serviceConfig.MaxReplicas - runningServiceInstancesCount

			if newNodesNeededCount > newNodesAllowedCount {
				log.Warnf("Scaling by %d nodes needed for service %s but only %d more nodes can be used",
					newNodesNeededCount,
					newNodesAllowedCount,
					serviceState.Service.Name)

				newNodesNeededCount = newNodesAllowedCount
			}

			if newNodesNeededCount == 0 {
				return
			}

			newNodes := getNewNodesForService(serviceState, cluster.GetRunningActiveNodes(), newNodesNeededCount)
			startServiceOnNodes(serviceConfig, newNodes)

			log.Infof("Started %d new instances for service %s because only %d instances are healthy", newNodesNeededCount, serviceState.Service.Name, healthyServiceNodesCount)

			delete(scaleOutStagingArea, serviceID)
			delete(scaleInStagingArea, serviceID)
		}

		return
	}

	if runningServiceInstancesCount == serviceConfig.MinReplicas {
		log.Infof("No scaling needed for service %s", serviceState.Service.Name)

		delete(scaleOutStagingArea, serviceID)
		delete(scaleInStagingArea, serviceID)

		return
	}

	// at this point we have more healthy instances than needed so we must scale in

	scaleInPeriod, _ := time.ParseDuration(serviceConfig.ScaleIn.Period)
	doScaleIn := false

	if scaleInPeriod.Seconds() != 0.0 {
		if s, ok := scaleInStagingArea[serviceID]; ok {
			if float64(time.Now().Unix()-s.StagedTimestamp) >= scaleInPeriod.Seconds() {
				doScaleIn = true
			}
		} else {
			scaleInStagingArea[serviceID] = types.ServiceStagedScaling{
				ServiceID:       serviceID,
				StagedTimestamp: time.Now().Unix(),
			}
		}
	} else {
		doScaleIn = true
	}

	if doScaleIn {
		extraNodesCount := healthyServiceNodesCount - serviceConfig.MinReplicas
		nodes := getLeastLoadedNodes(serviceState, extraNodesCount)

		stopServiceOnNodes(serviceConfig, nodes)

		delete(scaleOutStagingArea, serviceID)
		delete(scaleInStagingArea, serviceID)
	}
}

// getServiceState
func getServiceState(serviceName string) types.ServiceState {
	var result types.ServiceState

	clusterState := cluster.GetState()

	var serviceID string
	for _, s := range clusterState.Services {
		if s.Name == serviceName {
			serviceID = s.ID

			break
		}
	}

	// no service found with this name
	if serviceID == "" {
		return result
	}

	runningTasks := []types.RunningTask{}
	runningServiceInstances := []types.RunningServiceInstance{}
	for _, t := range clusterState.RunningTasks {
		if t.ServiceID == serviceID {
			containerStats := cluster.GetContainerStats(t.ContainerID)

			if containerStats == nil {
				continue
			}

			runningTasks = append(runningTasks, t)

			runningServiceInstance := types.RunningServiceInstance{
				Node:           clusterState.RunningActiveNodes[t.NodeID],
				ContainerStats: *containerStats,
			}

			runningServiceInstances = append(runningServiceInstances, runningServiceInstance)
		}
	}

	if len(runningTasks) == 0 {
		return result
	}

	result = types.ServiceState{
		Service:                 clusterState.Services[serviceID],
		RunningServiceInstances: runningServiceInstances,
	}

	return result
}

// getNewNodesForService
func getNewNodesForService(serviceState types.ServiceState, allNodes []types.Node, count int) (nodes []string) {
	nodes = []string{}

	serviceNodesMap := map[string]bool{}

	for _, r := range serviceState.RunningServiceInstances {
		serviceNodesMap[r.Node.ID] = true
	}

	for _, n := range allNodes {
		if _, isServiceOnNode := serviceNodesMap[n.ID]; !isServiceOnNode {
			nodes = append(nodes, n.ID)

			if len(nodes) == count {
				break
			}
		}
	}

	return nodes
}

// categorizeNodesForService
func categorizeNodesForService(serviceConfig types.ServiceConfig, serviceState types.ServiceState) (
	healthy []string, sick []string) {
	healthy = []string{}
	sick = []string{}

	for _, r := range serviceState.RunningServiceInstances {
		cpuOk := r.ContainerStats.Usage.CPU <= serviceConfig.ScaleOut.CPU
		memoryOk := r.ContainerStats.Usage.Memory <= serviceConfig.ScaleOut.Memory
		if cpuOk && memoryOk {
			healthy = append(healthy, r.Node.ID)
		} else {
			sick = append(sick, r.Node.ID)
		}
	}

	return
}

// getLeastLoadedNodes
func getLeastLoadedNodes(serviceState types.ServiceState, count int) (nodes []string) {
	nodes = []string{}

	runningServiceInstances := make([]types.RunningServiceInstance, len(serviceState.RunningServiceInstances))
	copy(runningServiceInstances, serviceState.RunningServiceInstances)

	sort.SliceStable(runningServiceInstances, func(i, j int) bool {
		rsi1 := runningServiceInstances[i]
		rsi2 := runningServiceInstances[j]

		if rsi1.ContainerStats.Usage.CPU == rsi2.ContainerStats.Usage.CPU {
			return rsi1.ContainerStats.Usage.Memory < rsi2.ContainerStats.Usage.Memory
		}

		return rsi1.ContainerStats.Usage.CPU < rsi2.ContainerStats.Usage.CPU
	})

	count = int(math.Max(float64(count), float64(len(runningServiceInstances))))

	for i := 0; i < count; i++ {
		nodes = append(nodes, runningServiceInstances[i].Node.ID)
	}

	return nodes
}

// startServiceOnNodes
func startServiceOnNodes(serviceConfig types.ServiceConfig, nodes []string) {
	if len(nodes) == 0 {
		return
	}

	log.Infof("starting service %s on %d nodes with label %s", serviceConfig.Name, len(nodes), serviceConfig.NodeLabel)

	for _, n := range nodes {
		cluster.AddLabelToNode(n, serviceConfig.NodeLabel, "1")
	}
}

// stopServiceOnNodes
func stopServiceOnNodes(serviceConfig types.ServiceConfig, nodes []string) {
	if len(nodes) == 0 {
		return
	}

	log.Infof("stopping service %s on %d nodes with label %s", serviceConfig.Name, len(nodes), serviceConfig.NodeLabel)

	for _, n := range nodes {
		cluster.RemoveLabelFromNode(n, serviceConfig.NodeLabel)
	}
}
