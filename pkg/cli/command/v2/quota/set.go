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

	"github.com/dingodb/dingofs-tools/pkg/base"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/v2/common"
	pbmdsv2 "github.com/dingodb/dingofs-tools/proto/dingofs/proto/mdsv2"

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	cmdCommon "github.com/dingodb/dingofs-tools/pkg/cli/command/v1/common"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	"github.com/spf13/cobra"
)

type SetQuotaCommand struct {
	basecmd.FinalDingoCmd
	Rpc *common.SetDirQuotaRpc
}

var _ basecmd.FinalDingoCmdFunc = (*SetQuotaCommand)(nil) // check interface

func NewSetQuotaCommand() *cobra.Command {
	setQuotaCmd := &SetQuotaCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:     "set",
			Short:   "set quota to a directory",
			Example: `$ dingo quota set --fsid 1 --path /quotadir --capacity 10 --inodes 100000`,
		},
	}
	basecmd.NewFinalDingoCli(&setQuotaCmd.FinalDingoCmd, setQuotaCmd)
	return setQuotaCmd.Cmd
}

func (setQuotaCmd *SetQuotaCommand) AddFlags() {
	config.AddRpcRetryTimesFlag(setQuotaCmd.Cmd)
	config.AddRpcRetryDelayFlag(setQuotaCmd.Cmd)
	config.AddRpcTimeoutFlag(setQuotaCmd.Cmd)
	config.AddFsMdsAddrFlag(setQuotaCmd.Cmd)
	config.AddFsIdUint32OptionFlag(setQuotaCmd.Cmd)
	config.AddFsNameStringOptionFlag(setQuotaCmd.Cmd)
	config.AddFsPathRequiredFlag(setQuotaCmd.Cmd)
	config.AddFsCapacityOptionalFlag(setQuotaCmd.Cmd)
	config.AddFsInodesOptionalFlag(setQuotaCmd.Cmd)
}

func (setQuotaCmd *SetQuotaCommand) Init(cmd *cobra.Command, args []string) error {
	// new prc
	mdsRpc, err := common.CreateNewMdsRpc(cmd, "SetDirQuota")
	if err != nil {
		return err
	}
	// check flags values
	capacity, inodes, quotaErr := cmdCommon.CheckAndGetQuotaValue(setQuotaCmd.Cmd)
	if quotaErr != nil {
		return quotaErr
	}
	fsId, fsErr := common.GetFsId(cmd)
	if fsErr != nil {
		return fsErr
	}
	path := config.GetFlagString(setQuotaCmd.Cmd, config.DINGOFS_QUOTA_PATH)
	if len(path) == 0 {
		return fmt.Errorf("path is required")
	}
	//get inodeid
	dirInodeId, inodeErr := common.GetDirPathInodeId(setQuotaCmd.Cmd, fsId, path)
	if inodeErr != nil {
		return inodeErr
	}
	// set request info
	request := &pbmdsv2.SetDirQuotaRequest{
		FsId:  fsId,
		Ino:   dirInodeId,
		Quota: &pbmdsv2.Quota{MaxBytes: capacity, MaxInodes: inodes},
	}
	setQuotaCmd.Rpc = &common.SetDirQuotaRpc{
		Info:    mdsRpc,
		Request: request,
	}

	// set table header
	header := []string{cobrautil.ROW_RESULT}
	setQuotaCmd.SetHeader(header)

	return nil
}

func (setQuotaCmd *SetQuotaCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&setQuotaCmd.FinalDingoCmd, setQuotaCmd)
}

func (setQuotaCmd *SetQuotaCommand) RunCommand(cmd *cobra.Command, args []string) error {
	// get rpc result
	response, errCmd := base.GetRpcResponse(setQuotaCmd.Rpc.Info, setQuotaCmd.Rpc)
	if errCmd.TypeCode() != cmderror.CODE_SUCCESS {
		return fmt.Errorf(errCmd.Message)
	}
	result := response.(*pbmdsv2.SetDirQuotaResponse)
	mdsErr := result.GetError()
	row := map[string]string{
		cobrautil.ROW_RESULT: cmderror.MDSV2Error(mdsErr).Message,
	}
	setQuotaCmd.TableNew.Append(cobrautil.Map2List(row, setQuotaCmd.Header))
	// to json
	res, errTranslate := output.MarshalProtoJson(result)
	if errTranslate != nil {
		return errTranslate
	}
	mapRes := res.(map[string]interface{})
	setQuotaCmd.Result = mapRes
	setQuotaCmd.Error = cmderror.ErrSuccess()

	return nil
}

func (setQuotaCmd *SetQuotaCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&setQuotaCmd.FinalDingoCmd)
}
