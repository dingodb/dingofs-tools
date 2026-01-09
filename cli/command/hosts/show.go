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

package hosts

import (
	"github.com/dingodb/dingofs-tools/cli/cli"
	cliutil "github.com/dingodb/dingofs-tools/internal/utils"
	"github.com/spf13/cobra"
)

type showOptions struct{}

func NewShowCommand(dingoadm *cli.DingoAdm) *cobra.Command {
	var options showOptions

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show hosts",
		Args:  cliutil.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShow(dingoadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	return cmd
}

func runShow(dingoadm *cli.DingoAdm, options showOptions) error {
	hosts := dingoadm.Hosts()
	if len(hosts) == 0 {
		dingoadm.WriteOutln("<empty hosts>")
	} else {
		dingoadm.WriteOut(hosts)
	}
	return nil
}
