// Copyright (c) 2024 dingodb.com, Inc. All Rights Reserved
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

package quota

import (
	"fmt"
	"syscall"

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	"github.com/dingodb/dingofs-tools/pkg/base"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	cmdCommon "github.com/dingodb/dingofs-tools/pkg/cli/command/v1/common"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	"github.com/dingodb/dingofs-tools/proto/dingofs/proto/metaserver"
	"github.com/spf13/cobra"
)

type ListQuotaCommand struct {
	basecmd.FinalDingoCmd
	Rpc *cmdCommon.ListQuotaRpc
}

var _ basecmd.FinalDingoCmdFunc = (*ListQuotaCommand)(nil) // check interface

func NewListQuotaCommand() *cobra.Command {
	listQuotaCmd := &ListQuotaCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:   "list",
			Short: "list all directory quotas of fileSystem by fsid",
			Example: `$ dingo quota list --fsid 1
$ dingo quota list --fsname dingofs`,
		},
	}
	basecmd.NewFinalDingoCli(&listQuotaCmd.FinalDingoCmd, listQuotaCmd)
	return listQuotaCmd.Cmd
}

func (listQuotaCmd *ListQuotaCommand) AddFlags() {
	config.AddRpcRetryTimesFlag(listQuotaCmd.Cmd)
	config.AddRpcTimeoutFlag(listQuotaCmd.Cmd)
	config.AddFsMdsAddrFlag(listQuotaCmd.Cmd)
	config.AddFsIdUint32OptionFlag(listQuotaCmd.Cmd)
	config.AddFsNameStringOptionFlag(listQuotaCmd.Cmd)
}

func (listQuotaCmd *ListQuotaCommand) Init(cmd *cobra.Command, args []string) error {
	_, getAddrErr := config.GetFsMdsAddrSlice(listQuotaCmd.Cmd)
	if getAddrErr.TypeCode() != cmderror.CODE_SUCCESS {
		listQuotaCmd.Error = getAddrErr
		return fmt.Errorf(getAddrErr.Message)
	}
	//check flags values
	fsId, fsErr := cmdCommon.GetFsId(cmd)
	if fsErr != nil {
		return fsErr
	}
	// get poolid copysetid
	partitionInfo, partErr := cmdCommon.GetPartitionInfo(listQuotaCmd.Cmd, fsId, config.ROOTINODEID)
	if partErr != nil {
		return partErr
	}
	poolId := partitionInfo.GetPoolId()
	copyetId := partitionInfo.GetCopysetId()
	request := &metaserver.LoadDirQuotasRequest{
		PoolId:    &poolId,
		CopysetId: &copyetId,
		FsId:      &fsId,
	}
	listQuotaCmd.Rpc = &cmdCommon.ListQuotaRpc{
		Request: request,
	}
	//get request addr leader
	addrs, addrErr := cmdCommon.GetLeaderPeerAddr(listQuotaCmd.Cmd, fsId, config.ROOTINODEID)
	if addrErr != nil {
		return addrErr
	}
	timeout := config.GetRpcTimeout(cmd)
	retrytimes := config.GetRpcRetryTimes(cmd)
	listQuotaCmd.Rpc.Info = base.NewRpc(addrs, timeout, retrytimes, "LoadDirQuotas")
	listQuotaCmd.Rpc.Info.RpcDataShow = config.GetFlagBool(listQuotaCmd.Cmd, config.VERBOSE)

	header := []string{cobrautil.ROW_ID, cobrautil.ROW_PATH, cobrautil.ROW_CAPACITY, cobrautil.ROW_USED, cobrautil.ROW_USED_PERCNET,
		cobrautil.ROW_INODES, cobrautil.ROW_INODES_IUSED, cobrautil.ROW_INODES_PERCENT}
	listQuotaCmd.SetHeader(header)

	return nil
}

func (listQuotaCmd *ListQuotaCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&listQuotaCmd.FinalDingoCmd, listQuotaCmd)
}

func (listQuotaCmd *ListQuotaCommand) RunCommand(cmd *cobra.Command, args []string) error {
	result, err := base.GetRpcResponse(listQuotaCmd.Rpc.Info, listQuotaCmd.Rpc)
	if err.TypeCode() != cmderror.CODE_SUCCESS {
		return err.ToError()
	}
	response := result.(*metaserver.LoadDirQuotasResponse)
	if statusCode := response.GetStatusCode(); statusCode != metaserver.MetaStatusCode_OK {
		return cmderror.ErrQuota(int(statusCode)).ToError()
	}
	//fill tables
	rows := make([]map[string]string, 0)
	for dirInode, quota := range response.GetQuotas() {
		row := make(map[string]string)
		quotaValueSlice := cmdCommon.ConvertQuotaToHumanizeValue(quota.GetMaxBytes(), quota.GetUsedBytes(), quota.GetMaxInodes(), quota.GetUsedInodes())
		dirPath, _, dirErr := cmdCommon.GetInodePath(listQuotaCmd.Cmd, listQuotaCmd.Rpc.Request.GetFsId(), dirInode)
		if dirErr == syscall.ENOENT {
			continue
		}
		if dirErr != nil {
			return dirErr
		}
		if dirPath == "" { // directory may be deleted,not show
			continue
		}
		row[cobrautil.ROW_ID] = fmt.Sprintf("%d", dirInode)
		row[cobrautil.ROW_PATH] = dirPath
		row[cobrautil.ROW_CAPACITY] = quotaValueSlice[0]
		row[cobrautil.ROW_USED] = quotaValueSlice[1]
		row[cobrautil.ROW_USED_PERCNET] = quotaValueSlice[2]
		row[cobrautil.ROW_INODES] = quotaValueSlice[3]
		row[cobrautil.ROW_INODES_IUSED] = quotaValueSlice[4]
		row[cobrautil.ROW_INODES_PERCENT] = quotaValueSlice[5]
		rows = append(rows, row)
	}
	list := cobrautil.ListMap2ListSortByKeys(rows, listQuotaCmd.Header, []string{cobrautil.ROW_PATH})
	listQuotaCmd.TableNew.AppendBulk(list)

	res, errTranslate := output.MarshalProtoJson(response)
	if errTranslate != nil {
		return errTranslate
	}
	mapRes := res.(map[string]interface{})
	listQuotaCmd.Result = mapRes
	listQuotaCmd.Error = cmderror.ErrSuccess()

	return nil
}

func (listQuotaCmd *ListQuotaCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&listQuotaCmd.FinalDingoCmd)
}
