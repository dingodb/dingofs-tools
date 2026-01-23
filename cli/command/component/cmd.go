/*
 * Copyright (c) 2025 dingodb.com, Inc. All Rights Reserved
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package component

import (
	"github.com/dingodb/dingocli/cli/cli"
	cliutil "github.com/dingodb/dingocli/internal/utils"
	"github.com/spf13/cobra"
)

func NewComponentCommand(dingocli *cli.DingoCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "component",
		Short:   "Manage components",
		GroupID: "DEPLOY",
		Args:    cliutil.NoArgs,
	}

	cmd.AddCommand(
		NewInstallCommand(dingocli),
		NewListCommand(dingocli),
		NewUninstallCommand(dingocli),
		NewUseCommand(dingocli),
		NewUpdateCommand(dingocli),
	)

	return cmd
}
