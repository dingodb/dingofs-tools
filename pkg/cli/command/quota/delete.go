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
	"github.com/dingodb/dingofs-tools/pkg/rpc"

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	"github.com/dingodb/dingofs-tools/pkg/base"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	pbmds "github.com/dingodb/dingofs-tools/proto/dingofs/proto/mds"
	"github.com/spf13/cobra"
)

type DeleteQuotaCommand struct {
	basecmd.FinalDingoCmd
	Rpc *rpc.DeleteDirQuotaRpc
}

var _ basecmd.FinalDingoCmdFunc = (*DeleteQuotaCommand)(nil) // check interface

func NewDeleteQuotaCommand() *cobra.Command {
	deleteQuotaCmd := &DeleteQuotaCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:   "delete",
			Short: "delete directory quota",
			Example: `$ dingo quota delete --fsid 1 --path /quotadir
$ dingo quota delete --fsname dingofs --path /quotadir`,
		},
	}
	basecmd.NewFinalDingoCli(&deleteQuotaCmd.FinalDingoCmd, deleteQuotaCmd)
	return deleteQuotaCmd.Cmd
}

func (deleteQuotaCmd *DeleteQuotaCommand) AddFlags() {
	config.AddRpcRetryTimesFlag(deleteQuotaCmd.Cmd)
	config.AddRpcRetryDelayFlag(deleteQuotaCmd.Cmd)
	config.AddRpcTimeoutFlag(deleteQuotaCmd.Cmd)
	config.AddFsMdsAddrFlag(deleteQuotaCmd.Cmd)
	config.AddFsIdUint32OptionFlag(deleteQuotaCmd.Cmd)
	config.AddFsNameStringOptionFlag(deleteQuotaCmd.Cmd)
	config.AddFsPathRequiredFlag(deleteQuotaCmd.Cmd)
}

func (deleteQuotaCmd *DeleteQuotaCommand) Init(cmd *cobra.Command, args []string) error {
	// check flags values
	fsId, fsErr := rpc.GetFsId(cmd)
	if fsErr != nil {
		return fsErr
	}
	path := config.GetFlagString(deleteQuotaCmd.Cmd, config.DINGOFS_QUOTA_PATH)
	if len(path) == 0 {
		return fmt.Errorf("path is required")
	}

	// get epoch id
	epoch, epochErr := rpc.GetFsEpochByFsId(cmd, fsId)
	if epochErr != nil {
		return epochErr
	}
	// create router
	routerErr := rpc.InitFsMDSRouter(cmd, fsId)
	if routerErr != nil {
		return routerErr
	}

	//get inodeid
	dirInodeId, inodeErr := rpc.GetDirPathInodeId(deleteQuotaCmd.Cmd, fsId, path, epoch)
	if inodeErr != nil {
		return inodeErr
	}

	// set request info
	endpoint := rpc.GetEndPoint(dirInodeId)
	mdsRpc := rpc.CreateNewMdsRpcWithEndPoint(cmd, endpoint, "DeleteDirQuota")
	deleteQuotaCmd.Rpc = &rpc.DeleteDirQuotaRpc{
		Info: mdsRpc,
		Request: &pbmds.DeleteDirQuotaRequest{
			Context: &pbmds.Context{Epoch: epoch},
			FsId:    fsId,
			Ino:     dirInodeId,
		},
	}

	// set table header
	header := []string{cobrautil.ROW_RESULT}
	deleteQuotaCmd.SetHeader(header)

	return nil
}

func (deleteQuotaCmd *DeleteQuotaCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&deleteQuotaCmd.FinalDingoCmd, deleteQuotaCmd)
}

func (deleteQuotaCmd *DeleteQuotaCommand) RunCommand(cmd *cobra.Command, args []string) error {
	// get rpc result
	response, errCmd := base.GetRpcResponse(deleteQuotaCmd.Rpc.Info, deleteQuotaCmd.Rpc)
	if errCmd.TypeCode() != cmderror.CODE_SUCCESS {
		return fmt.Errorf(errCmd.Message)
	}
	result := response.(*pbmds.DeleteDirQuotaResponse)
	mdsErr := result.GetError()
	row := map[string]string{
		cobrautil.ROW_RESULT: cmderror.MDSV2Error(mdsErr).Message,
	}
	deleteQuotaCmd.TableNew.Append(cobrautil.Map2List(row, deleteQuotaCmd.Header))
	// to json
	res, errTranslate := output.MarshalProtoJson(result)
	if errTranslate != nil {
		return errTranslate
	}
	mapRes := res.(map[string]interface{})
	deleteQuotaCmd.Result = mapRes
	deleteQuotaCmd.Error = cmderror.ErrSuccess()

	return nil
}

func (deleteQuotaCmd *DeleteQuotaCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&deleteQuotaCmd.FinalDingoCmd)
}
