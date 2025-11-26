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

package export

import (
	"fmt"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/dingossh"
	"github.com/dingodb/dingofs-tools/pkg/output"
	"github.com/spf13/cobra"
)

type RemoveCommand struct {
	basecmd.FinalDingoCmd
	sshClient   *dingossh.SSHClient
	shell       *dingossh.Shell
	execOptions dingossh.ExecOptions
}

var _ basecmd.FinalDingoCmdFunc = (*RemoveCommand)(nil) // check interface

func NewRemoveCommand() *cobra.Command {
	removeCmd := &RemoveCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:   "remove",
			Short: "remove nfs-ganesha export",
			Example: `
# local
$ dingo export remove --nfs.path /mnt/dingofs/export

# remote
$ dingo export remove --nfs.path /mnt/dingofs/export --ssh.host 192.168.1.99
$ dingo export remove --nfs.path /mnt/dingofs/export --ssh.host 192.168.1.99 --ssh.user root
`,
		},
	}
	basecmd.NewFinalDingoCli(&removeCmd.FinalDingoCmd, removeCmd)
	return removeCmd.Cmd
}

func (removeCmd *RemoveCommand) AddFlags() {
	config.AddNFSPathFlag(removeCmd.Cmd)
	config.AddSSHHostFlag(removeCmd.Cmd)
	config.AddSSHPortFlag(removeCmd.Cmd)
	config.AddSSHUserFlag(removeCmd.Cmd)
	config.AddSSHKeyFlag(removeCmd.Cmd)
}

func (removeCmd *RemoveCommand) Init(cmd *cobra.Command, args []string) error {
	header := []string{cobrautil.ROW_RESULT}
	removeCmd.SetHeader(header)
	removeCmd.Header = header

	local := true
	var sshClient *dingossh.SSHClient

	sshConfig, err := GetSSHConfig(cmd)
	if err != nil {
		return err
	}
	if sshConfig != nil {
		sshClient, err = dingossh.NewSSHClient(*sshConfig)
		if err != nil {
			return err
		}
		local = false
	}

	shell := dingossh.NewShell(sshClient)
	execOptions := dingossh.ExecOptions{ExecWithSudo: true, ExecInLocal: local, ExecTimeoutSec: 10}

	removeCmd.sshClient = sshClient
	removeCmd.shell = shell
	removeCmd.execOptions = execOptions

	return nil
}

func (removeCmd *RemoveCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&removeCmd.FinalDingoCmd, removeCmd)
}

func (removeCmd *RemoveCommand) RunCommand(cmd *cobra.Command, args []string) error {

	exportPath := config.GetFlagString(cmd, config.DINGOFS_NFS_PATH)

	// step 1: get export path inodeid
	inodeId, err := GetInodeId(removeCmd.shell, removeCmd.execOptions, exportPath)
	if err != nil {
		return err
	}

	// step2: remove export config file
	configFileName := GenerateFileName(inodeId, exportPath)
	removeCmd.shell.Remove(configFileName)
	removeCmd.shell.ClearOption().AddOption("-f")
	_, err = removeCmd.shell.Execute(removeCmd.execOptions)
	if err != nil {
		return fmt.Errorf("remove export config %s failed, err: %v", configFileName, err)
	}
	removeCmd.Logger.Infof("remove export config %s ok", configFileName)

	// step 3: notify nfs-ganesha reload new config
	ganeshaPid, err := GetGaneshaPID(removeCmd.shell, removeCmd.execOptions)
	if err != nil {
		return err
	}
	err = NotifyGaneshaReLoadConfig(removeCmd.shell, removeCmd.execOptions, ganeshaPid)
	if err != nil {
		return err
	}

	rows := make([]map[string]string, 0)
	row := make(map[string]string)
	row[cobrautil.ROW_RESULT] = "success"
	rows = append(rows, row)

	list := cobrautil.ListMap2ListSortByKeys(rows, removeCmd.Header, []string{cobrautil.ROW_RESULT})
	removeCmd.TableNew.AppendBulk(list)
	removeCmd.Result = rows

	return nil
}

func (removeCmd *RemoveCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&removeCmd.FinalDingoCmd)
}
