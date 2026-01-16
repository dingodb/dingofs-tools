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
	"github.com/dingodb/dingocli/internal/utils"
	"github.com/spf13/cobra"
)

const (
	DIFF_EXAMPLE = `Examples:
  $ dingocli config diff /path/to/topology.yaml  # Display difference for topology`
)

type diffOptions struct {
	filename string
}

func NewDiffCommand(dingocli *cli.DingoCli) *cobra.Command {
	var options diffOptions

	cmd := &cobra.Command{
		Use:     "diff TOPOLOGY",
		Short:   "Display difference for topology",
		Args:    utils.ExactArgs(1),
		Example: DIFF_EXAMPLE,
		RunE: func(cmd *cobra.Command, args []string) error {
			options.filename = args[0]
			return runDiff(dingocli, options)
		},
		DisableFlagsInUseLine: true,
	}

	return cmd
}

func runDiff(dingocli *cli.DingoCli, options diffOptions) error {
	// 1) data1: current cluster topology data
	data1 := dingocli.ClusterTopologyData()

	// 2) data2: topology in file
	if !utils.PathExist(options.filename) {
		return errno.ERR_TOPOLOGY_FILE_NOT_FOUND.
			F("%s: no such file", utils.AbsPath(options.filename))
	}
	data2, err := utils.ReadFile(options.filename)
	if err != nil {
		return errno.ERR_READ_TOPOLOGY_FILE_FAILED.E(err)
	}

	// 3) print difference
	diff := utils.Diff(data1, data2)
	dingocli.Out().Write([]byte(diff))
	return nil
}
