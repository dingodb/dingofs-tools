/*
 *  Copyright (c) 2022 NetEase Inc.
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

/*
 * Project: DingoCli
 * Created Date: 2022-05-09
 * Author: chengyi (Cyber-SiKu)
 */

package cli

import (
	"fmt"
	"os"

	"github.com/dingodb/dingofs-tools/pkg/cli/command/common/gateway"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/common/version"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/common/warmup"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/v1/check"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/v1/create"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/v1/delete"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/v1/list"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/v1/query"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/v1/quota"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/v1/stats"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/v1/status"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/v1/umount"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/v1/usage"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	cobratemplate "github.com/dingodb/dingofs-tools/internal/utils/template"
	quotaconfig "github.com/dingodb/dingofs-tools/pkg/cli/command/v1/config"
	v2Config "github.com/dingodb/dingofs-tools/pkg/cli/command/v2/config"
	v2Create "github.com/dingodb/dingofs-tools/pkg/cli/command/v2/create"
	v2Delete "github.com/dingodb/dingofs-tools/pkg/cli/command/v2/delete"
	v2List "github.com/dingodb/dingofs-tools/pkg/cli/command/v2/list"
	v2Query "github.com/dingodb/dingofs-tools/pkg/cli/command/v2/query"
	v2Quota "github.com/dingodb/dingofs-tools/pkg/cli/command/v2/quota"
	v2Stats "github.com/dingodb/dingofs-tools/pkg/cli/command/v2/stats"
	v2Status "github.com/dingodb/dingofs-tools/pkg/cli/command/v2/status"
	v2Umount "github.com/dingodb/dingofs-tools/pkg/cli/command/v2/umount"
	v2Usage "github.com/dingodb/dingofs-tools/pkg/cli/command/v2/usage"
)

func addSubCommands(cmd *cobra.Command) {
	cmd.AddCommand(
		version.NewVersionCommand(),
		gateway.NewGatewayCommand(),
		warmup.NewWarmupCommand(),
	)

	if config.MDSApiV2 {
		cmd.AddCommand(
			v2List.NewListCommand(),
			v2Create.NewCreateCommand(),
			v2Delete.NewDeleteCommand(),
			v2Status.NewStatusCommand(),
			v2Config.NewConfigCommand(),
			v2Query.NewQueryCommand(),
			v2Stats.NewStatsCommand(),
			v2Umount.NewUmountCommand(),
			v2Quota.NewQuotaCommand(),
			v2Usage.NewUsageCommand(),
		)
	} else {
		cmd.AddCommand(
			usage.NewUsageCommand(),
			list.NewListCommand(),
			status.NewStatusCommand(),
			umount.NewUmountCommand(),
			query.NewQueryCommand(),
			delete.NewDeleteCommand(),
			create.NewCreateCommand(),
			check.NewCheckCommand(),
			stats.NewStatsCommand(),
			quota.NewQuotaCommand(),
			quotaconfig.NewConfigCommand(),
		)
	}
}

func setupRootCommand(cmd *cobra.Command) {
	cmd.SetVersionTemplate("dingo {{.Version}}\n")
	cobratemplate.SetFlagErrorFunc(cmd)
	cobratemplate.SetHelpTemplate(cmd)
	cobratemplate.SetUsageTemplate(cmd)
}

func newDingoCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "dingo COMMAND [ARGS...]",
		Short:   "dingo is a tool for managing dingofs",
		Version: version.Version,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cobratemplate.ShowHelp(os.Stderr)(cmd, args)
			}
			return fmt.Errorf("dingo: '%s' is not a dingo command.\n"+
				"See 'dingo --help'", args[0])
		},
		SilenceUsage: false, // silence usage when an error occurs
		CompletionOptions: cobra.CompletionOptions{
			HiddenDefaultCmd: true,
		},
	}

	rootCmd.Flags().BoolP("version", "", false, "print dingo version")
	rootCmd.PersistentFlags().BoolP("help", "", false, "print help")
	rootCmd.PersistentFlags().StringP("conf", "", "", "config file (default is $HOME/.dingo/dingo.yaml or /etc/dingo/dingo.yaml)")
	config.AddShowErrorPFlag(rootCmd)
	rootCmd.PersistentFlags().BoolP("verbose", "", false, "show some extra info")
	viper.BindPFlag("useViper", rootCmd.PersistentFlags().Lookup("viper"))

	addSubCommands(rootCmd)
	setupRootCommand(rootCmd)

	return rootCmd
}

func Execute() {
	// for compatibility, dingo support mdsv2 and mdsv1, so we need to load different commands based on the MDS API version.
	// MDS API version can be set by environment variable MDS_API_VERSION or mds_api_version in config file.
	// if used mds_api_version parameter in config file with --conf flag,e.g.:
	// dingo list fs --conf dingo.yaml
	// we will need to parse cmd flags --conf to determine the MDS API version.

	var confFile string
	for i := 0; i < len(os.Args); i++ {
		if os.Args[i] == "--conf" {
			if i+1 < len(os.Args) {
				confFile = os.Args[i+1]
			}
		}
	}
	// initialize config
	config.InitConfig(confFile)

	res := newDingoCommand().Execute()
	if res != nil {
		os.Exit(1)
	}
}
