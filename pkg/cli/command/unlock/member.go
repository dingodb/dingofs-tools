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

package unlock

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
	UnlockMemberExample = `$ dingo unlock cachemember  --memberid 6ba7b810-9dad-11d1-80b4-00c04fd430c8 --ip 10.220.69.6 --port 10001`
)

type UnlockMemberCommand struct {
	basecmd.FinalDingoCmd
	Rpc      *rpc.UnlockCacheMemberRpc
	response *pbmds.UnLockMemberResponse
}

var _ basecmd.FinalDingoCmdFunc = (*UnlockMemberCommand)(nil) // check interface

func NewCacheMemberCommand() *cobra.Command {
	unlockMemberCmd := &UnlockMemberCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:     "cachemember",
			Short:   "unlock cache member",
			Example: UnlockMemberExample,
		},
	}

	basecmd.NewFinalDingoCli(&unlockMemberCmd.FinalDingoCmd, unlockMemberCmd)
	return unlockMemberCmd.Cmd
}

func (unlockMember *UnlockMemberCommand) AddFlags() {
	config.AddRpcRetryTimesFlag(unlockMember.Cmd)
	config.AddRpcRetryDelayFlag(unlockMember.Cmd)
	config.AddRpcTimeoutFlag(unlockMember.Cmd)
	config.AddFsMdsAddrFlag(unlockMember.Cmd)
	config.AddCacheMemberIdFlag(unlockMember.Cmd)
	config.AddCacheMemberIp(unlockMember.Cmd)
	config.AddCacheMemberPort(unlockMember.Cmd)
}

func (unlockMember *UnlockMemberCommand) Init(cmd *cobra.Command, args []string) error {
	header := []string{cobrautil.ROW_RESULT}
	unlockMember.SetHeader(header)

	return nil
}

func (unlockMember *UnlockMemberCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&unlockMember.FinalDingoCmd, unlockMember)
}

func (unlockMember *UnlockMemberCommand) RunCommand(cmd *cobra.Command, args []string) error {
	// new rpc
	mdsRpc, err := rpc.CreateNewMdsRpc(cmd, "UnlockMember")
	if err != nil {
		return err
	}
	// set request info
	memberId := config.GetFlagString(cmd, config.DINGOFS_CACHE_MEMBERID)
	ip := config.GetFlagString(cmd, config.DINGOFS_CACHE_IP)
	port := config.GetFlagUint32(cmd, config.DINGOFS_CACHE_PORT)

	unlockMember.Rpc = &rpc.UnlockCacheMemberRpc{
		Info: mdsRpc,
		Request: &pbmds.UnLockMemberRequest{
			MemberId: memberId,
			Ip:       ip,
			Port:     port,
		},
	}

	response, cmdErr := base.GetRpcResponse(unlockMember.Rpc.Info, unlockMember.Rpc)
	if cmdErr.TypeCode() != cmderror.CODE_SUCCESS {
		return cmdErr.ToError()
	}

	result := response.(*pbmds.UnLockMemberResponse)
	var message string
	mdsError := result.GetError()
	if mdsError.GetErrcode() != pbmdserr.Errno_OK {
		message = fmt.Sprintf("unlock cahce member %s, error: %s", memberId, mdsError.String())
	} else {
		message = cmderror.ErrSuccess().Message
	}

	row := map[string]string{
		cobrautil.ROW_RESULT: message,
	}
	unlockMember.TableNew.Append(cobrautil.Map2List(row, unlockMember.Header))

	// to json
	res, err := output.MarshalProtoJson(result)
	if err != nil {
		return err
	}
	mapRes := res.(map[string]interface{})
	unlockMember.Result = mapRes
	unlockMember.Error = cmderror.ErrSuccess()

	return nil
}

func (unlockMember *UnlockMemberCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&unlockMember.FinalDingoCmd)
}
