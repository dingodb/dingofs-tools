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
	"github.com/dingodb/dingocli/internal/errno"
	"github.com/dingodb/dingocli/internal/tui"
	cliutil "github.com/dingodb/dingocli/internal/utils"
	"github.com/spf13/cobra"
)

type auditOptions struct {
	tail    int
	verbose bool
}

func NewAuditCommand(dingocli *cli.DingoCli) *cobra.Command {
	var options auditOptions

	cmd := &cobra.Command{
		Use:   "audit [OPTIONS]",
		Short: "Show audit log of operation",
		Args:  cliutil.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAudit(dingocli, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.IntVarP(&options.tail, "tail", "n", 20, "Number of lines to show from the end of the logs (0 means all)")
	flags.BoolVarP(&options.verbose, "verbose", "v", false, "Verbose output for clusters")

	return cmd
}

func runAudit(dingocli *cli.DingoCli, options auditOptions) error {
	auditLogs, err := dingocli.Storage().GetAuditLogs()
	if err != nil {
		return errno.ERR_GET_AUDIT_LOGS_FAILE.E(err)
	}

	tail := options.tail
	if tail != 0 && tail > 0 && tail < len(auditLogs) {
		auditLogs = auditLogs[len(auditLogs)-tail:]
	}
	output := tui.FormatAuditLogs(auditLogs, options.verbose)
	dingocli.WriteOut(output)
	return nil
}
