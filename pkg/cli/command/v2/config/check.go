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
	"github.com/dingodb/dingofs-tools/pkg/cli/command/v2/common"

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	"github.com/dingodb/dingofs-tools/pkg/base"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	cmdCommon "github.com/dingodb/dingofs-tools/pkg/cli/command/v1/common"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	pbmdsv2 "github.com/dingodb/dingofs-tools/proto/dingofs/proto/mdsv2"
	"github.com/spf13/cobra"
)

type CheckQuotaCommand struct {
	basecmd.FinalDingoCmd
	Rpc *common.SetFsQuotaRpc
}

var _ basecmd.FinalDingoCmdFunc = (*CheckQuotaCommand)(nil) // check interface

func NewCheckQuotaCommand() *cobra.Command {
	checkQuotaCmd := &CheckQuotaCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:   "check",
			Short: "check quota of fs",
			Example: `$ dingo config check --fsid 1
$ dingo config check --fsid 1 --threads 8`,
		},
	}
	basecmd.NewFinalDingoCli(&checkQuotaCmd.FinalDingoCmd, checkQuotaCmd)
	return checkQuotaCmd.Cmd
}

func (checkQuotaCmd *CheckQuotaCommand) AddFlags() {
	config.AddRpcRetryTimesFlag(checkQuotaCmd.Cmd)
	config.AddRpcRetryDelayFlag(checkQuotaCmd.Cmd)
	config.AddRpcTimeoutFlag(checkQuotaCmd.Cmd)
	config.AddFsMdsAddrFlag(checkQuotaCmd.Cmd)
	config.AddFsIdUint32OptionFlag(checkQuotaCmd.Cmd)
	config.AddFsNameStringOptionFlag(checkQuotaCmd.Cmd)
	config.AddThreadsOptionFlag(checkQuotaCmd.Cmd)
	config.AddBoolOptionPFlag(checkQuotaCmd.Cmd, config.DINGOFS_QUOTA_REPAIR, "r", "repair inconsistent quota (default: false)")
}

func (checkQuotaCmd *CheckQuotaCommand) Init(cmd *cobra.Command, args []string) error {
	_, getAddrErr := config.GetFsMdsAddrSlice(checkQuotaCmd.Cmd)
	if getAddrErr.TypeCode() != cmderror.CODE_SUCCESS {
		checkQuotaCmd.Error = getAddrErr
		return fmt.Errorf(getAddrErr.Message)
	}

	return nil
}

func (checkQuotaCmd *CheckQuotaCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&checkQuotaCmd.FinalDingoCmd, checkQuotaCmd)
}

func (checkQuotaCmd *CheckQuotaCommand) RunCommand(cmd *cobra.Command, args []string) error {
	// get fs id
	fsId, idErr := common.GetFsId(cmd)
	if idErr != nil {
		return idErr
	}
	// get filesystem name
	fsName, nameErr := common.GetFsName(cmd)
	if nameErr != nil {
		return nameErr
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

	// get quota
	_, fsQuotaData, getErr := GetFsQuotaData(cmd, fsId)
	if getErr != nil {
		return getErr
	}
	fsQuota := fsQuotaData.GetQuota()

	// get root director real used space
	threads := config.GetFlagUint32(cmd, config.DINGOFS_THREADS)
	realUsedBytes, realUsedInodes, getErr := common.GetDirectorySizeAndInodes(checkQuotaCmd.Cmd, fsId, config.ROOTINODEID, true, epoch, threads)
	if getErr != nil {
		return getErr
	}

	checkResult, ok := cmdCommon.CheckQuota(fsQuota.GetMaxBytes(), fsQuota.GetUsedBytes(), fsQuota.GetMaxInodes(), fsQuota.GetUsedInodes(), realUsedBytes, realUsedInodes)
	repair := config.GetFlagBool(checkQuotaCmd.Cmd, config.DINGOFS_QUOTA_REPAIR)
	if repair && !ok { // inconsistent and need to repair
		// new prc
		mdsRpc, err := common.CreateNewMdsRpc(cmd, "setFsQuota")
		if err != nil {
			return err
		}
		// set request info
		request := &pbmdsv2.SetFsQuotaRequest{
			FsId:  fsId,
			Quota: &pbmdsv2.Quota{UsedBytes: realUsedBytes, UsedInodes: realUsedInodes},
		}

		setFsQuotaRpc := &common.SetFsQuotaRpc{
			Info:    mdsRpc,
			Request: request,
		}

		// get rpc result
		response, errCmd := base.GetRpcResponse(setFsQuotaRpc.Info, setFsQuotaRpc)
		if errCmd.TypeCode() != cmderror.CODE_SUCCESS {
			return fmt.Errorf(errCmd.Message)
		}
		result := response.(*pbmdsv2.SetFsQuotaResponse)
		mdsErr := result.GetError()

		// fill table
		row := map[string]string{
			cobrautil.ROW_RESULT: cmderror.MDSV2Error(mdsErr).Message,
		}
		header := []string{cobrautil.ROW_RESULT}
		checkQuotaCmd.SetHeader(header)
		checkQuotaCmd.TableNew.Append(cobrautil.Map2List(row, checkQuotaCmd.Header))

	} else {
		header := []string{cobrautil.ROW_FS_ID, cobrautil.ROW_FS_NAME, cobrautil.ROW_CAPACITY, cobrautil.ROW_USED, cobrautil.ROW_REAL_USED, cobrautil.ROW_INODES, cobrautil.ROW_INODES_IUSED, cobrautil.ROW_INODES_REAL_IUSED, cobrautil.ROW_STATUS}
		checkQuotaCmd.SetHeader(header)

		row := map[string]string{
			cobrautil.ROW_FS_ID:             fmt.Sprintf("%d", fsId),
			cobrautil.ROW_FS_NAME:           fsName,
			cobrautil.ROW_CAPACITY:          checkResult[0],
			cobrautil.ROW_USED:              checkResult[1],
			cobrautil.ROW_REAL_USED:         checkResult[2],
			cobrautil.ROW_INODES:            checkResult[3],
			cobrautil.ROW_INODES_IUSED:      checkResult[4],
			cobrautil.ROW_INODES_REAL_IUSED: checkResult[5],
			cobrautil.ROW_STATUS:            checkResult[6],
		}
		checkQuotaCmd.TableNew.Append(cobrautil.Map2List(row, checkQuotaCmd.Header))
	}

	return nil
}

func (checkQuotaCmd *CheckQuotaCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&checkQuotaCmd.FinalDingoCmd)
}
