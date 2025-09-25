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
	"github.com/dingodb/dingofs-tools/pkg/rpc"

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	pbmdsv2 "github.com/dingodb/dingofs-tools/proto/dingofs/proto/mdsv2"

	"github.com/dingodb/dingofs-tools/pkg/base"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	"github.com/spf13/cobra"
)

const (
	ListGroupExample = `$ dingo list cachegroup`
)

type CacheGroupCommand struct {
	basecmd.FinalDingoCmd
	Rpc *rpc.ListCacheGroupRpc
}

var _ basecmd.FinalDingoCmdFunc = (*CacheGroupCommand)(nil) // check interface

func NewCacheGroupCommand() *cobra.Command {
	return NewListCacheGroupCommand().Cmd
}

func NewListCacheGroupCommand() *CacheGroupCommand {
	cacheGroupCmd := &CacheGroupCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:     "cachegroup",
			Short:   "list remote cache groups",
			Example: ListGroupExample,
		},
	}

	basecmd.NewFinalDingoCli(&cacheGroupCmd.FinalDingoCmd, cacheGroupCmd)
	return cacheGroupCmd
}

func (cacheGroup *CacheGroupCommand) AddFlags() {
	config.AddRpcRetryTimesFlag(cacheGroup.Cmd)
	config.AddRpcRetryDelayFlag(cacheGroup.Cmd)
	config.AddRpcTimeoutFlag(cacheGroup.Cmd)
	config.AddFsMdsAddrFlag(cacheGroup.Cmd)
}

func (cacheGroup *CacheGroupCommand) Init(cmd *cobra.Command, args []string) error {
	header := []string{cobrautil.ROW_GROUP}
	cacheGroup.SetHeader(header)

	return nil
}

func (cacheGroup *CacheGroupCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&cacheGroup.FinalDingoCmd, cacheGroup)
}

func (cacheGroup *CacheGroupCommand) RunCommand(cmd *cobra.Command, args []string) error {
	// new rpc
	mdsRpc, err := rpc.CreateNewMdsRpc(cmd, "ListGroups")
	if err != nil {
		return err
	}
	// set request info
	cacheGroup.Rpc = &rpc.ListCacheGroupRpc{Info: mdsRpc, Request: &pbmdsv2.ListGroupsRequest{}}

	response, cmdErr := base.GetRpcResponse(cacheGroup.Rpc.Info, cacheGroup.Rpc)
	if cmdErr.TypeCode() != cmderror.CODE_SUCCESS {
		return cmdErr.ToError()
	}

	result := response.(*pbmdsv2.ListGroupsResponse)
	groups := result.GetGroupNames()
	if len(groups) == 0 {
		return fmt.Errorf("no cachegroup found")
	}

	rows := make([]map[string]string, 0)
	for _, group := range groups {
		row := make(map[string]string)
		row[cobrautil.ROW_GROUP] = group

		rows = append(rows, row)
	}
	list := cobrautil.ListMap2ListSortByKeys(rows, cacheGroup.Header, []string{cobrautil.ROW_GROUP})
	cacheGroup.TableNew.AppendBulk(list)

	// to json
	res, err := output.MarshalProtoJson(result)
	if err != nil {
		return err
	}
	mapRes := res.(map[string]interface{})
	cacheGroup.Result = mapRes
	cacheGroup.Error = cmderror.ErrSuccess()

	return nil
}

func (cacheGroup *CacheGroupCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&cacheGroup.FinalDingoCmd)
}
