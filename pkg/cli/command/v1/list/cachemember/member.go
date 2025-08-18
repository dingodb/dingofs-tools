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
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	rpc "github.com/dingodb/dingofs-tools/pkg/rpc/v1"
	pbCacheGroup "github.com/dingodb/dingofs-tools/proto/dingofs/proto/cachegroup"
	"github.com/spf13/cobra"
)

const (
	ListMemberExample = `$ dingo list cachemember
$ dingo list cachemember --group group1`
)

type CacheMemberCommand struct {
	basecmd.FinalDingoCmd
	Rpc      *rpc.ListCacheMemberRpc
	response *pbCacheGroup.ListMembersResponse
}

var _ basecmd.FinalDingoCmdFunc = (*CacheMemberCommand)(nil) // check interface

func NewCacheMemberCommand() *cobra.Command {
	cacheMemberCmd := &CacheMemberCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:     "cachemember",
			Short:   "list cache members in cachegroup",
			Example: ListMemberExample,
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
	config.AddCacheGroupOptionalFlag(cacheMember.Cmd)
}

func (cacheMember *CacheMemberCommand) Init(cmd *cobra.Command, args []string) error {
	header := []string{cobrautil.ROW_ID, cobrautil.ROW_IP, cobrautil.ROW_PORT, cobrautil.ROW_WEIGHT, cobrautil.ROW_CREATE_TIME, cobrautil.ROW_LASTONLINETIME, cobrautil.ROW_STATE, cobrautil.ROW_GROUP}
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
	rpcInfo := base.NewRpc(addrs, timeout, retryTimes, retryDelay, verbose, "ListMembers")

	groupName := config.GetFlagString(cmd, config.DINGOFS_CACHE_GROUP)
	request := pbCacheGroup.ListMembersRequest{}
	if len(groupName) > 0 {
		request.GroupName = &groupName
	}
	rpc := &rpc.ListCacheMemberRpc{
		Info:    rpcInfo,
		Request: &request,
	}

	response, cmdErr := base.GetRpcResponse(rpc.Info, rpc)
	if cmdErr.TypeCode() != cmderror.CODE_SUCCESS {
		return cmdErr.ToError()
	}

	result := response.(*pbCacheGroup.ListMembersResponse)
	if result.GetStatus() != pbCacheGroup.CacheGroupErrCode_CacheGroupOk {
		return fmt.Errorf("load members error: %s", result.GetStatus().String())
	}

	members := result.GetMembers()
	rows := make([]map[string]string, 0)
	for _, member := range members {
		row := make(map[string]string)
		row[cobrautil.ROW_ID] = member.GetId()
		ip := member.GetIp()
		port := member.GetPort()
		if len(ip) > 0 && port > 0 {
			row[cobrautil.ROW_IP] = member.GetIp()
			row[cobrautil.ROW_PORT] = fmt.Sprintf("%d", member.GetPort())
			row[cobrautil.ROW_WEIGHT] = fmt.Sprintf("%d", member.GetWeight())
			row[cobrautil.ROW_STATE] = cobrautil.TranslateCacheGroupMemberState(member.GetState())
			row[cobrautil.ROW_GROUP] = member.GetGroupName()

			// process create time
			seconds := int64(member.GetCreateTimeS())
			if seconds > 0 {
				createTime := time.Unix(seconds, 0)
				row[cobrautil.ROW_CREATE_TIME] = createTime.Format("2006-01-02 15:04:05.000")
			}
			// process online time
			ms := int64(member.GetLastOnlineTimeMs())
			if ms > 0 {
				sec := ms / 1000
				nsec := (ms % 1000) * 1000000
				onlineTime := time.Unix(sec, nsec)
				row[cobrautil.ROW_LASTONLINETIME] = onlineTime.Format("2006-01-02 15:04:05.000")
			}
		}

		rows = append(rows, row)
	}
	list := cobrautil.ListMap2ListSortByKeys(rows, cacheMember.Header, []string{cobrautil.ROW_GROUP, cobrautil.ROW_ID})
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
