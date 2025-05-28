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

package config

import (
	"fmt"
	"strconv"

	"github.com/dingodb/dingofs-tools/proto/dingofs/proto/metaserver"

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	"github.com/dingodb/dingofs-tools/pkg/base"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	cmdCommon "github.com/dingodb/dingofs-tools/pkg/cli/command/v1/common"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	"github.com/spf13/cobra"
)

type GetFsQuotaCommand struct {
	basecmd.FinalDingoCmd
	Rpc *cmdCommon.GetFsQuotaRpc
}

var _ basecmd.FinalDingoCmdFunc = (*GetFsQuotaCommand)(nil) // check interface

func NewGetFsQuotaCommand() *cobra.Command {
	fsQuotaCmd := &GetFsQuotaCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:   "get",
			Short: "get fs quota for dingofs",
			Example: `$ dingo config get --fsid 1 
$ dingo config fs --fsname dingofs
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
	_, getAddrErr := config.GetFsMdsAddrSlice(fsQuotaCmd.Cmd)
	if getAddrErr.TypeCode() != cmderror.CODE_SUCCESS {
		fsQuotaCmd.Error = getAddrErr
		return fmt.Errorf(getAddrErr.Message)
	}
	header := []string{cobrautil.ROW_FS_ID, cobrautil.ROW_FS_NAME, cobrautil.ROW_CAPACITY, cobrautil.ROW_USED, cobrautil.ROW_USED_PERCNET,
		cobrautil.ROW_INODES, cobrautil.ROW_INODES_IUSED, cobrautil.ROW_INODES_PERCENT}
	fsQuotaCmd.SetHeader(header)
	return nil
}

func (fsQuotaCmd *GetFsQuotaCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&fsQuotaCmd.FinalDingoCmd, fsQuotaCmd)
}

func (fsQuotaCmd *GetFsQuotaCommand) RunCommand(cmd *cobra.Command, args []string) error {

	fsId, fsErr := cmdCommon.GetFsId(fsQuotaCmd.Cmd)
	if fsErr != nil {
		return fsErr
	}
	request, response, err := GetFsQuotaData(fsQuotaCmd.Cmd, fsId)
	if err != nil {
		return err
	}
	fsQuota := response.GetQuota()
	quotaValueSlice := cmdCommon.ConvertQuotaToHumanizeValue(fsQuota.GetMaxBytes(), fsQuota.GetUsedBytes(), fsQuota.GetMaxInodes(), fsQuota.GetUsedInodes())
	//get filesystem name
	fsName, fsErr := cmdCommon.GetFsName(cmd)
	if fsErr != nil {
		return fsErr
	}
	row := map[string]string{
		cobrautil.ROW_FS_ID:          strconv.FormatUint(uint64(request.GetFsId()), 10),
		cobrautil.ROW_FS_NAME:        fsName,
		cobrautil.ROW_CAPACITY:       quotaValueSlice[0],
		cobrautil.ROW_USED:           quotaValueSlice[1],
		cobrautil.ROW_USED_PERCNET:   quotaValueSlice[2],
		cobrautil.ROW_INODES:         quotaValueSlice[3],
		cobrautil.ROW_INODES_IUSED:   quotaValueSlice[4],
		cobrautil.ROW_INODES_PERCENT: quotaValueSlice[5],
	}
	fsQuotaCmd.TableNew.Append(cobrautil.Map2List(row, fsQuotaCmd.Header))

	res, errTranslate := output.MarshalProtoJson(response)
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

func GetFsQuotaData(cmd *cobra.Command, fsId uint32) (*metaserver.GetFsQuotaRequest, *metaserver.GetFsQuotaResponse, error) {
	// get poolid copysetid
	partitionInfo, partErr := cmdCommon.GetPartitionInfo(cmd, fsId, config.ROOTINODEID)
	if partErr != nil {
		return nil, nil, partErr
	}
	poolId := partitionInfo.GetPoolId()
	copyetId := partitionInfo.GetCopysetId()
	request := &metaserver.GetFsQuotaRequest{
		PoolId:    &poolId,
		CopysetId: &copyetId,
		FsId:      &fsId,
	}
	requestRpc := &cmdCommon.GetFsQuotaRpc{
		Request: request,
	}
	addrs, addrErr := cmdCommon.GetLeaderPeerAddr(cmd, uint32(fsId), config.ROOTINODEID)
	if addrErr != nil {
		return nil, nil, addrErr
	}
	timeout := config.GetRpcTimeout(cmd)
	retrytimes := config.GetRpcRetryTimes(cmd)
	requestRpc.Info = base.NewRpc(addrs, timeout, retrytimes, "getFsQuota")
	requestRpc.Info.RpcDataShow = config.GetFlagBool(cmd, "verbose")

	result, err := base.GetRpcResponse(requestRpc.Info, requestRpc)
	if err.TypeCode() != cmderror.CODE_SUCCESS {
		return nil, nil, err.ToError()
	}
	response := result.(*metaserver.GetFsQuotaResponse)
	if statusCode := response.GetStatusCode(); statusCode != metaserver.MetaStatusCode_OK {
		return request, response, cmderror.ErrQuota(int(statusCode)).ToError()
	}
	return request, response, nil
}
