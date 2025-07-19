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
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/v2/common"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	"github.com/spf13/cobra"
)

const (
	dentryExample = `$ dingo list dentry --fsid 1 --inodeid 1
$ dingo list dentry --fsname dingofs --inodeid 1`
)

type DentryCommand struct {
	basecmd.FinalDingoCmd
}

var _ basecmd.FinalDingoCmdFunc = (*DentryCommand)(nil) // check interface

func NewDentryCommand() *cobra.Command {
	dentryCmd := &DentryCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:     "dentry",
			Short:   "list directory dentry",
			Example: dentryExample,
		},
	}
	basecmd.NewFinalDingoCli(&dentryCmd.FinalDingoCmd, dentryCmd)
	return dentryCmd.Cmd
}

func (dentryCommand *DentryCommand) AddFlags() {
	config.AddRpcRetryTimesFlag(dentryCommand.Cmd)
	config.AddRpcRetryDelayFlag(dentryCommand.Cmd)
	config.AddRpcTimeoutFlag(dentryCommand.Cmd)
	config.AddFsMdsAddrFlag(dentryCommand.Cmd)
	config.AddFsIdUint32OptionFlag(dentryCommand.Cmd)
	config.AddFsNameStringOptionFlag(dentryCommand.Cmd)
	config.AddInodeIdRequiredFlag(dentryCommand.Cmd)
}

func (dentryCommand *DentryCommand) Init(cmd *cobra.Command, args []string) error {
	header := []string{
		cobrautil.ROW_FS_ID, cobrautil.ROW_INODE_ID, cobrautil.ROW_NAME, cobrautil.ROW_PARENT, cobrautil.ROW_TYPE,
	}
	dentryCommand.SetHeader(header)

	return nil
}

func (dentryCommand *DentryCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&dentryCommand.FinalDingoCmd, dentryCommand)
}

func (dentryCommand *DentryCommand) RunCommand(cmd *cobra.Command, args []string) error {
	fsId, getError := common.GetFsId(cmd)
	if getError != nil {
		return getError
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

	inodeId := config.GetFlagUint64(cmd, config.DINGOFS_INODEID)
	entries, entErr := common.ListDentry(cmd, fsId, inodeId, epoch)
	if entErr != nil {
		return entErr
	}
	if len(entries) == 0 {
		fmt.Printf("no dentry in dingofs, fsid: %d, inodeid: %d\n", fsId, inodeId)
		return nil
	}

	rows := make([]map[string]string, 0)
	for _, entry := range entries {
		row := make(map[string]string)
		row[cobrautil.ROW_FS_ID] = fmt.Sprintf("%d", entry.GetFsId())
		row[cobrautil.ROW_INODE_ID] = fmt.Sprintf("%d", entry.GetIno())
		row[cobrautil.ROW_NAME] = entry.GetName()
		row[cobrautil.ROW_PARENT] = fmt.Sprintf("%d", entry.GetParent())
		row[cobrautil.ROW_TYPE] = entry.GetType().String()

		rows = append(rows, row)
	}

	list := cobrautil.ListMap2ListSortByKeys(rows, dentryCommand.Header, []string{})
	dentryCommand.TableNew.AppendBulk(list)
	dentryCommand.Result = entries
	dentryCommand.Error = cmderror.ErrSuccess()

	return nil
}

func (dentryCommand *DentryCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&dentryCommand.FinalDingoCmd)
}
