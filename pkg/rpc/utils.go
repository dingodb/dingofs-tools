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
	"fmt"
	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	"github.com/dingodb/dingofs-tools/pkg/base"
	"github.com/dingodb/dingofs-tools/pkg/common"
	"github.com/dingodb/dingofs-tools/pkg/config"
	pbmds "github.com/dingodb/dingofs-tools/proto/dingofs/proto/mds"
	"github.com/spf13/cobra"
	"sync"
)

var (
	mdsRouter common.MDSRouter
	routerMtx sync.RWMutex
)

func IsDir(inodeId uint64) bool {
	return (inodeId & 1) == 1
}

func IsFile(inodeId uint64) bool {
	return (inodeId & 1) == 0
}

func GetFsEpochByFsInfo(fsInfo *pbmds.FsInfo) uint64 {
	partitionPolicy := fsInfo.GetPartitionPolicy()

	return partitionPolicy.GetEpoch()
}

func GetFsEpochByFsId(cmd *cobra.Command, fsId uint32) (uint64, error) {
	fsInfo, err := GetFsInfo(cmd, fsId, "")
	if err != nil {
		return 0, err
	}

	epoch := GetFsEpochByFsInfo(fsInfo)

	return epoch, nil
}

func InitFsMDSRouter(cmd *cobra.Command, fsId uint32) error {
	routerMtx.Lock()
	defer routerMtx.Unlock()

	fsInfo, err := GetFsInfo(cmd, fsId, "")
	if err != nil {
		return err
	}

	mds, err2 := GetMDSList(cmd)
	if err2 != nil {
		return err2
	}

	mdsRouter = common.NewMDSRouter(fsInfo.GetPartitionPolicy().GetType())
	mdsRouter.Init(mds, fsInfo.GetPartitionPolicy())

	return nil
}

func GetFsMDSRouter() common.MDSRouter {
	routerMtx.RLock()
	defer routerMtx.RUnlock()

	return mdsRouter
}

func GetEndPoint(inodeId uint64) (endpoints []string) {
	mdsMeta, ok := GetFsMDSRouter().GetMDS(inodeId)
	if ok {
		location := mdsMeta.GetLocation()
		endpoint := fmt.Sprintf("%s:%d", location.Host, location.Port)
		endpoints = append(endpoints, endpoint)
		return
	}

	return nil
}

func CreateNewMdsRpcWithEndPoint(cmd *cobra.Command, endpoint []string, serviceName string) *base.Rpc {
	// new rpc
	timeout := config.GetRpcTimeout(cmd)
	retryTimes := config.GetRpcRetryTimes(cmd)
	retryDelay := config.GetRpcRetryDelay(cmd)
	verbose := config.GetFlagBool(cmd, config.VERBOSE)
	mdsRpc := base.NewRpc(endpoint, timeout, retryTimes, retryDelay, verbose, serviceName)

	return mdsRpc
}

// create new mds rpc
func CreateNewMdsRpc(cmd *cobra.Command, serviceName string) (*base.Rpc, error) {
	// get mds address
	endpoints, addr := config.GetFsMdsAddrSlice(cmd)
	if addr.TypeCode() != cmderror.CODE_SUCCESS {
		return nil, fmt.Errorf(addr.Message)
	}

	mdsRpc := CreateNewMdsRpcWithEndPoint(cmd, endpoints, serviceName)

	return mdsRpc, nil
}
