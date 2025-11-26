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
	"bytes"
	"fmt"
	"github.com/dingodb/dingofs-tools/internal/logger"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/dingossh"
	"github.com/dingodb/dingofs-tools/pkg/output"
	"github.com/spf13/cobra"
	"strings"
	"text/template"
)

const (
	NFS_EXPORT_TEMPLATE = `EXPORT
{
	Export_Id = {{.ExportID}};
	Path = {{.Path}};
	Pseudo = {{.Pseudo}};

	CLIENT {
		Clients = {{.Client}};
		Protocols = {{.Protocols}};
		Access_Type = {{.Access}};
		Squash = {{.Squash}};
		Sectype = {{.Sectype}};
	}

	FSAL {
		Name = {{.FSALName}};
	}
}`
)

type ExportConfig struct {
	ExportID  string
	Path      string
	Pseudo    string
	Protocols string
	Access    string
	Squash    string
	Sectype   string
	Client    string
	FSALName  string
}

type AddCommand struct {
	basecmd.FinalDingoCmd
	sshClient   *dingossh.SSHClient
	shell       *dingossh.Shell
	execOptions dingossh.ExecOptions
	exportPath  string
	exportConf  string
}

var _ basecmd.FinalDingoCmdFunc = (*AddCommand)(nil) // check interface

func NewAddCommand() *cobra.Command {
	addCmd := &AddCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:   "add",
			Short: "add nfs-ganesha export",
			Example: `
# local
$ dingo export add --nfs.path /mnt/dingofs/export --nfs.conf "*(Access_Type=RW)"
$ dingo export add --nfs.path /mnt/dingofs/export --nfs.conf "*(Access_Type=RW,Protocols=3:4,Squash=no_root_squash)"

# remote
$ dingo export add --nfs.path /mnt/dingofs/export --nfs.conf "*(Access_Type=RW)" --ssh.host 192.168.1.99
$ dingo export add --nfs.path /mnt/dingofs/export --nfs.conf "*(Access_Type=RW)" --ssh.host 192.168.1.99 --ssh.user root
`,
		},
	}
	basecmd.NewFinalDingoCli(&addCmd.FinalDingoCmd, addCmd)
	return addCmd.Cmd
}

func (addCmd *AddCommand) AddFlags() {
	config.AddNFSPathFlag(addCmd.Cmd)
	config.AddNFSConfFlag(addCmd.Cmd)
	config.AddSSHHostFlag(addCmd.Cmd)
	config.AddSSHPortFlag(addCmd.Cmd)
	config.AddSSHUserFlag(addCmd.Cmd)
	config.AddSSHKeyFlag(addCmd.Cmd)
}

func (addCmd *AddCommand) Init(cmd *cobra.Command, args []string) error {
	header := []string{cobrautil.ROW_RESULT}
	addCmd.SetHeader(header)
	addCmd.Header = header

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

	addCmd.sshClient = sshClient
	addCmd.shell = shell
	addCmd.execOptions = execOptions

	addCmd.exportPath = config.GetFlagString(cmd, config.DINGOFS_NFS_PATH)
	addCmd.exportConf = config.GetFlagString(cmd, config.DINGOFS_NFS_CONF)

	return nil
}

func (addCmd *AddCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&addCmd.FinalDingoCmd, addCmd)
}

func (addCmd *AddCommand) RunCommand(cmd *cobra.Command, args []string) error {

	//step 1: create directory for store export conf if not exists
	err := addCmd.GenerateExportStoragePath()
	if err != nil {
		return err
	}

	// step 2: check path if exported
	isExported := CheckPathIsExported(addCmd.shell, addCmd.execOptions, addCmd.exportPath, NFS_EXPORT_STORE_PATH)
	if isExported {
		return fmt.Errorf("path %s is already exported in %s", addCmd.exportPath, NFS_EXPORT_STORE_PATH)
	}

	//step 3: save export config file to NFS_EXPORT_STORE_PATH
	err = addCmd.SaveExportConfigFile(cmd)
	if err != nil {
		return err
	}

	//step 4: get current nfs-ganesha pid
	ganeshaPid, err := GetGaneshaPID(addCmd.shell, addCmd.execOptions)
	if err != nil {
		return err
	}

	// step 5: send SIGHUP signal to nfs-ganesha
	err = NotifyGaneshaReLoadConfig(addCmd.shell, addCmd.execOptions, ganeshaPid)
	if err != nil {
		return err
	}

	rows := make([]map[string]string, 0)
	row := make(map[string]string)
	row[cobrautil.ROW_RESULT] = "success"
	rows = append(rows, row)

	list := cobrautil.ListMap2ListSortByKeys(rows, addCmd.Header, []string{cobrautil.ROW_RESULT})
	addCmd.TableNew.AppendBulk(list)
	addCmd.Result = rows

	return nil
}

func (addCmd *AddCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&addCmd.FinalDingoCmd)
}

// input format :
// "*(Access_Type=RW,Protocols=3:4,Squash=no_root_squash)
// "192.168.1.1/24(Access_Type=RW,Protocols=3:4,Squash=no_root_squash)
func parseExportConfig(input string) (*ExportConfig, error) {
	exportConfig := &ExportConfig{
		Access:    "RW",
		Protocols: "3,4",
		Squash:    "no_root_squash",
		Sectype:   "sys",
		FSALName:  "VFS",
	}

	parts := strings.SplitN(input, "(", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid nfs config format: %s", input)
	}

	clientPart := strings.TrimSpace(parts[0])
	exportConfig.Client = clientPart

	// remove ")"
	configPart := strings.TrimSuffix(parts[1], ")")

	// pare key/value
	pairs := strings.Split(configPart, ",")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			continue
		}

		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		switch key {
		case "Access_Type":
			exportConfig.Access = value
		case "Protocols":
			//convert "3:4" to "3,4"
			exportConfig.Protocols = strings.ReplaceAll(value, ":", ",")
		case "Squash":
			exportConfig.Squash = value
		}
	}

	return exportConfig, nil
}

func (addCmd *AddCommand) GenerateExportStoragePath() error {
	addCmd.shell.ClearOption().AddOption("-d")
	addCmd.shell.Test(NFS_EXPORT_STORE_PATH)
	_, execErr := addCmd.shell.Execute(addCmd.execOptions)
	if execErr != nil {
		addCmd.shell.ClearOption()
		addCmd.shell.Mkdir(NFS_EXPORT_STORE_PATH)
		_, execErr = addCmd.shell.Execute(addCmd.execOptions)
		if execErr != nil {
			return fmt.Errorf("create nfs export directory %s failed, err: %v", NFS_EXPORT_STORE_PATH, execErr)
		}
		logger.GetLogger().Infof("create nfs export directory %s ok", NFS_EXPORT_STORE_PATH)
	} else {
		logger.GetLogger().Infof("nfs export directory %s already exists, ignore create", NFS_EXPORT_STORE_PATH)
	}

	return nil
}

func (addCmd *AddCommand) GenerateExportConfigFile(exportID string, exportPath string, exportCfg string) (string, error) {

	nfsConfig, err := parseExportConfig(exportCfg)
	if err != nil {
		return "", err
	}

	nfsConfig.ExportID = exportID
	nfsConfig.Path = exportPath
	nfsConfig.Pseudo = exportPath

	tmpl := template.Must(template.New("export").Parse(NFS_EXPORT_TEMPLATE))
	buffer := bytes.NewBufferString("")
	err = tmpl.Execute(buffer, nfsConfig)
	if err != nil {
		return "", fmt.Errorf("generate nfs export template failed, err: %v", err)
	}
	addCmd.Logger.Infof("export conf:\n%s\n", buffer.String())

	fileManager := dingossh.NewFileManager(nil)
	return fileManager.InstallTmpFile(buffer.String())
}

func (addCmd *AddCommand) SaveExportConfigFile(cmd *cobra.Command) error {
	inodeId, err := GetInodeId(addCmd.shell, addCmd.execOptions, addCmd.exportPath)
	if err != nil {
		return err
	}

	// generate export_id
	exportId, err := GenetateExportId(addCmd.shell, addCmd.execOptions, NFS_EXPORT_STORE_PATH)
	if err != nil {
		return err
	}
	tmpFileName, err := addCmd.GenerateExportConfigFile(fmt.Sprintf("%d", exportId), addCmd.exportPath, addCmd.exportConf)
	if err != nil {
		return err
	}
	newFileName := GenerateFileName(inodeId, addCmd.exportPath)

	if addCmd.execOptions.ExecInLocal {
		return addCmd.SaveToLocal(tmpFileName, newFileName)
	} else {
		return addCmd.SaveToRemote(tmpFileName, newFileName)
	}

}

func (addCmd *AddCommand) SaveToLocal(src string, dest string) error {
	addCmd.shell.ClearOption()
	addCmd.shell.Rename(src, dest)
	_, err := addCmd.shell.Execute(addCmd.execOptions)
	if err != nil {
		return fmt.Errorf("rename file %s to %s failed, err: %v", src, dest, err)
	}
	addCmd.Logger.Infof("rename file %s to %s ok", src, dest)

	return nil
}

func (addCmd *AddCommand) SaveToRemote(localFileName string, remoteFileName string) error {
	fileManager := dingossh.NewFileManager(addCmd.sshClient)
	// direct upload may be permission denied,e.g. fileManager.Upload(src, dest)
	// localFileName is in /tmp
	err := fileManager.Upload(localFileName, localFileName)
	if err != nil {
		return err
	}

	addCmd.shell.ClearOption()
	addCmd.shell.Rename(localFileName, remoteFileName)
	_, err = addCmd.shell.Execute(addCmd.execOptions)
	if err != nil {
		return fmt.Errorf("upload file %s to %s failed, err: %v", localFileName, remoteFileName, err)
	}
	addCmd.Logger.Infof("upload file %s to %s ok", localFileName, remoteFileName)

	return nil
}
