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

type DeleteQuotaCommand struct {
	basecmd.FinalDingoCmd
	Rpc *cmdCommon.DeleteQuotaRpc
}

var _ basecmd.FinalDingoCmdFunc = (*DeleteQuotaCommand)(nil) // check interface

func NewDeleteQuotaCommand() *cobra.Command {
	deleteQuotaCmd := &DeleteQuotaCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:     "delete",
			Short:   "delete quota of a directory",
			Example: `$ dingo quota delete --fsid 1 --path /quotadir`,
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
	_, getAddrErr := config.GetFsMdsAddrSlice(deleteQuotaCmd.Cmd)
	if getAddrErr.TypeCode() != cmderror.CODE_SUCCESS {
		deleteQuotaCmd.Error = getAddrErr
		return fmt.Errorf(getAddrErr.Message)
	}
	//check flags values
	fsId, fsErr := cmdCommon.GetFsId(cmd)
	if fsErr != nil {
		return fsErr
	}
	path := config.GetFlagString(deleteQuotaCmd.Cmd, config.DINGOFS_QUOTA_PATH)
	if len(path) == 0 {
		return fmt.Errorf("path is required")
	}
	//get inodeid
	dirInodeId, inodeErr := cmdCommon.GetDirPathInodeId(deleteQuotaCmd.Cmd, fsId, path)
	if inodeErr != nil {
		return inodeErr
	}
	// get poolid copysetid
	partitionInfo, partErr := cmdCommon.GetPartitionInfo(deleteQuotaCmd.Cmd, fsId, config.ROOTINODEID)
	if partErr != nil {
		return partErr
	}
	poolId := partitionInfo.GetPoolId()
	copyetId := partitionInfo.GetCopysetId()
	//set rpc request
	request := &metaserver.DeleteDirQuotaRequest{
		PoolId:     &poolId,
		CopysetId:  &copyetId,
		FsId:       &fsId,
		DirInodeId: &dirInodeId,
	}
	deleteQuotaCmd.Rpc = &cmdCommon.DeleteQuotaRpc{
		Request: request,
	}
	//get request addr leader
	addrs, addrErr := cmdCommon.GetLeaderPeerAddr(deleteQuotaCmd.Cmd, fsId, config.ROOTINODEID)
	if addrErr != nil {
		return addrErr
	}

	timeout := config.GetRpcTimeout(cmd)
	retrytimes := config.GetRpcRetryTimes(cmd)
	retryDelay := config.GetRpcRetryDelay(cmd)
	verbose := config.GetFlagBool(cmd, config.VERBOSE)
	deleteQuotaCmd.Rpc.Info = base.NewRpc(addrs, timeout, retrytimes, retryDelay, verbose, "DeleteDirQuota")

	header := []string{cobrautil.ROW_RESULT}
	deleteQuotaCmd.SetHeader(header)
	return nil
}

func (deleteQuotaCmd *DeleteQuotaCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&deleteQuotaCmd.FinalDingoCmd, deleteQuotaCmd)
}

func (deleteQuotaCmd *DeleteQuotaCommand) RunCommand(cmd *cobra.Command, args []string) error {
	result, err := base.GetRpcResponse(deleteQuotaCmd.Rpc.Info, deleteQuotaCmd.Rpc)
	if err.TypeCode() != cmderror.CODE_SUCCESS {
		return err.ToError()
	}
	response := result.(*metaserver.DeleteDirQuotaResponse)
	//mock rpc data
	//status := metaserver.MetaStatusCode_OK
	//response := &metaserver.DeleteDirQuotaResponse{
	//	StatusCode: &status,
	//}
	errQuota := cmderror.ErrQuota(int(response.GetStatusCode()))
	row := map[string]string{
		cobrautil.ROW_RESULT: errQuota.Message,
	}
	deleteQuotaCmd.TableNew.Append(cobrautil.Map2List(row, deleteQuotaCmd.Header))

	res, errTranslate := output.MarshalProtoJson(response)
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
