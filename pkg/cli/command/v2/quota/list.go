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

package quota

import (
	"fmt"
	"syscall"

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	"github.com/dingodb/dingofs-tools/pkg/base"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	cmdCommon "github.com/dingodb/dingofs-tools/pkg/cli/command/v1/common"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/v2/common"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	pbmdsv2error "github.com/dingodb/dingofs-tools/proto/dingofs/proto/error"
	pbmdsv2 "github.com/dingodb/dingofs-tools/proto/dingofs/proto/mdsv2"
	"github.com/spf13/cobra"
)

type ListQuotaCommand struct {
	basecmd.FinalDingoCmd
	Rpc  *common.ListDirQuotaRpc
	fsId uint32
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
	// new prc
	mdsRpc, err := common.CreateNewMdsRpc(cmd, "LoadDirQuotas")
	if err != nil {
		return err
	}
	// check flags values
	fsId, fsErr := common.GetFsId(cmd)
	if fsErr != nil {
		return fsErr
	}
	listQuotaCmd.fsId = fsId
	// set request info
	listQuotaCmd.Rpc = &common.ListDirQuotaRpc{
		Info: mdsRpc,
		Request: &pbmdsv2.LoadDirQuotasRequest{
			FsId: fsId},
	}

	header := []string{cobrautil.ROW_INODE_ID, cobrautil.ROW_PATH, cobrautil.ROW_CAPACITY, cobrautil.ROW_USED, cobrautil.ROW_USED_PERCNET,
		cobrautil.ROW_INODES, cobrautil.ROW_INODES_IUSED, cobrautil.ROW_INODES_PERCENT}
	listQuotaCmd.SetHeader(header)

	return nil
}

func (listQuotaCmd *ListQuotaCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&listQuotaCmd.FinalDingoCmd, listQuotaCmd)
}

func (listQuotaCmd *ListQuotaCommand) RunCommand(cmd *cobra.Command, args []string) error {
	// get rpc result
	response, errCmd := base.GetRpcResponse(listQuotaCmd.Rpc.Info, listQuotaCmd.Rpc)
	if errCmd.TypeCode() != cmderror.CODE_SUCCESS {
		return fmt.Errorf(errCmd.Message)
	}
	result := response.(*pbmdsv2.LoadDirQuotasResponse)
	if mdsErr := result.GetError(); mdsErr.GetErrcode() != pbmdsv2error.Errno_OK {
		return cmderror.MDSV2Error(mdsErr).ToError()
	}

	dirQuotas := result.GetQuotas()
	if len(dirQuotas) == 0 {
		fmt.Println("no directory quota in filesystem")
		return nil
	}
	//fill tables
	rows := make([]map[string]string, 0)
	for dirInode, quota := range dirQuotas {
		row := make(map[string]string)
		quotaValueSlice := cmdCommon.ConvertQuotaToHumanizeValue(quota.GetMaxBytes(), quota.GetUsedBytes(), quota.GetMaxInodes(), quota.GetUsedInodes())
		dirPath, _, dirErr := common.GetInodePath(listQuotaCmd.Cmd, listQuotaCmd.fsId, dirInode)
		if dirErr == syscall.ENOENT {
			continue
		}
		if dirErr != nil {
			return dirErr
		}
		if dirPath == "" { // directory may be deleted,not show
			continue
		}
		row[cobrautil.ROW_INODE_ID] = fmt.Sprintf("%d", dirInode)
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

	res, errTranslate := output.MarshalProtoJson(result)
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
