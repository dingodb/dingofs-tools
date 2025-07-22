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

type GetQuotaCommand struct {
	basecmd.FinalDingoCmd
	Rpc     *common.GetDirQuotaRpc
	fsId    uint32
	path    string
	inodeId uint64
	epoch   uint64
}

var _ basecmd.FinalDingoCmdFunc = (*GetQuotaCommand)(nil) // check interface

func NewGetQuotaDataCommand() *GetQuotaCommand {
	getQuotaCmd := &GetQuotaCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:   "get",
			Short: "get directory quota by path",
			Example: `$ dingo quota get --fsid 1 --path /quotadir
$ dingo quota get --fsname dingofs --path /quotadir`,
		},
	}
	basecmd.NewFinalDingoCli(&getQuotaCmd.FinalDingoCmd, getQuotaCmd)
	return getQuotaCmd
}

func NewGetQuotaCommand() *cobra.Command {
	return NewGetQuotaDataCommand().Cmd
}

func (getQuotaCmd *GetQuotaCommand) AddFlags() {
	config.AddRpcRetryTimesFlag(getQuotaCmd.Cmd)
	config.AddRpcRetryDelayFlag(getQuotaCmd.Cmd)
	config.AddRpcTimeoutFlag(getQuotaCmd.Cmd)
	config.AddFsMdsAddrFlag(getQuotaCmd.Cmd)
	config.AddFsIdUint32OptionFlag(getQuotaCmd.Cmd)
	config.AddFsNameStringOptionFlag(getQuotaCmd.Cmd)
	config.AddFsPathRequiredFlag(getQuotaCmd.Cmd)
}

func (getQuotaCmd *GetQuotaCommand) Init(cmd *cobra.Command, args []string) error {
	// check flags values
	fsId, fsErr := common.GetFsId(cmd)
	if fsErr != nil {
		return fsErr
	}
	path := config.GetFlagString(cmd, config.DINGOFS_QUOTA_PATH)
	if len(path) == 0 {
		return fmt.Errorf("path is required")
	}

	// get epoch id
	epoch, epochErr := common.GetFsEpochByFsId(cmd, fsId)
	if epochErr != nil {
		return epochErr
	}
	// create router
	routerErr := common.InitFsMDSRouter(cmd, fsId)
	if routerErr != nil {
		return routerErr
	}

	//get inodeid
	dirInodeId, inodeErr := common.GetDirPathInodeId(cmd, fsId, path, epoch)
	if inodeErr != nil {
		return inodeErr
	}

	getQuotaCmd.fsId = fsId
	getQuotaCmd.path = path
	getQuotaCmd.inodeId = dirInodeId
	getQuotaCmd.epoch = epoch

	header := []string{cobrautil.ROW_INODE_ID, cobrautil.ROW_PATH, cobrautil.ROW_CAPACITY, cobrautil.ROW_USED, cobrautil.ROW_USED_PERCNET,
		cobrautil.ROW_INODES, cobrautil.ROW_INODES_IUSED, cobrautil.ROW_INODES_PERCENT}
	getQuotaCmd.SetHeader(header)

	return nil
}

func (getQuotaCmd *GetQuotaCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&getQuotaCmd.FinalDingoCmd, getQuotaCmd)
}

func (getQuotaCmd *GetQuotaCommand) RunCommand(cmd *cobra.Command, args []string) error {
	_, response, err := GetDirQuotaData(cmd, getQuotaCmd.fsId, getQuotaCmd.inodeId, getQuotaCmd.epoch)
	if err != nil {
		return err
	}
	dirQuota := response.GetQuota()
	//fill table
	quotaValueSlice := cmdCommon.ConvertQuotaToHumanizeValue(uint64(dirQuota.GetMaxBytes()), dirQuota.GetUsedBytes(), uint64(dirQuota.GetMaxInodes()), dirQuota.GetUsedInodes())
	row := map[string]string{
		cobrautil.ROW_INODE_ID:       fmt.Sprintf("%d", getQuotaCmd.inodeId),
		cobrautil.ROW_PATH:           getQuotaCmd.path,
		cobrautil.ROW_CAPACITY:       quotaValueSlice[0],
		cobrautil.ROW_USED:           quotaValueSlice[1],
		cobrautil.ROW_USED_PERCNET:   quotaValueSlice[2],
		cobrautil.ROW_INODES:         quotaValueSlice[3],
		cobrautil.ROW_INODES_IUSED:   quotaValueSlice[4],
		cobrautil.ROW_INODES_PERCENT: quotaValueSlice[5],
	}
	getQuotaCmd.TableNew.Append(cobrautil.Map2List(row, getQuotaCmd.Header))

	//to json
	res, errTranslate := output.MarshalProtoJson(response)
	if errTranslate != nil {
		return errTranslate
	}
	mapRes := res.(map[string]interface{})
	getQuotaCmd.Result = mapRes
	getQuotaCmd.Error = cmderror.ErrSuccess()

	return nil
}

func (getQuotaCmd *GetQuotaCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&getQuotaCmd.FinalDingoCmd)
}

func GetDirQuotaData(cmd *cobra.Command, fsId uint32, dirInodeId uint64, epoch uint64) (*pbmdsv2.GetDirQuotaRequest, *pbmdsv2.GetDirQuotaResponse, error) {
	endpoint := common.GetEndPoint(dirInodeId)
	mdsRpc := common.CreateNewMdsRpcWithEndPoint(cmd, endpoint, "GetDirQuota")
	// set request info
	getQuotaRpc := &common.GetDirQuotaRpc{
		Info: mdsRpc,
		Request: &pbmdsv2.GetDirQuotaRequest{
			Context: &pbmdsv2.Context{Epoch: epoch},
			FsId:    fsId,
			Ino:     dirInodeId,
		},
	}
	// get rpc result
	response, errCmd := base.GetRpcResponse(getQuotaRpc.Info, getQuotaRpc)
	if errCmd.TypeCode() != cmderror.CODE_SUCCESS {
		return nil, nil, fmt.Errorf(errCmd.Message)
	}
	result := response.(*pbmdsv2.GetDirQuotaResponse)

	if mdsErr := result.GetError(); mdsErr.GetErrcode() != pbmdsv2error.Errno_OK {
		if mdsErr.GetErrcode() == pbmdsv2error.Errno_ENOT_FOUND {
			return nil, nil, fmt.Errorf("no quota for directory, inodeid: %d", dirInodeId)
		} else {
			return nil, nil, cmderror.MDSV2Error(mdsErr).ToError()
		}
	}

	return getQuotaRpc.Request, result, nil
}
