// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//	http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

// Package engine contains the core logic for managing tasks

package engine

import (
	"sync"

	"github.com/aws/amazon-ecs-agent/agent/ecs_client/model/ecs"
	"github.com/aws/amazon-ecs-agent/agent/logger"
	"github.com/aws/amazon-ecs-agent/agent/utils"
	"github.com/pkg/errors"
)

// HostResourceManager keeps account of each task in
type HostResourceManager struct {
	hostResource              map[string]ecs.Resource
	consumedResource          map[string]ecs.Resource
	hostResourceManagerRWLock sync.Mutex

	taskConsumed map[string]bool //task.arn to boolean whether host resources consumed or not
}

func (h *HostResourceManager) consume(taskArn string, resources map[string]ecs.Resource) (bool, error) {
	h.hostResourceManagerRWLock.Lock()
	defer h.hostResourceManagerRWLock.Unlock()

	ok, err := h.consumable(resources)
	if err != nil {
		logger.Info("Consume error")
		return false, err
	}
	if ok {
		// CPU
		*h.consumedResource["CPU"].IntegerValue += *resources["CPU"].IntegerValue
		// MEM
		*h.consumedResource["MEMORY"].IntegerValue += *resources["MEMORY"].IntegerValue
		// PORTS
		portsResource, ok := resources["PORTS"]
		if ok {
			taskPortsSlice := portsResource.StringSetValue
			for _, port := range taskPortsSlice {
				// Create a copy to assign it back as "PORTS"
				newPortResource := h.consumedResource["PORTS"]
				newPorts := append(h.consumedResource["PORTS"].StringSetValue, port)
				newPortResource.StringSetValue = newPorts
				h.consumedResource["PORTS"] = newPortResource
			}
		}
		// PORTS_UDP
		portsResource, ok = resources["PORTS_UDP"]
		if ok {
			taskPortsSlice := portsResource.StringSetValue
			for _, port := range taskPortsSlice {
				newPortResource := h.consumedResource["PORTS_UDP"]
				newPorts := append(h.consumedResource["PORTS_UDP"].StringSetValue, port)
				newPortResource.StringSetValue = newPorts
				h.consumedResource["PORTS_UDP"] = newPortResource
			}
		}
		// GPU
		*h.hostResource["GPU"].IntegerValue += *resources["GPU"].IntegerValue
		h.taskConsumed[taskArn] = true
		logger.Info("Consume true okay")
		return true, nil
	}
	logger.Info("Consume false okay")
	return false, nil
}

// Helper function for consume to check if resources are consumable with the current account
// we have for the host resources. Should not call host resource manager lock in this func
// return values
// false, nil -> did not consume, should stay pending
// false, err -> resources map has errors
// true, nil -> consumed
func (h *HostResourceManager) consumable(resources map[string]ecs.Resource) (bool, error) {
	// CPU
	// TODO: CPU might be optional, to verify
	cpuResource, ok := resources["CPU"]
	if ok {
		if *(h.hostResource["CPU"].IntegerValue) < *(h.consumedResource["CPU"].IntegerValue)+*(cpuResource.IntegerValue) {
			logger.Info("Unable to consume CPU")
			return false, nil
		}
	} else {
		return false, errors.New("No CPU in task resources")
	}

	// MEM
	memResource, ok := resources["MEMORY"]
	if ok {
		if *(h.hostResource["MEMORY"].IntegerValue) < *(h.consumedResource["MEMORY"].IntegerValue)+*(memResource.IntegerValue) {
			logger.Info("Unable to consume MEMORY")
			return false, nil
		}
	} else {
		return false, errors.New("No MEMORY in task resources")
	}

	// PORTS
	portsResource, ok := resources["PORTS"]
	if ok {
		taskPortsSlice := portsResource.StringSetValue
		// For each port in current resource object, check if it is already 'consumed'. This is
		// done by maintaining a list of consumed ports, i.e. list of all ports reserved by tasks
		// being accounted for by host resources manager.
		for _, port := range taskPortsSlice {
			for _, consumedPort := range h.consumedResource["PORTS"].StringSetValue {
				// If port is already reserved by some other task, this 'resources' object can not be consumed
				if *port == *consumedPort {
					logger.Info("Unable to consume PORTS")
					return false, nil
				}
			}
		}
	}

	// PORTS_UDP
	portsUDPResource, ok := resources["PORTS_UDP"]
	if ok {
		taskPortsUDPSlice := portsUDPResource.StringSetValue
		// Same for UDP ports
		for _, port := range taskPortsUDPSlice {
			for _, consumedPort := range h.consumedResource["PORTS_UDP"].StringSetValue {
				if *port == *consumedPort {
					logger.Info("Unable to consume PORTS_UDP")
					return false, nil
				}
			}
		}
	}
	// GPU
	gpuResouce, ok := resources["GPU"]
	if ok {
		if *(h.hostResource["GPU"].IntegerValue) < *(h.consumedResource["GPU"].IntegerValue)+*(gpuResouce.IntegerValue) {
			logger.Info("Unable to consume GPU")
			return false, nil
		}
	}
	logger.Info("Able to consume resources")
	return true, nil
}

func (h *HostResourceManager) release(taskArn string, resources map[string]ecs.Resource) {
	h.hostResourceManagerRWLock.Lock()
	defer h.hostResourceManagerRWLock.Unlock()

	if h.taskConsumed[taskArn] {
		// CPU
		*h.consumedResource["CPU"].IntegerValue -= *resources["CPU"].IntegerValue
		// MEM
		*h.consumedResource["MEMORY"].IntegerValue -= *resources["MEMORY"].IntegerValue
		// PORTS
		portsResource, ok := resources["PORTS"]
		if ok {
			taskPortsSlice := portsResource.StringSetValue
			// Start removing ports one by one
			for _, port := range taskPortsSlice {
				// Create a copy to assign it back as "PORTS"
				newPortResource := h.consumedResource["PORTS"]
				// Create a copy of ports slice, iterate and find the port index in this slice
				// then remove the port, create a new slice and assign it back to consumedResource["PORTS"]
				itrPortSlice := h.consumedResource["PORTS"].StringSetValue
				idx := 0
				for i, consumedPort := range itrPortSlice {
					if *consumedPort == *port {
						idx = i
					}
				}
				newPortSlice := append(itrPortSlice[:idx], itrPortSlice[idx+1:]...)
				newPortResource.StringSetValue = newPortSlice
				h.consumedResource["PORTS"] = newPortResource
			}
		}
		// PORTS_UDP
		// Above logic for UDP
		portsResource, ok = resources["PORTS_UDP"]
		if ok {
			taskPortsSlice := portsResource.StringSetValue
			for _, port := range taskPortsSlice {
				newPortResource := h.consumedResource["PORTS_UDP"]
				itrPortSlice := h.consumedResource["PORTS_UDP"].StringSetValue
				idx := 0
				for i, consumedPort := range itrPortSlice {
					if *consumedPort == *port {
						idx = i
					}
				}
				newPortSlice := append(itrPortSlice[:idx], itrPortSlice[idx+1:]...)
				newPortResource.StringSetValue = newPortSlice
				h.consumedResource["PORTS_UDP"] = newPortResource
			}
		}
		// GPU
		*h.consumedResource["GPU"].IntegerValue -= *resources["GPU"].IntegerValue
		delete(h.taskConsumed, taskArn)
	}
}

func NewHostResourceManager(hostResource []*ecs.Resource, totalGPU int64) HostResourceManager {
	resourceMap := make(map[string]ecs.Resource)
	consumedResourceMap := make(map[string]ecs.Resource)
	// assignes CPU, MEMORY, PORTS, PORTS_UDP from host
	for _, resource := range hostResource {
		resourceMap[*resource.Name] = *resource
	}
	resourceMap["GPU"] = ecs.Resource{
		Name:         utils.Strptr("GPU"),
		Type:         utils.Strptr("INTEGER"),
		IntegerValue: &totalGPU,
	}
	return HostResourceManager{
		hostResource:     resourceMap,
		consumedResource: consumedResourceMap,
	}
}
