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
 * Created Date: 2022-06-30
 * Author: chengyi (Cyber-SiKu)
 */

package topology

import (
	"context"
	"fmt"

	"github.com/dingodb/dingofs-tools/pkg/base"
	"github.com/dingodb/dingofs-tools/pkg/output"

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	"github.com/dingodb/dingofs-tools/proto/dingofs/proto/topology"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc"
)

type Server struct {
	Name         string `json:"name"`
	InternalIp   string `json:"internalip"`
	InternalPort uint32 `json:"internalport"`
	ExternalIp   string `json:"externalip"`
	ExternalPort uint32 `json:"externalport"`
	ZoneName     string `json:"zone"`
	PoolName     string `json:"pool"`
}

type DeleteServerRpc struct {
	Info           *base.Rpc
	Request        *topology.DeleteServerRequest
	topologyClient topology.TopologyServiceClient
}

func (dsRpc *DeleteServerRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	dsRpc.topologyClient = topology.NewTopologyServiceClient(cc)
}

func (dsRpc *DeleteServerRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := dsRpc.topologyClient.DeleteServer(ctx, dsRpc.Request)
	output.ShowRpcData(dsRpc.Request, response, dsRpc.Info.RpcDataShow)
	return response, err
}

var _ base.RpcFunc = (*DeleteServerRpc)(nil) // check interface

type CreateServerRpc struct {
	Info           *base.Rpc
	Request        *topology.ServerRegistRequest
	topologyClient topology.TopologyServiceClient
}

func (csRpc *CreateServerRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	csRpc.topologyClient = topology.NewTopologyServiceClient(cc)
}

func (csRpc *CreateServerRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := csRpc.topologyClient.RegistServer(ctx, csRpc.Request)
	output.ShowRpcData(csRpc.Request, response, csRpc.Info.RpcDataShow)
	return response, err
}

var _ base.RpcFunc = (*CreateServerRpc)(nil) // check interface

type ListZoneServerRpc struct {
	Info           *base.Rpc
	Request        *topology.ListZoneServerRequest
	topologyClient topology.TopologyServiceClient
}

func (lzsRpc *ListZoneServerRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	lzsRpc.topologyClient = topology.NewTopologyServiceClient(cc)
}

func (lzsRpc *ListZoneServerRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := lzsRpc.topologyClient.ListZoneServer(ctx, lzsRpc.Request)
	output.ShowRpcData(lzsRpc.Request, response, lzsRpc.Info.RpcDataShow)
	return response, err
}

var _ base.RpcFunc = (*ListZoneServerRpc)(nil) // check interface

func (tCmd *TopologyCommand) listZoneServer(zoneId uint32) (*topology.ListZoneServerResponse, *cmderror.CmdError) {
	request := &topology.ListZoneServerRequest{
		ZoneID: &zoneId,
	}
	tCmd.listZoneServerRpc = &ListZoneServerRpc{
		Request: request,
	}
	tCmd.listZoneServerRpc.Info = base.NewRpc(tCmd.addrs, tCmd.timeout, tCmd.retryTimes, tCmd.retryDelay, tCmd.verbose, "ListPoolZone")
	result, err := base.GetRpcResponse(tCmd.listZoneServerRpc.Info, tCmd.listZoneServerRpc)
	if err.TypeCode() != cmderror.CODE_SUCCESS {
		return nil, err
	}
	response := result.(*topology.ListZoneServerResponse)
	return response, cmderror.ErrSuccess()
}

func (tCmd *TopologyCommand) scanServers() *cmderror.CmdError {
	// scan server
	for _, zone := range tCmd.clusterZonesInfo {
		response, err := tCmd.listZoneServer(zone.GetZoneID())
		if err.TypeCode() != cmderror.CODE_SUCCESS {
			return err
		}
		if response.GetStatusCode() != topology.TopoStatusCode_TOPO_OK {
			return cmderror.ErrListPool(response.GetStatusCode())
		}
		tCmd.clusterServersInfo = append(tCmd.clusterServersInfo, response.GetServerInfo()...)
	}
	// update delete server
	compare := func(server Server, serverInfo *topology.ServerInfo) bool {
		return server.Name == serverInfo.GetHostName() && server.ZoneName == serverInfo.GetZoneName() && server.PoolName == serverInfo.GetPoolName()
	}
	for _, serverInfo := range tCmd.clusterServersInfo {
		index := slices.IndexFunc(tCmd.topology.Servers, func(server Server) bool {
			return compare(server, serverInfo)
		})
		if index == -1 {
			id := serverInfo.GetServerID()
			request := &topology.DeleteServerRequest{
				ServerID: &id,
			}
			tCmd.deleteServer = append(tCmd.deleteServer, request)
			row := make(map[string]string)
			row[cobrautil.ROW_NAME] = serverInfo.GetHostName()
			row[cobrautil.ROW_TYPE] = cobrautil.TYPE_SERVER
			row[cobrautil.ROW_OPERATION] = cobrautil.ROW_VALUE_DEL
			row[cobrautil.ROW_PARENT] = serverInfo.GetZoneName()
			tCmd.rows = append(tCmd.rows, row)
			tCmd.TableNew.Append(cobrautil.Map2List(row, tCmd.Header))
		}
	}

	// update create server
	for _, server := range tCmd.topology.Servers {
		index := slices.IndexFunc(tCmd.clusterServersInfo, func(serverInfo *topology.ServerInfo) bool {
			return compare(server, serverInfo)
		})
		newServer := server
		if index == -1 {
			request := &topology.ServerRegistRequest{
				HostName:     &newServer.Name,
				InternalIp:   &newServer.InternalIp,
				InternalPort: &newServer.InternalPort,
				ExternalIp:   &newServer.ExternalIp,
				ExternalPort: &newServer.ExternalPort,
				ZoneName:     &newServer.ZoneName,
				PoolName:     &newServer.PoolName,
			}
			tCmd.createServer = append(tCmd.createServer, request)
			row := make(map[string]string)
			row[cobrautil.ROW_NAME] = newServer.Name
			row[cobrautil.ROW_TYPE] = cobrautil.TYPE_SERVER
			row[cobrautil.ROW_OPERATION] = cobrautil.ROW_VALUE_ADD
			row[cobrautil.ROW_PARENT] = newServer.ZoneName
			tCmd.rows = append(tCmd.rows, row)
			tCmd.TableNew.Append(cobrautil.Map2List(row, tCmd.Header))
		}
	}

	return cmderror.ErrSuccess()
}

func (tCmd *TopologyCommand) removeServers() *cmderror.CmdError {
	tCmd.deleteServerRpc = &DeleteServerRpc{}
	tCmd.deleteServerRpc.Info = base.NewRpc(tCmd.addrs, tCmd.timeout, tCmd.retryTimes, tCmd.retryDelay, tCmd.verbose, "DeleteServer")
	for _, delReuest := range tCmd.deleteServer {
		tCmd.deleteServerRpc.Request = delReuest
		result, err := base.GetRpcResponse(tCmd.deleteServerRpc.Info, tCmd.deleteServerRpc)
		if err.TypeCode() != cmderror.CODE_SUCCESS {
			return err
		}
		response := result.(*topology.DeleteServerResponse)
		if response.GetStatusCode() != topology.TopoStatusCode_TOPO_OK {
			return cmderror.ErrDeleteTopology(response.GetStatusCode(), cobrautil.TYPE_SERVER, fmt.Sprintf("%d", delReuest.GetServerID()))
		}
	}
	return cmderror.ErrSuccess()
}

func (tCmd *TopologyCommand) createServers() *cmderror.CmdError {
	tCmd.createServerRpc = &CreateServerRpc{}
	tCmd.createServerRpc.Info = base.NewRpc(tCmd.addrs, tCmd.timeout, tCmd.retryTimes, tCmd.retryDelay, tCmd.verbose, "RegisterServer")
	for _, crtReuest := range tCmd.createServer {
		tCmd.createServerRpc.Request = crtReuest
		result, err := base.GetRpcResponse(tCmd.createServerRpc.Info, tCmd.createServerRpc)
		if err.TypeCode() != cmderror.CODE_SUCCESS {
			return err
		}
		response := result.(*topology.ServerRegistResponse)
		if response.GetStatusCode() != topology.TopoStatusCode_TOPO_OK {
			return cmderror.ErrCreateTopology(response.GetStatusCode(), cobrautil.TYPE_SERVER, crtReuest.GetHostName())
		}
	}
	return cmderror.ErrSuccess()
}
