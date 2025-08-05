// Copyright (c) 2025 dingodb.com, Inc. All Rights Reserved
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cachemember

import (
	"fmt"

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	"github.com/dingodb/dingofs-tools/pkg/base"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/v1/common"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	"github.com/dingodb/dingofs-tools/proto/dingofs/proto/cachegroup"
	"github.com/spf13/cobra"
)

const (
	SetMemberExample = `$ dingo set cachemember --memberid 1 --weight 100`
)

type ReweightMemberCommand struct {
	basecmd.FinalDingoCmd
	Rpc      *common.ReweightMemberRpc
	response *cachegroup.ReweightMemberResponse
}

var _ basecmd.FinalDingoCmdFunc = (*ReweightMemberCommand)(nil) // check interface

func NewCacheMemberCommand() *cobra.Command {
	return NewListCacheMemberCommand().Cmd
}

func NewListCacheMemberCommand() *ReweightMemberCommand {
	reweightMemberCmd := &ReweightMemberCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:     "cachemember",
			Short:   "set remote cachegroup member attribute",
			Example: SetMemberExample,
		},
	}

	basecmd.NewFinalDingoCli(&reweightMemberCmd.FinalDingoCmd, reweightMemberCmd)
	return reweightMemberCmd
}

func (reweightMember *ReweightMemberCommand) AddFlags() {
	config.AddRpcRetryTimesFlag(reweightMember.Cmd)
	config.AddRpcRetryDelayFlag(reweightMember.Cmd)
	config.AddRpcTimeoutFlag(reweightMember.Cmd)
	config.AddFsMdsAddrFlag(reweightMember.Cmd)
	config.AddCacheMemberId(reweightMember.Cmd)
	config.AddCacheMemberWeight(reweightMember.Cmd)
}

func (reweightMember *ReweightMemberCommand) Init(cmd *cobra.Command, args []string) error {
	header := []string{cobrautil.ROW_RESULT}
	reweightMember.SetHeader(header)

	return nil
}

func (reweightMember *ReweightMemberCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&reweightMember.FinalDingoCmd, reweightMember)
}

func (reweightMember *ReweightMemberCommand) RunCommand(cmd *cobra.Command, args []string) error {
	addrs, addrErr := config.GetFsMdsAddrSlice(reweightMember.Cmd)
	if addrErr.TypeCode() != cmderror.CODE_SUCCESS {
		reweightMember.Error = addrErr
		return fmt.Errorf(addrErr.Message)
	}

	timeout := config.GetRpcTimeout(cmd)
	retryTimes := config.GetRpcRetryTimes(cmd)
	retryDelay := config.GetRpcRetryDelay(cmd)
	verbose := config.GetFlagBool(cmd, config.VERBOSE)
	rpcInfo := base.NewRpc(addrs, timeout, retryTimes, retryDelay, verbose, "LoadMembers")

	memberId := config.GetFlagUint64(cmd, config.DINGOFS_CACHE_MEMBERID)
	weight := config.GetFlagUint32(cmd, config.DINGOFS_CACHE_WEIGHT)

	rpc := &common.ReweightMemberRpc{
		Info: rpcInfo,
		Request: &cachegroup.ReweightMemberRequest{
			MemberId: &memberId,
			Weight:   &weight,
		},
	}

	response, cmdErr := base.GetRpcResponse(rpc.Info, rpc)
	if cmdErr.TypeCode() != cmderror.CODE_SUCCESS {
		return cmdErr.ToError()
	}

	result := response.(*cachegroup.ReweightMemberResponse)
	var message string
	if result.GetStatus() != cachegroup.CacheGroupErrCode_CacheGroupOk {
		message = fmt.Sprintf("reweight cahce member %d error: %s", memberId, result.GetStatus().String())
	} else {
		message = cmderror.ErrSuccess().Message
	}

	row := map[string]string{
		cobrautil.ROW_RESULT: message,
	}
	reweightMember.TableNew.Append(cobrautil.Map2List(row, reweightMember.Header))

	// to json
	res, err := output.MarshalProtoJson(result)
	if err != nil {
		return err
	}
	mapRes := res.(map[string]interface{})
	reweightMember.Result = mapRes
	reweightMember.Error = cmderror.ErrSuccess()

	return nil
}

func (reweightMember *ReweightMemberCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&reweightMember.FinalDingoCmd)
}
