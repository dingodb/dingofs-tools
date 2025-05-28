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

package config

import (
	"fmt"

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	"github.com/dingodb/dingofs-tools/pkg/base"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	cmdcommon "github.com/dingodb/dingofs-tools/pkg/cli/command/v1/common"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/v2/common"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	pbmdsv2error "github.com/dingodb/dingofs-tools/proto/dingofs/proto/error"
	pbmdsv2 "github.com/dingodb/dingofs-tools/proto/dingofs/proto/mdsv2"
	"github.com/spf13/cobra"
)

type GetFsQuotaCommand struct {
	basecmd.FinalDingoCmd
	Rpc *common.GetFsQuotaRpc
}

var _ basecmd.FinalDingoCmdFunc = (*GetFsQuotaCommand)(nil) // check interface

func NewGetFsQuotaCommand() *cobra.Command {
	fsQuotaCmd := &GetFsQuotaCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:   "get",
			Short: "get fs quota for dingofs",
			Example: `$ dingo config get --fsid 1 
$ dingo config get --fsname dingofs
`,
		},
	}
	basecmd.NewFinalDingoCli(&fsQuotaCmd.FinalDingoCmd, fsQuotaCmd)
	return fsQuotaCmd.Cmd
}

func (fsQuotaCmd *GetFsQuotaCommand) AddFlags() {
	config.AddRpcRetryTimesFlag(fsQuotaCmd.Cmd)
	config.AddRpcTimeoutFlag(fsQuotaCmd.Cmd)
	config.AddFsMdsAddrFlag(fsQuotaCmd.Cmd)
	config.AddFsIdUint32OptionFlag(fsQuotaCmd.Cmd)
	config.AddFsNameStringOptionFlag(fsQuotaCmd.Cmd)
}

func (fsQuotaCmd *GetFsQuotaCommand) Init(cmd *cobra.Command, args []string) error {
	return nil
}

func (fsQuotaCmd *GetFsQuotaCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&fsQuotaCmd.FinalDingoCmd, fsQuotaCmd)
}

func (fsQuotaCmd *GetFsQuotaCommand) RunCommand(cmd *cobra.Command, args []string) error {
	// get fs id
	fsId, fsErr := common.GetFsId(cmd)
	if fsErr != nil {
		return fsErr
	}
	// get filesystem name
	fsName, fsErr := common.GetFsName(cmd)
	if fsErr != nil {
		return fsErr
	}
	// get quota
	_, result, getErr := GetFsQuotaData(cmd, fsId)
	if getErr != nil {
		return getErr
	}

	mdsErr := result.GetError()
	row := make(map[string]string, 0)
	if mdsErr.GetErrcode() == pbmdsv2error.Errno_OK {
		// set table header
		header := []string{cobrautil.ROW_FS_ID, cobrautil.ROW_FS_NAME, cobrautil.ROW_CAPACITY, cobrautil.ROW_USED, cobrautil.ROW_USED_PERCNET,
			cobrautil.ROW_INODES, cobrautil.ROW_INODES_IUSED, cobrautil.ROW_INODES_PERCENT}
		fsQuotaCmd.SetHeader(header)

		fsQuota := result.GetQuota()
		quotaValueSlice := cmdcommon.ConvertQuotaToHumanizeValue(fsQuota.GetMaxBytes(), fsQuota.GetUsedBytes(), fsQuota.GetMaxInodes(), fsQuota.GetUsedInodes())
		// fill table
		row = map[string]string{
			cobrautil.ROW_FS_ID:          fmt.Sprintf("%d", fsId),
			cobrautil.ROW_FS_NAME:        fsName,
			cobrautil.ROW_CAPACITY:       quotaValueSlice[0],
			cobrautil.ROW_USED:           quotaValueSlice[1],
			cobrautil.ROW_USED_PERCNET:   quotaValueSlice[2],
			cobrautil.ROW_INODES:         quotaValueSlice[3],
			cobrautil.ROW_INODES_IUSED:   quotaValueSlice[4],
			cobrautil.ROW_INODES_PERCENT: quotaValueSlice[5],
		}

	} else {
		header := []string{cobrautil.ROW_FS_NAME, cobrautil.ROW_RESULT}
		fsQuotaCmd.SetHeader(header)
		row = map[string]string{
			cobrautil.ROW_FS_NAME: fsName,
			cobrautil.ROW_RESULT:  cmderror.MDSV2Error(mdsErr).Message,
		}

	}
	fsQuotaCmd.TableNew.Append(cobrautil.Map2List(row, fsQuotaCmd.Header))

	// to json
	res, errTranslate := output.MarshalProtoJson(result)
	if errTranslate != nil {
		return errTranslate
	}
	mapRes := res.(map[string]interface{})
	fsQuotaCmd.Result = mapRes
	fsQuotaCmd.Error = cmderror.ErrSuccess()

	return nil
}

func (fsQuotaCmd *GetFsQuotaCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&fsQuotaCmd.FinalDingoCmd)
}

func GetFsQuotaData(cmd *cobra.Command, fsId uint32) (*pbmdsv2.GetFsQuotaRequest, *pbmdsv2.GetFsQuotaResponse, error) {
	// new prc
	mdsRpc, err := common.CreateNewMdsRpc(cmd, "getFsQuota")
	if err != nil {
		return nil, nil, err
	}
	// set request info
	request := &pbmdsv2.GetFsQuotaRequest{
		FsId: fsId,
	}
	requestRpc := &common.GetFsQuotaRpc{
		Info:    mdsRpc,
		Request: request,
	}
	// get rpc result
	response, errCmd := base.GetRpcResponse(requestRpc.Info, requestRpc)
	if errCmd.TypeCode() != cmderror.CODE_SUCCESS {
		return nil, nil, fmt.Errorf(errCmd.Message)
	}
	result := response.(*pbmdsv2.GetFsQuotaResponse)

	return request, result, nil
}
