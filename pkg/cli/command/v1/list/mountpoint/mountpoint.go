/*
 *  Copyright (c) 2022 NetEase Inc.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

/*
 * Project: DingoCli
 * Created Date: 2022-06-10
 * Author: chengyi (Cyber-SiKu)
 */

package mountpoint

import (
	"fmt"

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/v1/list/fs"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	mds "github.com/dingodb/dingofs-tools/proto/dingofs/proto/mds"
	"github.com/spf13/cobra"
)

const (
	mountpointExample = `$ dingo list mountpoint`
)

type MountpointCommand struct {
	basecmd.FinalDingoCmd
	fsInfo *mds.ListClusterFsInfoResponse
	number uint64
}

var _ basecmd.FinalDingoCmdFunc = (*MountpointCommand)(nil) // check interface

func NewMountpointCommand() *cobra.Command {
	mpCmd := &MountpointCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:     "mountpoint",
			Short:   "list all mountpoint of the dingofs",
			Example: mountpointExample,
		},
	}
	basecmd.NewFinalDingoCli(&mpCmd.FinalDingoCmd, mpCmd)
	return mpCmd.Cmd
}

func (mpCmd *MountpointCommand) AddFlags() {
	config.AddRpcRetryTimesFlag(mpCmd.Cmd)
	config.AddRpcRetryDelayFlag(mpCmd.Cmd)
	config.AddRpcTimeoutFlag(mpCmd.Cmd)
	config.AddFsMdsAddrFlag(mpCmd.Cmd)
}

func (mpCmd *MountpointCommand) Init(cmd *cobra.Command, args []string) error {
	var fsInfoErr *cmderror.CmdError
	mpCmd.fsInfo, fsInfoErr = fs.GetClusterFsInfo(mpCmd.Cmd)
	if fsInfoErr.TypeCode() != cmderror.CODE_SUCCESS {
		mpCmd.Error = fsInfoErr
		return fmt.Errorf(fsInfoErr.Message)
	}

	header := []string{cobrautil.ROW_FS_ID, cobrautil.ROW_FS_NAME, cobrautil.ROW_MOUNTPOINT}
	mpCmd.SetHeader(header)
	mpCmd.TableNew.SetAutoMergeCells(true)

	return nil
}

func (mpCmd *MountpointCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&mpCmd.FinalDingoCmd, mpCmd)
}

func (mpCmd *MountpointCommand) RunCommand(cmd *cobra.Command, args []string) error {
	mpCmd.updateTable()
	return nil
}

func (mpCmd *MountpointCommand) updateTable() {
	fssInfo := mpCmd.fsInfo.GetFsInfo()
	rows := make([]map[string]string, 0)
	for _, fsInfo := range fssInfo {
		if len(fsInfo.GetMountpoints()) == 0 {
			continue
		}
		for _, mountpoint := range fsInfo.GetMountpoints() {
			mpCmd.number++
			row := make(map[string]string)
			row[cobrautil.ROW_FS_ID] = fmt.Sprintf("%d", fsInfo.GetFsId())
			row[cobrautil.ROW_FS_NAME] = fsInfo.GetFsName()
			mountpointStr := fmt.Sprintf("%s:%d:%s", mountpoint.GetHostname(), mountpoint.GetPort(), mountpoint.GetPath())
			row[cobrautil.ROW_MOUNTPOINT] = mountpointStr
			rows = append(rows, row)
		}
	}
	list := cobrautil.ListMap2ListSortByKeys(rows, mpCmd.Header, []string{cobrautil.ROW_FS_ID})
	mpCmd.TableNew.AppendBulk(list)
	mpCmd.Result = rows
}

func (mpCmd *MountpointCommand) ResultPlainOutput() error {
	if mpCmd.number == 0 {
		fmt.Println("no mountpoint in dingofs")
	}
	return output.FinalCmdOutputPlain(&mpCmd.FinalDingoCmd)
}
