/*
 *  Copyright (c) 2021 NetEase Inc.
 * 	Copyright (c) 2024 dingodb.com Inc.
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

/*
 * Project: CurveAdm
 * Created Date: 2021-10-15
 * Author: Jingli Chen (Wine93)
 *
 * Project: dingoadm
 * Author: dongwei (jackblack369)
 */

// __SIGN_BY_WINE93__

package config

import (
	"encoding/json"

	"github.com/dingodb/dingofs-tools/cli/cli"
	"github.com/dingodb/dingofs-tools/internal/configure"
	"github.com/dingodb/dingofs-tools/internal/errno"
	cliutil "github.com/dingodb/dingofs-tools/internal/utils"
	"github.com/spf13/cobra"
)

type showOptions struct {
	showPool bool
}

func NewShowCommand(dingoadm *cli.DingoAdm) *cobra.Command {
	var options showOptions

	cmd := &cobra.Command{
		Use:   "show [OPTIONS]",
		Short: "Show cluster topology",
		Args:  cliutil.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShow(dingoadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.BoolVarP(&options.showPool, "pool", "p", false, "Show cluster pool information")

	return cmd
}

func decodePoolJSON(data string) (string, error) {
	pool := configure.DingoFsClusterTopo{}
	err := json.Unmarshal([]byte(data), &pool)
	if err != nil {
		return "", errno.ERR_DECODE_CLUSTER_POOL_JSON_FAILED.E(err)
	}
	bytes, err := json.MarshalIndent(pool, "", "    ")
	if err != nil {
		return "", errno.ERR_DECODE_CLUSTER_POOL_JSON_FAILED.E(err)
	}
	return string(bytes), nil
}

func runShow(dingoadm *cli.DingoAdm, options showOptions) error {
	// 1) check whether cluster exist
	if dingoadm.ClusterId() == -1 {
		return errno.ERR_NO_CLUSTER_SPECIFIED
	} else if len(dingoadm.ClusterTopologyData()) == 0 {
		dingoadm.WriteOutln("<empty topology>")
		return nil
	}

	// 2) display cluster topology
	if !options.showPool {
		dingoadm.WriteOut("%s", dingoadm.ClusterTopologyData())
		return nil
	}

	// 3) OR display cluster pool information
	if len(dingoadm.ClusterPoolData()) == 0 {
		dingoadm.WriteOutln("<empty pool>")
		return nil
	}
	data, err := decodePoolJSON(dingoadm.ClusterPoolData())
	if err != nil {
		return err
	}
	dingoadm.WriteOutln(data)
	return nil
}
