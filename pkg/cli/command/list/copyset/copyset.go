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
 * Created Date: 2022-06-25
 * Author: chengyi (Cyber-SiKu)
 */

package copyset

import (
	"context"
	"fmt"
	"strings"

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	"github.com/dingodb/dingofs-tools/proto/dingofs/proto/topology"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc"
)

const (
	copysetExample = `$ dingo list copyset`
)

const ()

type ListCopysetRpc struct {
	Info           *basecmd.Rpc
	Request        *topology.ListCopysetInfoRequest
	topologyClient topology.TopologyServiceClient
}

var _ basecmd.RpcFunc = (*ListCopysetRpc)(nil) // check interface

type CopysetCommand struct {
	basecmd.FinalDingoCmd
	Rpc      *ListCopysetRpc
	response *topology.ListCopysetInfoResponse
}

var _ basecmd.FinalDingoCmdFunc = (*CopysetCommand)(nil) // check interface

func (cRpc *ListCopysetRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	cRpc.topologyClient = topology.NewTopologyServiceClient(cc)
}

func (cRpc *ListCopysetRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := cRpc.topologyClient.ListCopysetInfo(ctx, cRpc.Request)
	output.ShowRpcData(cRpc.Request, response, cRpc.Info.RpcDataShow)
	return response, err
}

func NewCopysetCommand() *cobra.Command {
	return NewListCopysetCommand().Cmd
}

func NewListCopysetCommand() *CopysetCommand {
	copysetCmd := &CopysetCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:     "copyset",
			Short:   "list all copyset info of the dingofs",
			Example: copysetExample,
		},
	}
	basecmd.NewFinalDingoCli(&copysetCmd.FinalDingoCmd, copysetCmd)
	return copysetCmd
}

func (cCmd *CopysetCommand) AddFlags() {
	config.AddRpcRetryTimesFlag(cCmd.Cmd)
	config.AddRpcTimeoutFlag(cCmd.Cmd)
	config.AddFsMdsAddrFlag(cCmd.Cmd)
}

func (cCmd *CopysetCommand) Init(cmd *cobra.Command, args []string) error {
	addrs, addrErr := config.GetFsMdsAddrSlice(cCmd.Cmd)
	if addrErr.TypeCode() != cmderror.CODE_SUCCESS {
		cCmd.Error = addrErr
		return fmt.Errorf(addrErr.Message)
	}
	cCmd.Rpc = &ListCopysetRpc{}
	cCmd.Rpc.Request = &topology.ListCopysetInfoRequest{}
	timeout := config.GetRpcTimeout(cmd)
	retrytimes := config.GetRpcRetryTimes(cmd)
	cCmd.Rpc.Info = basecmd.NewRpc(addrs, timeout, retrytimes, "ListCopysetInfo")
	cCmd.Rpc.Info.RpcDataShow = config.GetFlagBool(cCmd.Cmd, "verbose")

	header := []string{cobrautil.ROW_KEY, cobrautil.ROW_COPYSET_ID, cobrautil.ROW_POOL_ID, cobrautil.ROW_EPOCH, cobrautil.ROW_LEADER_PEER, cobrautil.ROW_FOLLOWER_PEER, cobrautil.ROW_PEER_NUMBER}
	cCmd.SetHeader(header)
	index_pool := slices.Index(header, cobrautil.ROW_POOL_ID)
	index_leader := slices.Index(header, cobrautil.ROW_LEADER_PEER)
	cCmd.TableNew.SetAutoMergeCellsByColumnIndex([]int{index_pool, index_leader})

	return nil
}

func (cCmd *CopysetCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&cCmd.FinalDingoCmd, cCmd)
}

func (cCmd *CopysetCommand) RunCommand(cmd *cobra.Command, args []string) error {
	response, errCmd := basecmd.GetRpcResponse(cCmd.Rpc.Info, cCmd.Rpc)
	if errCmd.TypeCode() != cmderror.CODE_SUCCESS {
		return errCmd.ToError()
	}
	cCmd.response = response.(*topology.ListCopysetInfoResponse)
	res, err := output.MarshalProtoJson(cCmd.response)
	if err != nil {
		return err
	}
	mapRes := res.(map[string]interface{})
	cCmd.Result = mapRes
	cCmd.updateTable()
	cCmd.Error = cmderror.ErrSuccess()
	return nil
}

func (cCmd *CopysetCommand) updateTable() {
	rows := make([]map[string]string, 0)
	copysets := cCmd.response.GetCopysetValues()
	for _, copyset := range copysets {
		info := copyset.GetCopysetInfo()
		code := copyset.GetStatusCode()
		row := make(map[string]string)
		key := cobrautil.GetCopysetKey(uint64(info.GetPoolId()), uint64(info.GetCopysetId()))
		row[cobrautil.ROW_KEY] = fmt.Sprintf("%d", key)
		row[cobrautil.ROW_COPYSET_ID] = fmt.Sprintf("%d", info.GetCopysetId())
		row[cobrautil.ROW_POOL_ID] = fmt.Sprintf("%d", info.GetPoolId())
		row[cobrautil.ROW_EPOCH] = fmt.Sprintf("%d", info.GetEpoch())
		if code != topology.TopoStatusCode_TOPO_OK && info.GetLeaderPeer() == nil {
			row[cobrautil.ROW_LEADER_PEER] = ""
			row[cobrautil.ROW_PEER_NUMBER] = fmt.Sprintf("%d", len(info.GetPeers()))
		} else {
			row[cobrautil.ROW_LEADER_PEER] = info.GetLeaderPeer().String()

			//all peers
			peerNum := 0
			all_peers := info.GetPeers()
			leader_id := info.GetLeaderPeer().GetId()
			var follower_peers string
			for _, peer := range all_peers {
				if peer != nil {
					peerNum++
					if peer.GetId() != leader_id {
						follower_peers += strings.Replace(peer.String(), "\r\n", " ", -1) + "\n"
					}
				}
			}
			if len(follower_peers) > 0 {
				row[cobrautil.ROW_FOLLOWER_PEER] = follower_peers[:len(follower_peers)-1] //remove last \n
			}
			row[cobrautil.ROW_FOLLOWER_PEER] = follower_peers
			row[cobrautil.ROW_PEER_NUMBER] = fmt.Sprintf("%d", peerNum)
		}
		rows = append(rows, row)
	}
	list := cobrautil.ListMap2ListSortByKeys(rows, cCmd.Header, []string{cobrautil.ROW_LEADER_PEER, cobrautil.ROW_KEY})
	cCmd.TableNew.AppendBulk(list)
}

func (cCmd *CopysetCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&cCmd.FinalDingoCmd)
}

func GetCopysetsInfos(caller *cobra.Command) (*topology.ListCopysetInfoResponse, *cmderror.CmdError) {
	listCopyset := NewListCopysetCommand()
	listCopyset.Cmd.SetArgs([]string{
		fmt.Sprintf("--%s", config.FORMAT), config.FORMAT_NOOUT,
	})
	config.AlignFlagsValue(caller, listCopyset.Cmd, []string{config.RPCRETRYTIMES, config.RPCTIMEOUT, config.DINGOFS_MDSADDR})
	listCopyset.Cmd.SilenceErrors = true
	err := listCopyset.Cmd.Execute()
	if err != nil {
		retErr := cmderror.ErrListCopyset()
		retErr.Format(err)
		return nil, retErr
	}
	return listCopyset.response, cmderror.ErrSuccess()
}
