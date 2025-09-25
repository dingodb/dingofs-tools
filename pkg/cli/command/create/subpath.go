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

package create

import (
	"fmt"
	"github.com/dingodb/dingofs-tools/pkg/rpc"
	"path/filepath"

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	"github.com/dingodb/dingofs-tools/pkg/base"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	pbmdsv2error "github.com/dingodb/dingofs-tools/proto/dingofs/proto/error"
	"github.com/dingodb/dingofs-tools/proto/dingofs/proto/mdsv2"
	pbmdsv2 "github.com/dingodb/dingofs-tools/proto/dingofs/proto/mdsv2"
	"github.com/spf13/cobra"
)

const (
	DirectoryLength = 4096
	Mode            = 16877 // os.ModeDir | 0755
)

type InodeParam struct {
	fsId   uint32
	parent uint64
	length uint64
	uid    uint32
	gid    uint32
	mode   uint32
	rdev   uint64
	name   string
	epoch  uint64
}

type SubPathCommand struct {
	basecmd.FinalDingoCmd
	inodeParam InodeParam
}

var _ basecmd.FinalDingoCmdFunc = (*SubPathCommand)(nil) // check interface

func NewSubPathCommand() *cobra.Command {
	subPathCmd := &SubPathCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:   "subpath",
			Short: "create sub directory in dingofs",
			Example: `$ dingo create subpath --fsid 1 --path /path1
$ dingo create subpath --fsname dingofs --path /path1/path2
$ dingo create subpath --fsname dingofs --path /path1/path2 --uid 1000 -gid 1000`,
		},
	}
	basecmd.NewFinalDingoCli(&subPathCmd.FinalDingoCmd, subPathCmd)
	return subPathCmd.Cmd
}

func (subPathCmd *SubPathCommand) AddFlags() {
	config.AddRpcRetryTimesFlag(subPathCmd.Cmd)
	config.AddRpcRetryDelayFlag(subPathCmd.Cmd)
	config.AddRpcTimeoutFlag(subPathCmd.Cmd)
	config.AddFsMdsAddrFlag(subPathCmd.Cmd)
	config.AddFsIdUint32OptionFlag(subPathCmd.Cmd)
	config.AddFsNameStringOptionFlag(subPathCmd.Cmd)
	config.AddFsPathRequiredFlag(subPathCmd.Cmd)
	config.AddUidOptionalFlag(subPathCmd.Cmd)
	config.AddGidOptionalFlag(subPathCmd.Cmd)
}

func (subPathCmd *SubPathCommand) Init(cmd *cobra.Command, args []string) error {
	// get and process cmd value
	// TODO: new path instead of DINGOFS_QUOTA_PATH
	path := config.GetFlagString(subPathCmd.Cmd, config.DINGOFS_QUOTA_PATH)
	if len(path) == 0 {
		return fmt.Errorf("path is required")
	}
	fsId, fsErr := rpc.GetFsId(cmd)
	if fsErr != nil {
		return fsErr
	}

	path = filepath.Clean(path)
	parentPathName := filepath.Dir(path)
	subPathName := filepath.Base(path)
	uid := config.GetFlagUint32(cmd, config.DINGOFS_SUBPATH_UID)
	gid := config.GetFlagUint32(cmd, config.DINGOFS_SUBPATH_GID)

	// get epoch id
	epoch, epochErr := rpc.GetFsEpochByFsId(cmd, fsId)
	if epochErr != nil {
		return epochErr
	}
	// create router
	routerErr := rpc.InitFsMDSRouter(cmd, fsId)
	if routerErr != nil {
		return routerErr
	}

	// get parent
	parentInodeId, inodeErr := rpc.GetDirPathInodeId(subPathCmd.Cmd, fsId, parentPathName, epoch)
	if inodeErr != nil {
		return inodeErr
	}

	subPathCmd.inodeParam = InodeParam{
		fsId:   fsId,
		parent: parentInodeId,
		length: DirectoryLength,
		uid:    uid,
		gid:    gid,
		mode:   Mode,
		rdev:   0,
		name:   subPathName,
		epoch:  epoch,
	}

	header := []string{cobrautil.ROW_RESULT}
	subPathCmd.Header = header
	subPathCmd.SetHeader(header)

	return nil
}

func (subPathCmd *SubPathCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&subPathCmd.FinalDingoCmd, subPathCmd)
}

func (subPathCmd *SubPathCommand) RunCommand(cmd *cobra.Command, args []string) error {
	errCreatePath := cmderror.Success()
	defer func() {
		rows := make([]map[string]string, 0)
		row := make(map[string]string)
		row[cobrautil.ROW_RESULT] = errCreatePath.Message
		rows = append(rows, row)
		list := cobrautil.ListMap2ListSortByKeys(rows, subPathCmd.Header, []string{cobrautil.ROW_RESULT})
		subPathCmd.TableNew.AppendBulk(list)

		subPathCmd.Result = rows
		subPathCmd.Error = errCreatePath
	}()

	// check subpath is exist
	if subPathCmd.CheckPathIsExist(cmd) {
		return nil
	}

	errCreatePath = subPathCmd.MkDir(cmd, subPathCmd.inodeParam)

	// TODO: maybe update fs or dir quota usage here
	return nil
}

func (subPathCmd *SubPathCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&subPathCmd.FinalDingoCmd)
}

func (subPathCmd *SubPathCommand) CheckPathIsExist(cmd *cobra.Command) bool {
	entries, entErr := rpc.ListDentry(cmd, subPathCmd.inodeParam.fsId, subPathCmd.inodeParam.parent, subPathCmd.inodeParam.epoch)
	if entErr != nil {
		return false
	}
	for _, entry := range entries {
		if entry.GetName() == subPathCmd.inodeParam.name {
			return true
		}
	}

	return false
}

func (subPathCmd *SubPathCommand) MkDir(cmd *cobra.Command, inodeParam InodeParam) *cmderror.CmdError {
	// new prc request
	endpoint := rpc.GetEndPoint(inodeParam.parent)
	mdsRpc := rpc.CreateNewMdsRpcWithEndPoint(cmd, endpoint, "MkDir")

	mkDirRpc := &rpc.MkDirRpc{
		Info: mdsRpc,
		Request: &mdsv2.MkDirRequest{
			Context: &pbmdsv2.Context{Epoch: inodeParam.epoch},
			FsId:    inodeParam.fsId,
			Name:    inodeParam.name,
			Length:  inodeParam.length,
			Uid:     inodeParam.uid,
			Gid:     inodeParam.gid,
			Mode:    inodeParam.mode,
			Parent:  inodeParam.parent,
		},
	}

	// get rpc result
	response, errCmd := base.GetRpcResponse(mkDirRpc.Info, mkDirRpc)
	if errCmd.TypeCode() != cmderror.CODE_SUCCESS {
		return errCmd
	}
	result := response.(*pbmdsv2.MkDirResponse)
	if mdsErr := result.GetError(); mdsErr.GetErrcode() != pbmdsv2error.Errno_OK {
		return cmderror.MDSV2Error(mdsErr)
	}

	return cmderror.Success()
}
