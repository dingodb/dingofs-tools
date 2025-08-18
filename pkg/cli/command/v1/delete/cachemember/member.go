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
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	rpc "github.com/dingodb/dingofs-tools/pkg/rpc/v1"
	pbCacheGroup "github.com/dingodb/dingofs-tools/proto/dingofs/proto/cachegroup"
	"github.com/spf13/cobra"
)

const (
	DeleteMemberExample = `
$ dingo delete cachemember --group group1 --memberid 6ba7b810-9dad-11d1-80b4-00c04fd430c8
$ dingo delete cachemember --group group1 --ip 10.220.69.6 --port 10001`
)

type CacheMemberCommand struct {
	basecmd.FinalDingoCmd
	Rpc      *rpc.LeaveCacheMemberRpc
	response *pbCacheGroup.LeaveCacheGroupResponse
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
	config.AddCacheGroupFlag(cacheMember.Cmd)
	config.AddCacheMemberIdOptionalFlag(cacheMember.Cmd)
	config.AddCacheMemberIpOptionalFlag(cacheMember.Cmd)
	config.AddCacheMemberPortOptionalFlag(cacheMember.Cmd)
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
	addrs, addrErr := config.GetFsMdsAddrSlice(cacheMember.Cmd)
	if addrErr.TypeCode() != cmderror.CODE_SUCCESS {
		cacheMember.Error = addrErr
		return fmt.Errorf(addrErr.Message)
	}

	timeout := config.GetRpcTimeout(cmd)
	retryTimes := config.GetRpcRetryTimes(cmd)
	retryDelay := config.GetRpcRetryDelay(cmd)
	verbose := config.GetFlagBool(cmd, config.VERBOSE)
	rpcInfo := base.NewRpc(addrs, timeout, retryTimes, retryDelay, verbose, "LeaveCacheMember")

	groupName := config.GetFlagString(cmd, config.DINGOFS_CACHE_GROUP)
	memberId := config.GetFlagString(cmd, config.DINGOFS_CACHE_MEMBERID)
	ip := config.GetFlagString(cmd, config.DINGOFS_CACHE_IP)
	port := config.GetFlagUint32(cmd, config.DINGOFS_CACHE_PORT)

	if (len(memberId) == 0) && (len(ip) == 0 || port == 0) {
		return fmt.Errorf("memberId or (ip,port) is invalid")
	}

	request := pbCacheGroup.LeaveCacheGroupRequest{GroupName: &groupName}
	if len(memberId) > 0 {
		request.MemberId = &memberId
	} else {
		request.Ip = &ip
		request.Port = &port
	}

	rpc := &rpc.LeaveCacheMemberRpc{
		Info:    rpcInfo,
		Request: &request,
	}

	response, cmdErr := base.GetRpcResponse(rpc.Info, rpc)
	if cmdErr.TypeCode() != cmderror.CODE_SUCCESS {
		return cmdErr.ToError()
	}

	result := response.(*pbCacheGroup.LeaveCacheGroupResponse)
	dingoCacheErr := cmderror.ErrDingoCacheRequest(result.GetStatus())
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
	cacheMember.Error = dingoCacheErr

	return nil
}

func (cacheMember *CacheMemberCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&cacheMember.FinalDingoCmd)
}
