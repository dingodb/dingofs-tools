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

package client

import (
	"github.com/dingodb/dingocli/cli/cli"
	"github.com/dingodb/dingocli/internal/configure/topology"
	"github.com/dingodb/dingocli/internal/errno"
	"github.com/dingodb/dingocli/internal/tools"
	"github.com/dingodb/dingocli/internal/utils"
	"github.com/spf13/cobra"
)

type enterOptions struct {
	id string
}

func NewEnterCommand(dingocli *cli.DingoCli) *cobra.Command {
	var options enterOptions

	cmd := &cobra.Command{
		Use:   "enter ID",
		Short: "Enter client container",
		Args:  utils.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.id = args[0]
			return runEnter(dingocli, options)
		},
		DisableFlagsInUseLine: true,
	}

	return cmd
}

func runEnter(dingocli *cli.DingoCli, options enterOptions) error {
	// 1) get container id
	clients, err := dingocli.Storage().GetClient(options.id)
	if err != nil {
		return err
	} else if len(clients) != 1 {
		return errno.ERR_NO_CLIENT_MATCHED
	}

	// 2) attch remote container
	client := clients[0]
	home := "/dingofs"
	if client.Kind == topology.KIND_DINGOFS {
		home = "/dingofs/client"
	}
	return tools.AttachRemoteContainer(dingocli, client.Host, client.ContainerId, home)
}
