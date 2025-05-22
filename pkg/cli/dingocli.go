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

	"github.com/dingodb/dingofs-tools/pkg/cli/command/check"
	quotaconfig "github.com/dingodb/dingofs-tools/pkg/cli/command/config"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/create"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/delete"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/gateway"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/list"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/query"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/quota"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/stats"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/status"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/umount"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/usage"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/warmup"
	"github.com/dingodb/dingofs-tools/pkg/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	cobratemplate "github.com/dingodb/dingofs-tools/internal/utils/template"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/version"
)

func addSubCommands(cmd *cobra.Command) {
	cmd.AddCommand(
		usage.NewUsageCommand(),
		list.NewListCommand(),
		status.NewStatusCommand(),
		umount.NewUmountCommand(),
		query.NewQueryCommand(),
		delete.NewDeleteCommand(),
		create.NewCreateCommand(),
		check.NewCheckCommand(),
		warmup.NewWarmupCommand(),
		stats.NewStatsCommand(),
		quota.NewQuotaCommand(),
		quotaconfig.NewConfigCommand(),
		gateway.NewGatewayCommand(),
		version.NewVersionCommand(),
	)
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
			if cmd.HasParent() {
				skipCommands := map[string]bool{
					"help":    true,
					"version": true,
				}
				if !skipCommands[cmd.Name()] {
					config.InitConfig()
				}
			}
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
	rootCmd.PersistentFlags().StringVarP(&config.ConfPath, "conf", "", "", "config file (default is $HOME/.dingo/dingo.yaml or /etc/dingo/dingo.yaml)")
	config.AddShowErrorPFlag(rootCmd)
	rootCmd.PersistentFlags().BoolP("verbose", "", false, "show some extra info")
	viper.BindPFlag("useViper", rootCmd.PersistentFlags().Lookup("viper"))

	addSubCommands(rootCmd)
	setupRootCommand(rootCmd)

	return rootCmd
}

func Execute() {
	res := newDingoCommand().Execute()
	if res != nil {
		os.Exit(1)
	}
}
