// Copyright (c) 2025 dingodb.com, Inc. All Rights Reserved
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cli

import (
	"fmt"
	"os"

	"github.com/dingodb/dingofs-tools/internal/logger"
	cobratemplate "github.com/dingodb/dingofs-tools/internal/utils/template"
	cmdConfig "github.com/dingodb/dingofs-tools/pkg/cli/command/config"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/create"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/delete"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/fuse"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/gateway"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/leave"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/list"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/query"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/quota"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/set"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/stats"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/status"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/umount"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/unlock"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/usage"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/version"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/warmup"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func addSubCommands(cmd *cobra.Command) {
	cmd.AddCommand(
		version.NewVersionCommand(),
		gateway.NewGatewayCommand(),
		fuse.NewFuseCommand(),
		list.NewListCommand(),
		create.NewCreateCommand(),
		delete.NewDeleteCommand(),
		status.NewStatusCommand(),
		cmdConfig.NewConfigCommand(),
		query.NewQueryCommand(),
		stats.NewStatsCommand(),
		umount.NewUmountCommand(),
		quota.NewQuotaCommand(),
		usage.NewUsageCommand(),
		set.NewSetCommand(),
		unlock.NewUnlockCommand(),
		leave.NewLeaveCommand(),
		warmup.NewWarmupCommand(),
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
		Use:   "dingo COMMAND [ARGS...]",
		Short: "dingo is a tool for managing dingofs",
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

	rootCmd.PersistentFlags().BoolP("help", "", false, "print help")
	rootCmd.PersistentFlags().StringP("conf", "", "", "config file (default is $HOME/.dingo/dingo.yaml or /etc/dingo/dingo.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "", false, "show some extra info")
	config.AddShowErrorPFlag(rootCmd)
	config.AddLogfilePFlag(rootCmd)
	config.AddLoglevelPFlag(rootCmd)
	config.AddLogfmtPFlag(rootCmd)
	viper.BindPFlag("useViper", rootCmd.PersistentFlags().Lookup("viper"))

	addSubCommands(rootCmd)
	setupRootCommand(rootCmd)

	return rootCmd
}

func Execute() {
	// for compatibility, dingo support mds and mdsv1, so we need to load different commands based on the MDS API version.
	// MDS API version can be set by environment variable MDS_API_VERSION or mds_api_version in config file.
	// if used mds_api_version parameter in config file with --conf flag,e.g.:
	// dingo list fs --conf dingo.yaml
	// we will need to parse cmd flags --conf to determine the MDS API version.

	logfile := logger.DEFAULT_LOG_FILE
	loglevel := logger.DEFAULT_LOG_LEVEL
	logfmt := logger.DEFAULT_LOG_FORMAT

	var confFile string
	for i := 0; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--conf":
			if i+1 < len(os.Args) {
				confFile = os.Args[i+1]
				i++ // Skip the next argument since we've consumed it
			}
		case "--logfile":
			if i+1 < len(os.Args) {
				logfile = os.Args[i+1]
				i++
			}
		case "--loglevel":
			if i+1 < len(os.Args) {
				loglevel = os.Args[i+1]
				i++
			}
		case "--logfmt":
			if i+1 < len(os.Args) {
				logfmt = os.Args[i+1]
				i++
			}
		default:

		}
	}
	// init glogal logger
	logger.InitGlobalLogger(logger.WithLogFile(logfile), logger.WithLogLevel(loglevel), logger.WithFormat(logfmt))
	// initialize config
	config.InitConfig(confFile)

	res := newDingoCommand().Execute()
	if res != nil {
		os.Exit(1)
	}
}
