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
 * Created Date: 2022-06-06
 * Author: chengyi (Cyber-SiKu)
 */

package status

import (
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/v1/status/cluster"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/v1/status/copyset"
	etcd "github.com/dingodb/dingofs-tools/pkg/cli/command/v1/status/etcd"
	mds "github.com/dingodb/dingofs-tools/pkg/cli/command/v1/status/mds"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/v1/status/metaserver"
	"github.com/spf13/cobra"
)

type StatusCommand struct {
	basecmd.MidDingoCmd
}

var _ basecmd.MidDingoCmdFunc = (*StatusCommand)(nil) // check interface

func (statusCmd *StatusCommand) AddSubCommands() {
	statusCmd.Cmd.AddCommand(
		mds.NewMdsCommand(),
		metaserver.NewMetaserverCommand(),
		etcd.NewEtcdCommand(),
		copyset.NewCopysetCommand(),
		cluster.NewClusterCommand(),
	)
}

func NewStatusCommand() *cobra.Command {
	statusCmd := &StatusCommand{
		basecmd.MidDingoCmd{
			Use:   "status",
			Short: "get the status of dingofs",
		},
	}
	return basecmd.NewMidDingoCli(&statusCmd.MidDingoCmd, statusCmd)
}
