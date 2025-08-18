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

package register

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
	RegisterMemberExample = `$ dingo register cachemember --ip 10.220.69.6 --port 10001 --memberid 6ba7b810-9dad-11d1-80b4-00c04fd430c8`
)

type RegisterMemberCommand struct {
	basecmd.FinalDingoCmd
	Rpc *rpc.RegisterCacheMemberRpc
}

var _ basecmd.FinalDingoCmdFunc = (*RegisterMemberCommand)(nil) // check interface

func NewRegisterCacheMemberCommand() *cobra.Command {
	registerMemberCmd := &RegisterMemberCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:     "cachemember",
			Short:   "register cache member",
			Example: RegisterMemberExample,
		},
	}

	basecmd.NewFinalDingoCli(&registerMemberCmd.FinalDingoCmd, registerMemberCmd)
	return registerMemberCmd.Cmd
}

func (registerMember *RegisterMemberCommand) AddFlags() {
	config.AddRpcRetryTimesFlag(registerMember.Cmd)
	config.AddRpcRetryDelayFlag(registerMember.Cmd)
	config.AddRpcTimeoutFlag(registerMember.Cmd)
	config.AddFsMdsAddrFlag(registerMember.Cmd)
	config.AddCacheMemberIdFlag(registerMember.Cmd)
	config.AddCacheMemberIp(registerMember.Cmd)
	config.AddCacheMemberPort(registerMember.Cmd)
}

func (registerMember *RegisterMemberCommand) Init(cmd *cobra.Command, args []string) error {
	header := []string{cobrautil.ROW_RESULT}
	registerMember.SetHeader(header)

	return nil
}

func (registerMember *RegisterMemberCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&registerMember.FinalDingoCmd, registerMember)
}

func (registerMember *RegisterMemberCommand) RunCommand(cmd *cobra.Command, args []string) error {
	addrs, addrErr := config.GetFsMdsAddrSlice(registerMember.Cmd)
	if addrErr.TypeCode() != cmderror.CODE_SUCCESS {
		registerMember.Error = addrErr
		return fmt.Errorf(addrErr.Message)
	}

	timeout := config.GetRpcTimeout(cmd)
	retryTimes := config.GetRpcRetryTimes(cmd)
	retryDelay := config.GetRpcRetryDelay(cmd)
	verbose := config.GetFlagBool(cmd, config.VERBOSE)
	rpcInfo := base.NewRpc(addrs, timeout, retryTimes, retryDelay, verbose, "RegisterMember")

	ip := config.GetFlagString(cmd, config.DINGOFS_CACHE_IP)
	port := config.GetFlagUint32(cmd, config.DINGOFS_CACHE_PORT)
	memberId := config.GetFlagString(cmd, config.DINGOFS_CACHE_MEMBERID)

	rpc := &rpc.RegisterCacheMemberRpc{
		Info: rpcInfo,
		Request: &pbCacheGroup.RegisterMemberRequest{
			Ip:     &ip,
			Port:   &port,
			WantId: &memberId,
		},
	}

	response, cmdErr := base.GetRpcResponse(rpc.Info, rpc)
	if cmdErr.TypeCode() != cmderror.CODE_SUCCESS {
		return cmdErr.ToError()
	}

	result := response.(*pbCacheGroup.RegisterMemberResponse)
	var message string
	if result.GetStatus() != pbCacheGroup.CacheGroupErrCode_CacheGroupOk {
		message = fmt.Sprintf("register cahce member[%s:%d,%s] error: %s", ip, port, memberId, result.GetStatus().String())
	} else {
		message = cmderror.ErrSuccess().Message
	}

	row := map[string]string{
		cobrautil.ROW_RESULT: message,
	}
	registerMember.TableNew.Append(cobrautil.Map2List(row, registerMember.Header))

	// to json
	res, err := output.MarshalProtoJson(result)
	if err != nil {
		return err
	}
	mapRes := res.(map[string]interface{})
	registerMember.Result = mapRes
	registerMember.Error = cmderror.ErrSuccess()

	return nil
}

func (registerMember *RegisterMemberCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&registerMember.FinalDingoCmd)
}
