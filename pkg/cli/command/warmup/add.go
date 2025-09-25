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

package warmup

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	addExample = `
 # warmup all files in warmup.lst,file must in dingofs 
$ dingo warmup add --filelist /mnt/warmup.lst

 # warmup one file 
$ dingo warmup add /mnt/bigfile.bin

 # warmup all files in directory model 
$ dingo warmup add /mnt/model`
)

type AddCommand struct {
	basecmd.FinalDingoCmd
	Mountpoint *mountinfo.MountInfo
	Path       string // path in user system
	Single     bool
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
	fileListPath := config.GetFileListOptionFlag(aCmd.Cmd)
	if fileListPath == "" && len(args) == 0 {
		cmd.SilenceUsage = false
		return fmt.Errorf("no warmup file is specified")
	} else if fileListPath != "" {
		aCmd.Path = fileListPath
	} else {
		aCmd.Path = args[0]
		aCmd.Single = true
	}

	absPath, _ := filepath.Abs(aCmd.Path)
	cleanAbsPath := filepath.Clean(absPath)
	aCmd.Path = cleanAbsPath

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
		if strings.HasPrefix(cleanAbsPath, mountpoint.MountPoint) {
			aCmd.Mountpoint = mountpoint
			break
		}
	}
	if aCmd.Mountpoint == nil {
		return fmt.Errorf("[%s] is not saved in dingofs", aCmd.Path)
	}

	return nil
}

func (aCmd *AddCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&aCmd.FinalDingoCmd, aCmd)
}

func (aCmd *AddCommand) RunCommand(cmd *cobra.Command, args []string) error {
	var inodesStr string
	if aCmd.Single {
		inodeId, err := cobrautil.GetFileInode(aCmd.Path)
		if err != nil {
			return err
		}
		inodesStr = fmt.Sprintf("%d", inodeId)
	} else {
		inodes, err := cobrautil.GetInodesAsString(aCmd.Path)
		if err != nil {
			return err
		}
		inodesStr = inodes
	}

	err := unix.Setxattr(aCmd.Path, DINGOFS_WARMUP_OP_XATTR, []byte(inodesStr), 0)
	if err == unix.ENOTSUP || err == unix.EOPNOTSUPP {
		return fmt.Errorf("filesystem does not support extended attributes")
	} else if err != nil {
		setErr := cmderror.ErrSetxattr()
		setErr.Format(DINGOFS_WARMUP_OP_XATTR, err.Error())
		return setErr.ToError()
	}
	if !config.GetDaemonFlag(aCmd.Cmd) {
		time.Sleep(1 * time.Second) //wait for 1s
		GetWarmupProgress(aCmd.Cmd, aCmd.Path)
	} else {
		aCmd.TableNew.Append([]string{"success"})
	}

	return nil
}

func (aCmd *AddCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&aCmd.FinalDingoCmd)
}
