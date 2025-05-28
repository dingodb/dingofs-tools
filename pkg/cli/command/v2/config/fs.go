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

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	"github.com/dingodb/dingofs-tools/pkg/base"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	cmdcommon "github.com/dingodb/dingofs-tools/pkg/cli/command/v1/common"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/v2/common"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	pbmdsv2 "github.com/dingodb/dingofs-tools/proto/dingofs/proto/mdsv2"

	"github.com/spf13/cobra"
)

type SetFsQuotaCommand struct {
	basecmd.FinalDingoCmd
	Rpc *common.SetFsQuotaRpc
}

var _ basecmd.FinalDingoCmdFunc = (*SetFsQuotaCommand)(nil) // check interface

func NewSetFsQuotaCommand() *cobra.Command {
	fsQuotaCmd := &SetFsQuotaCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:   "fs",
			Short: "set fs quota for dingofs",
			Example: `$ dingo config fs --fsid 1 --capacity 10 --inodes 1000
$ dingo config fs --fsid 1 --capacity 10
$ dingo config fs --fsname dingofs --capacity 10 --inodes 1000
`,
		},
	}
	basecmd.NewFinalDingoCli(&fsQuotaCmd.FinalDingoCmd, fsQuotaCmd)
	return fsQuotaCmd.Cmd
}

func (fsQuotaCmd *SetFsQuotaCommand) AddFlags() {
	config.AddRpcRetryTimesFlag(fsQuotaCmd.Cmd)
	config.AddRpcTimeoutFlag(fsQuotaCmd.Cmd)
	config.AddFsMdsAddrFlag(fsQuotaCmd.Cmd)
	config.AddFsIdUint32OptionFlag(fsQuotaCmd.Cmd)
	config.AddFsNameStringOptionFlag(fsQuotaCmd.Cmd)
	config.AddFsCapacityOptionalFlag(fsQuotaCmd.Cmd)
	config.AddFsInodesOptionalFlag(fsQuotaCmd.Cmd)
}

func (fsQuotaCmd *SetFsQuotaCommand) Init(cmd *cobra.Command, args []string) error {
	// new prc
	mdsRpc, err := common.CreateNewMdsRpc(cmd, "setFsQuota")
	if err != nil {
		return err
	}
	// check flags values
	capacity, inodes, quotaErr := cmdcommon.CheckAndGetQuotaValue(fsQuotaCmd.Cmd)
	if quotaErr != nil {
		return quotaErr
	}
	// get fs id
	fsId, fsErr := common.GetFsId(cmd)
	if fsErr != nil {
		return fsErr
	}
	// set request info
	request := &pbmdsv2.SetFsQuotaRequest{
		FsId:  fsId,
		Quota: &pbmdsv2.Quota{MaxBytes: capacity, MaxInodes: inodes},
	}
	fsQuotaCmd.Rpc = &common.SetFsQuotaRpc{
		Info:    mdsRpc,
		Request: request,
	}
	// set table header
	header := []string{cobrautil.ROW_RESULT}
	fsQuotaCmd.SetHeader(header)

	return nil
}

func (fsQuotaCmd *SetFsQuotaCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&fsQuotaCmd.FinalDingoCmd, fsQuotaCmd)
}

func (fsQuotaCmd *SetFsQuotaCommand) RunCommand(cmd *cobra.Command, args []string) error {
	// get rpc result
	response, errCmd := base.GetRpcResponse(fsQuotaCmd.Rpc.Info, fsQuotaCmd.Rpc)
	if errCmd.TypeCode() != cmderror.CODE_SUCCESS {
		return fmt.Errorf(errCmd.Message)
	}
	result := response.(*pbmdsv2.SetFsQuotaResponse)
	mdsErr := result.GetError()
	row := map[string]string{
		cobrautil.ROW_RESULT: cmderror.MDSV2Error(mdsErr).Message,
	}
	fsQuotaCmd.TableNew.Append(cobrautil.Map2List(row, fsQuotaCmd.Header))
	// to json
	res, errTranslate := output.MarshalProtoJson(result)
	if errTranslate != nil {
		return errTranslate
	}
	mapRes := res.(map[string]interface{})
	fsQuotaCmd.Result = mapRes
	fsQuotaCmd.Error = cmderror.ErrSuccess()

	return nil
}

func (fsQuotaCmd *SetFsQuotaCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&fsQuotaCmd.FinalDingoCmd)
}
