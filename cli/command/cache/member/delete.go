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

	"github.com/dingodb/dingofs-tools/cli/cli"
	"github.com/dingodb/dingofs-tools/internal/common"
	"github.com/dingodb/dingofs-tools/internal/errno"
	"github.com/dingodb/dingofs-tools/internal/output"
	"github.com/dingodb/dingofs-tools/internal/rpc"
	"github.com/dingodb/dingofs-tools/internal/utils"
	pbmdserror "github.com/dingodb/dingofs-tools/proto/dingofs/proto/error"
	"github.com/dingodb/dingofs-tools/proto/dingofs/proto/mds"
	"github.com/spf13/cobra"
)

const (
	CACHEMEMBER_DELETE_EXAMPLE = `Examples:
   $ dingo cache member delete 6ba7b810-9dad-11d1-80b4-00c04fd430c8`
)

type deleteOptions struct {
	memberid  string
	format    string
	noConfirm bool
}

func NewCacheMemberDeleteCommand(dingoadm *cli.DingoAdm) *cobra.Command {
	var options deleteOptions

	cmd := &cobra.Command{
		Use:     "delete MEMBERID [OPTIONS]",
		Short:   "delete cache member",
		Args:    utils.ExactArgs(1),
		Example: CACHEMEMBER_DELETE_EXAMPLE,
		RunE: func(cmd *cobra.Command, args []string) error {
			utils.ReadCommandConfig(cmd)

			options.memberid = args[0]
			options.format = utils.GetStringFlag(cmd, utils.FORMAT)
			options.noConfirm = utils.GetBoolFlag(cmd, utils.DINGOFS_NOCONFIRM)

			output.SetShow(utils.GetBoolFlag(cmd, utils.VERBOSE))

			return runDelete(cmd, dingoadm, options)
		},
		SilenceUsage:          false,
		DisableFlagsInUseLine: true,
	}

	utils.SetFlagErrorFunc(cmd)

	// add flags
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

func runDelete(cmd *cobra.Command, dingoadm *cli.DingoAdm, options deleteOptions) error {
	// new rpc
	mdsRpc, err := rpc.CreateNewMdsRpc(cmd, "DeleteCacheMember")
	if err != nil {
		return err
	}

	outputResult := &common.OutputResult{
		Error: errno.ERR_OK,
	}
	// set request info
	deleteRpc := &rpc.DeleteCacheMemberRpc{
		Info: mdsRpc,
		Request: &mds.DeleteMemberRequest{
			MemberId: options.memberid,
		},
	}

	if !options.noConfirm && !utils.AskConfirmation(fmt.Sprintf("Are you sure to delete cachemember %s?", options.memberid), options.memberid) {
		return fmt.Errorf("abort delete cachemember")
	}

	// get rpc result
	response, rpcError := rpc.GetRpcResponse(deleteRpc.Info, deleteRpc)
	if rpcError.GetCode() != errno.ERR_OK.GetCode() {
		outputResult.Error = rpcError
	} else {
		result := response.(*mds.DeleteMemberResponse)
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
	fmt.Printf("Successfully delete cachemember %s\n", options.memberid)

	return nil
}
