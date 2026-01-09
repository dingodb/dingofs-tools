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

	"github.com/dingodb/dingofs-tools/cli/cli"
	"github.com/dingodb/dingofs-tools/internal/configure"
	"github.com/dingodb/dingofs-tools/internal/task/step"
	"github.com/dingodb/dingofs-tools/internal/task/task"
)

func NewPullImageTask(dingoadm *cli.DingoAdm, cfg *configure.MonitorConfig) (*task.Task, error) {
	image := cfg.GetImage()
	host := cfg.GetHost()
	hc, err := dingoadm.GetHost(host)
	if err != nil {
		return nil, err
	}

	// new task
	subname := fmt.Sprintf("host=%s image=%s", host, image)
	t := task.NewTask("Pull Image", subname, hc.GetSSHConfig())
	// add step to task
	t.AddStep(&step.PullImage{
		Image:       image,
		ExecOptions: dingoadm.ExecOptions(),
	})
	return t, nil
}
