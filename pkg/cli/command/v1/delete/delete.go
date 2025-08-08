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

package delete

import (
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/v1/delete/cachemember"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/v1/delete/fs"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/v1/delete/metaserver"
	"github.com/spf13/cobra"
)

type DeleteCommand struct {
	basecmd.MidDingoCmd
}

var _ basecmd.MidDingoCmdFunc = (*DeleteCommand)(nil) // check interface

func (deleteCmd *DeleteCommand) AddSubCommands() {
	deleteCmd.Cmd.AddCommand(
		fs.NewFsCommand(),
		metaserver.NewDeleteMetaServerCommand(),
		cachemember.NewDeleteCacheMemberCommand(),
	)
}

func NewDeleteCommand() *cobra.Command {
	deleteCmd := &DeleteCommand{
		basecmd.MidDingoCmd{
			Use:   "delete",
			Short: "delete resources in the dingofs",
		},
	}
	return basecmd.NewMidDingoCli(&deleteCmd.MidDingoCmd, deleteCmd)
}
