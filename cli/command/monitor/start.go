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
	"github.com/dingodb/dingocli/cli/cli"
	"github.com/dingodb/dingocli/internal/configure"
	"github.com/dingodb/dingocli/internal/errno"
	"github.com/dingodb/dingocli/internal/playbook"
	tui "github.com/dingodb/dingocli/internal/tui/common"
	cliutil "github.com/dingodb/dingocli/internal/utils"
	"github.com/spf13/cobra"
)

var (
	MONITOR_START_STEPS = []int{
		playbook.START_MONITOR_SERVICE,
	}
)

type startOptions struct {
	id   string
	role string
	host string
}

func NewStartCommand(dingocli *cli.DingoCli) *cobra.Command {
	var options startOptions
	cmd := &cobra.Command{
		Use:   "start [OPTIONS]",
		Short: "Start monitor service",
		Args:  cliutil.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStart(dingocli, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVar(&options.id, "id", "*", "Specify monitor service id")
	flags.StringVar(&options.role, "role", "*", "Specify monitor service role")
	flags.StringVar(&options.host, "host", "*", "Specify monitor service host")

	return cmd
}

func genStartPlaybook(dingocli *cli.DingoCli,
	mcs []*configure.MonitorConfig,
	options startOptions) (*playbook.Playbook, error) {
	mcs = configure.FilterMonitorConfig(dingocli, mcs, configure.FilterMonitorOption{
		Id:   options.id,
		Role: options.role,
		Host: options.host,
	})
	if len(mcs) == 0 {
		return nil, errno.ERR_NO_SERVICES_MATCHED
	}

	steps := MONITOR_START_STEPS
	pb := playbook.NewPlaybook(dingocli)
	for _, step := range steps {
		pb.AddStep(&playbook.PlaybookStep{
			Type:    step,
			Configs: mcs,
		})
	}
	return pb, nil
}

func runStart(dingocli *cli.DingoCli, options startOptions) error {
	// 1) parse monitor configure
	mcs, err := configure.ParseMonitor(dingocli)
	if err != nil {
		return err
	}

	// 2) generate start playbook
	pb, err := genStartPlaybook(dingocli, mcs, options)
	if err != nil {
		return err
	}

	// 3) confirm by user
	if pass := tui.ConfirmYes(tui.PromptStartService(options.id, options.role, options.host)); !pass {
		dingocli.WriteOut(tui.PromptCancelOpetation("start monitor service"))
		return errno.ERR_CANCEL_OPERATION
	}

	// 4) run playground
	return pb.Run()
}
