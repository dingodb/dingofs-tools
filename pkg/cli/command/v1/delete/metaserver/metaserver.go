// Copyright (c) 2024 dingodb.com, Inc. All Rights Reserved
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

package metaserver

import (
	"context"
	"fmt"

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	"github.com/dingodb/dingofs-tools/pkg/base"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	"github.com/dingodb/dingofs-tools/proto/dingofs/proto/topology"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

const (
	deleteExample = `$ dingo delete metaserver --metaserverid 1`
)

type DeleteMetaServerRpc struct {
	Info           *base.Rpc
	Request        *topology.DeleteMetaServerRequest
	topologyClient topology.TopologyServiceClient
}

var _ base.RpcFunc = (*DeleteMetaServerRpc)(nil) // check interface

type DeleteMetaServerCommand struct {
	basecmd.FinalDingoCmd
	Rpc *DeleteMetaServerRpc
}

var _ basecmd.FinalDingoCmdFunc = (*DeleteMetaServerCommand)(nil) // check interface

func (deleteMetaServerRpc *DeleteMetaServerRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	deleteMetaServerRpc.topologyClient = topology.NewTopologyServiceClient(cc)
}

func (deleteMetaServerRpc *DeleteMetaServerRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := deleteMetaServerRpc.topologyClient.DeleteMetaServer(ctx, deleteMetaServerRpc.Request)
	output.ShowRpcData(deleteMetaServerRpc.Request, response, deleteMetaServerRpc.Info.RpcDataShow)
	return response, err
}

func NewDeleteMetaServerCommand() *cobra.Command {
	metaServerCmd := &DeleteMetaServerCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:     "metaserver",
			Short:   "delete metaserver from topology",
			Example: deleteExample,
		},
	}
	basecmd.NewFinalDingoCli(&metaServerCmd.FinalDingoCmd, metaServerCmd)
	return metaServerCmd.Cmd
}

func (metaServerCmd *DeleteMetaServerCommand) AddFlags() {
	config.AddRpcRetryTimesFlag(metaServerCmd.Cmd)
	config.AddRpcRetryDelayFlag(metaServerCmd.Cmd)
	config.AddRpcTimeoutFlag(metaServerCmd.Cmd)
	config.AddFsMdsAddrFlag(metaServerCmd.Cmd)
	config.AddMetaserverIdFlag(metaServerCmd.Cmd)
	config.AddNoConfirmOptionFlag(metaServerCmd.Cmd)
}

func (metaServerCmd *DeleteMetaServerCommand) Init(cmd *cobra.Command, args []string) error {
	addrs, addrErr := config.GetFsMdsAddrSlice(metaServerCmd.Cmd)
	if addrErr.TypeCode() != cmderror.CODE_SUCCESS {
		metaServerCmd.Error = addrErr
		return fmt.Errorf(addrErr.Message)
	}

	header := []string{cobrautil.ROW_RESULT}
	metaServerCmd.SetHeader(header)

	metaServerID := config.GetFlagUint32(metaServerCmd.Cmd, config.DINGOFS_METASERVERID)
	request := &topology.DeleteMetaServerRequest{
		MetaServerID: &metaServerID,
	}
	metaServerCmd.Rpc = &DeleteMetaServerRpc{
		Request: request,
	}

	timeout := config.GetRpcTimeout(cmd)
	retrytimes := config.GetRpcRetryTimes(cmd)
	retryDelay := config.GetRpcRetryDelay(cmd)
	verbose := config.GetFlagBool(cmd, config.VERBOSE)
	metaServerCmd.Rpc.Info = base.NewRpc(addrs, timeout, retrytimes, retryDelay, verbose, "DeleteMetaServer")

	return nil
}

func (metaServerCmd *DeleteMetaServerCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&metaServerCmd.FinalDingoCmd, metaServerCmd)
}

func (metaServerCmd *DeleteMetaServerCommand) RunCommand(cmd *cobra.Command, args []string) error {
	metaServerIdStr := fmt.Sprintf("%d", metaServerCmd.Rpc.Request.GetMetaServerID())
	if !config.GetFlagBool(metaServerCmd.Cmd, config.DINGOFS_NOCONFIRM) && !cobrautil.AskConfirmation(fmt.Sprintf("Are you sure to delete metaserver %s?", metaServerIdStr), "yes") {
		return fmt.Errorf("abort delete metaserver")
	}

	result, err := base.GetRpcResponse(metaServerCmd.Rpc.Info, metaServerCmd.Rpc)
	if err.TypeCode() != cmderror.CODE_SUCCESS {
		return err.ToError()
	}
	response := result.(*topology.DeleteMetaServerResponse)
	if response.GetStatusCode() != topology.TopoStatusCode_TOPO_OK {
		err := cmderror.ErrDeleteTopology(response.GetStatusCode(), cobrautil.TYPE_METASERVER, metaServerIdStr).ToError()
		row := map[string]string{
			cobrautil.ROW_RESULT: err.Error(),
		}
		metaServerCmd.TableNew.Append(cobrautil.Map2List(row, metaServerCmd.Header))
	} else {
		row := map[string]string{
			cobrautil.ROW_RESULT: "success",
		}
		metaServerCmd.TableNew.Append(cobrautil.Map2List(row, metaServerCmd.Header))
		metaServerCmd.Error = cmderror.ErrSuccess()
	}
	return nil
}

func (metaServerCmd *DeleteMetaServerCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&metaServerCmd.FinalDingoCmd)
}
