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

package fuse

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	"github.com/spf13/cobra"
)

type FuseConnCommand struct {
	basecmd.FinalDingoCmd
}

var _ basecmd.FinalDingoCmdFunc = (*FuseConnCommand)(nil) // check interface

func NewFuseConnCommand() *cobra.Command {
	fuseConnCmd := &FuseConnCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:   "connection",
			Short: "show fuse connection info",
			Example: `$ dingo fuse connection --mountpoint /mnt/dingofs
this command run with root privileges`,
		},
	}
	basecmd.NewFinalDingoCli(&fuseConnCmd.FinalDingoCmd, fuseConnCmd)
	return fuseConnCmd.Cmd
}

func (fuseConnCmd *FuseConnCommand) AddFlags() {
	config.AddMountpointRequiredFlag(fuseConnCmd.Cmd)
}

func (fuseConnCmd *FuseConnCommand) Init(cmd *cobra.Command, args []string) error {
	header := []string{cobrautil.ROW_FUSE_CONNECTION, cobrautil.ROW_FUSE_WAITING}
	fuseConnCmd.SetHeader(header)
	fuseConnCmd.Header = header

	return nil
}

func (fuseConnCmd *FuseConnCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&fuseConnCmd.FinalDingoCmd, fuseConnCmd)
}

func (fuseConnCmd *FuseConnCommand) RunCommand(cmd *cobra.Command, args []string) error {
	mountPoint := config.GetFlagString(cmd, config.DINGOFS_MOUNTPOINT)
	var st syscall.Stat_t
	if err := syscall.Stat(mountPoint, &st); err != nil {
		return err
	}
	if st.Ino != 1 {
		return fmt.Errorf("path %s is invalid mountpoint", mountPoint)
	}
	dev := uint64(st.Dev)
	connectionPath := fmt.Sprintf("/sys/fs/fuse/connections/%d", GetFuseConnectionId(dev))
	content, err := os.ReadFile(connectionPath + "/waiting")
	if err != nil {
		return err
	}
	waiting := strings.ReplaceAll(string(content), "\n", "")
	//fill table
	rows := make([]map[string]string, 0)
	row := make(map[string]string)
	row[cobrautil.ROW_FUSE_CONNECTION] = connectionPath
	row[cobrautil.ROW_FUSE_WAITING] = waiting

	rows = append(rows, row)
	list := cobrautil.ListMap2ListSortByKeys(rows, fuseConnCmd.Header, []string{})
	fuseConnCmd.TableNew.AppendBulk(list)

	fuseConnCmd.Error = cmderror.ErrSuccess()

	return nil
}

func (fuseConnCmd *FuseConnCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&fuseConnCmd.FinalDingoCmd)
}

func GetFuseConnectionId(dev uint64) uint32 {
	minor := dev & 0xff
	minor |= (dev >> 12) & 0xffffff00
	return uint32(minor)
}
