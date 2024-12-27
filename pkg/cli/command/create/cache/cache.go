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
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

const (
	cacheExample = `$ dingo create cacheCluster`
)

type CreateCacheRpc struct {
	Info           *basecmd.Rpc
	Request        *topology.RegistMemcacheClusterRequest
	topologyClient topology.TopologyServiceClient
}

var _ basecmd.RpcFunc = (*CreateCacheRpc)(nil) // check interface

type CacheCommand struct {
	basecmd.FinalDingoCmd
	Rpc *CreateCacheRpc
}

var _ basecmd.FinalDingoCmdFunc = (*CacheCommand)(nil) // check interface

func (cRpc *CreateCacheRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	cRpc.topologyClient = topology.NewTopologyServiceClient(cc)
}

func (cRpc *CreateCacheRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := cRpc.topologyClient.RegistMemcacheCluster(ctx, cRpc.Request)
	output.ShowRpcData(cRpc.Request, response, cRpc.Info.RpcDataShow)
	return response, err
}

func NewCacheCommand() *cobra.Command {
	return NewCreateCacheCommand().Cmd
}

func NewCreateCacheCommand() *CacheCommand {
	cacheCmd := &CacheCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:     "cachecluster",
			Short:   "register a memcache cluster in the dingofs",
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
	config.AddFsServersRequiredFlag(cCmd.Cmd)
}

func (cCmd *CacheCommand) Init(cmd *cobra.Command, args []string) error {
	addrs, addrErr := config.GetFsMdsAddrSlice(cCmd.Cmd)
	if addrErr.TypeCode() != cmderror.CODE_SUCCESS {
		cCmd.Error = addrErr
		return fmt.Errorf(addrErr.Message)
	}
	cCmd.Rpc = &CreateCacheRpc{}
	cCmd.Rpc.Request = &topology.RegistMemcacheClusterRequest{}
	servers, addrErr := config.GetAddrSlice(cCmd.Cmd, config.DINGOFS_SERVERS)
	if addrErr.TypeCode() != cmderror.CODE_SUCCESS {
		return addrErr.ToError()
	}
	for _, server := range servers {
		ip, port, addrErr := cobrautil.Addr2IpPort(server)
		if addrErr.TypeCode() != cmderror.CODE_SUCCESS {
			return addrErr.ToError()
		}
		serverInfo := topology.MemcacheServerInfo{
			Ip:   &ip,
			Port: &port,
		}
		cCmd.Rpc.Request.Servers = append(cCmd.Rpc.Request.Servers, &serverInfo)
	}
	timeout := viper.GetDuration(config.VIPER_GLOBALE_RPCTIMEOUT)
	retrytimes := viper.GetInt32(config.VIPER_GLOBALE_RPCRETRYTIMES)
	cCmd.Rpc.Info = basecmd.NewRpc(addrs, timeout, retrytimes, "ResistMemcacheCluster")
	cCmd.Rpc.Info.RpcDataShow = config.GetFlagBool(cCmd.Cmd, "verbose")

	header := []string{cobrautil.ROW_ID, cobrautil.ROW_RESULT}
	cCmd.SetHeader(header)

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
	result := response.(*topology.RegistMemcacheClusterResponse)
	res, err := output.MarshalProtoJson(result)
	if err != nil {
		return err
	}
	cCmd.Result = res
	cCmd.Error = cmderror.ErrCreateCacheClusterRpc(result.GetStatusCode())

	row := make(map[string]string)
	if cCmd.Error.TypeCode() == cmderror.CODE_SUCCESS {
		row[cobrautil.ROW_ID] = fmt.Sprintf("%d", result.GetClusterId())
	} else {
		row[cobrautil.ROW_ID] = "nil"
	}
	row[cobrautil.ROW_RESULT] = cCmd.Error.Message
	list := cobrautil.Map2List(row, cCmd.Header)
	cCmd.TableNew.Append(list)
	cCmd.Cmd.SilenceErrors = true
	return nil
}

func (cCmd *CacheCommand) ResultPlainOutput() error {
	if cCmd.TableNew.NumLines() == 0 {
		fmt.Println("no fs in cluster")
	}
	return output.FinalCmdOutputPlain(&cCmd.FinalDingoCmd)
}
