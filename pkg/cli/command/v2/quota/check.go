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

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	"github.com/dingodb/dingofs-tools/pkg/base"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	cmdCommon "github.com/dingodb/dingofs-tools/pkg/cli/command/v1/common"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/v2/common"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	pbmdsv2 "github.com/dingodb/dingofs-tools/proto/dingofs/proto/mdsv2"
	"github.com/spf13/cobra"
)

type CheckQuotaCommand struct {
	basecmd.FinalDingoCmd
	Rpc *common.CheckDirQuotaRpc
}

var _ basecmd.FinalDingoCmdFunc = (*CheckQuotaCommand)(nil) // check interface

func NewCheckQuotaCommand() *cobra.Command {
	checkQuotaCmd := &CheckQuotaCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:   "check",
			Short: "check directory quota consistency ",
			Example: `$ dingo quota check --fsid 1 --path /quotadir
$ dingo quota check --fsname 1 --path /quotadir`,
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
	config.AddFsPathRequiredFlag(checkQuotaCmd.Cmd)
	config.AddBoolOptionPFlag(checkQuotaCmd.Cmd, config.DINGOFS_QUOTA_REPAIR, "r", "repair inconsistent quota (default: false)")
}

func (checkQuotaCmd *CheckQuotaCommand) Init(cmd *cobra.Command, args []string) error {
	return nil
}

func (checkQuotaCmd *CheckQuotaCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&checkQuotaCmd.FinalDingoCmd, checkQuotaCmd)
}

func (checkQuotaCmd *CheckQuotaCommand) RunCommand(cmd *cobra.Command, args []string) error {
	// check flags values
	fsId, fsErr := common.GetFsId(cmd)
	if fsErr != nil {
		return fsErr
	}
	path := config.GetFlagString(cmd, config.DINGOFS_QUOTA_PATH)
	if len(path) == 0 {
		return fmt.Errorf("path is required")
	}
	//get inodeid
	dirInodeId, inodeErr := common.GetDirPathInodeId(cmd, fsId, path)
	if inodeErr != nil {
		return inodeErr
	}
	_, dirQuotaResponse, err := GetDirQuotaData(cmd, fsId, dirInodeId)
	if err != nil {
		return err
	}
	// get inode by path
	dirInode, dirErr := common.GetDirPathInodeId(cmd, fsId, path)
	if dirErr != nil {
		return dirErr
	}
	// get real used space
	realUsedBytes, realUsedInodes, getErr := common.GetDirectorySizeAndInodes(checkQuotaCmd.Cmd, fsId, dirInode, false)
	if getErr != nil {
		return getErr
	}

	dirQuota := dirQuotaResponse.GetQuota()
	checkResult, ok := cmdCommon.CheckQuota(dirQuota.GetMaxBytes(), dirQuota.GetUsedBytes(), dirQuota.GetMaxInodes(), dirQuota.GetUsedInodes(), realUsedBytes, realUsedInodes)
	repair := config.GetFlagBool(checkQuotaCmd.Cmd, config.DINGOFS_QUOTA_REPAIR)
	if repair && !ok { // inconsistent and need to repair
		// new prc
		mdsRpc, err := common.CreateNewMdsRpc(cmd, "SetDirQuota")
		if err != nil {
			return err
		}
		request := &pbmdsv2.SetDirQuotaRequest{
			FsId:  fsId,
			Ino:   dirInodeId,
			Quota: &pbmdsv2.Quota{UsedBytes: realUsedBytes, UsedInodes: realUsedInodes},
		}
		checkQuotaCmd.Rpc = &common.CheckDirQuotaRpc{
			Info:    mdsRpc,
			Request: request,
		}
		// get rpc result
		response, errCmd := base.GetRpcResponse(checkQuotaCmd.Rpc.Info, checkQuotaCmd.Rpc)
		if errCmd.TypeCode() != cmderror.CODE_SUCCESS {
			return fmt.Errorf(errCmd.Message)
		}
		result := response.(*pbmdsv2.SetDirQuotaResponse)
		mdsErr := result.GetError()
		//set header
		header := []string{cobrautil.ROW_RESULT}
		checkQuotaCmd.SetHeader(header)
		// fill table
		row := map[string]string{
			cobrautil.ROW_RESULT: cmderror.MDSV2Error(mdsErr).Message,
		}
		checkQuotaCmd.TableNew.Append(cobrautil.Map2List(row, checkQuotaCmd.Header))

	} else {
		header := []string{cobrautil.ROW_INODE_ID, cobrautil.ROW_NAME, cobrautil.ROW_CAPACITY, cobrautil.ROW_USED, cobrautil.ROW_REAL_USED, cobrautil.ROW_INODES, cobrautil.ROW_INODES_IUSED, cobrautil.ROW_INODES_REAL_IUSED, cobrautil.ROW_STATUS}
		checkQuotaCmd.SetHeader(header)
		row := map[string]string{
			cobrautil.ROW_INODE_ID:          fmt.Sprintf("%d", dirInodeId),
			cobrautil.ROW_NAME:              path,
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
