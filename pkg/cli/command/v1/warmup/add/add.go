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

package add

import (
	"errors"
	"fmt"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/v1/warmup/query"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	mountinfo "github.com/cilium/cilium/pkg/mountinfo"
	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
)

const (
	addExample = `$ dingo warmup add --filelist /mnt/warmup/0809.list # warmup the file(dir) saved in /mnt/warmup/0809.list
$ dingo warmup add /mnt/warmup # warmup all files in /mnt/warmup`
)

const (
	DINGOFS_WARMUP_OP_XATTR      = "dingofs.warmup.op"
	DINGOFS_WARMUP_OP_ADD_SINGLE = "add\nsingle\n%s\n%s"
	DINGOFS_WARMUP_OP_ADD_LIST   = "add\nlist\n%s\n%s"
)

var STORAGE_TYPE = map[string]string{
	"disk": "disk",
	"mem":  "kvclient",
}

type AddCommand struct {
	basecmd.FinalDingoCmd
	Mountpoint   *mountinfo.MountInfo
	Path         string // path in user system
	DingofsPath  string // path in dingofs
	Single       bool   // warmup a single file or directory
	StorageType  string // warmup storage type
	ConvertFails []string
}

var _ basecmd.FinalDingoCmdFunc = (*AddCommand)(nil) // check interface

func NewAddWarmupCommand() *AddCommand {
	aCmd := &AddCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:     "add",
			Short:   "tell client to warmup files(directories) to local",
			Example: addExample,
		},
	}
	basecmd.NewFinalDingoCli(&aCmd.FinalDingoCmd, aCmd)
	return aCmd
}

func NewAddCommand() *cobra.Command {
	return NewAddWarmupCommand().Cmd
}

func (aCmd *AddCommand) AddFlags() {
	config.AddFileListOptionFlag(aCmd.Cmd)
	config.AddDaemonOptionPFlag(aCmd.Cmd)
	config.AddStorageOptionFlag(aCmd.Cmd)
}

func (aCmd *AddCommand) Init(cmd *cobra.Command, args []string) error {
	if config.GetDaemonFlag(aCmd.Cmd) {
		header := []string{cobrautil.ROW_RESULT}
		aCmd.SetHeader(header)
	}

	// check has dingofs mountpoint
	mountpoints, err := cobrautil.GetDingoFSMountPoints()
	if err.TypeCode() != cmderror.CODE_SUCCESS {
		aCmd.Error = err
		return err.ToError()
	} else if len(mountpoints) == 0 {
		return errors.New("no dingofs mountpoint found")
	}

	// check args
	aCmd.Single = false
	fileList := config.GetFileListOptionFlag(aCmd.Cmd)
	if fileList == "" && len(args) == 0 {
		cmd.SilenceUsage = false
		return fmt.Errorf("no --filelist or file(dir) specified")
	} else if fileList != "" {
		aCmd.Path = fileList
	} else {
		aCmd.Path = args[0]
		aCmd.Single = true
	}

	// check file is exist
	info, errStat := os.Stat(aCmd.Path)
	if errStat != nil {
		if os.IsNotExist(errStat) {
			return fmt.Errorf("[%s]: no such file or directory", aCmd.Path)
		} else {
			return fmt.Errorf("stat [%s] fail: %s", aCmd.Path, errStat.Error())
		}
	} else if !aCmd.Single && info.IsDir() {
		// --filelist must be a file
		return fmt.Errorf("[%s]: must be a file", aCmd.Path)
	}

	aCmd.Mountpoint = nil
	for _, mountpoint := range mountpoints {
		absPath, _ := filepath.Abs(aCmd.Path)
		rel, err := filepath.Rel(mountpoint.MountPoint, absPath)
		if err == nil && !strings.HasPrefix(rel, "..") {
			// found the mountpoint
			if aCmd.Mountpoint == nil ||
				len(aCmd.Mountpoint.MountPoint) < len(mountpoint.MountPoint) {
				// Prevent the dingofs directory from being mounted under the dingofs directory
				// /a/b/c:
				// test-1 mount in /a
				// test-1 mount in /a/b
				// warmup /a/b/c.
				aCmd.Mountpoint = mountpoint
				aCmd.DingofsPath = cobrautil.Path2DingofsPath(aCmd.Path, mountpoint)
			}
		}
	}
	if aCmd.Mountpoint == nil {
		return fmt.Errorf("[%s] is not saved in dingofs", aCmd.Path)
	}

	// check storage type
	aCmd.StorageType = STORAGE_TYPE[config.GetStorageFlag(aCmd.Cmd)]
	if aCmd.StorageType == "" {
		return fmt.Errorf("[%s] is not support storage type", aCmd.StorageType)
	}

	return nil
}

func (aCmd *AddCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&aCmd.FinalDingoCmd, aCmd)
}

func (aCmd *AddCommand) convertFilelist() *cmderror.CmdError {
	data, err := ioutil.ReadFile(aCmd.Path)
	if err != nil {
		readErr := cmderror.ErrReadFile()
		readErr.Format(aCmd.Path, err.Error())
		return readErr
	}
	if len(data) == 0 {
		readErr := cmderror.ErrReadFile()
		readErr.Format(aCmd.Path, "file is empty")
		return readErr
	}

	lines := strings.Split(string(data), "\n")
	validPath := ""
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		rel, err := filepath.Rel(aCmd.Mountpoint.MountPoint, line)
		if err == nil && !strings.HasPrefix(rel, "..") {
			// convert to dingofs path
			dingofsAbspath := cobrautil.Path2DingofsPath(line, aCmd.Mountpoint)
			validPath += (dingofsAbspath + "\n")
		} else {
			convertFail := fmt.Sprintf("[%s] is not saved in dingofs", line)
			aCmd.ConvertFails = append(aCmd.ConvertFails, convertFail)
		}
	}
	if validPath == "" {
		readErr := cmderror.ErrReadFile()
		readErr.Format(aCmd.Path, "not find valid path in filelist")
		return readErr
	}
	if err = ioutil.WriteFile(aCmd.Path, []byte(validPath), 0644); err != nil {
		writeErr := cmderror.ErrWriteFile()
		writeErr.Format(aCmd.Path, err.Error())
	}
	return cmderror.ErrSuccess()
}

func (aCmd *AddCommand) RunCommand(cmd *cobra.Command, args []string) error {
	xattr := DINGOFS_WARMUP_OP_ADD_SINGLE
	if !aCmd.Single {
		convertErr := aCmd.convertFilelist()
		if convertErr.TypeCode() != cmderror.CODE_SUCCESS {
			return convertErr.ToError()
		}
		xattr = DINGOFS_WARMUP_OP_ADD_LIST
	}
	value := fmt.Sprintf(xattr, aCmd.DingofsPath, aCmd.StorageType)
	err := unix.Setxattr(aCmd.Path, DINGOFS_WARMUP_OP_XATTR, []byte(value), 0)
	if err == unix.ENOTSUP || err == unix.EOPNOTSUPP {
		return fmt.Errorf("filesystem does not support extended attributes")
	} else if err != nil {
		setErr := cmderror.ErrSetxattr()
		setErr.Format(DINGOFS_WARMUP_OP_XATTR, err.Error())
		return setErr.ToError()
	}
	if !config.GetDaemonFlag(aCmd.Cmd) {
		query.GetWarmupProgress(aCmd.Cmd, aCmd.Path)
	} else {
		aCmd.TableNew.Append([]string{"success"})
	}
	return nil
}

func (aCmd *AddCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&aCmd.FinalDingoCmd)
}
