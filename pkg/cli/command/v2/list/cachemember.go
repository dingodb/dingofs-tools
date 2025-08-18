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

package list

import (
	"fmt"
	"time"

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	rpc "github.com/dingodb/dingofs-tools/pkg/rpc/v2"
	pbmdsv2 "github.com/dingodb/dingofs-tools/proto/dingofs/proto/mdsv2"

	"github.com/dingodb/dingofs-tools/pkg/base"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/v2/common"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	"github.com/spf13/cobra"
)

const (
	ListMemberExample = `$ dingo list cachemember
$ dingo list cachemember --group group1`
)

type CacheMemberCommand struct {
	basecmd.FinalDingoCmd
	Rpc *rpc.ListCacheMemberRpc
}

var _ basecmd.FinalDingoCmdFunc = (*CacheMemberCommand)(nil) // check interface

func NewCacheMemberCommand() *cobra.Command {
	cacheMemberCmd := &CacheMemberCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:     "cachemember",
			Short:   "list all cachemembers",
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
	header := []string{cobrautil.ROW_MEMBERID, cobrautil.ROW_IP, cobrautil.ROW_PORT, cobrautil.ROW_WEIGHT, cobrautil.ROW_LOCKED, cobrautil.ROW_CREATE_TIME, cobrautil.ROW_LASTONLINETIME, cobrautil.ROW_STATE, cobrautil.ROW_GROUP}
	cacheMember.SetHeader(header)

	return nil
}

func (cacheMember *CacheMemberCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&cacheMember.FinalDingoCmd, cacheMember)
}

func (cacheMember *CacheMemberCommand) RunCommand(cmd *cobra.Command, args []string) error {
	// new rpc
	mdsRpc, err := common.CreateNewMdsRpc(cmd, "ListMembers")
	if err != nil {
		return err
	}

	request := pbmdsv2.ListMembersRequest{}
	if cmd.Flag(config.DINGOFS_CACHE_GROUP).Changed {
		groupName := config.GetFlagString(cmd, config.DINGOFS_CACHE_GROUP)
		request.GroupName = &groupName
	}

	cacheMember.Rpc = &rpc.ListCacheMemberRpc{
		Info:    mdsRpc,
		Request: &request,
	}
	// set request info
	response, cmdErr := base.GetRpcResponse(cacheMember.Rpc.Info, cacheMember.Rpc)
	if cmdErr.TypeCode() != cmderror.CODE_SUCCESS {
		return cmdErr.ToError()
	}

	result := response.(*pbmdsv2.ListMembersResponse)
	members := result.GetMembers()
	if len(members) == 0 {
		return fmt.Errorf("no cachemember found")
	}

	rows := make([]map[string]string, 0)
	for _, member := range members {
		row := make(map[string]string)
		row[cobrautil.ROW_MEMBERID] = member.GetMemberId()
		ip := member.GetIp()
		port := member.GetPort()
		if len(ip) > 0 && port > 0 {
			row[cobrautil.ROW_IP] = member.GetIp()
			row[cobrautil.ROW_PORT] = fmt.Sprintf("%d", member.GetPort())
			row[cobrautil.ROW_WEIGHT] = fmt.Sprintf("%d", member.GetWeight())
			row[cobrautil.ROW_STATE] = cobrautil.TranslateCacheGroupMemberState2(member.GetState())
			row[cobrautil.ROW_GROUP] = member.GetGroupName()
			row[cobrautil.ROW_LOCKED] = fmt.Sprintf("%v", member.GetLocked())

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
