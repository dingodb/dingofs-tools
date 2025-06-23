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

package fs

import (
	"fmt"

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	cmdCommon "github.com/dingodb/dingofs-tools/pkg/cli/command/common"

	"github.com/dustin/go-humanize"

	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	"github.com/spf13/cobra"
)

type FsUageCommand struct {
	basecmd.FinalDingoCmd
}

var _ basecmd.FinalDingoCmdFunc = (*FsUageCommand)(nil) // check interface

const (
	UsageExample = `
# get usage by fsid
$ dingo usage fs --fsid 1

# get usage by fsname
$ dingo usage fs --fsname dingofs1,

# get all usage
$ dingo usage fs`
)

func NewFsUsageCommand() *cobra.Command {
	fsUsageCmd := &FsUageCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:     "fs",
			Short:   "Get the usage of the file system",
			Example: UsageExample,
		},
	}
	basecmd.NewFinalDingoCli(&fsUsageCmd.FinalDingoCmd, fsUsageCmd)
	return fsUsageCmd.Cmd
}

func (fsUsageCmd *FsUageCommand) AddFlags() {
	config.AddRpcRetryTimesFlag(fsUsageCmd.Cmd)
	config.AddRpcTimeoutFlag(fsUsageCmd.Cmd)
	config.AddFsMdsAddrFlag(fsUsageCmd.Cmd)
	config.AddFsIdUint32OptionFlag(fsUsageCmd.Cmd)
	config.AddFsNameStringOptionFlag(fsUsageCmd.Cmd)
	config.AddHumanizeOptionFlag(fsUsageCmd.Cmd)
}

func (fsUsageCmd *FsUageCommand) Init(cmd *cobra.Command, args []string) error {
	header := []string{cobrautil.ROW_FS_ID, cobrautil.ROW_FS_NAME, cobrautil.ROW_USED, cobrautil.ROW_INODES_IUSED}
	fsUsageCmd.Header = header
	fsUsageCmd.SetHeader(header)

	return nil
}

func (fsUsageCmd *FsUageCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&fsUsageCmd.FinalDingoCmd, fsUsageCmd)
}

func (fsUsageCmd *FsUageCommand) RunCommand(cmd *cobra.Command, args []string) error {
	errGetFsUsage := cmderror.Success()
	defer func() { fsUsageCmd.Error = errGetFsUsage }()

	fsIds := make([]uint32, 0)
	fsNames := make([]string, 0)

	if !fsUsageCmd.Cmd.Flag(config.DINGOFS_FSID).Changed && !fsUsageCmd.Cmd.Flag(config.DINGOFS_FSNAME).Changed {
		// get all filesystem info
		fsInfos, err := cmdCommon.ListAllFsInfo(cmd)
		if err != nil {
			errGetFsUsage = cmderror.ErrGetFsUsage()
			errGetFsUsage.Format(err.Error())
			return nil
		}

		for _, fsInfo := range fsInfos {
			fsIds = append(fsIds, fsInfo.GetFsId())
			fsNames = append(fsNames, fsInfo.GetFsName())
		}

	} else {
		fsId, err := cmdCommon.GetFsId(fsUsageCmd.Cmd)
		if err != nil {
			errGetFsUsage = cmderror.ErrGetFsUsage()
			errGetFsUsage.Format(err.Error())
			return nil
		}
		fsName, err := cmdCommon.GetFsName(fsUsageCmd.Cmd)
		if err != nil {
			errGetFsUsage = cmderror.ErrGetFsUsage()
			errGetFsUsage.Format(err.Error())
			return nil
		}

		fsIds = append(fsIds, fsId)
		fsNames = append(fsNames, fsName)
	}

	if len(fsIds) == 0 {
		errGetFsUsage = cmderror.ErrGetFsUsage()
		errGetFsUsage.Format("no data found")
		return nil
	}

	var totalUsed int64 = 0
	var totalIUsed int64 = 0

	isHumanize := config.GetFlagBool(fsUsageCmd.Cmd, config.DINGOFS_HUMANIZE)
	rows := make([]map[string]string, 0)
	for idx, fsId := range fsIds {
		row := make(map[string]string)
		//get real used space
		realUsedBytes, realUsedInodes, err := cmdCommon.GetDirectorySizeAndInodes(fsUsageCmd.Cmd, fsId, config.ROOTINODEID, true)
		if err != nil {
			errGetFsUsage = cmderror.ErrGetFsUsage()
			errGetFsUsage.Format(err.Error())
			return nil
		}

		totalUsed += realUsedBytes
		totalIUsed += realUsedInodes

		row[cobrautil.ROW_FS_ID] = fmt.Sprintf("%d", fsId)
		row[cobrautil.ROW_FS_NAME] = fsNames[idx]
		if isHumanize {
			row[cobrautil.ROW_USED] = humanize.IBytes(uint64(realUsedBytes))
			row[cobrautil.ROW_INODES_IUSED] = humanize.Comma(int64(realUsedInodes))
		} else {
			row[cobrautil.ROW_USED] = fmt.Sprintf("%d", realUsedBytes)
			row[cobrautil.ROW_INODES_IUSED] = fmt.Sprintf("%d", realUsedInodes)
		}

		rows = append(rows, row)
	}

	// set footer
	if len(rows) > 1 {
		var footer []string
		if isHumanize {
			footer = []string{"Total ", "-", humanize.IBytes(uint64(totalUsed)), humanize.Comma(int64(totalIUsed))}

		} else {
			footer = []string{"Total ", "-", fmt.Sprintf("%d", totalUsed), fmt.Sprintf("%d", totalIUsed)}

		}
		fsUsageCmd.TableNew.SetFooter(footer)
	}

	list := cobrautil.ListMap2ListSortByKeys(rows, fsUsageCmd.Header, []string{cobrautil.ROW_FS_ID})
	fsUsageCmd.TableNew.AppendBulk(list)

	fsUsageCmd.Result = rows

	return nil
}

func (fsUsageCmd *FsUageCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&fsUsageCmd.FinalDingoCmd)
}
