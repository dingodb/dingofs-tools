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

	"github.com/dingodb/dingocli/cli/cli"
	"github.com/dingodb/dingocli/internal/common"
	"github.com/dingodb/dingocli/internal/errno"
	"github.com/dingodb/dingocli/internal/output"
	"github.com/dingodb/dingocli/internal/rpc"
	"github.com/dingodb/dingocli/internal/utils"
	pbmdserror "github.com/dingodb/dingocli/proto/dingofs/proto/error"
	"github.com/dingodb/dingocli/proto/dingofs/proto/mds"
	"github.com/spf13/cobra"
)

const (
	CACHEMEMBER_LEAVE_EXAMPLE = `Examples:
   $ dingo cache member leave --group group1 --memberid 6ba7b810-9dad-11d1-80b4-00c04fd430c8 --ip 10.220.69.6 --port 10001`
)

type leaveOptions struct {
	group    string
	memberid string
	ip       string
	port     uint32
	format   string
}

func NewCacheMemberLeaveCommand(dingocli *cli.DingoCli) *cobra.Command {
	var options leaveOptions

	cmd := &cobra.Command{
		Use:     "leave [OPTIONS]",
		Short:   "leave cache member from group",
		Args:    utils.NoArgs,
		Example: CACHEMEMBER_LEAVE_EXAMPLE,
		RunE: func(cmd *cobra.Command, args []string) error {
			utils.ReadCommandConfig(cmd)

			options.group = utils.GetStringFlag(cmd, utils.DINGOFS_CACHE_GROUP)
			options.memberid = utils.GetStringFlag(cmd, utils.DINGOFS_CACHE_MEMBERID)
			options.ip = utils.GetStringFlag(cmd, utils.DINGOFS_CACHE_IP)
			options.port = utils.GetUint32Flag(cmd, utils.DINGOFS_CACHE_PORT)

			options.format = utils.GetStringFlag(cmd, utils.FORMAT)

			output.SetShow(utils.GetBoolFlag(cmd, utils.VERBOSE))

			return runLeave(cmd, dingocli, options)
		},
		SilenceUsage:          false,
		DisableFlagsInUseLine: true,
	}

	utils.SetFlagErrorFunc(cmd)

	// add flags
	utils.AddStringRequiredFlag(cmd, utils.DINGOFS_CACHE_GROUP, "Cache group id")
	utils.AddStringRequiredFlag(cmd, utils.DINGOFS_CACHE_MEMBERID, "Cache member id")
	utils.AddStringRequiredFlag(cmd, utils.DINGOFS_CACHE_IP, "Cache member ip")
	utils.AddUint32RequiredFlag(cmd, utils.DINGOFS_CACHE_PORT, "Cache member port")

	utils.AddBoolFlag(cmd, utils.DINGOFS_NOCONFIRM, "Do not confirm the command")
	utils.AddBoolFlag(cmd, utils.VERBOSE, "Show more debug info")
	utils.AddFormatFlag(cmd)
	utils.AddConfigFileFlag(cmd)

	utils.AddDurationFlag(cmd, utils.RPCTIMEOUT, "RPC timeout")
	utils.AddDurationFlag(cmd, utils.RPCRETRYDElAY, "RPC retry delay")
	utils.AddUint32Flag(cmd, utils.RPCRETRYTIMES, "RPC retry times")

	utils.AddStringFlag(cmd, utils.DINGOFS_MDSADDR, "Specify mds address")

	return cmd
}

func runLeave(cmd *cobra.Command, dingocli *cli.DingoCli, options leaveOptions) error {
	// new rpc
	mdsRpc, err := rpc.CreateNewMdsRpc(cmd, "LeaveCacheMember")
	if err != nil {
		return err
	}

	outputResult := &common.OutputResult{
		Error: errno.ERR_OK,
	}
	// set request info
	leaveRpc := &rpc.LeaveCacheMemberRpc{
		Info: mdsRpc,
		Request: &mds.LeaveCacheGroupRequest{
			GroupName: options.group,
			MemberId:  options.memberid,
			Ip:        options.ip,
			Port:      options.port,
		},
	}

	// get rpc result
	response, rpcError := rpc.GetRpcResponse(leaveRpc.Info, leaveRpc)
	if rpcError.GetCode() != errno.ERR_OK.GetCode() {
		outputResult.Error = rpcError
	} else {
		result := response.(*mds.LeaveCacheGroupResponse)
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
	fmt.Printf("Successfully leave cachemember %s\n", options.memberid)

	return nil
}
