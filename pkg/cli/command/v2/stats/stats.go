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

package stats

import (
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	"github.com/spf13/cobra"
)

type StatsCommand struct {
	basecmd.MidDingoCmd
}

var _ basecmd.MidDingoCmdFunc = (*StatsCommand)(nil) // check interface

func (statsCmd *StatsCommand) AddSubCommands() {
	statsCmd.Cmd.AddCommand(
		NewMountpointCommand(),
		NewClusterCommand(),
	)
}

func NewStatsCommand() *cobra.Command {
	statsCmd := &StatsCommand{
		basecmd.MidDingoCmd{
			Use:   "stats",
			Short: "monitor filesystem performance",
		},
	}
	return basecmd.NewMidDingoCli(&statsCmd.MidDingoCmd, statsCmd)
}
