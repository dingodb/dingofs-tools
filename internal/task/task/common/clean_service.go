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

package common

import (
	"fmt"
	"strings"

	"github.com/dingodb/dingocli/cli/cli"
	comm "github.com/dingodb/dingocli/internal/common"
	"github.com/dingodb/dingocli/internal/configure/topology"
	"github.com/dingodb/dingocli/internal/storage"
	"github.com/dingodb/dingocli/internal/task/context"
	"github.com/dingodb/dingocli/internal/task/step"
	"github.com/dingodb/dingocli/internal/task/task"
	tui "github.com/dingodb/dingocli/internal/tui/common"
	"github.com/dingodb/dingocli/internal/utils"
	"github.com/dingodb/dingocli/pkg/module"
	"github.com/fatih/color"
)

const (
	SIGNATURE_CONTAINER_REMOVED = "No such container"
)

type (
	Step2CleanContainer struct {
		ServiceId   string
		ContainerId string
		Storage     *storage.Storage
		ExecOptions module.ExecOptions
	}
)

func (s *Step2CleanContainer) Execute(ctx *context.Context) error {
	containerId := s.ContainerId
	if len(containerId) == 0 { // container not created
		return nil
	} else if containerId == comm.CLEANED_CONTAINER_ID { // container has removed
		return nil
	}

	cli := ctx.Module().DockerCli().RemoveContainer(s.ContainerId)
	cli.Execute(s.ExecOptions)
	//out, err := cli.Execute(s.ExecOptions)

	// container has removed
	//if err != nil && !strings.Contains(out, SIGNATURE_CONTAINER_REMOVED) {
	//	// fmt.Printf("container removed error: %s", out)
	//	return err
	//}
	return s.Storage.SetContainId(s.ServiceId, comm.CLEANED_CONTAINER_ID)
}

func getCleanFiles(clean map[string]bool, dc *topology.DeployConfig) []string {
	files := []string{}
	for item := range clean {
		switch item {
		case comm.CLEAN_ITEM_LOG:
			files = append(files, dc.GetLogDir())
		case comm.CLEAN_ITEM_DATA:
			files = append(files, dc.GetDataDir())
		case comm.CLEAN_ITEM_RAFT:
			if dc.GetRole() == topology.ROLE_COORDINATOR ||
				dc.GetRole() == topology.ROLE_STORE ||
				dc.GetRole() == topology.ROLE_DINGODB_DOCUMENT ||
				dc.GetRole() == topology.ROLE_DINGODB_INDEX {
				files = append(files, dc.GetDingoRaftDir())
			}
		case comm.CLEAN_ITEM_DOC:
			if dc.GetRole() == topology.ROLE_DINGODB_DOCUMENT {
				files = append(files, dc.GetDingoStoreDocDir())
			}
		case comm.CLEAN_ITEM_VECTOR:
			if dc.GetRole() == topology.ROLE_DINGODB_INDEX {
				files = append(files, dc.GetDingoStoreVectorDir())
			}
		}
	}
	return files
}

func NewCleanServiceTask(dingocli *cli.DingoCli, dc *topology.DeployConfig) (*task.Task, error) {
	serviceId := dingocli.GetServiceId(dc.GetId())
	containerId, err := dingocli.GetContainerId(serviceId)
	if containerId == comm.CLEANED_CONTAINER_ID {
		// container has removed, no need to clean
		dingocli.Storage().SetContainId(serviceId, comm.CLEANED_CONTAINER_ID)
		dingocli.WriteOutln("%s clean service: host=%s role=%s ", color.YellowString("[SKIP]"), dc.GetHost(), dc.GetRole())
		return nil, nil
	}
	if dingocli.IsSkip(dc) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	hc, err := dingocli.GetHost(dc.GetHost())
	if err != nil {
		return nil, err
	}

	if dc.GetRole() == topology.ROLE_FS_MDS_CLI {
		skipTmp := dingocli.MemStorage().Get(comm.KEY_SKIP_MDSV2_CLI)
		if skipTmp != nil && skipTmp.(bool) {
			return nil, nil
		}
	}

	// new task
	only := dingocli.MemStorage().Get(comm.KEY_CLEAN_ITEMS).([]string)
	subname := fmt.Sprintf("host=%s role=%s containerId=%s clean=%s",
		dc.GetHost(), dc.GetRole(), tui.TrimContainerId(containerId), strings.Join(only, ","))
	t := task.NewTask("Clean Service", subname, hc.GetSSHConfig())

	// add step to task
	clean := utils.Slice2Map(only)
	files := getCleanFiles(clean, dc) // directorys which need cleaned

	t.AddStep(&step.RemoveFile{
		Files:       files,
		ExecOptions: dingocli.ExecOptions(),
	})
	if clean[comm.CLEAN_ITEM_CONTAINER] {

		// var status string
		// t.AddStep(&step.InspectContainer{
		// ContainerId: containerId,
		// Format:      "'{{.State.Status}}'",
		// Out:         &status,
		// ExecOptions: dingocli.ExecOptions(),
		// })
		// if err != nil {
		// return errno.ERR_CONTAINER_NOT_EXISTED.S(*s.Out)
		// } else if status != "running" {
		// return errno.ERR_CONTAINER_IS_ABNORMAL.
		// F("host=%s role=%s containerId=%s",
		// s.Host, s.Role, tui.TrimContainerId(s.ContainerId))
		// }
		// return nil

		t.AddStep(&Step2CleanContainer{
			ServiceId:   serviceId,
			ContainerId: containerId,
			Storage:     dingocli.Storage(),
			ExecOptions: dingocli.ExecOptions(),
		})
	}

	return t, nil
}
