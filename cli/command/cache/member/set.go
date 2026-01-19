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
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

const (
	CACHEMEMBER_SET_EXAMPLE = `Examples:
   $ dingo cache member set --memberid 6ba7b810-9dad-11d1-80b4-00c04fd430c8 --ip 10.220.69.6 --port 10001 --weight 90`
)

type setOptions struct {
	memberid string
	ip       string
	port     uint32
	weight   uint32
	format   string
}

func NewCacheMemberSetCommand(dingocli *cli.DingoCli) *cobra.Command {
	var options setOptions

	cmd := &cobra.Command{
		Use:     "set --memberid MEMBERID --ip IP --port PORT --weight WEIGHT [OPTIONS]",
		Short:   "set cache member weight",
		Args:    utils.NoArgs,
		Example: CACHEMEMBER_SET_EXAMPLE,
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			utils.ReadCommandConfig(cmd)

			options.memberid = utils.GetStringFlag(cmd, utils.DINGOFS_CACHE_MEMBERID)
			options.ip = utils.GetStringFlag(cmd, utils.DINGOFS_CACHE_IP)
			options.port = utils.GetUint32Flag(cmd, utils.DINGOFS_CACHE_PORT)
			options.weight, err = cmd.Flags().GetUint32("weight")
			if err != nil {
				return err
			}
			options.format = utils.GetStringFlag(cmd, utils.FORMAT)

			output.SetShow(utils.GetBoolFlag(cmd, utils.VERBOSE))

			return runSet(cmd, dingocli, options)
		},
		SilenceUsage:          false,
		DisableFlagsInUseLine: true,
	}

	utils.SetFlagErrorFunc(cmd)

	// add flags
	utils.AddStringRequiredFlag(cmd, utils.DINGOFS_CACHE_MEMBERID, "Cache member id")
	utils.AddStringRequiredFlag(cmd, utils.DINGOFS_CACHE_IP, "Cache member ip")
	utils.AddUint32RequiredFlag(cmd, utils.DINGOFS_CACHE_PORT, "Cache member port")
	cmd.Flags().Uint32("weight", 100, "Cache member weight"+color.RedString("[required]"))
	cmd.MarkFlagRequired("weight")

	utils.AddBoolFlag(cmd, utils.VERBOSE, "Show more debug info")
	utils.AddFormatFlag(cmd)
	utils.AddConfigFileFlag(cmd)

	utils.AddDurationFlag(cmd, utils.RPCTIMEOUT, "RPC timeout")
	utils.AddDurationFlag(cmd, utils.RPCRETRYDElAY, "RPC retry delay")
	utils.AddUint32Flag(cmd, utils.RPCRETRYTIMES, "RPC retry times")

	utils.AddStringFlag(cmd, utils.DINGOFS_MDSADDR, "Specify mds address")

	return cmd
}

func runSet(cmd *cobra.Command, dingocli *cli.DingoCli, options setOptions) error {
	// new rpc
	mdsRpc, err := rpc.CreateNewMdsRpc(cmd, "ReWeightMember")
	if err != nil {
		return err
	}

	outputResult := &common.OutputResult{
		Error: errno.ERR_OK,
	}
	// set request info
	reWeightRpc := &rpc.ReWeightMemberRpc{
		Info: mdsRpc,
		Request: &mds.ReweightMemberRequest{
			MemberId: options.memberid,
			Ip:       options.ip,
			Port:     options.port,
			Weight:   options.weight,
		},
	}

	// get rpc result
	response, rpcError := rpc.GetRpcResponse(reWeightRpc.Info, reWeightRpc)
	if rpcError.GetCode() != errno.ERR_OK.GetCode() {
		outputResult.Error = rpcError
	} else {
		result := response.(*mds.ReweightMemberResponse)
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
	fmt.Printf("Successfully reweight cachemember %s to %d\n", options.memberid, options.weight)

	return nil
}
