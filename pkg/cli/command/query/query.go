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
 * Created Date: 2022-06-15
 * Author: chengyi (Cyber-SiKu)
 */

package query

import (
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/query/copyset"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/query/dirtree"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/query/fs"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/query/inode"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/query/metaserver"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/query/partition"
	"github.com/spf13/cobra"
)

type QueryCommand struct {
	basecmd.MidDingoCmd
}

var _ basecmd.MidDingoCmdFunc = (*QueryCommand)(nil) // check interface

func (queryCmd *QueryCommand) AddSubCommands() {
	queryCmd.Cmd.AddCommand(
		fs.NewFsCommand(),
		metaserver.NewMetaserverCommand(),
		partition.NewPartitionCommand(),
		copyset.NewCopysetCommand(),
		inode.NewInodeCommand(),
		dirtree.NewDirTreeCommand(),
	)
}

func NewQueryCommand() *cobra.Command {
	queryCmd := &QueryCommand{
		basecmd.MidDingoCmd{
			Use:   "query",
			Short: "query resources in the dingofs",
		},
	}
	return basecmd.NewMidDingoCli(&queryCmd.MidDingoCmd, queryCmd)
}
