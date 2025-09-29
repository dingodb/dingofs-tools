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

package rpc

import (
	"context"
	"fmt"
	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	"github.com/dingodb/dingofs-tools/pkg/base"
	"github.com/dingodb/dingofs-tools/pkg/common"
	"github.com/dingodb/dingofs-tools/pkg/config"
	pbmdserror "github.com/dingodb/dingofs-tools/proto/dingofs/proto/error"
	pbmds "github.com/dingodb/dingofs-tools/proto/dingofs/proto/mds"
	"github.com/spf13/cobra"
	"log"
	"path"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
)

// get mds list
func GetMDSList(cmd *cobra.Command) ([]*pbmds.MDS, error) {
	// new prc
	mdsRpc, err := CreateNewMdsRpc(cmd, "GetMDSList")
	if err != nil {
		return nil, err
	}
	getMDSRpc := &GetMDSRpc{
		Info:    mdsRpc,
		Request: &pbmds.GetMDSListRequest{},
	}

	// get rpc result
	response, errCmd := base.GetRpcResponse(getMDSRpc.Info, getMDSRpc)
	if errCmd.TypeCode() != cmderror.CODE_SUCCESS {
		return nil, fmt.Errorf(errCmd.Message)
	}
	result := response.(*pbmds.GetMDSListResponse)
	if mdsErr := result.GetError(); mdsErr.GetErrcode() != pbmdserror.Errno_OK {
		return nil, cmderror.MDSV2Error(mdsErr).ToError()
	}

	return result.GetMdses(), nil
}

// retrieve fsid from command-line parameters,if not set, get by GetFsInfo via fsname
func GetFsId(cmd *cobra.Command) (uint32, error) {
	fsId, fsName, fsErr := common.GetFsInfoFlagValue(cmd)
	if fsErr != nil {
		return 0, fsErr
	}
	// fsId is not set,need to get fsId by fsName (fsName -> fsId)
	if fsId == 0 {
		fsInfo, fsErr := GetFsInfo(cmd, 0, fsName)
		if fsErr != nil {
			return 0, fsErr
		}
		fsId = fsInfo.GetFsId()
		if fsId == 0 {
			return 0, fmt.Errorf("fsid is invalid")
		}
	}

	return fsId, nil
}

// retrieve fsid from command-line parameters,if not set, get by GetFsInfo via fsid
func GetFsName(cmd *cobra.Command) (string, error) {
	fsId, fsName, fsErr := common.GetFsInfoFlagValue(cmd)
	if fsErr != nil {
		return "", fsErr
	}
	if len(fsName) == 0 { // fsName is not set,need to get fsName by fsId (fsId->fsName)
		fsInfo, fsErr := GetFsInfo(cmd, fsId, "")
		if fsErr != nil {
			return "", fsErr
		}
		fsName = fsInfo.GetFsName()
		if len(fsName) == 0 {
			return "", fmt.Errorf("fsName is invalid")
		}
	}

	return fsName, nil
}

// list filesystem info
func ListFsInfo(cmd *cobra.Command) ([]*pbmds.FsInfo, error) {
	// new prc
	mdsRpc, err := CreateNewMdsRpc(cmd, "ListFsInfo")
	if err != nil {
		return nil, err
	}
	// set request info
	listFsRpc := &ListFsInfoRpc{Info: mdsRpc, Request: &pbmds.ListFsInfoRequest{}}
	// get rpc result
	response, errCmd := base.GetRpcResponse(listFsRpc.Info, listFsRpc)
	if errCmd.TypeCode() != cmderror.CODE_SUCCESS {
		return nil, fmt.Errorf(errCmd.Message)
	}
	result := response.(*pbmds.ListFsInfoResponse)
	if mdsErr := result.GetError(); mdsErr.GetErrcode() != pbmdserror.Errno_OK {
		return nil, cmderror.MDSV2Error(mdsErr).ToError()
	}

	fsInfos := result.GetFsInfos()
	// fill fs meta cache
	for _, fsInfo := range fsInfos {
		fsMetaCache.SetFsInfo(fsInfo)
	}

	return fsInfos, nil
}

// get fsinfo by fsid or fsname
func GetFsInfo(cmd *cobra.Command, fsId uint32, fsName string) (*pbmds.FsInfo, error) {
	// first read from cache
	fsInfo, ok := fsMetaCache.GetFsInfo(fsId)
	if ok {
		return fsInfo, nil
	}
	// new prc
	mdsRpc, err := CreateNewMdsRpc(cmd, "GetFsInfo")
	if err != nil {
		return nil, err
	}
	// set request info
	var getFsRpc *GetFsRpc
	if fsId > 0 {
		getFsRpc = &GetFsRpc{Info: mdsRpc, Request: &pbmds.GetFsInfoRequest{FsId: fsId}}
	} else {
		getFsRpc = &GetFsRpc{Info: mdsRpc, Request: &pbmds.GetFsInfoRequest{FsName: fsName}}
	}
	// get rpc result
	response, errCmd := base.GetRpcResponse(getFsRpc.Info, getFsRpc)
	if errCmd.TypeCode() != cmderror.CODE_SUCCESS {
		return nil, fmt.Errorf(errCmd.Message)
	}
	result := response.(*pbmds.GetFsInfoResponse)
	if mdsErr := result.GetError(); mdsErr.GetErrcode() != pbmdserror.Errno_OK {
		return nil, cmderror.MDSV2Error(mdsErr).ToError()
	}

	fsInfo = result.GetFsInfo()
	fsMetaCache.SetFsInfo(fsInfo)

	return fsInfo, nil
}

// GetDentry
func GetDentry(cmd *cobra.Command, fsId uint32, parentId uint64, name string, epoch uint64) (*pbmds.Dentry, error) {
	endpoint := GetEndPoint(parentId)
	if len(endpoint) == 0 {
		return nil, fmt.Errorf("endpoint is null")
	}
	// new prc
	mdsRpc := CreateNewMdsRpcWithEndPoint(cmd, endpoint, "GetDentry")
	// set request info
	getDentryRpc := &GetDentryRpc{
		Info: mdsRpc,
		Request: &pbmds.GetDentryRequest{
			Context: &pbmds.Context{Epoch: epoch},
			FsId:    fsId,
			Parent:  parentId,
			Name:    name,
		},
	}
	// get rpc result
	response, errCmd := base.GetRpcResponse(getDentryRpc.Info, getDentryRpc)
	if errCmd.TypeCode() != cmderror.CODE_SUCCESS {
		return nil, fmt.Errorf(errCmd.Message)
	}
	result := response.(*pbmds.GetDentryResponse)
	if mdsErr := result.GetError(); mdsErr.GetErrcode() != pbmdserror.Errno_OK {
		return nil, cmderror.MDSV2Error(mdsErr).ToError()
	}
	return result.GetDentry(), nil
}

func DeleteFile(cmd *cobra.Command, fsId uint32, parentId uint64, name string, epoch uint64) error {
	endpoint := GetEndPoint(parentId)
	if len(endpoint) == 0 {
		return fmt.Errorf("endpoint is null")
	}
	// new prc
	mdsRpc := CreateNewMdsRpcWithEndPoint(cmd, endpoint, "UnLink")
	// set request info
	unlinkFileRpc := &UnlinkFileRpc{
		Info: mdsRpc,
		Request: &pbmds.UnLinkRequest{
			Context: &pbmds.Context{Epoch: epoch},
			FsId:    fsId,
			Parent:  parentId,
			Name:    name,
		},
	}
	// get rpc result
	response, errCmd := base.GetRpcResponse(unlinkFileRpc.Info, unlinkFileRpc)
	if errCmd.TypeCode() != cmderror.CODE_SUCCESS {
		return fmt.Errorf(errCmd.Message)
	}
	result := response.(*pbmds.UnLinkResponse)
	if mdsErr := result.GetError(); mdsErr.GetErrcode() != pbmdserror.Errno_OK {
		return cmderror.MDSV2Error(mdsErr).ToError()
	}

	return nil
}

func DeleteDirectory(cmd *cobra.Command, fsId uint32, parentId uint64, name string, epoch uint64) error {
	endpoint := GetEndPoint(parentId)
	if len(endpoint) == 0 {
		return fmt.Errorf("endpoint is null")
	}
	// new prc
	mdsRpc := CreateNewMdsRpcWithEndPoint(cmd, endpoint, "Rmdir")
	// set request info
	rmDirRpc := &RmDirRpc{
		Info: mdsRpc,
		Request: &pbmds.RmDirRequest{
			Context: &pbmds.Context{Epoch: epoch},
			FsId:    fsId,
			Parent:  parentId,
			Name:    name,
		},
	}
	// get rpc result
	response, errCmd := base.GetRpcResponse(rmDirRpc.Info, rmDirRpc)
	if errCmd.TypeCode() != cmderror.CODE_SUCCESS {
		return fmt.Errorf(errCmd.Message)
	}
	result := response.(*pbmds.RmDirResponse)
	if mdsErr := result.GetError(); mdsErr.GetErrcode() != pbmdserror.Errno_OK {
		return cmderror.MDSV2Error(mdsErr).ToError()
	}

	return nil
}

// parse directory path -> inodeId
func GetDirPathInodeId(cmd *cobra.Command, fsId uint32, path string, epoch uint64) (uint64, error) {
	if path == "/" {
		return config.ROOTINODEID, nil
	}
	inodeId := config.ROOTINODEID

	for path != "" {
		names := strings.SplitN(path, "/", 2)
		if names[0] != "" {
			dentry, err := GetDentry(cmd, fsId, inodeId, names[0], epoch)
			if err != nil {
				return 0, err
			}
			if dentry.GetType() != pbmds.FileType_DIRECTORY {
				return 0, syscall.ENOTDIR
			}
			inodeId = dentry.GetIno()
		}
		if len(names) == 1 {
			break
		}
		path = names[1]
	}
	return inodeId, nil
}

// get inode
func GetInode(cmd *cobra.Command, fsId uint32, inodeId uint64, parent uint64, epoch uint64) (*pbmds.Inode, error) {
	var endpoint []string
	requestContext := &pbmds.Context{Epoch: epoch}

	if IsFile(inodeId) && parent > 0 { // file: get endpoint by parent
		endpoint = GetEndPoint(parent)
	} else {
		endpoint = GetEndPoint(inodeId) // directory: get endpoint by self inodeid
	}
	if len(endpoint) == 0 {
		return nil, fmt.Errorf("endpoint is null")
	}
	if IsFile(inodeId) && parent == 0 { // file but parent is not set, bypass cache
		requestContext.IsBypassCache = true
	}
	// new prc
	mdsRpc := CreateNewMdsRpcWithEndPoint(cmd, endpoint, "GetInode")

	// set request info
	getInodeRpc := &GetInodeRpc{
		Info: mdsRpc,
		Request: &pbmds.GetInodeRequest{
			Context: requestContext,
			FsId:    fsId,
			Ino:     inodeId,
		},
	}
	// get rpc result
	response, errCmd := base.GetRpcResponse(getInodeRpc.Info, getInodeRpc)
	if errCmd.TypeCode() != cmderror.CODE_SUCCESS {
		return nil, fmt.Errorf(errCmd.Message)
	}
	result := response.(*pbmds.GetInodeResponse)
	if mdsErr := result.GetError(); mdsErr.GetErrcode() != pbmdserror.Errno_OK {
		return nil, cmderror.MDSV2Error(mdsErr).ToError()
	}

	return result.GetInode(), nil
}

// list dentry
func ListDentry(cmd *cobra.Command, fsId uint32, inodeId uint64, epoch uint64) ([]*pbmds.Dentry, error) {
	endpoint := GetEndPoint(inodeId)
	if len(endpoint) == 0 {
		return nil, fmt.Errorf("endpoint is null")
	}
	// new prc
	mdsRpc := CreateNewMdsRpcWithEndPoint(cmd, endpoint, "ListDentry")
	// set request info
	listDentryRpc := &ListDentryRpc{
		Info: mdsRpc,
		Request: &pbmds.ListDentryRequest{
			Context: &pbmds.Context{Epoch: epoch},
			FsId:    fsId,
			Parent:  inodeId,
		},
	}
	// get rpc result
	response, errCmd := base.GetRpcResponse(listDentryRpc.Info, listDentryRpc)
	if errCmd.TypeCode() != cmderror.CODE_SUCCESS {
		return nil, fmt.Errorf(errCmd.Message)
	}
	result := response.(*pbmds.ListDentryResponse)
	if mdsErr := result.GetError(); mdsErr.GetErrcode() != pbmdserror.Errno_OK {
		return nil, cmderror.MDSV2Error(mdsErr).ToError()
	}

	return result.GetDentries(), nil
}

// get dir path
func GetInodePath(cmd *cobra.Command, fsId uint32, inodeId uint64, epoch uint64) (string, string, error) {
	reverse := func(s []string) {
		for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
			s[i], s[j] = s[j], s[i]
		}
	}
	if inodeId == config.ROOTINODEID {
		return "/", fmt.Sprintf("%d", config.ROOTINODEID), nil
	}
	var names []string
	var inodes []string
	for inodeId != config.ROOTINODEID {
		inode, inodeErr := GetInode(cmd, fsId, inodeId, 0, epoch)
		if inodeErr != nil {
			return "", "", inodeErr
		}
		//do list entry rpc
		parentIds := inode.GetParents()
		parentId := parentIds[0]
		entries, entryErr := ListDentry(cmd, fsId, parentId, epoch)
		if entryErr != nil {
			return "", "", entryErr
		}
		for _, e := range entries {
			if e.GetIno() == inodeId {
				names = append(names, e.GetName())
				inodes = append(inodes, fmt.Sprintf("%d", inodeId))
				break
			}
		}
		inodeId = parentId
	}
	if len(names) == 0 { //directory may be deleted
		return "", "", nil
	}
	names = append(names, "/")                                     // add root
	inodes = append(inodes, fmt.Sprintf("%d", config.ROOTINODEID)) // add root
	reverse(names)
	reverse(inodes)

	return path.Join(names...), path.Join(inodes...), nil
}

// get directory size and inodes by inode
func GetDirSummarySize(cmd *cobra.Command, fsId uint32, inode uint64, summary *common.Summary, concurrent chan struct{},
	ctx context.Context, cancel context.CancelFunc, isFsCheck bool, inodeMap *sync.Map, epoch uint64) error {
	var err error
	entries, entErr := ListDentry(cmd, fsId, inode, epoch)
	if entErr != nil {
		return entErr
	}
	var wg sync.WaitGroup
	var errCh = make(chan error, 1)
	for _, entry := range entries {
		if entry.GetType() == pbmds.FileType_FILE {
			inodeAttr, err := GetInode(cmd, fsId, entry.GetIno(), entry.GetParent(), epoch)
			if err != nil {
				return err
			}
			if isFsCheck && inodeAttr.GetNlink() >= 2 { //filesystem check, hardlink is ignored
				if _, ok := inodeMap.LoadOrStore(inodeAttr.GetIno(), struct{}{}); ok {
					continue
				}
			}
			atomic.AddUint64(&summary.Length, inodeAttr.GetLength())
		}
		atomic.AddUint64(&summary.Inodes, 1)
		if entry.GetType() != pbmds.FileType_DIRECTORY {
			continue
		}
		select {
		case err := <-errCh:
			cancel()
			return err
		case <-ctx.Done():
			return fmt.Errorf("cancel scan directory for other goroutine error")
		case concurrent <- struct{}{}:
			wg.Add(1)
			go func(e *pbmds.Dentry) {
				defer wg.Done()
				sumErr := GetDirSummarySize(cmd, fsId, e.GetIno(), summary, concurrent, ctx, cancel, isFsCheck, inodeMap, epoch)
				<-concurrent
				if sumErr != nil {
					select {
					case errCh <- sumErr:
					default:
					}
				}
			}(entry)
		default:
			if sumErr := GetDirSummarySize(cmd, fsId, entry.GetIno(), summary, concurrent, ctx, cancel, isFsCheck, inodeMap, epoch); sumErr != nil {
				return sumErr
			}
		}
	}
	wg.Wait()
	select {
	case err = <-errCh:
	default:
	}

	return err
}

// get directory size and inodes by path name
func GetDirectorySizeAndInodes(cmd *cobra.Command, fsId uint32, dirInode uint64, isFsCheck bool, epoch uint64, threads uint32) (int64, int64, error) {
	log.Printf("start to summary directory statistics, inode[%d]", dirInode)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	summary := &common.Summary{0, 0}
	concurrent := make(chan struct{}, threads)
	var inodeMap *sync.Map = &sync.Map{}

	sumErr := GetDirSummarySize(cmd, fsId, dirInode, summary, concurrent, ctx, cancel, isFsCheck, inodeMap, epoch)
	if sumErr != nil {
		return 0, 0, sumErr
	}

	log.Printf("end summary directory statistics, inode[%d],inodes[%d],size[%d]", dirInode, summary.Inodes, summary.Length)

	// add root inode
	atomic.AddUint64(&summary.Inodes, 1)
	return int64(summary.Length), int64(summary.Inodes), nil
}
