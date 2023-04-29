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
	"github.com/aws/amazon-ecs-agent/agent/utils"
)

// HostResourceManager keeps account of each task in
type HostResourceManager struct {
	hostResource              map[string]ecs.Resource
	consumedResource          map[string]ecs.Resource
	hostResourceManagerRWLock sync.Mutex

	taskConsumed map[string]bool //task.arn to boolean whether host resources consumed or not
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
