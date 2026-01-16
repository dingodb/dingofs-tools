/*
 * 	Copyright (c) 2025 dingodb.com Inc.
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
	"encoding/json"

	"github.com/dingodb/dingocli/cli/cli"
	"github.com/dingodb/dingocli/internal/configure"
	"github.com/dingodb/dingocli/internal/errno"
	cliutil "github.com/dingodb/dingocli/internal/utils"
	"github.com/spf13/cobra"
)

func NewShowCommand(dingocli *cli.DingoCli) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "show [OPTIONS]",
		Short: "Show cluster topology",
		Args:  cliutil.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShow(dingocli)
		},
		DisableFlagsInUseLine: true,
	}

	return cmd
}

func decodeMonitorJSON(data string) (string, error) {
	monitor := configure.Monitor{}
	err := json.Unmarshal([]byte(data), &monitor)
	if err != nil {
		return "", errno.ERR_DECODE_CLUSTER_POOL_JSON_FAILED.E(err)
	}
	bytes, err := json.MarshalIndent(monitor, "", "    ")
	if err != nil {
		return "", errno.ERR_DECODE_CLUSTER_POOL_JSON_FAILED.E(err)
	}
	return string(bytes), nil
}

func runShow(dingocli *cli.DingoCli) error {
	// 1) check whether cluster exist
	if len(dingocli.Monitor().Monitor) == 0 {
		dingocli.WriteOutln("<empty monitor>")
		return nil
	}

	//data, err := decodeMonitorJSON(dingocli.Monitor().Monitor)
	//if err != nil {
	//	return err
	//}
	dingocli.WriteOutln(dingocli.Monitor().Monitor)
	return nil
}
