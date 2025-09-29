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

package set

import (
	"fmt"
	"github.com/dingodb/dingofs-tools/pkg/rpc"

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	pbmdserr "github.com/dingodb/dingofs-tools/proto/dingofs/proto/error"
	pbmds "github.com/dingodb/dingofs-tools/proto/dingofs/proto/mds"

	"github.com/dingodb/dingofs-tools/pkg/base"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	"github.com/spf13/cobra"
)

const (
	SetMemberExample = `$ dingo set cachemember  --memberid 6ba7b810-9dad-11d1-80b4-00c04fd430c8 --ip 10.220.69.6 --port 10001 --weight 90`
)

type ReweightMemberCommand struct {
	basecmd.FinalDingoCmd
	Rpc      *rpc.ReWeightMemberRpc
	response *pbmds.ReweightMemberResponse
}

var _ basecmd.FinalDingoCmdFunc = (*ReweightMemberCommand)(nil) // check interface

func NewCacheMemberCommand() *cobra.Command {
	reweightMemberCmd := &ReweightMemberCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:     "cachemember",
			Short:   "set remote cachegroup member attribute",
			Example: SetMemberExample,
		},
	}

	basecmd.NewFinalDingoCli(&reweightMemberCmd.FinalDingoCmd, reweightMemberCmd)
	return reweightMemberCmd.Cmd
}

func (reweightMember *ReweightMemberCommand) AddFlags() {
	config.AddRpcRetryTimesFlag(reweightMember.Cmd)
	config.AddRpcRetryDelayFlag(reweightMember.Cmd)
	config.AddRpcTimeoutFlag(reweightMember.Cmd)
	config.AddFsMdsAddrFlag(reweightMember.Cmd)
	config.AddCacheMemberIdFlag(reweightMember.Cmd)
	config.AddCacheMemberIp(reweightMember.Cmd)
	config.AddCacheMemberPort(reweightMember.Cmd)
	config.AddCacheMemberWeightFlag(reweightMember.Cmd)
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
	// new rpc
	mdsRpc, err := rpc.CreateNewMdsRpc(cmd, "ReweightMember")
	if err != nil {
		return err
	}
	// set request info
	memberId := config.GetFlagString(cmd, config.DINGOFS_CACHE_MEMBERID)
	ip := config.GetFlagString(cmd, config.DINGOFS_CACHE_IP)
	port := config.GetFlagUint32(cmd, config.DINGOFS_CACHE_PORT)
	weight := config.GetFlagUint32(cmd, config.DINGOFS_CACHE_WEIGHT)

	reweightMember.Rpc = &rpc.ReWeightMemberRpc{
		Info: mdsRpc,
		Request: &pbmds.ReweightMemberRequest{
			MemberId: memberId,
			Ip:       ip,
			Port:     port,
			Weight:   weight,
		},
	}

	response, cmdErr := base.GetRpcResponse(reweightMember.Rpc.Info, reweightMember.Rpc)
	if cmdErr.TypeCode() != cmderror.CODE_SUCCESS {
		return cmdErr.ToError()
	}

	result := response.(*pbmds.ReweightMemberResponse)
	var message string
	mdsError := result.GetError()
	if mdsError.GetErrcode() != pbmdserr.Errno_OK {
		message = fmt.Sprintf("reweight cahce member %s, error: %s", memberId, mdsError.String())
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
