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

package ecsdaemonmanager

import (
	"fmt"

	apicontainer "github.com/aws/amazon-ecs-agent/agent/api/container"
	apitask "github.com/aws/amazon-ecs-agent/agent/api/task"
	"github.com/aws/amazon-ecs-agent/agent/config"
	"github.com/aws/amazon-ecs-agent/agent/taskresource"
	apicontainerstatus "github.com/aws/amazon-ecs-agent/ecs-agent/api/container/status"
	apitaskstatus "github.com/aws/amazon-ecs-agent/ecs-agent/api/task/status"
)

type Manager interface {
	GetPauseContainerName() string
	CreatePauseContainer(cfg *config.Config) *apicontainer.Container
	GetPauseTask() *apitask.Task
}

type ecsDaemonManager struct {
	pauseContainerName string
	address            string

	pauseContainer *apicontainer.Container
	pauseTask      *apitask.Task
}

func NewEcsDaemonManager() *ecsDaemonManager {
	return &ecsDaemonManager{}
}

func (edm *ecsDaemonManager) GetPauseContainerName() string {
	return edm.pauseContainerName
}

func (edm *ecsDaemonManager) CreatePauseContainer(cfg *config.Config) *apicontainer.Container {
	if edm.pauseContainer == nil {
		edm.initializePauseContainer(cfg)
	}
	return edm.pauseContainer
}

func (edm *ecsDaemonManager) initializePauseContainer(cfg *config.Config) {
	edm.pauseContainer = apicontainer.NewContainerWithSteadyState(apicontainerstatus.ContainerResourcesProvisioned)
	edm.pauseContainer.TransitionDependenciesMap = make(map[apicontainerstatus.ContainerStatus]apicontainer.TransitionDependencySet)
	edm.pauseContainer.Name = edm.pauseContainerName
	edm.pauseContainer.Image = fmt.Sprintf("%s:%s", cfg.PauseContainerImageName, cfg.PauseContainerTag)
	edm.pauseContainer.Essential = true
	edm.pauseContainer.Type = apicontainer.ContainerCNIPause

	edm.pauseTask = &apitask.Task{
		Arn:                 fmt.Sprintf("arn:::::/daemon-pause-task"),
		DesiredStatusUnsafe: apitaskstatus.TaskRunning,
		Containers:          []*apicontainer.Container{edm.pauseContainer},
		LaunchType:          "EC2",
		NetworkMode:         "none",
		ResourcesMapUnsafe:  make(map[string][]taskresource.TaskResource),
		IsInternal:          true,
	}

	edm.pauseTask.Containers = append(edm.pauseTask.Containers, edm.pauseContainer)
}

func (edm *ecsDaemonManager) GetPauseTask() *apitask.Task {
	return edm.pauseTask
}
