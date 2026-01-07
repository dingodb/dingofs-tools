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

package group

import (
	"github.com/dingodb/dingofs-tools/cli/cli"
	"github.com/dingodb/dingofs-tools/internal/common"
	"github.com/dingodb/dingofs-tools/internal/errno"
	"github.com/dingodb/dingofs-tools/internal/output"
	"github.com/dingodb/dingofs-tools/internal/rpc"
	"github.com/dingodb/dingofs-tools/internal/table"
	"github.com/dingodb/dingofs-tools/internal/utils"

	pbmdserror "github.com/dingodb/dingofs-tools/proto/dingofs/proto/error"
	"github.com/dingodb/dingofs-tools/proto/dingofs/proto/mds"
	"github.com/spf13/cobra"
)

const (
	CACHEGROUP_LIST_EXAMPLE = `Examples:
   $ dingo cache group list`
)

type listOptions struct {
	format string
}

func NewCacheGroupListCommand(dingoadm *cli.DingoAdm) *cobra.Command {
	var options listOptions

	cmd := &cobra.Command{
		Use:     "list [OPTIONS]",
		Short:   "list all remote cachegroup",
		Args:    utils.NoArgs,
		Example: CACHEGROUP_LIST_EXAMPLE,
		RunE: func(cmd *cobra.Command, args []string) error {
			utils.ReadCommandConfig(cmd)

			options.format = utils.GetStringFlag(cmd, utils.FORMAT)

			output.SetShow(utils.GetBoolFlag(cmd, utils.VERBOSE))

			return runList(cmd, dingoadm, options)
		},
		SilenceUsage:          false,
		DisableFlagsInUseLine: true,
	}

	utils.SetFlagErrorFunc(cmd)

	// add flags
	utils.AddBoolFlag(cmd, utils.VERBOSE, "Show more debug info")
	utils.AddConfigFileFlag(cmd)
	utils.AddFormatFlag(cmd)

	utils.AddDurationFlag(cmd, utils.RPCTIMEOUT, "RPC timeout")
	utils.AddDurationFlag(cmd, utils.RPCRETRYDElAY, "RPC retry delay")
	utils.AddUint32Flag(cmd, utils.RPCRETRYTIMES, "RPC retry times")

	utils.AddStringFlag(cmd, utils.DINGOFS_MDSADDR, "Specify mds address")

	return cmd
}

func runList(cmd *cobra.Command, dingoadm *cli.DingoAdm, options listOptions) error {
	outputResult := &common.OutputResult{
		Error: errno.ERR_OK,
	}
	// new rpc
	mdsRpc, err := rpc.CreateNewMdsRpc(cmd, "ListGroups")
	if err != nil {
		return err
	}

	// set request info
	listRpc := &rpc.ListCacheGroupRpc{
		Info:    mdsRpc,
		Request: &mds.ListGroupsRequest{},
	}

	// get rpc result
	var result *mds.ListGroupsResponse
	response, rpcError := rpc.GetRpcResponse(listRpc.Info, listRpc)
	if rpcError.GetCode() != errno.ERR_OK.GetCode() {
		outputResult.Error = rpcError
	} else {
		result = response.(*mds.ListGroupsResponse)
		if mdsErr := result.GetError(); mdsErr.GetErrcode() != pbmdserror.Errno_OK {
			outputResult.Error = errno.ERR_RPC_FAILED.S(mdsErr.String())
		}
		outputResult.Result = result
	}

	// print result
	if options.format == "json" {
		return output.OutputJson(outputResult)
	}

	if outputResult.Error.GetCode() != errno.ERR_OK.GetCode() {
		return outputResult.Error
	}

	// set table header
	header := []string{common.ROW_GROUP}
	table.SetHeader(header)
	// fill table
	groups := result.GetGroupNames()
	rows := make([]map[string]string, 0)
	for _, group := range groups {
		row := make(map[string]string)
		row[common.ROW_GROUP] = group

		rows = append(rows, row)
	}

	list := table.ListMap2ListSortByKeys(rows, header, []string{common.ROW_GROUP})
	table.AppendBulk(list)
	table.RenderWithNoData("no cachegroup in cluster")

	return nil
}
