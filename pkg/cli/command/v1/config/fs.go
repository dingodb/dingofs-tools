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

type ConfigFsQuotaCommand struct {
	basecmd.FinalDingoCmd
	Rpc *cmdCommon.SetFsQuotaRpc
}

var _ basecmd.FinalDingoCmdFunc = (*ConfigFsQuotaCommand)(nil) // check interface

func NewConfigFsQuotaCommand() *cobra.Command {
	fsQuotaCmd := &ConfigFsQuotaCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:   "fs",
			Short: "config fs quota for dingofs",
			Example: `$ dingo config fs --mdsaddr "10.20.61.2:6700,172.20.61.3:6700,10.20.61.4:6700" --fsid 1 --capacity 10 --inodes 1000
$ dingo config fs --fsid 1 --capacity 10 --inodes 1000
$ dingo config fs --fsname dingofs --capacity 10 --inodes 1000
`,
		},
	}
	basecmd.NewFinalDingoCli(&fsQuotaCmd.FinalDingoCmd, fsQuotaCmd)
	return fsQuotaCmd.Cmd
}

func (fsQuotaCmd *ConfigFsQuotaCommand) AddFlags() {
	config.AddRpcRetryTimesFlag(fsQuotaCmd.Cmd)
	config.AddRpcTimeoutFlag(fsQuotaCmd.Cmd)
	config.AddFsMdsAddrFlag(fsQuotaCmd.Cmd)
	config.AddFsIdUint32OptionFlag(fsQuotaCmd.Cmd)
	config.AddFsNameStringOptionFlag(fsQuotaCmd.Cmd)
	config.AddFsCapacityOptionalFlag(fsQuotaCmd.Cmd)
	config.AddFsInodesOptionalFlag(fsQuotaCmd.Cmd)
}

func (fsQuotaCmd *ConfigFsQuotaCommand) Init(cmd *cobra.Command, args []string) error {
	_, getAddrErr := config.GetFsMdsAddrSlice(fsQuotaCmd.Cmd)
	if getAddrErr.TypeCode() != cmderror.CODE_SUCCESS {
		fsQuotaCmd.Error = getAddrErr
		return fmt.Errorf(getAddrErr.Message)
	}
	// check flags values
	capacity, inodes, quotaErr := cmdCommon.CheckAndGetQuotaValue(fsQuotaCmd.Cmd)
	if quotaErr != nil {
		return quotaErr
	}
	// get fs id
	fsId, fsErr := cmdCommon.GetFsId(cmd)
	if fsErr != nil {
		return fsErr
	}
	// get poolid copysetid
	partitionInfo, partErr := cmdCommon.GetPartitionInfo(fsQuotaCmd.Cmd, fsId, config.ROOTINODEID)
	if partErr != nil {
		return partErr
	}
	poolId := partitionInfo.GetPoolId()
	copyetId := partitionInfo.GetCopysetId()
	//set request info
	request := &metaserver.SetFsQuotaRequest{
		FsId:      &fsId,
		PoolId:    &poolId,
		CopysetId: &copyetId,
		Quota:     &metaserver.Quota{MaxBytes: &capacity, MaxInodes: &inodes},
	}
	fsQuotaCmd.Rpc = &cmdCommon.SetFsQuotaRpc{
		Request: request,
	}
	// get leader
	addrs, addrErr := cmdCommon.GetLeaderPeerAddr(fsQuotaCmd.Cmd, fsId, config.ROOTINODEID)
	if addrErr != nil {
		return addrErr
	}
	timeout := config.GetRpcTimeout(cmd)
	retrytimes := config.GetRpcRetryTimes(cmd)
	fsQuotaCmd.Rpc.Info = base.NewRpc(addrs, timeout, retrytimes, "setFsQuota")
	fsQuotaCmd.Rpc.Info.RpcDataShow = config.GetFlagBool(fsQuotaCmd.Cmd, "verbose")

	header := []string{cobrautil.ROW_RESULT}
	fsQuotaCmd.SetHeader(header)
	return nil
}

func (fsQuotaCmd *ConfigFsQuotaCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&fsQuotaCmd.FinalDingoCmd, fsQuotaCmd)
}

func (fsQuotaCmd *ConfigFsQuotaCommand) RunCommand(cmd *cobra.Command, args []string) error {
	result, err := base.GetRpcResponse(fsQuotaCmd.Rpc.Info, fsQuotaCmd.Rpc)
	if err.TypeCode() != cmderror.CODE_SUCCESS {
		return err.ToError()
	}
	response := result.(*metaserver.SetFsQuotaResponse)
	errQuota := cmderror.ErrQuota(int(response.GetStatusCode()))
	row := map[string]string{
		cobrautil.ROW_RESULT: errQuota.Message,
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

func (fsQuotaCmd *ConfigFsQuotaCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&fsQuotaCmd.FinalDingoCmd)
}
