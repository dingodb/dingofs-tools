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
 * Created Date: 2022-05-30
 * Author: chengyi (Cyber-SiKu)
 */

package list

import (
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/v1/list/cachegroup"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/v1/list/cachemember"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/v1/list/copyset"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/v1/list/dentry"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/v1/list/fs"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/v1/list/mountpoint"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/v1/list/partition"

	topology "github.com/dingodb/dingofs-tools/pkg/cli/command/v1/list/topology"
	"github.com/spf13/cobra"
)

type ListCommand struct {
	basecmd.MidDingoCmd
}

var _ basecmd.MidDingoCmdFunc = (*ListCommand)(nil) // check interface

func (listCmd *ListCommand) AddSubCommands() {
	listCmd.Cmd.AddCommand(
		topology.NewTopologyCommand(),
		fs.NewFsCommand(),
		mountpoint.NewMountpointCommand(),
		partition.NewPartitionCommand(),
		copyset.NewCopysetCommand(),
		dentry.NewDentryCommand(),
		cachemember.NewCacheMemberCommand(),
		cachegroup.NewCacheGroupCommand(),
	)
}

func NewListCommand() *cobra.Command {
	listCmd := &ListCommand{
		basecmd.MidDingoCmd{
			Use:   "list",
			Short: "list resources in the dingofs",
		},
	}
	return basecmd.NewMidDingoCli(&listCmd.MidDingoCmd, listCmd)
}
