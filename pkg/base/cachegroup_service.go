package base

import (
	"context"
	"github.com/dingodb/dingofs-tools/pkg/output"
	pbCacheGgroup "github.com/dingodb/dingofs-tools/proto/dingofs/proto/cachegroup"
	"google.golang.org/grpc"
)

type ListCacheGroupRpc struct {
	Info             *Rpc
	Request          *pbCacheGgroup.ListGroupsRequest
	cacheGroupClient pbCacheGgroup.CacheGroupMemberServiceClient
}

type ListCacheMemberRpc struct {
	Info             *Rpc
	Request          *pbCacheGgroup.ListMembersRequest
	cacheGroupClient pbCacheGgroup.CacheGroupMemberServiceClient
}

type ReWeightMemberRpc struct {
	Info             *Rpc
	Request          *pbCacheGgroup.ReweightMemberRequest
	cacheGroupClient pbCacheGgroup.CacheGroupMemberServiceClient
}

type LeaveCacheMemberRpc struct {
	Info             *Rpc
	Request          *pbCacheGgroup.LeaveCacheGroupRequest
	cacheGroupClient pbCacheGgroup.CacheGroupMemberServiceClient
}

type RegisterCacheMemberRpc struct {
	Info             *Rpc
	Request          *pbCacheGgroup.RegisterMemberRequest
	cacheGroupClient pbCacheGgroup.CacheGroupMemberServiceClient
}

type DeregisterCacheMemberRpc struct {
	Info             *Rpc
	Request          *pbCacheGgroup.DeregisterMemberRequest
	cacheGroupClient pbCacheGgroup.CacheGroupMemberServiceClient
}

var _ RpcFunc = (*ListCacheGroupRpc)(nil)      // check interface
var _ RpcFunc = (*ListCacheMemberRpc)(nil)     // check interface
var _ RpcFunc = (*ReWeightMemberRpc)(nil)      // check interface
var _ RpcFunc = (*LeaveCacheMemberRpc)(nil)    // check interface
var _ RpcFunc = (*RegisterCacheMemberRpc)(nil) // check interface

func (listCacheGroup *ListCacheGroupRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	listCacheGroup.cacheGroupClient = pbCacheGgroup.NewCacheGroupMemberServiceClient(cc)
}

func (listCacheGroup *ListCacheGroupRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := listCacheGroup.cacheGroupClient.ListGroups(ctx, listCacheGroup.Request)
	output.ShowRpcData(listCacheGroup.Request, response, listCacheGroup.Info.RpcDataShow)
	return response, err
}

func (listCacheMember *ListCacheMemberRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	listCacheMember.cacheGroupClient = pbCacheGgroup.NewCacheGroupMemberServiceClient(cc)
}

func (listCacheMember *ListCacheMemberRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := listCacheMember.cacheGroupClient.ListMembers(ctx, listCacheMember.Request)
	output.ShowRpcData(listCacheMember.Request, response, listCacheMember.Info.RpcDataShow)
	return response, err
}

func (rewightMember *ReWeightMemberRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	rewightMember.cacheGroupClient = pbCacheGgroup.NewCacheGroupMemberServiceClient(cc)
}

func (rewightMember *ReWeightMemberRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := rewightMember.cacheGroupClient.ReweightMember(ctx, rewightMember.Request)
	output.ShowRpcData(rewightMember.Request, response, rewightMember.Info.RpcDataShow)
	return response, err
}

func (leaveMember *LeaveCacheMemberRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	leaveMember.cacheGroupClient = pbCacheGgroup.NewCacheGroupMemberServiceClient(cc)
}

func (leaveMember *LeaveCacheMemberRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := leaveMember.cacheGroupClient.LeaveCacheGroup(ctx, leaveMember.Request)
	output.ShowRpcData(leaveMember.Request, response, leaveMember.Info.RpcDataShow)
	return response, err
}

func (registerMember *RegisterCacheMemberRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	registerMember.cacheGroupClient = pbCacheGgroup.NewCacheGroupMemberServiceClient(cc)
}

func (registerMember *RegisterCacheMemberRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := registerMember.cacheGroupClient.RegisterMember(ctx, registerMember.Request)
	output.ShowRpcData(registerMember.Request, response, registerMember.Info.RpcDataShow)
	return response, err
}

func (deregisterMember *DeregisterCacheMemberRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	deregisterMember.cacheGroupClient = pbCacheGgroup.NewCacheGroupMemberServiceClient(cc)
}

func (deregisterMember *DeregisterCacheMemberRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := deregisterMember.cacheGroupClient.DeregisterMember(ctx, deregisterMember.Request)
	output.ShowRpcData(deregisterMember.Request, response, deregisterMember.Info.RpcDataShow)
	return response, err
}
