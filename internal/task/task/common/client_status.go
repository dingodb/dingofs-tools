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

	"github.com/dingodb/dingocli/cli/cli"
	comm "github.com/dingodb/dingocli/internal/common"
	"github.com/dingodb/dingocli/internal/errno"
	"github.com/dingodb/dingocli/internal/storage"
	"github.com/dingodb/dingocli/internal/task/context"
	"github.com/dingodb/dingocli/internal/task/step"
	"github.com/dingodb/dingocli/internal/task/task"
	tui "github.com/dingodb/dingocli/internal/tui/common"
	"github.com/dingodb/dingocli/internal/utils"
)

const (
	MISSING_CLIENT_CONFIG = "-"
)

type (
	step2InitClientStatus struct {
		client     storage.Client
		cfgPath    *string
		memStorage *utils.SafeMap
	}

	step2FormatClientStatus struct {
		client     storage.Client
		status     *string
		memStorage *utils.SafeMap
	}

	ClientStatus struct {
		Id          string
		Host        string
		Kind        string
		ContainerId string
		Status      string
		AuxInfo     string
		CfgPath     string
	}
)

func dumpCfg(dingocli *cli.DingoCli, id string, cfgPath *string) step.LambdaType {
	return func(ctx *context.Context) error {
		*cfgPath = MISSING_CLIENT_CONFIG
		cfgs, err := dingocli.Storage().GetClientConfig(id)
		if err != nil {
			return errno.ERR_SELECT_CLIENT_CONFIG_FAILED.E(err)
		} else if len(cfgs) == 0 {
			return nil
		}

		data := cfgs[0].Data
		path := utils.RandFilename("/tmp")
		err = utils.WriteFile(path, data, 0644)
		if err != nil {
			return errno.ERR_WRITE_FILE_FAILED.E(err)
		}

		*cfgPath = path
		return nil
	}
}

// TODO(P0): init client status
func setClientStatus(memStorage *utils.SafeMap, id string, status ClientStatus) {
	memStorage.TX(func(kv *utils.SafeMap) error {
		m := map[string]ClientStatus{}
		v := kv.Get(comm.KEY_ALL_CLIENT_STATUS)
		if v != nil {
			m = v.(map[string]ClientStatus)
		}
		m[id] = status
		kv.Set(comm.KEY_ALL_CLIENT_STATUS, m)
		return nil
	})
}

func (s *step2InitClientStatus) Execute(ctx *context.Context) error {
	client := s.client
	id := client.Id
	setClientStatus(s.memStorage, id, ClientStatus{
		Id:          client.Id,
		Host:        client.Host,
		Kind:        client.Kind,
		ContainerId: client.ContainerId,
		Status:      comm.CLIENT_STATUS_UNKNOWN,
		AuxInfo:     client.AuxInfo,
		CfgPath:     *s.cfgPath,
	})
	return nil
}

func (s *step2FormatClientStatus) Execute(ctx *context.Context) error {
	status := *s.status
	if len(status) == 0 { // container losed
		status = comm.CLIENT_STATUS_LOSED
	}

	id := s.client.Id
	s.memStorage.TX(func(kv *utils.SafeMap) error {
		v := kv.Get(comm.KEY_ALL_CLIENT_STATUS)
		m := v.(map[string]ClientStatus)

		// update the status
		s := m[id]
		s.Status = status
		m[id] = s
		kv.Set(comm.KEY_ALL_CLIENT_STATUS, m)
		return nil
	})
	return nil
}

func NewInitClientStatusTask(dingocli *cli.DingoCli, v interface{}) (*task.Task, error) {
	client := v.(storage.Client)

	t := task.NewTask("Init Client Status", "", nil)

	var cfgPath string
	t.AddStep(&step.Lambda{
		Lambda: dumpCfg(dingocli, client.Id, &cfgPath),
	})
	t.AddStep(&step2InitClientStatus{
		client:     client,
		cfgPath:    &cfgPath,
		memStorage: dingocli.MemStorage(),
	})

	return t, nil
}

func NewGetClientStatusTask(dingocli *cli.DingoCli, v interface{}) (*task.Task, error) {
	client := v.(storage.Client)
	hc, err := dingocli.GetHost(client.Host)
	if err != nil {
		return nil, err
	}

	containerId := client.ContainerId
	subname := fmt.Sprintf("host=%s kind=%s containerId=%s",
		hc.GetHost(), client.Kind, tui.TrimContainerId(containerId))
	t := task.NewTask("Get Client Status", subname, hc.GetSSHConfig())

	// add step
	var status string
	t.AddStep(&step.ListContainers{
		ShowAll:     true,
		Format:      `"{{.Status}}"`,
		Filter:      fmt.Sprintf("id=%s", containerId),
		Out:         &status,
		ExecOptions: dingocli.ExecOptions(),
	})
	t.AddStep(&step2FormatClientStatus{
		client:     client,
		status:     &status,
		memStorage: dingocli.MemStorage(),
	})

	return t, nil
}
