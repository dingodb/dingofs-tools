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

package member

import (
	"fmt"
	"time"

	"github.com/dingodb/dingocli/cli/cli"
	"github.com/dingodb/dingocli/internal/common"
	"github.com/dingodb/dingocli/internal/errno"
	"github.com/dingodb/dingocli/internal/output"
	"github.com/dingodb/dingocli/internal/rpc"
	"github.com/dingodb/dingocli/internal/table"
	"github.com/dingodb/dingocli/internal/utils"

	pbmdserror "github.com/dingodb/dingocli/proto/dingofs/proto/error"
	"github.com/dingodb/dingocli/proto/dingofs/proto/mds"
	"github.com/spf13/cobra"
)

const (
	CACHEGROUP_LIST_EXAMPLE = `Examples:
   $ dingo cache member list`
)

type listOptions struct {
	group  string
	format string
}

func NewCacheMemberListCommand(dingocli *cli.DingoCli) *cobra.Command {
	var options listOptions

	cmd := &cobra.Command{
		Use:     "list [OPTIONS]",
		Short:   "list all cache members",
		Args:    utils.NoArgs,
		Example: CACHEGROUP_LIST_EXAMPLE,
		RunE: func(cmd *cobra.Command, args []string) error {
			utils.ReadCommandConfig(cmd)

			options.group = utils.GetStringFlag(cmd, utils.DINGOFS_CACHE_GROUP)
			options.format = utils.GetStringFlag(cmd, utils.FORMAT)

			output.SetShow(utils.GetBoolFlag(cmd, utils.VERBOSE))

			return runList(cmd, dingocli, options)
		},
		SilenceUsage:          false,
		DisableFlagsInUseLine: true,
	}

	utils.SetFlagErrorFunc(cmd)

	// add flags
	utils.AddStringFlag(cmd, utils.DINGOFS_CACHE_GROUP, "Cachegroup name")

	utils.AddBoolFlag(cmd, utils.VERBOSE, "Show more debug info")
	utils.AddConfigFileFlag(cmd)
	utils.AddFormatFlag(cmd)

	utils.AddDurationFlag(cmd, utils.RPCTIMEOUT, "RPC timeout")
	utils.AddDurationFlag(cmd, utils.RPCRETRYDElAY, "RPC retry delay")
	utils.AddUint32Flag(cmd, utils.RPCRETRYTIMES, "RPC retry times")

	utils.AddStringFlag(cmd, utils.DINGOFS_MDSADDR, "Specify mds address")

	return cmd
}

func runList(cmd *cobra.Command, dingocli *cli.DingoCli, options listOptions) error {
	outputResult := &common.OutputResult{
		Error: errno.ERR_OK,
	}
	// new rpc
	mdsRpc, err := rpc.CreateNewMdsRpc(cmd, "ListMembers")
	if err != nil {
		return err
	}

	// set request info
	request := mds.ListMembersRequest{}
	if len(options.group) != 0 {
		request.GroupName = &options.group
	}
	listRpc := &rpc.ListCacheMemberRpc{
		Info:    mdsRpc,
		Request: &request,
	}

	// get rpc result
	var result *mds.ListMembersResponse
	response, rpcError := rpc.GetRpcResponse(listRpc.Info, listRpc)
	if rpcError.GetCode() != errno.ERR_OK.GetCode() {
		outputResult.Error = rpcError
	} else {
		result = response.(*mds.ListMembersResponse)
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
	header := []string{common.ROW_MEMBERID, common.ROW_IP, common.ROW_PORT, common.ROW_WEIGHT, common.ROW_LOCKED, common.ROW_CREATE_TIME, common.ROW_LASTONLINETIME, common.ROW_STATE, common.ROW_GROUP}
	table.SetHeader(header)
	// fill table
	members := result.GetMembers()
	rows := make([]map[string]string, 0)
	for _, member := range members {
		row := make(map[string]string)
		row[common.ROW_MEMBERID] = member.GetMemberId()
		ip := member.GetIp()
		port := member.GetPort()
		if len(ip) > 0 && port > 0 {
			row[common.ROW_IP] = member.GetIp()
			row[common.ROW_PORT] = fmt.Sprintf("%d", member.GetPort())
			row[common.ROW_WEIGHT] = fmt.Sprintf("%d", member.GetWeight())
			row[common.ROW_STATE] = utils.TranslateCacheGroupMemberState(member.GetState())
			row[common.ROW_GROUP] = member.GetGroupName()
			row[common.ROW_LOCKED] = fmt.Sprintf("%v", member.GetLocked())

			// process create time
			seconds := int64(member.GetCreateTimeS())
			if seconds > 0 {
				createTime := time.Unix(seconds, 0)
				row[common.ROW_CREATE_TIME] = createTime.Format("2006-01-02 15:04:05.000")
			}
			// process online time
			ms := int64(member.GetLastOnlineTimeMs())
			if ms > 0 {
				sec := ms / 1000
				nsec := (ms % 1000) * 1000000
				onlineTime := time.Unix(sec, nsec)
				row[common.ROW_LASTONLINETIME] = onlineTime.Format("2006-01-02 15:04:05.000")
			}
		}

		rows = append(rows, row)
	}

	list := table.ListMap2ListSortByKeys(rows, header, []string{common.ROW_GROUP})
	table.AppendBulk(list)
	table.RenderWithNoData("no cachemember in cluster")

	return nil
}
