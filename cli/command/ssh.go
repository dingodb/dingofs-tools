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
	"github.com/dingodb/dingocli/cli/cli"
	"github.com/dingodb/dingocli/internal/tools"
	cliutil "github.com/dingodb/dingocli/internal/utils"
	"github.com/spf13/cobra"
)

type sshOptions struct {
	host   string
	become bool
}

func NewSSHCommand(dingocli *cli.DingoCli) *cobra.Command {
	var options sshOptions

	cmd := &cobra.Command{
		Use:     "ssh HOST [OPTIONS]",
		Short:   "Connect remote host",
		GroupID: "UTILS",
		Args:    cliutil.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.host = args[0]
			return runSSH(dingocli, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.BoolVarP(&options.become, "become", "b", false, "Connect remote host with become")

	return cmd
}

func runSSH(dingocli *cli.DingoCli, options sshOptions) error {
	return tools.AttachRemoteHost(dingocli, options.host, options.become)
}
