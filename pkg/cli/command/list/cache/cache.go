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
 * Created Date: 2022-10-21
 * Author: chengyi (Cyber-SiKu)
 */

package cache

import (
	"context"
	"fmt"

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	"github.com/dingodb/dingofs-tools/proto/dingofs/proto/topology"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

const (
	cacheExample = `$ dingo list cacheCluster`
)

type ListCacheRpc struct {
	Info           *basecmd.Rpc
	Request        *topology.ListMemcacheClusterRequest
	topologyClient topology.TopologyServiceClient
}

var _ basecmd.RpcFunc = (*ListCacheRpc)(nil) // check interface

type CacheCommand struct {
	basecmd.FinalDingoCmd
	Rpc *ListCacheRpc
}

var _ basecmd.FinalDingoCmdFunc = (*CacheCommand)(nil) // check interface

func (lRpc *ListCacheRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	lRpc.topologyClient = topology.NewTopologyServiceClient(cc)
}

func (lRpc *ListCacheRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := lRpc.topologyClient.ListMemcacheCluster(ctx, lRpc.Request)
	output.ShowRpcData(lRpc.Request, response, lRpc.Info.RpcDataShow)
	return response, err
}

func NewCacheCommand() *cobra.Command {
	return NewListCacheCommand().Cmd
}

func NewListCacheCommand() *CacheCommand {
	cacheCmd := &CacheCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:     "cachecluster",
			Short:   "list all memcache cluster in the dingofs",
			Example: cacheExample,
		},
	}

	basecmd.NewFinalDingoCli(&cacheCmd.FinalDingoCmd, cacheCmd)
	return cacheCmd
}

func (cCmd *CacheCommand) AddFlags() {
	config.AddRpcRetryTimesFlag(cCmd.Cmd)
	config.AddRpcTimeoutFlag(cCmd.Cmd)
	config.AddFsMdsAddrFlag(cCmd.Cmd)
}

func (cCmd *CacheCommand) Init(cmd *cobra.Command, args []string) error {
	addrs, addrErr := config.GetFsMdsAddrSlice(cCmd.Cmd)
	if addrErr.TypeCode() != cmderror.CODE_SUCCESS {
		cCmd.Error = addrErr
		return fmt.Errorf(addrErr.Message)
	}
	cCmd.Rpc = &ListCacheRpc{
		Request: &topology.ListMemcacheClusterRequest{},
	}
	timeout := config.GetRpcTimeout(cmd)
	retrytimes := config.GetRpcRetryTimes(cmd)
	cCmd.Rpc.Info = basecmd.NewRpc(addrs, timeout, retrytimes, "ListMemcacheCluster")
	cCmd.Rpc.Info.RpcDataShow = config.GetFlagBool(cCmd.Cmd, "verbose")

	header := []string{cobrautil.ROW_ID, cobrautil.ROW_SERVER}
	cCmd.SetHeader(header)
	cCmd.TableNew.SetAutoMergeCellsByColumnIndex(cobrautil.GetIndexSlice(
		cCmd.Header, []string{cobrautil.ROW_ID},
	))

	return nil
}

func (cCmd *CacheCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&cCmd.FinalDingoCmd, cCmd)
}

func (cCmd *CacheCommand) RunCommand(cmd *cobra.Command, args []string) error {
	response, errCmd := basecmd.GetRpcResponse(cCmd.Rpc.Info, cCmd.Rpc)
	if errCmd.TypeCode() != cmderror.CODE_SUCCESS {
		return fmt.Errorf(errCmd.Message)
	}
	result := response.(*topology.ListMemcacheClusterResponse)
	res, err := output.MarshalProtoJson(result)
	if err != nil {
		return err
	}
	cCmd.Result = res
	cCmd.Error = cmderror.ErrListMemcacheCluster(result.GetStatusCode())

	rows := make([]map[string]string, 0)
	for _, cluster := range result.GetMemcacheClusters() {
		for _, server := range cluster.GetServers() {
			row := make(map[string]string)
			row[cobrautil.ROW_ID] = fmt.Sprintf("%d", cluster.GetClusterId())
			row[cobrautil.ROW_SERVER] = fmt.Sprintf("%s:%d", server.GetIp(), server.GetPort())
			rows = append(rows, row)
		}
	}
	list := cobrautil.ListMap2ListSortByKeys(rows, cCmd.Header, []string{cobrautil.ROW_ID})
	cCmd.TableNew.AppendBulk(list)
	return nil
}

func (cCmd *CacheCommand) ResultPlainOutput() error {
	if cCmd.TableNew.NumLines() == 0 {
		fmt.Println("no memcache Cluster in the dingofs")
	}
	return output.FinalCmdOutputPlain(&cCmd.FinalDingoCmd)
}
