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

package unregister

import (
	"fmt"

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	"github.com/dingodb/dingofs-tools/pkg/base"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	pbCacheGroup "github.com/dingodb/dingofs-tools/proto/dingofs/proto/cachegroup"
	"github.com/spf13/cobra"
)

const (
	UnRegisterMemberExample = `$ dingo unregister cachemember --ip 10.220.69.6 --port 10001`
)

type UnRegisterMemberCommand struct {
	basecmd.FinalDingoCmd
	Rpc *base.UnRegisterCacheMemberRpc
}

var _ basecmd.FinalDingoCmdFunc = (*UnRegisterMemberCommand)(nil) // check interface

func NewUnRegisterCacheMemberCommand() *cobra.Command {
	UnRegisterMemberCmd := &UnRegisterMemberCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:     "cachemember",
			Short:   "unregister cache member",
			Example: UnRegisterMemberExample,
		},
	}

	basecmd.NewFinalDingoCli(&UnRegisterMemberCmd.FinalDingoCmd, UnRegisterMemberCmd)
	return UnRegisterMemberCmd.Cmd
}

func (unRegisterMember *UnRegisterMemberCommand) AddFlags() {
	config.AddRpcRetryTimesFlag(unRegisterMember.Cmd)
	config.AddRpcRetryDelayFlag(unRegisterMember.Cmd)
	config.AddRpcTimeoutFlag(unRegisterMember.Cmd)
	config.AddFsMdsAddrFlag(unRegisterMember.Cmd)
	config.AddCacheMemberIp(unRegisterMember.Cmd)
	config.AddCacheMemberPort(unRegisterMember.Cmd)
}

func (unRegisterMember *UnRegisterMemberCommand) Init(cmd *cobra.Command, args []string) error {
	header := []string{cobrautil.ROW_RESULT}
	unRegisterMember.SetHeader(header)

	return nil
}

func (unRegisterMember *UnRegisterMemberCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&unRegisterMember.FinalDingoCmd, unRegisterMember)
}

func (unRegisterMember *UnRegisterMemberCommand) RunCommand(cmd *cobra.Command, args []string) error {
	addrs, addrErr := config.GetFsMdsAddrSlice(unRegisterMember.Cmd)
	if addrErr.TypeCode() != cmderror.CODE_SUCCESS {
		unRegisterMember.Error = addrErr
		return fmt.Errorf(addrErr.Message)
	}

	timeout := config.GetRpcTimeout(cmd)
	retryTimes := config.GetRpcRetryTimes(cmd)
	retryDelay := config.GetRpcRetryDelay(cmd)
	verbose := config.GetFlagBool(cmd, config.VERBOSE)
	rpcInfo := base.NewRpc(addrs, timeout, retryTimes, retryDelay, verbose, "UnregisterMember")

	ip := config.GetFlagString(cmd, config.DINGOFS_CACHE_IP)
	port := config.GetFlagUint32(cmd, config.DINGOFS_CACHE_PORT)

	rpc := &base.UnRegisterCacheMemberRpc{
		Info: rpcInfo,
		Request: &pbCacheGroup.DeregisterMemberRequest{
			Ip:   &ip,
			Port: &port,
		},
	}

	response, cmdErr := base.GetRpcResponse(rpc.Info, rpc)
	if cmdErr.TypeCode() != cmderror.CODE_SUCCESS {
		return cmdErr.ToError()
	}

	result := response.(*pbCacheGroup.DeregisterMemberResponse)
	var message string
	if result.GetStatus() != pbCacheGroup.CacheGroupErrCode_CacheGroupOk {
		message = fmt.Sprintf("unregister cache member[%s:%d] error: %s", ip, port, result.GetStatus().String())
	} else {
		message = cmderror.ErrSuccess().Message
	}

	row := map[string]string{
		cobrautil.ROW_RESULT: message,
	}
	unRegisterMember.TableNew.Append(cobrautil.Map2List(row, unRegisterMember.Header))

	// to json
	res, err := output.MarshalProtoJson(result)
	if err != nil {
		return err
	}
	mapRes := res.(map[string]interface{})
	unRegisterMember.Result = mapRes
	unRegisterMember.Error = cmderror.ErrSuccess()

	return nil
}

func (unRegisterMember *UnRegisterMemberCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&unRegisterMember.FinalDingoCmd)
}
