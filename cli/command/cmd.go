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

package command

import (
	"fmt"

	"github.com/dingodb/dingofs-tools/cli/command/export"
	"github.com/dingodb/dingofs-tools/cli/command/fs"
	"github.com/dingodb/dingofs-tools/cli/command/mds"
	"github.com/dingodb/dingofs-tools/cli/command/quota"
	"github.com/dingodb/dingofs-tools/cli/command/stats"
	"github.com/dingodb/dingofs-tools/cli/command/subpath"
	"github.com/dingodb/dingofs-tools/cli/command/warmup"

	"github.com/dingodb/dingofs-tools/cli/cli"
	"github.com/dingodb/dingofs-tools/cli/command/cache"
	"github.com/dingodb/dingofs-tools/cli/command/client"
	"github.com/dingodb/dingofs-tools/cli/command/cluster"
	"github.com/dingodb/dingofs-tools/cli/command/config"
	"github.com/dingodb/dingofs-tools/cli/command/hosts"
	"github.com/dingodb/dingofs-tools/cli/command/monitor"
	"github.com/dingodb/dingofs-tools/internal/errno"
	tools "github.com/dingodb/dingofs-tools/internal/tools/upgrade"
	cliutil "github.com/dingodb/dingofs-tools/internal/utils"
	"github.com/spf13/cobra"
)

var dingoadmExample = `Examples:
  $ dingoadm playground run --kind dingofs  # Run a dingoFS playground quickly
  $ dingoadm cluster add c1                 # Add a cluster named 'c1'
  $ dingoadm deploy                         # Deploy current cluster
  $ dingoadm stop                           # Stop current cluster service
  $ dingoadm clean                          # Clean current cluster
  $ dingoadm enter 6ff561598c6f             # Enter specified service container
  $ dingoadm -u                             # Upgrade dingoadm itself to the latest version`

type rootOptions struct {
	debug   bool
	upgrade bool
}

func addSubCommands(cmd *cobra.Command, dingoadm *cli.DingoAdm) {
	cmd.AddCommand(
		client.NewClientCommand(dingoadm),   // dingoadm client
		cluster.NewClusterCommand(dingoadm), // dingoadm cluster ...
		config.NewConfigCommand(dingoadm),   // dingoadm config ...
		hosts.NewHostsCommand(dingoadm),     // dingoadm hosts ...
		monitor.NewMonitorCommand(dingoadm), // dingoadm monitor ...
		fs.NewFSCommand(dingoadm),           // dingoadm fs ...
		subpath.NewSubpathCommand(dingoadm), // dingoadm subpath ...
		// cachegroup.NewCacheGroupCommand(dingoadm),   // dingoadm cachegroup ...
		// cachemember.NewCacheMemberCommand(dingoadm), // dingoadm cachemember...
		cache.NewCacheCommand(dingoadm),   // dingoadm cache ...
		stats.NewStatsCommand(dingoadm),   // dingoadm stats...
		warmup.NewWarmupCommand(dingoadm), //dingoadm warmup...
		export.NewExportCommand(dingoadm), //dingoadm export...
		quota.NewQuotaCommand(dingoadm),   // dingoadm quota ...
		mds.NewMDSCommand(dingoadm),       // dingoadm mds ...

		NewAuditCommand(dingoadm),      // dingoadm audit
		NewCleanCommand(dingoadm),      // dingoadm clean
		NewCompletionCommand(dingoadm), // dingoadm completion
		NewDeployCommand(dingoadm),     // dingoadm deploy
		NewEnterCommand(dingoadm),      // dingoadm enter
		NewExecCommand(dingoadm),       // dingoadm exec
		NewPrecheckCommand(dingoadm),   // dingoadm precheck
		NewRestartCommand(dingoadm),    // dingoadm restart
		NewStartCommand(dingoadm),      // dingoadm start
		NewStatusCommand(dingoadm),     // dingoadm status
		NewStopCommand(dingoadm),       // dingoadm stop
		NewUpgradeCommand(dingoadm),    // dingoadm upgrade
		// commonly used shorthands
		hosts.NewSSHCommand(dingoadm),      // dingoadm ssh
		hosts.NewPlaybookCommand(dingoadm), // dingoadm playbook
		client.NewMountCommand(dingoadm),   // dingoadm mount
		client.NewUmountCommand(dingoadm),  // dingoadm umount
	)
}

func setupRootCommand(cmd *cobra.Command, dingoadm *cli.DingoAdm) {
	cmd.SetVersionTemplate(`{{with .Name}}{{printf "%s " .}}{{end}}{{printf "Version %s" .Version}}
Copyright 2025 dingodb.com Inc.
Licensed under the Apache License, Version 2.0
http://www.apache.org/licenses/LICENSE-2.0
`)
	cliutil.SetFlagErrorFunc(cmd)
	cliutil.SetHelpTemplate(cmd)
	cliutil.SetUsageTemplate(cmd)
	cliutil.SetErr(cmd, dingoadm)
}

func NewDingoAdmCommand(dingoadm *cli.DingoAdm) *cobra.Command {
	var options rootOptions

	cmd := &cobra.Command{
		Use:   "dingoadm [OPTIONS] COMMAND [ARGS...]",
		Short: "Deploy and manage dingoFS cluster",
		// Version: fmt.Sprintf("dingoadm v%s, build %s", cli.Version, cli.CommitId),
		Version: fmt.Sprintf("%s (commit: %s) \nBuild Date: %s", cli.Version, cli.CommitId, cli.BuildTime),
		Example: dingoadmExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			if options.debug {
				return errno.List()
			} else if options.upgrade {
				return tools.Upgrade2Latest(cli.CommitId)
			} else if len(args) == 0 {
				return cliutil.ShowHelp(dingoadm.Err())(cmd, args)
			}

			return fmt.Errorf("dingoadm: '%s' is not a dingoadm command.\n"+
				"See 'dingoadm --help'", args[0])
		},
		SilenceUsage:          true, // silence usage when an error occurs
		DisableFlagsInUseLine: true,
	}

	cmd.Flags().BoolP("version", "v", false, "Print version information and quit")
	cmd.PersistentFlags().BoolP("help", "h", false, "Print usage")
	cmd.Flags().BoolVarP(&options.debug, "debug", "d", false, "Print debug information")
	cmd.Flags().BoolVarP(&options.upgrade, "upgrade", "u", false, "Upgrade dingoadm itself to the latest version")

	addSubCommands(cmd, dingoadm)
	setupRootCommand(cmd, dingoadm)

	return cmd
}
