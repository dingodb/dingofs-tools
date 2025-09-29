package rpc

import (
	"context"
	"github.com/dingodb/dingofs-tools/pkg/base"
	"github.com/dingodb/dingofs-tools/pkg/output"
	pbmds "github.com/dingodb/dingofs-tools/proto/dingofs/proto/mds"
	"google.golang.org/grpc"
)

// rpc services
type GetMDSRpc struct {
	Info      *base.Rpc
	Request   *pbmds.GetMDSListRequest
	mdsClient pbmds.MDSServiceClient
}

type CreateFsRpc struct {
	Info      *base.Rpc
	Request   *pbmds.CreateFsRequest
	mdsClient pbmds.MDSServiceClient
}

type DeleteFsRpc struct {
	Info      *base.Rpc
	Request   *pbmds.DeleteFsRequest
	mdsClient pbmds.MDSServiceClient
}

type ListFsRpc struct {
	Info      *base.Rpc
	Request   *pbmds.ListFsInfoRequest
	mdsClient pbmds.MDSServiceClient
}

type GetFsRpc struct {
	Info      *base.Rpc
	Request   *pbmds.GetFsInfoRequest
	mdsClient pbmds.MDSServiceClient
}

type GetMdsRpc struct {
	Info      *base.Rpc
	Request   *pbmds.GetMDSListRequest
	mdsClient pbmds.MDSServiceClient
}

type SetFsQuotaRpc struct {
	Info      *base.Rpc
	Request   *pbmds.SetFsQuotaRequest
	mdsClient pbmds.MDSServiceClient
}

type GetFsQuotaRpc struct {
	Info      *base.Rpc
	Request   *pbmds.GetFsQuotaRequest
	mdsClient pbmds.MDSServiceClient
}

type GetInodeRpc struct {
	Info      *base.Rpc
	Request   *pbmds.GetInodeRequest
	mdsClient pbmds.MDSServiceClient
}

type MkDirRpc struct {
	Info      *base.Rpc
	Request   *pbmds.MkDirRequest
	mdsClient pbmds.MDSServiceClient
}

type GetDentryRpc struct {
	Info      *base.Rpc
	Request   *pbmds.GetDentryRequest
	mdsClient pbmds.MDSServiceClient
}

type ListDentryRpc struct {
	Info      *base.Rpc
	Request   *pbmds.ListDentryRequest
	mdsClient pbmds.MDSServiceClient
}

type GetFsStatsRpc struct {
	Info      *base.Rpc
	Request   *pbmds.GetFsStatsRequest
	mdsClient pbmds.MDSServiceClient
}

type UmountFsRpc struct {
	Info      *base.Rpc
	Request   *pbmds.UmountFsRequest
	mdsClient pbmds.MDSServiceClient
}

type SetDirQuotaRpc struct {
	Info      *base.Rpc
	Request   *pbmds.SetDirQuotaRequest
	mdsClient pbmds.MDSServiceClient
}

type GetDirQuotaRpc struct {
	Info      *base.Rpc
	Request   *pbmds.GetDirQuotaRequest
	mdsClient pbmds.MDSServiceClient
}

type ListDirQuotaRpc struct {
	Info      *base.Rpc
	Request   *pbmds.LoadDirQuotasRequest
	mdsClient pbmds.MDSServiceClient
}

type DeleteDirQuotaRpc struct {
	Info      *base.Rpc
	Request   *pbmds.DeleteDirQuotaRequest
	mdsClient pbmds.MDSServiceClient
}

type CheckDirQuotaRpc struct {
	Info      *base.Rpc
	Request   *pbmds.SetDirQuotaRequest
	mdsClient pbmds.MDSServiceClient
}

type ListFsInfoRpc struct {
	Info      *base.Rpc
	Request   *pbmds.ListFsInfoRequest
	mdsClient pbmds.MDSServiceClient
}

type UnlinkFileRpc struct {
	Info      *base.Rpc
	Request   *pbmds.UnLinkRequest
	mdsClient pbmds.MDSServiceClient
}

type RmDirRpc struct {
	Info      *base.Rpc
	Request   *pbmds.RmDirRequest
	mdsClient pbmds.MDSServiceClient
}

// check interface
var _ base.RpcFunc = (*GetMdsRpc)(nil)         // check interface
var _ base.RpcFunc = (*CreateFsRpc)(nil)       // check interface
var _ base.RpcFunc = (*DeleteFsRpc)(nil)       // check interface
var _ base.RpcFunc = (*ListFsRpc)(nil)         // check interface
var _ base.RpcFunc = (*GetFsRpc)(nil)          // check interface
var _ base.RpcFunc = (*GetMdsRpc)(nil)         // check interface
var _ base.RpcFunc = (*SetFsQuotaRpc)(nil)     // check interface
var _ base.RpcFunc = (*GetFsQuotaRpc)(nil)     // check interface
var _ base.RpcFunc = (*GetInodeRpc)(nil)       // check interface
var _ base.RpcFunc = (*MkDirRpc)(nil)          // check interface
var _ base.RpcFunc = (*GetDentryRpc)(nil)      // check interface
var _ base.RpcFunc = (*ListDentryRpc)(nil)     // check interface
var _ base.RpcFunc = (*GetFsStatsRpc)(nil)     // check interface
var _ base.RpcFunc = (*UmountFsRpc)(nil)       // check interface
var _ base.RpcFunc = (*SetDirQuotaRpc)(nil)    // check interface
var _ base.RpcFunc = (*GetDirQuotaRpc)(nil)    // check interface
var _ base.RpcFunc = (*ListDirQuotaRpc)(nil)   // check interface
var _ base.RpcFunc = (*DeleteDirQuotaRpc)(nil) // check interface
var _ base.RpcFunc = (*CheckDirQuotaRpc)(nil)  // check interface
var _ base.RpcFunc = (*CheckDirQuotaRpc)(nil)  // check interface
var _ base.RpcFunc = (*UnlinkFileRpc)(nil)     // check interface
var _ base.RpcFunc = (*RmDirRpc)(nil)          // check interface

func (mdsFs *GetMDSRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	mdsFs.mdsClient = pbmds.NewMDSServiceClient(cc)
}

func (mdsFs *GetMDSRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := mdsFs.mdsClient.GetMDSList(ctx, mdsFs.Request)
	output.ShowRpcData(mdsFs.Request, response, mdsFs.Info.RpcDataShow)
	return response, err
}

func (createFs *CreateFsRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	createFs.mdsClient = pbmds.NewMDSServiceClient(cc)
}

func (createFs *CreateFsRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := createFs.mdsClient.CreateFs(ctx, createFs.Request)
	output.ShowRpcData(createFs.Request, response, createFs.Info.RpcDataShow)
	return response, err
}

func (deleteFs *DeleteFsRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	deleteFs.mdsClient = pbmds.NewMDSServiceClient(cc)
}

func (deleteFs *DeleteFsRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := deleteFs.mdsClient.DeleteFs(ctx, deleteFs.Request)
	output.ShowRpcData(deleteFs.Request, response, deleteFs.Info.RpcDataShow)
	return response, err
}

func (listFs *ListFsRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	listFs.mdsClient = pbmds.NewMDSServiceClient(cc)
}

func (listFs *ListFsRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := listFs.mdsClient.ListFsInfo(ctx, listFs.Request)
	output.ShowRpcData(listFs.Request, response, listFs.Info.RpcDataShow)
	return response, err
}

func (getFs *GetFsRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	getFs.mdsClient = pbmds.NewMDSServiceClient(cc)
}

func (getFs *GetFsRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := getFs.mdsClient.GetFsInfo(ctx, getFs.Request)
	output.ShowRpcData(getFs.Request, response, getFs.Info.RpcDataShow)
	return response, err
}

func (getMds *GetMdsRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	getMds.mdsClient = pbmds.NewMDSServiceClient(cc)
}

func (getMds *GetMdsRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := getMds.mdsClient.GetMDSList(ctx, getMds.Request)
	output.ShowRpcData(getMds.Request, response, getMds.Info.RpcDataShow)
	return response, err
}

func (setFsQuota *SetFsQuotaRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	setFsQuota.mdsClient = pbmds.NewMDSServiceClient(cc)
}

func (setFsQuota *SetFsQuotaRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := setFsQuota.mdsClient.SetFsQuota(ctx, setFsQuota.Request)
	output.ShowRpcData(setFsQuota.Request, response, setFsQuota.Info.RpcDataShow)
	return response, err
}

func (getFsQuota *GetFsQuotaRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	getFsQuota.mdsClient = pbmds.NewMDSServiceClient(cc)
}

func (getFsQuota *GetFsQuotaRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := getFsQuota.mdsClient.GetFsQuota(ctx, getFsQuota.Request)
	output.ShowRpcData(getFsQuota.Request, response, getFsQuota.Info.RpcDataShow)
	return response, err
}

func (getInode *GetInodeRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	getInode.mdsClient = pbmds.NewMDSServiceClient(cc)
}

func (getInode *GetInodeRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := getInode.mdsClient.GetInode(ctx, getInode.Request)
	output.ShowRpcData(getInode.Request, response, getInode.Info.RpcDataShow)
	return response, err
}

func (mkDir *MkDirRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	mkDir.mdsClient = pbmds.NewMDSServiceClient(cc)
}

func (mkDir *MkDirRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := mkDir.mdsClient.MkDir(ctx, mkDir.Request)
	output.ShowRpcData(mkDir.Request, response, mkDir.Info.RpcDataShow)
	return response, err
}

func (listDentry *ListDentryRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	listDentry.mdsClient = pbmds.NewMDSServiceClient(cc)
}

func (listDentry *ListDentryRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := listDentry.mdsClient.ListDentry(ctx, listDentry.Request)
	output.ShowRpcData(listDentry.Request, response, listDentry.Info.RpcDataShow)
	return response, err
}

func (getDentry *GetDentryRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	getDentry.mdsClient = pbmds.NewMDSServiceClient(cc)
}

func (getDentry *GetDentryRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := getDentry.mdsClient.GetDentry(ctx, getDentry.Request)
	output.ShowRpcData(getDentry.Request, response, getDentry.Info.RpcDataShow)
	return response, err
}

func (getFsStats *GetFsStatsRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	getFsStats.mdsClient = pbmds.NewMDSServiceClient(cc)
}

func (getFsStats *GetFsStatsRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := getFsStats.mdsClient.GetFsStats(ctx, getFsStats.Request)
	output.ShowRpcData(getFsStats.Request, response, getFsStats.Info.RpcDataShow)
	return response, err
}

func (umountFs *UmountFsRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	umountFs.mdsClient = pbmds.NewMDSServiceClient(cc)
}

func (umountFs *UmountFsRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := umountFs.mdsClient.UmountFs(ctx, umountFs.Request)
	output.ShowRpcData(umountFs.Request, response, umountFs.Info.RpcDataShow)
	return response, err
}

func (setDirQuota *SetDirQuotaRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	setDirQuota.mdsClient = pbmds.NewMDSServiceClient(cc)
}

func (setDirQuota *SetDirQuotaRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := setDirQuota.mdsClient.SetDirQuota(ctx, setDirQuota.Request)
	output.ShowRpcData(setDirQuota.Request, response, setDirQuota.Info.RpcDataShow)
	return response, err
}

func (getDirQuota *GetDirQuotaRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	getDirQuota.mdsClient = pbmds.NewMDSServiceClient(cc)
}

func (getDirQuota *GetDirQuotaRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := getDirQuota.mdsClient.GetDirQuota(ctx, getDirQuota.Request)
	output.ShowRpcData(getDirQuota.Request, response, getDirQuota.Info.RpcDataShow)
	return response, err
}

func (listDirQuota *ListDirQuotaRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	listDirQuota.mdsClient = pbmds.NewMDSServiceClient(cc)
}

func (listDirQuota *ListDirQuotaRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := listDirQuota.mdsClient.LoadDirQuotas(ctx, listDirQuota.Request)
	output.ShowRpcData(listDirQuota.Request, response, listDirQuota.Info.RpcDataShow)
	return response, err
}

func (deleteDirQuota *DeleteDirQuotaRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	deleteDirQuota.mdsClient = pbmds.NewMDSServiceClient(cc)
}

func (deleteDirQuota *DeleteDirQuotaRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := deleteDirQuota.mdsClient.DeleteDirQuota(ctx, deleteDirQuota.Request)
	output.ShowRpcData(deleteDirQuota.Request, response, deleteDirQuota.Info.RpcDataShow)
	return response, err
}

func (checkDirQuota *CheckDirQuotaRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	checkDirQuota.mdsClient = pbmds.NewMDSServiceClient(cc)
}

func (checkDirQuota *CheckDirQuotaRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := checkDirQuota.mdsClient.SetDirQuota(ctx, checkDirQuota.Request)
	output.ShowRpcData(checkDirQuota.Request, response, checkDirQuota.Info.RpcDataShow)
	return response, err
}

func (listFsInfo *ListFsInfoRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	listFsInfo.mdsClient = pbmds.NewMDSServiceClient(cc)
}

func (listFsInfo *ListFsInfoRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := listFsInfo.mdsClient.ListFsInfo(ctx, listFsInfo.Request)
	output.ShowRpcData(listFsInfo.Request, response, listFsInfo.Info.RpcDataShow)
	return response, err
}

func (unlinkFile *UnlinkFileRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	unlinkFile.mdsClient = pbmds.NewMDSServiceClient(cc)
}

func (unlinkFile *UnlinkFileRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := unlinkFile.mdsClient.UnLink(ctx, unlinkFile.Request)
	output.ShowRpcData(unlinkFile.Request, response, unlinkFile.Info.RpcDataShow)
	return response, err
}

func (rmDir *RmDirRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	rmDir.mdsClient = pbmds.NewMDSServiceClient(cc)
}

func (rmDir *RmDirRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := rmDir.mdsClient.RmDir(ctx, rmDir.Request)
	output.ShowRpcData(rmDir.Request, response, rmDir.Info.RpcDataShow)
	return response, err
}
