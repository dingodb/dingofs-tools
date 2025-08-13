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

package deregister

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
	deRegisterMemberExample = `$ dingo deregister cachemember --ip 10.220.69.6 --port 10001`
)

type deRegisterMemberCommand struct {
	basecmd.FinalDingoCmd
	Rpc *base.DeregisterCacheMemberRpc
}

var _ basecmd.FinalDingoCmdFunc = (*deRegisterMemberCommand)(nil) // check interface

func NewDeregisterCacheMemberCommand() *cobra.Command {
	deRegisterMemberCmd := &deRegisterMemberCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:     "cachemember",
			Short:   "deregister cache member",
			Example: deRegisterMemberExample,
		},
	}

	basecmd.NewFinalDingoCli(&deRegisterMemberCmd.FinalDingoCmd, deRegisterMemberCmd)
	return deRegisterMemberCmd.Cmd
}

func (deRegisterMember *deRegisterMemberCommand) AddFlags() {
	config.AddRpcRetryTimesFlag(deRegisterMember.Cmd)
	config.AddRpcRetryDelayFlag(deRegisterMember.Cmd)
	config.AddRpcTimeoutFlag(deRegisterMember.Cmd)
	config.AddFsMdsAddrFlag(deRegisterMember.Cmd)
	config.AddCacheMemberIp(deRegisterMember.Cmd)
	config.AddCacheMemberPort(deRegisterMember.Cmd)
}

func (deRegisterMember *deRegisterMemberCommand) Init(cmd *cobra.Command, args []string) error {
	header := []string{cobrautil.ROW_RESULT}
	deRegisterMember.SetHeader(header)

	return nil
}

func (deRegisterMember *deRegisterMemberCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&deRegisterMember.FinalDingoCmd, deRegisterMember)
}

func (deRegisterMember *deRegisterMemberCommand) RunCommand(cmd *cobra.Command, args []string) error {
	addrs, addrErr := config.GetFsMdsAddrSlice(deRegisterMember.Cmd)
	if addrErr.TypeCode() != cmderror.CODE_SUCCESS {
		deRegisterMember.Error = addrErr
		return fmt.Errorf(addrErr.Message)
	}

	timeout := config.GetRpcTimeout(cmd)
	retryTimes := config.GetRpcRetryTimes(cmd)
	retryDelay := config.GetRpcRetryDelay(cmd)
	verbose := config.GetFlagBool(cmd, config.VERBOSE)
	rpcInfo := base.NewRpc(addrs, timeout, retryTimes, retryDelay, verbose, "DeregisterMember")

	ip := config.GetFlagString(cmd, config.DINGOFS_CACHE_IP)
	port := config.GetFlagUint32(cmd, config.DINGOFS_CACHE_PORT)

	rpc := &base.DeregisterCacheMemberRpc{
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
	deRegisterMember.TableNew.Append(cobrautil.Map2List(row, deRegisterMember.Header))

	// to json
	res, err := output.MarshalProtoJson(result)
	if err != nil {
		return err
	}
	mapRes := res.(map[string]interface{})
	deRegisterMember.Result = mapRes
	deRegisterMember.Error = cmderror.ErrSuccess()

	return nil
}

func (deRegisterMember *deRegisterMemberCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&deRegisterMember.FinalDingoCmd)
}
