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
 * Created Date: 2022-06-11
 * Author: chengyi (Cyber-SiKu)
 */

package umount

import (
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	"github.com/spf13/cobra"
)

type UmountCommand struct {
	basecmd.MidDingoCmd
}

var _ basecmd.MidDingoCmdFunc = (*UmountCommand)(nil) // check interface

func (umountCmd *UmountCommand) AddSubCommands() {
	umountCmd.Cmd.AddCommand(
		NewFsCommand(),
	)
}

func NewUmountCommand() *cobra.Command {
	umountCmd := &UmountCommand{
		basecmd.MidDingoCmd{
			Use:   "umount",
			Short: "umount fs in the dingofs",
		},
	}
	return basecmd.NewMidDingoCli(&umountCmd.MidDingoCmd, umountCmd)
}
