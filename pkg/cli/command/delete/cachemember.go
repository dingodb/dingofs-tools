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

package delete

import (
	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	"github.com/dingodb/dingofs-tools/pkg/rpc"
	pbmdsv2 "github.com/dingodb/dingofs-tools/proto/dingofs/proto/mdsv2"

	"github.com/dingodb/dingofs-tools/pkg/base"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	"github.com/spf13/cobra"
)

const (
	DeleteMemberExample = `
$ dingo delete cachemember --memberid 6ba7b810-9dad-11d1-80b4-00c04fd430c8`
)

type CacheMemberCommand struct {
	basecmd.FinalDingoCmd
	Rpc      *rpc.DeleteCacheMemberRpc
	response *pbmdsv2.DeleteMemberResponse
}

var _ basecmd.FinalDingoCmdFunc = (*CacheMemberCommand)(nil) // check interface

func NewDeleteCacheMemberCommand() *cobra.Command {
	cacheMemberCmd := &CacheMemberCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:     "cachemember",
			Short:   "delete cachegroup member",
			Example: DeleteMemberExample,
		},
	}

	basecmd.NewFinalDingoCli(&cacheMemberCmd.FinalDingoCmd, cacheMemberCmd)
	return cacheMemberCmd.Cmd
}

func (cacheMember *CacheMemberCommand) AddFlags() {
	config.AddRpcRetryTimesFlag(cacheMember.Cmd)
	config.AddRpcRetryDelayFlag(cacheMember.Cmd)
	config.AddRpcTimeoutFlag(cacheMember.Cmd)
	config.AddFsMdsAddrFlag(cacheMember.Cmd)
	config.AddCacheMemberIdFlag(cacheMember.Cmd)
}

func (cacheMember *CacheMemberCommand) Init(cmd *cobra.Command, args []string) error {
	header := []string{cobrautil.ROW_RESULT}
	cacheMember.SetHeader(header)

	return nil
}

func (cacheMember *CacheMemberCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&cacheMember.FinalDingoCmd, cacheMember)
}

func (cacheMember *CacheMemberCommand) RunCommand(cmd *cobra.Command, args []string) error {
	// new rpc
	mdsRpc, err := rpc.CreateNewMdsRpc(cmd, "DeleteCacheMember")
	if err != nil {
		return err
	}
	// set request info
	memberId := config.GetFlagString(cmd, config.DINGOFS_CACHE_MEMBERID)

	cacheMember.Rpc = &rpc.DeleteCacheMemberRpc{
		Info: mdsRpc,
		Request: &pbmdsv2.DeleteMemberRequest{
			MemberId: memberId,
		},
	}

	response, cmdErr := base.GetRpcResponse(cacheMember.Rpc.Info, cacheMember.Rpc)
	if cmdErr.TypeCode() != cmderror.CODE_SUCCESS {
		return cmdErr.ToError()
	}

	result := response.(*pbmdsv2.DeleteMemberResponse)
	dingoCacheErr := cmderror.MDSV2Error(result.GetError())
	row := map[string]string{
		cobrautil.ROW_RESULT: dingoCacheErr.Message,
	}
	cacheMember.TableNew.Append(cobrautil.Map2List(row, cacheMember.Header))

	// to json
	res, err := output.MarshalProtoJson(result)
	if err != nil {
		return err
	}
	mapRes := res.(map[string]interface{})
	cacheMember.Result = mapRes
	cacheMember.Error = cmderror.ErrSuccess()

	return nil
}

func (cacheMember *CacheMemberCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&cacheMember.FinalDingoCmd)
}
