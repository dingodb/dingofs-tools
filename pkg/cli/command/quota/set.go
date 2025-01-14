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
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	cmdCommon "github.com/dingodb/dingofs-tools/pkg/cli/command/common"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	"github.com/dingodb/dingofs-tools/proto/dingofs/proto/metaserver"
	"github.com/spf13/cobra"
)

type SetQuotaCommand struct {
	basecmd.FinalDingoCmd
	Rpc *cmdCommon.SetQuotaRpc
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
	config.AddRpcTimeoutFlag(setQuotaCmd.Cmd)
	config.AddFsMdsAddrFlag(setQuotaCmd.Cmd)
	config.AddFsIdUint32OptionFlag(setQuotaCmd.Cmd)
	config.AddFsNameStringOptionFlag(setQuotaCmd.Cmd)
	config.AddFsPathRequiredFlag(setQuotaCmd.Cmd)
	config.AddFsCapacityOptionalFlag(setQuotaCmd.Cmd)
	config.AddFsInodesOptionalFlag(setQuotaCmd.Cmd)
}

func (setQuotaCmd *SetQuotaCommand) Init(cmd *cobra.Command, args []string) error {
	_, getAddrErr := config.GetFsMdsAddrSlice(setQuotaCmd.Cmd)
	if getAddrErr.TypeCode() != cmderror.CODE_SUCCESS {
		setQuotaCmd.Error = getAddrErr
		return fmt.Errorf(getAddrErr.Message)
	}
	//check flags values
	capacity, inodes, quotaErr := cmdCommon.CheckAndGetQuotaValue(setQuotaCmd.Cmd)
	if quotaErr != nil {
		return quotaErr
	}
	fsId, fsErr := cmdCommon.GetFsId(cmd)
	if fsErr != nil {
		return fsErr
	}
	path := config.GetFlagString(setQuotaCmd.Cmd, config.DINGOFS_QUOTA_PATH)
	if len(path) == 0 {
		return fmt.Errorf("path is required")
	}
	//get inodeid
	dirInodeId, inodeErr := cmdCommon.GetDirPathInodeId(setQuotaCmd.Cmd, fsId, path)
	if inodeErr != nil {
		return inodeErr
	}
	// get directory real used
	realUsedBytes, realUsedInodes, getErr := cmdCommon.GetDirectorySizeAndInodes(setQuotaCmd.Cmd, fsId, dirInodeId, false)
	if getErr != nil {
		return getErr
	}
	// get poolid copysetid
	partitionInfo, partErr := cmdCommon.GetPartitionInfo(setQuotaCmd.Cmd, fsId, config.ROOTINODEID)
	if partErr != nil {
		return partErr
	}
	poolId := partitionInfo.GetPoolId()
	copyetId := partitionInfo.GetCopysetId()

	//set rpc request
	request := &metaserver.SetDirQuotaRequest{
		PoolId:     &poolId,
		CopysetId:  &copyetId,
		FsId:       &fsId,
		DirInodeId: &dirInodeId,
		Quota: &metaserver.Quota{
			MaxBytes:   &capacity,
			MaxInodes:  &inodes,
			UsedBytes:  &realUsedBytes,
			UsedInodes: &realUsedInodes,
		},
	}
	setQuotaCmd.Rpc = &cmdCommon.SetQuotaRpc{
		Request: request,
	}
	//get request addr leader
	addrs, addrErr := cmdCommon.GetLeaderPeerAddr(setQuotaCmd.Cmd, fsId, config.ROOTINODEID)
	if addrErr != nil {
		return addrErr
	}
	timeout := config.GetRpcTimeout(cmd)
	retrytimes := config.GetRpcRetryTimes(cmd)
	setQuotaCmd.Rpc.Info = basecmd.NewRpc(addrs, timeout, retrytimes, "SetDirQuota")
	setQuotaCmd.Rpc.Info.RpcDataShow = config.GetFlagBool(setQuotaCmd.Cmd, config.VERBOSE)

	header := []string{cobrautil.ROW_RESULT}
	setQuotaCmd.SetHeader(header)
	return nil
}

func (setQuotaCmd *SetQuotaCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&setQuotaCmd.FinalDingoCmd, setQuotaCmd)
}

func (setQuotaCmd *SetQuotaCommand) RunCommand(cmd *cobra.Command, args []string) error {
	result, err := basecmd.GetRpcResponse(setQuotaCmd.Rpc.Info, setQuotaCmd.Rpc)
	if err.TypeCode() != cmderror.CODE_SUCCESS {
		return err.ToError()
	}
	response := result.(*metaserver.SetDirQuotaResponse)
	errQuota := cmderror.ErrQuota(int(response.GetStatusCode()))
	row := map[string]string{
		cobrautil.ROW_RESULT: errQuota.Message,
	}
	setQuotaCmd.TableNew.Append(cobrautil.Map2List(row, setQuotaCmd.Header))

	res, errTranslate := output.MarshalProtoJson(response)
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
