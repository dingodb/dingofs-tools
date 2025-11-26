package export

import (
	"errors"
	"fmt"
	"github.com/dingodb/dingofs-tools/internal/logger"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/dingossh"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"os/exec"
	"os/user"
	"path"
	"strconv"
	"strings"
)

const (
	NFS_GANESHA_NAME      = "ganesha.nfsd"
	NFS_EXPORT_STORE_PATH = "/etc/ganesha/export.d"
)

func GetInodeId(shell *dingossh.Shell, options dingossh.ExecOptions, path string) (uint64, error) {
	shell.List(path)
	shell.ClearOption().AddOption("-di")
	execResult, execErr := shell.Execute(options)
	if execErr != nil {
		return 0, fmt.Errorf("get path %s failed, err: %v", path, execErr)
	}
	// execResult: "inode_id path"
	parts := strings.Fields(execResult)
	inodeId, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return 0, err
	}

	logger.GetLogger().Infof("export path: %s, inodeid: %d", path, inodeId)

	return inodeId, nil
}

func GetGaneshaPID(shell *dingossh.Shell, options dingossh.ExecOptions) (int, error) {
	shell.Pgrep(NFS_GANESHA_NAME)
	shell.ClearOption()
	//execResult: pid
	execResult, execErr := shell.Execute(options)
	if execErr != nil {
		return 0, fmt.Errorf("get nfs-ganesha pid: failed, err: %v", execErr)
	}

	pid, err := strconv.ParseInt(strings.ReplaceAll(execResult, "\n", ""), 10, 32)
	if err != nil {
		return 0, err
	}

	logger.GetLogger().Infof("nfs-ganesha process pid: %d", pid)

	return int(pid), nil
}

// generate config file name  /etc/ganesha/export.d/<inode-id>-<pathname>.conf,e.g.: 536932699-nfs01.conf
func GenerateFileName(inodeId uint64, exportPath string) string {
	configFileName := fmt.Sprintf("%d-%s.conf", inodeId, path.Base(exportPath))
	return fmt.Sprintf("%s/%s", NFS_EXPORT_STORE_PATH, configFileName)
}

func NotifyGaneshaReLoadConfig(shell *dingossh.Shell, options dingossh.ExecOptions, pid int) error {
	shell.ClearOption().AddOption("-SIGHUP")
	shell.Kill(pid)
	_, execErr := shell.Execute(options)
	if execErr != nil {
		return fmt.Errorf("send SIGHUP signal to nfs-ganesha failed, pid: %d, err: %v", pid, execErr)
	}

	return nil
}

// check export path is already exported
func CheckPathIsExported(shell *dingossh.Shell, options dingossh.ExecOptions, exportPath string, storePath string) bool {
	shell.ClearOption().AddOption("-rh")
	pattern := fmt.Sprintf("\"Path.*=.*%s\"", exportPath)
	shell.Grep(pattern, storePath)
	_, execErr := shell.Execute(options)
	if execErr == nil {
		return true
	} else {
		return false
	}
}

// each EXPORT must have a unique Export_Id
func GenetateExportId(shell *dingossh.Shell, options dingossh.ExecOptions, storePath string) (int, error) {

	shell.ClearOption().AddOption("-rh")
	shell.Grep("\"Export_Id.*=\"", storePath)
	// execResult:
	//Export_Id = 1;
	//Export_Id = 2;
	//Export_Id = 3;
	execResult, execErr := shell.Execute(options)
	if execErr != nil {
		var exitErr *exec.ExitError   // local exec error
		var remoteExit *ssh.ExitError // remote exec error
		if (errors.As(execErr, &exitErr) && exitErr.ExitCode() == 1) || errors.As(execErr, &remoteExit) && remoteExit.ExitStatus() == 1 {
			logger.GetLogger().Info("not find export config file, export id start from 1")
			return 1, nil
		}
	}

	input := strings.TrimSpace(execResult)
	exportIds := strings.Split(input, ";")
	// no find export config file, begin from 1
	if len(exportIds) == 0 {
		return 1, nil
	}

	var maxExportId int = 0
	// pare key/value
	for _, pair := range exportIds {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			continue
		}

		value := strings.TrimSpace(kv[1])
		exportId, err := strconv.Atoi(value)
		if err != nil {
			return 0, err
		}
		if exportId > maxExportId {
			maxExportId = exportId
		}
	}

	return maxExportId + 1, nil
}

func GetSSHConfig(cmd *cobra.Command) (*dingossh.SSHConfig, error) {
	if cmd.Flag(config.DINGOFS_SSH_HOST).Changed { // remote
		currentUser, err := user.Current()
		if err != nil {
			return nil, err
		}
		sshUser := currentUser.Username
		if cmd.Flag(config.DINGOFS_SSH_USER).Changed {
			sshUser = config.GetFlagString(cmd, config.DINGOFS_SSH_USER)
		}
		sshKey := fmt.Sprintf("%s/.ssh/id_rsa", currentUser.HomeDir)
		if cmd.Flag(config.DINGOFS_SSH_KEY).Changed {
			sshKey = config.GetFlagString(cmd, config.DINGOFS_SSH_KEY)
		}

		sshHost := config.GetFlagString(cmd, config.DINGOFS_SSH_HOST)
		sshPort := config.GetFlagUint32(cmd, config.DINGOFS_SSH_PORT)

		sshConfig := &dingossh.SSHConfig{
			User:              sshUser,
			Host:              sshHost,
			Port:              uint(sshPort),
			PrivateKeyPath:    sshKey,
			ForwardAgent:      false,
			ConnectTimeoutSec: 10,
			ConnectRetries:    5,
		}

		return sshConfig, nil
	} else {
		return nil, nil
	}
}
