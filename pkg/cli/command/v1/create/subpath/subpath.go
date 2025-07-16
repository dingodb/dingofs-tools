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

package subpath

import (
	"fmt"
	"path/filepath"
	"time"

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	"github.com/dingodb/dingofs-tools/pkg/base"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	cmdCommon "github.com/dingodb/dingofs-tools/pkg/cli/command/v1/common"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	"github.com/dingodb/dingofs-tools/proto/dingofs/proto/metaserver"
	"github.com/spf13/cobra"
)

const (
	DirectoryLength = 4096
	Mode            = 16877 // os.ModeDir | 0755
)

type RequestInfo struct {
	rpcInfo       *base.Rpc
	poolId        uint32
	copysetId     uint32
	partitionId   uint32
	txId          uint64
	fsId          uint32
	inodeId       uint64
	parentInodeId uint64
}

type InodeParam struct {
	fsId     uint32
	parent   uint64
	length   uint64
	uid      uint32
	gid      uint32
	mode     uint32
	fileType metaserver.FsFileType
	rdev     uint64
}

type SubPathCommand struct {
	basecmd.FinalDingoCmd
	fsId          uint32
	parentInodeId uint64
	pathName      string
	uid           uint32
	gid           uint32
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
	fsId, fsErr := cmdCommon.GetFsId(cmd)
	if fsErr != nil {
		return fsErr
	}

	path = filepath.Clean(path)
	parentPathName := filepath.Dir(path)
	subPathName := filepath.Base(path)
	subPathCmd.uid = config.GetFlagUint32(cmd, config.DINGOFS_SUBPATH_UID)
	subPathCmd.gid = config.GetFlagUint32(cmd, config.DINGOFS_SUBPATH_GID)

	parentInodeId, inodeErr := cmdCommon.GetDirPathInodeId(subPathCmd.Cmd, fsId, parentPathName)
	if inodeErr != nil {
		return inodeErr
	}

	subPathCmd.fsId = fsId
	subPathCmd.parentInodeId = parentInodeId
	subPathCmd.pathName = subPathName

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

	// get request addr leader
	addrs, addrErr := cmdCommon.GetLeaderPeerAddr(subPathCmd.Cmd, subPathCmd.fsId, subPathCmd.parentInodeId)
	if addrErr != nil {
		errCreatePath = cmderror.ErrQueryCopyset()
		errCreatePath.Format(addrErr.Error())
		return nil
	}

	// create rpc info
	timeout := config.GetRpcTimeout(cmd)
	retryTimes := config.GetRpcRetryTimes(cmd)
	retryDelay := config.GetRpcRetryDelay(cmd)
	verbose := config.GetFlagBool(subPathCmd.Cmd, config.VERBOSE)
	rpcInfo := base.NewRpc(addrs, timeout, retryTimes, retryDelay, verbose, "")

	// get copyset info
	partitionInfo, partErr := cmdCommon.GetPartitionInfo(subPathCmd.Cmd, subPathCmd.fsId, subPathCmd.parentInodeId)
	if partErr != nil {
		errCreatePath = cmderror.ErrGetFsPartition()
		errCreatePath.Format(partErr.Error())
		return nil
	}
	poolId := partitionInfo.GetPoolId()
	copysetId := partitionInfo.GetCopysetId()
	partitionId := partitionInfo.GetPartitionId()
	txId := partitionInfo.GetTxId()

	// create common request
	request := &RequestInfo{
		rpcInfo:     rpcInfo,
		poolId:      poolId,
		copysetId:   copysetId,
		partitionId: partitionId,
		txId:        txId,
		fsId:        subPathCmd.fsId,
	}

	// create inode
	request.rpcInfo.RpcFuncName = "CreateInode"
	request.parentInodeId = subPathCmd.parentInodeId
	newInodeId, createInodeErr := subPathCmd.CreateInode(cmd, request)
	if createInodeErr != nil {
		errCreatePath = createInodeErr
		return nil
	}

	// create dentry
	request.rpcInfo.RpcFuncName = "CreateDentry"
	request.parentInodeId = subPathCmd.parentInodeId
	request.inodeId = newInodeId //new create inode
	createDentryErr := subPathCmd.CreateDentry(cmd, request)
	if createDentryErr != nil {
		errCreatePath = createDentryErr
		return nil
	}

	// update parent inode mtime,ctime
	request.rpcInfo.RpcFuncName = "UpdateInode"
	request.parentInodeId = subPathCmd.parentInodeId
	updateInodeErr := subPathCmd.UpdateInodeAttr(cmd, request)
	if updateInodeErr != nil {
		errCreatePath = updateInodeErr
		return nil
	}
	// TODO: maybe update fs or dir quota usage here

	return nil
}

func (subPathCmd *SubPathCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&subPathCmd.FinalDingoCmd)
}

func (subPathCmd *SubPathCommand) CheckPathIsExist(cmd *cobra.Command) bool {
	entries, entErr := cmdCommon.ListDentry(cmd, subPathCmd.fsId, subPathCmd.parentInodeId)
	if entErr != nil {
		return false
	}
	for _, entry := range entries {
		if entry.GetName() == subPathCmd.pathName {
			return true
		}
	}

	return false
}

func (subPathCmd *SubPathCommand) CreateInode(cmd *cobra.Command, request *RequestInfo) (uint64, *cmderror.CmdError) {
	inodeParam := InodeParam{
		fsId:     request.fsId,
		parent:   request.parentInodeId,
		length:   DirectoryLength,
		uid:      subPathCmd.uid,
		gid:      subPathCmd.gid,
		mode:     Mode,
		fileType: metaserver.FsFileType_TYPE_DIRECTORY,
		rdev:     0,
	}
	createInodeRpc := &cmdCommon.CreateInodeRpc{
		Info: request.rpcInfo,
		Request: &metaserver.CreateInodeRequest{
			PoolId:      &request.poolId,
			CopysetId:   &request.copysetId,
			PartitionId: &request.partitionId,
			FsId:        &inodeParam.fsId,
			Length:      &inodeParam.length,
			Uid:         &inodeParam.uid,
			Gid:         &inodeParam.gid,
			Mode:        &inodeParam.mode,
			Type:        &inodeParam.fileType,
			Parent:      &inodeParam.parent,
		},
	}

	// create inode rpc request
	createInodeResult, rpcErr := base.GetRpcResponse(createInodeRpc.Info, createInodeRpc)
	if rpcErr.TypeCode() != cmderror.CODE_SUCCESS {
		return 0, rpcErr
	}

	createInodeResponse := createInodeResult.(*metaserver.CreateInodeResponse)
	if statusCode := createInodeResponse.GetStatusCode(); statusCode != metaserver.MetaStatusCode_OK {
		return 0, cmderror.ErrMetaServerRequest(int(statusCode))
	}
	// Obtain the Inode attr that successfully created
	newInode := createInodeResponse.GetInode()

	return newInode.GetInodeId(), nil
}

func (subPathCmd *SubPathCommand) DeleteInode(cmd *cobra.Command, request *RequestInfo) *cmderror.CmdError {
	deleteInodeRpc := &cmdCommon.DeleteInodeRpc{
		Info: request.rpcInfo,
		Request: &metaserver.DeleteInodeRequest{
			PoolId:      &request.poolId,
			CopysetId:   &request.copysetId,
			PartitionId: &request.partitionId,
			FsId:        &request.fsId,
			InodeId:     &request.inodeId, // new created inodeId
		},
	}
	deleteInodeResult, rpcErr := base.GetRpcResponse(deleteInodeRpc.Info, deleteInodeRpc)
	if rpcErr.TypeCode() != cmderror.CODE_SUCCESS {
		return rpcErr
	}
	deleteInodeResponse := deleteInodeResult.(*metaserver.DeleteInodeResponse)
	if statusCode := deleteInodeResponse.GetStatusCode(); statusCode != metaserver.MetaStatusCode_OK {
		return cmderror.ErrMetaServerRequest(int(statusCode))
	}

	return nil
}

func (subPathCmd *SubPathCommand) CreateDentry(cmd *cobra.Command, request *RequestInfo) *cmderror.CmdError {
	fileType := metaserver.FsFileType_TYPE_DIRECTORY
	dentry := metaserver.Dentry{
		FsId:          &request.fsId,
		InodeId:       &request.inodeId, // new Inode id
		ParentInodeId: &request.parentInodeId,
		Name:          &subPathCmd.pathName,
		TxId:          &request.txId,
		Type:          &fileType,
	}
	createDentryRpc := &cmdCommon.CreateDentryRpc{
		Info: request.rpcInfo,
		Request: &metaserver.CreateDentryRequest{
			PoolId:      &request.poolId,
			CopysetId:   &request.copysetId,
			PartitionId: &request.partitionId,
			Dentry:      &dentry,
		},
	}

	// create dentry rpc request, delete inode if failed
	createDentryResult, rpcErr := base.GetRpcResponse(createDentryRpc.Info, createDentryRpc)
	if rpcErr.TypeCode() != cmderror.CODE_SUCCESS {
		return rpcErr
	}

	createDentryResponse := createDentryResult.(*metaserver.CreateDentryResponse)
	if statusCode := createDentryResponse.GetStatusCode(); statusCode != metaserver.MetaStatusCode_OK {
		createDentryErr := cmderror.ErrMetaServerRequest(int(statusCode))

		request.rpcInfo.RpcFuncName = "DeleteInode"
		deleteInodeErr := subPathCmd.DeleteInode(cmd, request)
		if deleteInodeErr != nil {
			//  multiple errors
			var vecErrs []*cmderror.CmdError
			vecErrs = append(vecErrs, createDentryErr, deleteInodeErr)
			allErr := cmderror.MergeCmdError(vecErrs)

			// return create dentry and delete inode error
			return allErr
		}

		// return create dentry error
		return createDentryErr
	}

	return nil
}

func (subPathCmd *SubPathCommand) UpdateInodeAttr(cmd *cobra.Command, request *RequestInfo) *cmderror.CmdError {
	now := time.Now()
	tv_sec := uint64(now.Unix())
	tv_nsec := uint32(now.Nanosecond())

	updateInodeRpc := &cmdCommon.UpdateInodeRpc{
		Info: request.rpcInfo,
		Request: &metaserver.UpdateInodeRequest{
			PoolId:      &request.poolId,
			CopysetId:   &request.copysetId,
			PartitionId: &request.partitionId,
			FsId:        &request.fsId,
			InodeId:     &request.parentInodeId,
			Ctime:       &tv_sec,
			CtimeNs:     &tv_nsec,
			Mtime:       &tv_sec,
			MtimeNs:     &tv_nsec,
		},
	}

	// create dentry rpc request
	updateInodeResult, rpcErr := base.GetRpcResponse(updateInodeRpc.Info, updateInodeRpc)
	if rpcErr.TypeCode() != cmderror.CODE_SUCCESS {
		return rpcErr
	}

	updateInodeResponse := updateInodeResult.(*metaserver.UpdateInodeResponse)
	if statusCode := updateInodeResponse.GetStatusCode(); statusCode != metaserver.MetaStatusCode_OK {
		return cmderror.ErrMetaServerRequest(int(statusCode))
	}

	return nil
}
