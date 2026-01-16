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

package config

import (
	"github.com/dingodb/dingocli/cli/cli"
	"github.com/dingodb/dingocli/internal/errno"
	cliutil "github.com/dingodb/dingocli/internal/utils"
	"github.com/spf13/cobra"
)

type showOptions struct {
	showPool bool
}

func NewShowCommand(dingocli *cli.DingoCli) *cobra.Command {
	var options showOptions

	cmd := &cobra.Command{
		Use:   "show [OPTIONS]",
		Short: "Show cluster topology",
		Args:  cliutil.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShow(dingocli, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.BoolVarP(&options.showPool, "pool", "p", false, "Show cluster pool information")

	return cmd
}

func runShow(dingocli *cli.DingoCli, options showOptions) error {
	// 1) check whether cluster exist
	if dingocli.ClusterId() == -1 {
		return errno.ERR_NO_CLUSTER_SPECIFIED
	} else if len(dingocli.ClusterTopologyData()) == 0 {
		dingocli.WriteOutln("<empty topology>")
		return nil
	}

	dingocli.WriteOut("%s", dingocli.ClusterTopologyData())
	return nil
}
