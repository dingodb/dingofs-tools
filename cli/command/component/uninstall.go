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
	"os"
	"path/filepath"

	"github.com/dingodb/dingocli/cli/cli"
	"github.com/dingodb/dingocli/internal/component"
	"github.com/dingodb/dingocli/internal/utils"

	"github.com/spf13/cobra"
)

const (
	COMPONENT_UN_EXAMPLE = `Examples:
  # Uninstall the specific version a component
  $ dingo component uninstall dingo-client:v1.2.0"

  # Uninstall all version of specific component
  $ dingo component uninstall dingo-client --all"`
)

type uninstallOptions struct {
	component string
	all       bool
	force     bool
}

func NewUninstallCommand(dingocli *cli.DingoCli) *cobra.Command {
	var options uninstallOptions

	cmd := &cobra.Command{
		Use:     "uninstall <component1>[:version] [OPTIONS]",
		Short:   "uninstall components",
		Args:    utils.ExactArgs(1),
		Example: COMPONENT_UN_EXAMPLE,
		RunE: func(cmd *cobra.Command, args []string) error {
			options.component = args[0]

			return runUninstall(cmd, dingocli, &options)
		},
		SilenceUsage:          false,
		DisableFlagsInUseLine: true,
	}

	utils.SetFlagErrorFunc(cmd)

	cmd.Flags().BoolVar(&options.all, "all", false, "Uninstall all versions of a component")
	cmd.Flags().BoolVar(&options.force, "force", false, "Force uninstall even if the component is active")

	return cmd
}

func runUninstall(cmd *cobra.Command, dingocli *cli.DingoCli, options *uninstallOptions) error {

	componentManager, err := component.NewComponentManager()
	if err != nil {
		return err
	}
	name, version := component.ParseComponentVersion(options.component)

	if options.all {
		if version != "" {
			return fmt.Errorf("cannot specify version when --all is set")
		}

		removedComponents, err := componentManager.RemoveComponents(name, true)
		if err != nil {
			return err
		}

		fmt.Printf("Successfully removed components: \n")
		for _, comp := range removedComponents {
			os.Remove(filepath.Join(comp.Path, comp.Name))
			fmt.Printf("  %s:%s \n", comp.Name, comp.Version)
		}

		return nil
	}

	version = utils.Ternary(version == "", component.LASTEST_VERSION, version)
	// remove one component
	if err := componentManager.RemoveComponent(name, version, options.force, true); err != nil {
		return err
	}

	fmt.Printf("Successfully removed component: %s:%s\n", name, version)

	return nil
}
