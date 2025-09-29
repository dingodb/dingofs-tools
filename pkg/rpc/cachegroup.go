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

	"github.com/dingodb/dingofs-tools/pkg/base"
	"github.com/dingodb/dingofs-tools/pkg/output"
	"github.com/dingodb/dingofs-tools/proto/dingofs/proto/mds"
	"google.golang.org/grpc"
)

type ListCacheGroupRpc struct {
	Info             *base.Rpc
	Request          *mds.ListGroupsRequest
	cacheGroupClient mds.MDSServiceClient
}

type ListCacheMemberRpc struct {
	Info             *base.Rpc
	Request          *mds.ListMembersRequest
	cacheGroupClient mds.MDSServiceClient
}

type ReWeightMemberRpc struct {
	Info             *base.Rpc
	Request          *mds.ReweightMemberRequest
	cacheGroupClient mds.MDSServiceClient
}

type LeaveCacheMemberRpc struct {
	Info             *base.Rpc
	Request          *mds.LeaveCacheGroupRequest
	cacheGroupClient mds.MDSServiceClient
}

type DeleteCacheMemberRpc struct {
	Info             *base.Rpc
	Request          *mds.DeleteMemberRequest
	cacheGroupClient mds.MDSServiceClient
}

type UnlockCacheMemberRpc struct {
	Info             *base.Rpc
	Request          *mds.UnLockMemberRequest
	cacheGroupClient mds.MDSServiceClient
}

var _ base.RpcFunc = (*ListCacheGroupRpc)(nil)    // check interface
var _ base.RpcFunc = (*ListCacheMemberRpc)(nil)   // check interface
var _ base.RpcFunc = (*ReWeightMemberRpc)(nil)    // check interface
var _ base.RpcFunc = (*LeaveCacheMemberRpc)(nil)  // check interface
var _ base.RpcFunc = (*DeleteCacheMemberRpc)(nil) // check interface
var _ base.RpcFunc = (*UnlockCacheMemberRpc)(nil) // check interface

func (listCacheGroup *ListCacheGroupRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	listCacheGroup.cacheGroupClient = mds.NewMDSServiceClient(cc)
}

func (listCacheGroup *ListCacheGroupRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := listCacheGroup.cacheGroupClient.ListGroups(ctx, listCacheGroup.Request)
	output.ShowRpcData(listCacheGroup.Request, response, listCacheGroup.Info.RpcDataShow)
	return response, err
}

func (listCacheMember *ListCacheMemberRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	listCacheMember.cacheGroupClient = mds.NewMDSServiceClient(cc)
}

func (listCacheMember *ListCacheMemberRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := listCacheMember.cacheGroupClient.ListMembers(ctx, listCacheMember.Request)
	output.ShowRpcData(listCacheMember.Request, response, listCacheMember.Info.RpcDataShow)
	return response, err
}

func (rewightMember *ReWeightMemberRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	rewightMember.cacheGroupClient = mds.NewMDSServiceClient(cc)
}

func (rewightMember *ReWeightMemberRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := rewightMember.cacheGroupClient.ReweightMember(ctx, rewightMember.Request)
	output.ShowRpcData(rewightMember.Request, response, rewightMember.Info.RpcDataShow)
	return response, err
}

func (leaveMember *LeaveCacheMemberRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	leaveMember.cacheGroupClient = mds.NewMDSServiceClient(cc)
}

func (leaveMember *LeaveCacheMemberRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := leaveMember.cacheGroupClient.LeaveCacheGroup(ctx, leaveMember.Request)
	output.ShowRpcData(leaveMember.Request, response, leaveMember.Info.RpcDataShow)
	return response, err
}

func (deleteMember *DeleteCacheMemberRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	deleteMember.cacheGroupClient = mds.NewMDSServiceClient(cc)
}

func (deleteMember *DeleteCacheMemberRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := deleteMember.cacheGroupClient.DeleteMember(ctx, deleteMember.Request)
	output.ShowRpcData(deleteMember.Request, response, deleteMember.Info.RpcDataShow)
	return response, err
}

func (unlockMember *UnlockCacheMemberRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	unlockMember.cacheGroupClient = mds.NewMDSServiceClient(cc)
}

func (unlockMember *UnlockCacheMemberRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := unlockMember.cacheGroupClient.UnlockMember(ctx, unlockMember.Request)
	output.ShowRpcData(unlockMember.Request, response, unlockMember.Info.RpcDataShow)
	return response, err
}
