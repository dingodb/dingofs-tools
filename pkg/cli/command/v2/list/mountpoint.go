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

package list

import (
	"fmt"

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	"github.com/dingodb/dingofs-tools/pkg/base"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/v2/common"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	pbmdsv2error "github.com/dingodb/dingofs-tools/proto/dingofs/proto/error"
	pbmdsv2 "github.com/dingodb/dingofs-tools/proto/dingofs/proto/mdsv2"
	"github.com/spf13/cobra"
)

const (
	mountPointExample = `$ dingo list mountpoint`
)

type MountPointCommand struct {
	basecmd.FinalDingoCmd
	Rpc    *common.ListFsRpc
	number uint64
}

var _ basecmd.FinalDingoCmdFunc = (*MountPointCommand)(nil) // check interface

func NewMountPointCommand() *cobra.Command {
	mpCmd := &MountPointCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:     "mountpoint",
			Short:   "list all mountpoint of the dingofs",
			Example: mountPointExample,
		},
	}
	basecmd.NewFinalDingoCli(&mpCmd.FinalDingoCmd, mpCmd)
	return mpCmd.Cmd
}

func (mpCmd *MountPointCommand) AddFlags() {
	config.AddRpcRetryTimesFlag(mpCmd.Cmd)
	config.AddRpcTimeoutFlag(mpCmd.Cmd)
	config.AddFsMdsAddrFlag(mpCmd.Cmd)
}

func (mpCmd *MountPointCommand) Init(cmd *cobra.Command, args []string) error {
	// new rpc
	mdsRpc, err := common.CreateNewMdsRpc(cmd, "ListFsInfo")
	if err != nil {
		return err
	}
	// set request info
	mpCmd.Rpc = &common.ListFsRpc{Info: mdsRpc, Request: &pbmdsv2.ListFsInfoRequest{}}
	// set table header
	header := []string{cobrautil.ROW_FS_ID, cobrautil.ROW_FS_NAME, cobrautil.ROW_FS_CLIENTID, cobrautil.ROW_MOUNTPOINT, cobrautil.ROW_FS_CTO}
	mpCmd.SetHeader(header)
	mpCmd.TableNew.SetAutoMergeCells(true)

	return nil

}

func (mpCmd *MountPointCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&mpCmd.FinalDingoCmd, mpCmd)
}

func (mpCmd *MountPointCommand) RunCommand(cmd *cobra.Command, args []string) error {
	// get rpc request
	response, errCmd := base.GetRpcResponse(mpCmd.Rpc.Info, mpCmd.Rpc)
	if errCmd.TypeCode() != cmderror.CODE_SUCCESS {
		return fmt.Errorf(errCmd.Message)
	}
	result := response.(*pbmdsv2.ListFsInfoResponse)
	if mdsErr := result.GetError(); mdsErr.GetErrcode() != pbmdsv2error.Errno_OK {
		return cmderror.MDSV2Error(mdsErr).ToError()
	}
	// fill table
	fsInfos := result.GetFsInfos()
	rows := make([]map[string]string, 0)
	for _, fsInfo := range fsInfos {
		if len(fsInfo.GetMountPoints()) == 0 {
			continue
		}
		for _, mountPoint := range fsInfo.GetMountPoints() {
			mpCmd.number++
			row := make(map[string]string)
			row[cobrautil.ROW_FS_ID] = fmt.Sprintf("%d", fsInfo.GetFsId())
			row[cobrautil.ROW_FS_NAME] = fsInfo.GetFsName()
			row[cobrautil.ROW_FS_CLIENTID] = mountPoint.GetClientId()
			row[cobrautil.ROW_MOUNTPOINT] = fmt.Sprintf("%s:%d:%s", mountPoint.GetHostname(), mountPoint.GetPort(), mountPoint.GetPath())
			row[cobrautil.ROW_FS_CTO] = fmt.Sprintf("%v", mountPoint.GetCto())
			rows = append(rows, row)
		}
	}
	list := cobrautil.ListMap2ListSortByKeys(rows, mpCmd.Header, []string{cobrautil.ROW_FS_ID})
	mpCmd.TableNew.AppendBulk(list)
	// to json
	res, err := output.MarshalProtoJson(result)
	if err != nil {
		return err
	}
	mapRes := res.(map[string]interface{})
	mpCmd.Result = mapRes
	mpCmd.Error = cmderror.ErrSuccess()

	return nil
}

func (mpCmd *MountPointCommand) ResultPlainOutput() error {
	if mpCmd.number == 0 {
		fmt.Println("no mountpoint in dingofs")
	}
	return output.FinalCmdOutputPlain(&mpCmd.FinalDingoCmd)
}
