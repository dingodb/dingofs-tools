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
	"time"

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
	ListMemberExample = `$ dingo list cachemember --group group1`
)

type CacheMemberCommand struct {
	basecmd.FinalDingoCmd
	Rpc      *common.ListCacheMemberRpc
	response *cachegroup.LoadMembersResponse
}

var _ basecmd.FinalDingoCmdFunc = (*CacheMemberCommand)(nil) // check interface

func NewCacheMemberCommand() *cobra.Command {
	return NewListCacheMemberCommand().Cmd
}

func NewListCacheMemberCommand() *CacheMemberCommand {
	cacheMemberCmd := &CacheMemberCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:     "cachemember",
			Short:   "list cache members in cachegroup",
			Example: ListMemberExample,
		},
	}

	basecmd.NewFinalDingoCli(&cacheMemberCmd.FinalDingoCmd, cacheMemberCmd)
	return cacheMemberCmd
}

func (cacheMember *CacheMemberCommand) AddFlags() {
	config.AddRpcRetryTimesFlag(cacheMember.Cmd)
	config.AddRpcRetryDelayFlag(cacheMember.Cmd)
	config.AddRpcTimeoutFlag(cacheMember.Cmd)
	config.AddFsMdsAddrFlag(cacheMember.Cmd)
	config.AddCacheGroup(cacheMember.Cmd)
}

func (cacheMember *CacheMemberCommand) Init(cmd *cobra.Command, args []string) error {
	header := []string{cobrautil.ROW_ID, cobrautil.ROW_IP, cobrautil.ROW_PORT, cobrautil.ROW_WEIGHT, cobrautil.ROW_LASTONLINETIME, cobrautil.ROW_STATE}
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
	rpcInfo := base.NewRpc(addrs, timeout, retryTimes, retryDelay, verbose, "LoadMembers")

	groupName := config.GetFlagString(cmd, config.DINGOFS_CACHE_GROUP)
	rpc := &common.ListCacheMemberRpc{
		Info: rpcInfo,
		Request: &cachegroup.LoadMembersRequest{
			GroupName: &groupName,
		},
	}

	response, cmdErr := base.GetRpcResponse(rpc.Info, rpc)
	if cmdErr.TypeCode() != cmderror.CODE_SUCCESS {
		return cmdErr.ToError()
	}

	result := response.(*cachegroup.LoadMembersResponse)
	if result.GetStatus() != cachegroup.CacheGroupErrCode_CacheGroupOk {
		return fmt.Errorf("load members error: %s", result.GetStatus().String())
	}

	members := result.GetMembers()
	rows := make([]map[string]string, 0)
	for _, member := range members {
		row := make(map[string]string)
		row[cobrautil.ROW_ID] = fmt.Sprintf("%d", member.GetId())
		row[cobrautil.ROW_IP] = member.GetIp()
		row[cobrautil.ROW_PORT] = fmt.Sprintf("%d", member.GetPort())
		row[cobrautil.ROW_WEIGHT] = fmt.Sprintf("%d", member.GetWeight())
		ms := int64(member.GetLastOnlineTimeMs())
		sec := ms / 1000
		nsec := (ms % 1000) * 1000000
		t := time.Unix(sec, nsec)
		row[cobrautil.ROW_LASTONLINETIME] = t.Format("2006-01-02 15:04:05.000")
		row[cobrautil.ROW_STATE] = cobrautil.TranslateCacheGroupMemberState(member.GetState())

		rows = append(rows, row)
	}
	list := cobrautil.ListMap2ListSortByKeys(rows, cacheMember.Header, []string{cobrautil.ROW_ID})
	cacheMember.TableNew.AppendBulk(list)

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
