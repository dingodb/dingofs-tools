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
	"strconv"

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

type GetQuotaCommand struct {
	basecmd.FinalDingoCmd
	Rpc      *cmdCommon.GetQuotaRpc
	Path     string
	Response *metaserver.GetDirQuotaResponse
}

var _ basecmd.FinalDingoCmdFunc = (*GetQuotaCommand)(nil) // check interface

func NewGetQuotaDataCommand() *GetQuotaCommand {
	getQuotaCmd := &GetQuotaCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:     "get",
			Short:   "get quota of a directory",
			Example: `$ dingo quota get --fsid 1 --path /quotadir`,
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
	_, getAddrErr := config.GetFsMdsAddrSlice(getQuotaCmd.Cmd)
	if getAddrErr.TypeCode() != cmderror.CODE_SUCCESS {
		getQuotaCmd.Error = getAddrErr
		return fmt.Errorf(getAddrErr.Message)
	}
	//check flags values
	fsId, fsErr := cmdCommon.GetFsId(getQuotaCmd.Cmd)
	if fsErr != nil {
		return fsErr
	}
	path := config.GetFlagString(getQuotaCmd.Cmd, config.DINGOFS_QUOTA_PATH)
	if len(path) == 0 {
		return fmt.Errorf("path is required")
	}
	getQuotaCmd.Path = path
	//get inodeid
	dirInodeId, inodeErr := cmdCommon.GetDirPathInodeId(getQuotaCmd.Cmd, fsId, getQuotaCmd.Path)
	if inodeErr != nil {
		return inodeErr
	}
	// get poolid copysetid
	partitionInfo, partErr := cmdCommon.GetPartitionInfo(getQuotaCmd.Cmd, fsId, config.ROOTINODEID)
	if partErr != nil {
		return partErr
	}
	poolId := partitionInfo.GetPoolId()
	copyetId := partitionInfo.GetCopysetId()
	//set rpc request
	request := &metaserver.GetDirQuotaRequest{
		PoolId:     &poolId,
		CopysetId:  &copyetId,
		FsId:       &fsId,
		DirInodeId: &dirInodeId,
	}
	getQuotaCmd.Rpc = &cmdCommon.GetQuotaRpc{
		Request: request,
	}
	//get request addr leader
	addrs, addrErr := cmdCommon.GetLeaderPeerAddr(getQuotaCmd.Cmd, fsId, config.ROOTINODEID)
	if addrErr != nil {
		return addrErr
	}

	timeout := config.GetRpcTimeout(cmd)
	retrytimes := config.GetRpcRetryTimes(cmd)
	retryDelay := config.GetRpcRetryDelay(cmd)
	verbose := config.GetFlagBool(cmd, config.VERBOSE)
	getQuotaCmd.Rpc.Info = base.NewRpc(addrs, timeout, retrytimes, retryDelay, verbose, "GetDirQuota")

	header := []string{cobrautil.ROW_ID, cobrautil.ROW_PATH, cobrautil.ROW_CAPACITY, cobrautil.ROW_USED, cobrautil.ROW_USED_PERCNET,
		cobrautil.ROW_INODES, cobrautil.ROW_INODES_IUSED, cobrautil.ROW_INODES_PERCENT}
	getQuotaCmd.SetHeader(header)
	return nil
}

func (getQuotaCmd *GetQuotaCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&getQuotaCmd.FinalDingoCmd, getQuotaCmd)
}

func (getQuotaCmd *GetQuotaCommand) RunCommand(cmd *cobra.Command, args []string) error {
	result, err := base.GetRpcResponse(getQuotaCmd.Rpc.Info, getQuotaCmd.Rpc)
	if err.TypeCode() != cmderror.CODE_SUCCESS {
		return err.ToError()
	}
	response := result.(*metaserver.GetDirQuotaResponse)
	getQuotaCmd.Response = response

	if statusCode := response.GetStatusCode(); statusCode != metaserver.MetaStatusCode_OK {
		return cmderror.ErrQuota(int(statusCode)).ToError()
	}
	quota := response.GetQuota()
	quotaValueSlice := cmdCommon.ConvertQuotaToHumanizeValue(quota.GetMaxBytes(), quota.GetUsedBytes(), quota.GetMaxInodes(), quota.GetUsedInodes())
	row := map[string]string{
		cobrautil.ROW_ID:             strconv.FormatUint(getQuotaCmd.Rpc.Request.GetDirInodeId(), 10),
		cobrautil.ROW_PATH:           getQuotaCmd.Path,
		cobrautil.ROW_CAPACITY:       quotaValueSlice[0],
		cobrautil.ROW_USED:           quotaValueSlice[1],
		cobrautil.ROW_USED_PERCNET:   quotaValueSlice[2],
		cobrautil.ROW_INODES:         quotaValueSlice[3],
		cobrautil.ROW_INODES_IUSED:   quotaValueSlice[4],
		cobrautil.ROW_INODES_PERCENT: quotaValueSlice[5],
	}
	getQuotaCmd.TableNew.Append(cobrautil.Map2List(row, getQuotaCmd.Header))

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

func GetDirQuotaData(caller *cobra.Command) (*metaserver.GetDirQuotaRequest, *metaserver.GetDirQuotaResponse, error) {
	dirQuotaDataCmd := NewGetQuotaDataCommand()
	dirQuotaDataCmd.Cmd.SetArgs([]string{"--format", config.FORMAT_NOOUT})
	config.AlignFlagsValue(caller, dirQuotaDataCmd.Cmd, []string{
		config.RPCRETRYTIMES, config.RPCTIMEOUT, config.DINGOFS_MDSADDR, config.DINGOFS_FSID, config.DINGOFS_FSNAME,
		config.DINGOFS_QUOTA_PATH,
	})
	dirQuotaDataCmd.Cmd.SilenceErrors = true
	err := dirQuotaDataCmd.Cmd.Execute()
	if err != nil {
		return nil, nil, err
	}
	return dirQuotaDataCmd.Rpc.Request, dirQuotaDataCmd.Response, nil
}
