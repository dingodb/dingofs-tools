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

package delete

import (
	"fmt"

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	"github.com/dingodb/dingofs-tools/pkg/base"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/v2/common"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	pbmdsv2 "github.com/dingodb/dingofs-tools/proto/dingofs/proto/mdsv2"

	"github.com/spf13/cobra"
)

const (
	fsExample = `$ dingo delete fs --fsname dingofs`
)

type FsCommand struct {
	basecmd.FinalDingoCmd
	Rpc *common.DeleteFsRpc
}

var _ basecmd.FinalDingoCmdFunc = (*FsCommand)(nil) // check interface

func NewDeleteFsCommand() *cobra.Command {
	fsCmd := &FsCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:     "fs",
			Short:   "delete fs from dingofs",
			Example: fsExample,
		},
	}
	basecmd.NewFinalDingoCli(&fsCmd.FinalDingoCmd, fsCmd)
	return fsCmd.Cmd
}

func (fCmd *FsCommand) AddFlags() {
	config.AddRpcRetryTimesFlag(fCmd.Cmd)
	config.AddRpcRetryDelayFlag(fCmd.Cmd)
	config.AddRpcTimeoutFlag(fCmd.Cmd)
	config.AddFsMdsAddrFlag(fCmd.Cmd)
	config.AddFsNameRequiredFlag(fCmd.Cmd)
	config.AddNoConfirmOptionFlag(fCmd.Cmd)
}

func (fCmd *FsCommand) Init(cmd *cobra.Command, args []string) error {
	// new rpc
	mdsRpc, err := common.CreateNewMdsRpc(cmd, "DeleteFs")
	if err != nil {
		return err
	}
	// set request info
	fsName := config.GetFlagString(fCmd.Cmd, config.DINGOFS_FSNAME)
	fCmd.Rpc = &common.DeleteFsRpc{
		Info: mdsRpc,
		Request: &pbmdsv2.DeleteFsRequest{
			FsName: fsName,
		},
	}
	// set table header
	header := []string{cobrautil.ROW_FS_NAME, cobrautil.ROW_RESULT}
	fCmd.SetHeader(header)

	return nil
}

func (fCmd *FsCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&fCmd.FinalDingoCmd, fCmd)
}

func (fCmd *FsCommand) RunCommand(cmd *cobra.Command, args []string) error {
	fsName := fCmd.Rpc.Request.GetFsName()
	if !config.GetFlagBool(fCmd.Cmd, config.DINGOFS_NOCONFIRM) && !cobrautil.AskConfirmation(fmt.Sprintf("Are you sure to delete fs %s?", fsName), fsName) {
		return fmt.Errorf("abort delete fs")
	}
	// get rpc result
	response, errCmd := base.GetRpcResponse(fCmd.Rpc.Info, fCmd.Rpc)
	if errCmd.TypeCode() != cmderror.CODE_SUCCESS {
		return fmt.Errorf(errCmd.Message)
	}
	result := response.(*pbmdsv2.DeleteFsResponse)
	mdsErr := result.GetError()
	row := map[string]string{
		cobrautil.ROW_FS_NAME: fCmd.Rpc.Request.GetFsName(),
		cobrautil.ROW_RESULT:  cmderror.MDSV2Error(mdsErr).Message,
	}
	fCmd.TableNew.Append(cobrautil.Map2List(row, fCmd.Header))
	// to json
	res, errTranslate := output.MarshalProtoJson(result)
	if errTranslate != nil {
		return errTranslate
	}
	fCmd.Result = res
	fCmd.Error = cmderror.ErrSuccess()

	return nil
}

func (fCmd *FsCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&fCmd.FinalDingoCmd)
}
