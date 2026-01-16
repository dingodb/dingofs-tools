/*
 * 	Copyright (c) 2025 dingodb.com Inc.
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
	comm "github.com/dingodb/dingocli/internal/common"
	"github.com/dingodb/dingocli/internal/configure"
	"github.com/dingodb/dingocli/internal/errno"
	"github.com/dingodb/dingocli/internal/playbook"
	tui "github.com/dingodb/dingocli/internal/tui/common"
	cliutil "github.com/dingodb/dingocli/internal/utils"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	UPGRADE_PLAYBOOK_STEPS = []int{
		playbook.PULL_MONITOR_IMAGE,
		playbook.STOP_MONITOR_SERVICE,
		playbook.CLEAN_MONITOR_SERVICE,
		playbook.CREATE_MONITOR_CONTAINER,
		playbook.START_MONITOR_SERVICE,
		playbook.SYNC_HOSTS_MAPPING,
	}
)

type upgradeOptions struct {
	id            string
	role          string
	host          string
	force         bool
	useLocalImage bool
}

func NewUpgradeCommand(dingocli *cli.DingoCli) *cobra.Command {
	var options upgradeOptions

	cmd := &cobra.Command{
		Use:   "upgrade [OPTIONS]",
		Short: "Upgrade monitor service",
		Args:  cliutil.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpgrade(dingocli, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVar(&options.id, "id", "*", "Specify service id")
	flags.StringVar(&options.role, "role", "*", "Specify service role")
	flags.StringVar(&options.host, "host", "*", "Specify service host")
	flags.BoolVarP(&options.force, "force", "f", false, "Never prompt")
	flags.BoolVar(&options.useLocalImage, "local", false, "Use local image")

	return cmd
}

func genUpgradePlaybook(dingocli *cli.DingoCli,
	mcs []*configure.MonitorConfig,
	options upgradeOptions) (*playbook.Playbook, error) {
	steps := UPGRADE_PLAYBOOK_STEPS
	if options.useLocalImage {
		// remove PULL_IMAGE step
		for i, item := range steps {
			if item == playbook.PULL_MONITOR_IMAGE {
				steps = append(steps[:i], steps[i+1:]...)
				break
			}
		}
	}
	pb := playbook.NewPlaybook(dingocli)
	for _, step := range steps {

		pb.AddStep(&playbook.PlaybookStep{
			Type:    step,
			Configs: mcs,
			Options: map[string]interface{}{
				comm.KEY_CLEAN_ITEMS:      []string{comm.CLEAN_ITEM_CONTAINER},
				comm.KEY_CLEAN_BY_RECYCLE: true,
				comm.KEY_UPGRADE_FLAG:     true,
			},
		})
	}
	return pb, nil
}

func displayTitle(dingocli *cli.DingoCli, mcs []*configure.MonitorConfig, options upgradeOptions) {
	total := len(mcs)
	if options.force {
		dingocli.WriteOutln(color.YellowString("Upgrade %d services at once", total))
	} else {
		dingocli.WriteOutln(color.YellowString("Upgrade %d services one by one", total))
	}
	dingocli.WriteOutln(tui.PromptUpgradeService(options.id, options.role, options.host))
}

func upgradeAtOnce(dingocli *cli.DingoCli, mcs []*configure.MonitorConfig, options upgradeOptions) error {
	// 1) display upgrade title
	displayTitle(dingocli, mcs, options)

	// 2) generate upgrade playbook
	pb, err := genUpgradePlaybook(dingocli, mcs, options)
	if err != nil {
		return err
	}

	// 3) run playbook
	err = pb.Run()
	if err != nil {
		return err
	}

	// 4) print success prompt
	dingocli.WriteOutln("")
	dingocli.WriteOutln(color.GreenString("Upgrade %d services success :)", len(mcs)))
	return nil
}

func upgradeOneByOne(dingocli *cli.DingoCli, mcs []*configure.MonitorConfig, options upgradeOptions) error {
	// 1) display upgrade title
	displayTitle(dingocli, mcs, options)

	// 2) upgrade service one by one
	total := len(mcs)
	for i, mc := range mcs {
		// 2.1) confirm by user
		dingocli.WriteOutln("")
		dingocli.WriteOutln("Upgrade %s service:", color.BlueString("%d/%d", i+1, total))
		dingocli.WriteOutln("  + host=%s  role=%s  image=%s", mc.GetHost(), mc.GetRole(), mc.GetImage())
		if pass := tui.ConfirmYes(tui.DEFAULT_CONFIRM_PROMPT); !pass {
			dingocli.WriteOut(tui.PromptCancelOpetation("upgrade service"))
			return errno.ERR_CANCEL_OPERATION
		}

		// 2.2) generate upgrade playbook
		pb, err := genUpgradePlaybook(dingocli, []*configure.MonitorConfig{mc}, options)
		if err != nil {
			return err
		}

		// 2.3) run playbook
		err = pb.Run()
		if err != nil {
			return err
		}

		// 2.4) print success prompt
		dingocli.WriteOutln("")
		dingocli.WriteOutln(color.GreenString("Upgrade %d/%d sucess :)"), i+1, total)
	}
	return nil
}

func runUpgrade(dingocli *cli.DingoCli, options upgradeOptions) error {
	// 1) parse monitor configure
	mcs, err := configure.ParseMonitor(dingocli)
	if err != nil {
		return err
	}

	// 2) filter deploy config
	mcs = configure.FilterMonitorConfig(dingocli, mcs, configure.FilterMonitorOption{
		Id:   options.id,
		Role: options.role,
		Host: options.host,
	})
	if len(mcs) == 0 {
		return errno.ERR_NO_SERVICES_MATCHED
	}

	// 3.1) upgrade service at once
	if options.force {
		return upgradeAtOnce(dingocli, mcs, options)
	}

	// 3.2) OR upgrade service one by one
	return upgradeOneByOne(dingocli, mcs, options)
}
