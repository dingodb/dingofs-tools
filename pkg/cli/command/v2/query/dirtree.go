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

package query

import (
	"fmt"

	"github.com/dingodb/dingofs-tools/pkg/cli/command/v2/common"

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	"github.com/spf13/cobra"
)

const (
	dirTreeExample = `$ dingo query dirtree --fsid 1 --inodeid 102`
)

type DirTreeCommand struct {
	basecmd.FinalDingoCmd
}

var _ basecmd.FinalDingoCmdFunc = (*DirTreeCommand)(nil) // check interface

func NewDirTreeCommand() *cobra.Command {
	dirTreeCmd := &DirTreeCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:     "dirtree",
			Short:   "show directory tree",
			Example: dirTreeExample,
		},
	}
	basecmd.NewFinalDingoCli(&dirTreeCmd.FinalDingoCmd, dirTreeCmd)
	return dirTreeCmd.Cmd
}

func (dirTreeCommand *DirTreeCommand) AddFlags() {
	config.AddRpcRetryTimesFlag(dirTreeCommand.Cmd)
	config.AddRpcTimeoutFlag(dirTreeCommand.Cmd)
	config.AddFsMdsAddrFlag(dirTreeCommand.Cmd)
	config.AddFsIdUint32OptionFlag(dirTreeCommand.Cmd)
	config.AddFsNameStringOptionFlag(dirTreeCommand.Cmd)
	config.AddInodeIdRequiredFlag(dirTreeCommand.Cmd)
}

func (dirTreeCommand *DirTreeCommand) Init(cmd *cobra.Command, args []string) error {

	return nil
}

func (dirTreeCommand *DirTreeCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&dirTreeCommand.FinalDingoCmd, dirTreeCommand)
}

func (dirTreeCommand *DirTreeCommand) RunCommand(cmd *cobra.Command, args []string) error {
	fsId, getError := common.GetFsId(cmd)
	if getError != nil {
		return getError
	}
	inodeId := config.GetFlagUint64(cmd, config.DINGOFS_INODEID)

	namePath, inodePath, err := common.GetInodePath(cmd, fsId, inodeId)
	if err != nil {
		return err
	}
	fmt.Printf("-- name  path:	%s\n", namePath)
	fmt.Printf("-- inode path:	%s\n", inodePath)

	dirTreeCommand.Error = cmderror.ErrSuccess()

	return nil
}

func (dirTreeCommand *DirTreeCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&dirTreeCommand.FinalDingoCmd)
}
