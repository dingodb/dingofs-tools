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
 * Created Date: 2022-06-23
 * Author: chengyi (Cyber-SiKu)
 */

package check

import (
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/v1/check/chunk"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/v1/check/copyset"

	"github.com/spf13/cobra"
)

type CheckCommand struct {
	basecmd.MidDingoCmd
}

var _ basecmd.MidDingoCmdFunc = (*CheckCommand)(nil) // check interface

func (checkCmd *CheckCommand) AddSubCommands() {
	checkCmd.Cmd.AddCommand(
		copyset.NewCopysetCommand(),
		chunk.NewChunkCommand(),
	)
}

func NewCheckCommand() *cobra.Command {
	checkCmd := &CheckCommand{
		basecmd.MidDingoCmd{
			Use:   "check",
			Short: "checkout the health of resources in dingofs",
		},
	}
	return basecmd.NewMidDingoCli(&checkCmd.MidDingoCmd, checkCmd)
}
