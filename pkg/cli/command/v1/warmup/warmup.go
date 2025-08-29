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
 * Created Date: 2022-08-10
 * Author: chengyi (Cyber-SiKu)
 */

package warmup

import (
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/v1/warmup/add"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/v1/warmup/query"
	"github.com/spf13/cobra"
)

type WarmupCommand struct {
	basecmd.MidDingoCmd
}

var _ basecmd.MidDingoCmdFunc = (*WarmupCommand)(nil) // check interface

func (warmupCmd *WarmupCommand) AddSubCommands() {
	warmupCmd.Cmd.AddCommand(
		add.NewAddCommand(),
		query.NewQueryCommand(),
	)
}

func NewWarmupCommand() *cobra.Command {
	warmupCmd := &WarmupCommand{
		basecmd.MidDingoCmd{
			Use:   "warmup",
			Short: "add warmup file to local in dingofs",
		},
	}
	return basecmd.NewMidDingoCli(&warmupCmd.MidDingoCmd, warmupCmd)
}
