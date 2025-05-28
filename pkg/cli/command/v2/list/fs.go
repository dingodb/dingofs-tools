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
	"slices"

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	"github.com/dingodb/dingofs-tools/pkg/base"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/v2/common"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	pbmdsv2error "github.com/dingodb/dingofs-tools/proto/dingofs/proto/error"
	pbmdsv2 "github.com/dingodb/dingofs-tools/proto/dingofs/proto/mdsv2"

	"github.com/spf13/cobra"
)

const (
	fsExample = `$ dingo list fs`
)

type FsCommand struct {
	basecmd.FinalDingoCmd
	Rpc *common.ListFsRpc
}

var _ basecmd.FinalDingoCmdFunc = (*FsCommand)(nil) // check interface

func NewFsCommand() *cobra.Command {
	return NewListFsCommand().Cmd
}

func NewListFsCommand() *FsCommand {
	fsCmd := &FsCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:     "fs",
			Short:   "list all fs info in the dingofs",
			Example: fsExample,
		},
	}

	basecmd.NewFinalDingoCli(&fsCmd.FinalDingoCmd, fsCmd)
	return fsCmd
}

func (fCmd *FsCommand) AddFlags() {
	config.AddRpcRetryTimesFlag(fCmd.Cmd)
	config.AddRpcTimeoutFlag(fCmd.Cmd)
	config.AddFsMdsAddrFlag(fCmd.Cmd)
}

func (fCmd *FsCommand) Init(cmd *cobra.Command, args []string) error {
	// new rpc
	mdsRpc, err := common.CreateNewMdsRpc(cmd, "ListFsInfo")
	if err != nil {
		return err
	}
	// set request info
	fCmd.Rpc = &common.ListFsRpc{Info: mdsRpc, Request: &pbmdsv2.ListFsInfoRequest{}}
	// set table header
	header := []string{cobrautil.ROW_FS_ID, cobrautil.ROW_FS_NAME, cobrautil.ROW_STATUS, cobrautil.ROW_BLOCKSIZE, cobrautil.ROW_CHUNK_SIZE, cobrautil.ROW_STORAGE_TYPE, cobrautil.ROW_PARTITION_TYPE, cobrautil.ROW_OWNER, cobrautil.ROW_MOUNT_NUM, cobrautil.ROW_UUID}
	fCmd.SetHeader(header)
	indexOwner := slices.Index(header, cobrautil.ROW_OWNER)
	indexType := slices.Index(header, cobrautil.ROW_STORAGE_TYPE)
	fCmd.TableNew.SetAutoMergeCellsByColumnIndex([]int{indexOwner, indexType})

	return nil
}

func (fCmd *FsCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&fCmd.FinalDingoCmd, fCmd)
}

func (fCmd *FsCommand) RunCommand(cmd *cobra.Command, args []string) error {
	// get rpc result
	response, errCmd := base.GetRpcResponse(fCmd.Rpc.Info, fCmd.Rpc)
	if errCmd.TypeCode() != cmderror.CODE_SUCCESS {
		return fmt.Errorf(errCmd.Message)
	}
	result := response.(*pbmdsv2.ListFsInfoResponse)
	if mdsErr := result.GetError(); mdsErr.GetErrcode() != pbmdsv2error.Errno_OK {
		return cmderror.MDSV2Error(mdsErr).ToError()
	}
	// fill table
	rows := make([]map[string]string, 0)
	for _, fsInfo := range result.GetFsInfos() {
		row := make(map[string]string)
		row[cobrautil.ROW_FS_ID] = fmt.Sprintf("%d", fsInfo.GetFsId())
		row[cobrautil.ROW_FS_NAME] = fsInfo.GetFsName()
		row[cobrautil.ROW_STATUS] = fsInfo.GetStatus().String()
		row[cobrautil.ROW_BLOCKSIZE] = fmt.Sprintf("%d", fsInfo.GetBlockSize())
		row[cobrautil.ROW_CHUNK_SIZE] = fmt.Sprintf("%d", fsInfo.GetChunkSize())
		row[cobrautil.ROW_STORAGE_TYPE] = fsInfo.GetFsType().String()
		row[cobrautil.ROW_PARTITION_TYPE] = common.ConvertPbPartitionTypeToString(fsInfo.GetPartitionPolicy().GetType())
		row[cobrautil.ROW_OWNER] = fsInfo.GetOwner()
		row[cobrautil.ROW_MOUNT_NUM] = fmt.Sprintf("%d", len(fsInfo.GetMountPoints()))
		row[cobrautil.ROW_UUID] = fsInfo.GetUuid()
		rows = append(rows, row)
	}
	list := cobrautil.ListMap2ListSortByKeys(rows, fCmd.Header, []string{cobrautil.ROW_OWNER, cobrautil.ROW_STORAGE_TYPE, cobrautil.ROW_ID})
	fCmd.TableNew.AppendBulk(list)
	// to json
	res, err := output.MarshalProtoJson(result)
	if err != nil {
		return err
	}
	mapRes := res.(map[string]interface{})
	fCmd.Result = mapRes
	fCmd.Error = cmderror.ErrSuccess()

	return nil
}

func (fCmd *FsCommand) ResultPlainOutput() error {
	if fCmd.TableNew.NumLines() == 0 {
		fmt.Println("no fs in cluster")
	}
	return output.FinalCmdOutputPlain(&fCmd.FinalDingoCmd)
}
