/*
 * Copyright (c) 2026 dingodb.com, Inc. All Rights Reserved
*
*  Licensed under the Apache License, Version 2.0 (the "License");
*  you may not use this file except in compliance with the License.
*  You may obtain a copy of the License at
*
*      http://www.apache.org/licenses/LICENSE-2.0
*
*  Unless required by applicable law or agreed to in writing, software
*  distributed under the License is distributed on an "AS IS" BASIS,
*  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
*  See the License for the specific language governing permissions and
*  limitations under the License.
*/

package monitor

import (
	"fmt"

	"github.com/dingodb/dingocli/cli/cli"
	"github.com/dingodb/dingocli/internal/configure"
	"github.com/dingodb/dingocli/internal/task/step"
	"github.com/dingodb/dingocli/internal/task/task"
	"github.com/dingodb/dingocli/internal/task/task/common"
	tui "github.com/dingodb/dingocli/internal/tui/common"
)

func NewStopServiceTask(dingocli *cli.DingoCli, cfg *configure.MonitorConfig) (*task.Task, error) {
	serviceId := dingocli.GetServiceId(cfg.GetId())
	containerId, err := dingocli.GetContainerId(serviceId)
	if IsSkip(cfg, []string{ROLE_MONITOR_CONF}) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	hc, err := dingocli.GetHost(cfg.GetHost())
	if err != nil {
		return nil, err
	}

	// new task
	subname := fmt.Sprintf("host=%s role=%s containerId=%s",
		cfg.GetHost(), cfg.GetRole(), tui.TrimContainerId(containerId))
	t := task.NewTask("Stop Service", subname, hc.GetSSHConfig())

	// add step to task
	var out string
	role, host := cfg.GetRole(), cfg.GetHost()
	t.AddStep(&step.ListContainers{
		ShowAll:     true,
		Format:      `"{{.ID}}"`,
		Filter:      fmt.Sprintf("id=%s", containerId),
		Out:         &out,
		ExecOptions: dingocli.ExecOptions(),
	})
	t.AddStep(&step.Lambda{
		Lambda: common.CheckContainerExist(host, role, containerId, &out),
	})
	t.AddStep(&step.StopContainer{
		ContainerId: containerId,
		Out:         &out,
		ExecOptions: dingocli.ExecOptions(),
	})
	return t, nil
}
