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
	"fmt"

	"github.com/dingodb/dingocli/cli/cli"
	"github.com/dingodb/dingocli/internal/component"
	"github.com/dingodb/dingocli/internal/utils"

	"github.com/spf13/cobra"
)

const (
	COMPONENT_USE_EXAMPLE = `Examples:
  # Use the specific version as default
  $ dingo component use dingo-client:v1.2.0"
  `
)

type useOptions struct {
	component string
}

func NewUseCommand(dingocli *cli.DingoCli) *cobra.Command {
	var options useOptions

	cmd := &cobra.Command{
		Use:     "use <component1>:[version] [OPTIONS]",
		Short:   "set default version",
		Args:    utils.ExactArgs(1),
		Example: COMPONENT_USE_EXAMPLE,
		RunE: func(cmd *cobra.Command, args []string) error {
			options.component = args[0]

			return runUse(cmd, dingocli, &options)
		},
		SilenceUsage:          false,
		DisableFlagsInUseLine: true,
	}

	utils.SetFlagErrorFunc(cmd)

	return cmd
}

func runUse(cmd *cobra.Command, dingocli *cli.DingoCli, options *useOptions) error {
	componentManager, err := component.NewComponentManager()
	if err != nil {
		return err
	}

	name, version := component.ParseComponentVersion(options.component)
	version = utils.Ternary(version == "", component.LASTEST_VERSION, version)
	if err := componentManager.SetDefaultVersion(name, version); err != nil {
		return err
	}

	if err := componentManager.SaveInstalledComponents(); err != nil {
		return err
	}

	fmt.Printf("Successfully use %s:%s as default version\n", name, version)

	return nil
}
