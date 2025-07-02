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
 * Created Date: 2022-06-13
 * Author: chengyi (Cyber-SiKu)
 */

package metadata

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
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

type MetadataRpc struct {
	Info           *base.Rpc
	Request        *topology.StatMetadataUsageRequest
	topologyClient topology.TopologyServiceClient
}

var _ base.RpcFunc = (*MetadataRpc)(nil) // check interface

type MetadataCommand struct {
	basecmd.FinalDingoCmd
	Rpc      *MetadataRpc
	response *topology.StatMetadataUsageResponse
}

var _ basecmd.FinalDingoCmdFunc = (*MetadataCommand)(nil) // check interface

func (mRpc *MetadataRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	mRpc.topologyClient = topology.NewTopologyServiceClient(cc)
}

func (mRpc *MetadataRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := mRpc.topologyClient.StatMetadataUsage(ctx, mRpc.Request)
	output.ShowRpcData(mRpc.Request, response, mRpc.Info.RpcDataShow)
	return response, err
}

const (
	metadataExample = `$ dingofs usage metadata`
)

func NewMetadataCommand() *cobra.Command {
	fsCmd := &MetadataCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:     "metadata",
			Short:   "get the usage of metadata in dingofs",
			Example: metadataExample,
		},
	}
	basecmd.NewFinalDingoCli(&fsCmd.FinalDingoCmd, fsCmd)
	return fsCmd.Cmd
}

func (mCmd *MetadataCommand) AddFlags() {
	config.AddRpcRetryTimesFlag(mCmd.Cmd)
	config.AddRpcRetryDelayFlag(mCmd.Cmd)
	config.AddRpcTimeoutFlag(mCmd.Cmd)
	config.AddFsMdsAddrFlag(mCmd.Cmd)
}

func (mCmd *MetadataCommand) Init(cmd *cobra.Command, args []string) error {
	addrs, addrErr := config.GetFsMdsAddrSlice(mCmd.Cmd)
	if addrErr.TypeCode() != cmderror.CODE_SUCCESS {
		mCmd.Error = addrErr
		return fmt.Errorf(addrErr.Message)
	}

	mCmd.Rpc = &MetadataRpc{}
	mCmd.Rpc.Request = &topology.StatMetadataUsageRequest{}
	timeout := config.GetRpcTimeout(cmd)
	retrytimes := config.GetRpcRetryTimes(cmd)
	retryDelay := config.GetRpcRetryDelay(cmd)
	verbose := config.GetFlagBool(cmd, config.VERBOSE)
	mCmd.Rpc.Info = base.NewRpc(addrs, timeout, retrytimes, retryDelay, verbose, "StatMetadataUsage")

	header := []string{cobrautil.ROW_METASERVER_ADDR, cobrautil.ROW_TOTAL, cobrautil.ROW_USED, cobrautil.ROW_LEFT}
	mCmd.SetHeader(header)

	return nil
}

func (mCmd *MetadataCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&mCmd.FinalDingoCmd, mCmd)
}

func (mCmd *MetadataCommand) RunCommand(cmd *cobra.Command, args []string) error {
	response, errCmd := base.GetRpcResponse(mCmd.Rpc.Info, mCmd.Rpc)
	if errCmd.TypeCode() != cmderror.CODE_SUCCESS {
		return fmt.Errorf(errCmd.Message)
	}
	mCmd.Error = errCmd
	mCmd.response = response.(*topology.StatMetadataUsageResponse)
	res, err := output.MarshalProtoJson(mCmd.response)
	if err != nil {
		return err
	}
	mapRes := res.(map[string]interface{})
	// update uint
	data := mapRes["metadataUsages"].([]interface{})
	for _, v := range data {
		vm := v.(map[string]interface{})
		vm["uint"] = "Byte"
	}
	mCmd.Result = mapRes
	mCmd.updateTable()
	return nil
}

func (mCmd *MetadataCommand) updateTable() {
	rows := make([]map[string]string, 0)
	for _, md := range mCmd.response.GetMetadataUsages() {
		row := make(map[string]string)
		row[cobrautil.ROW_METASERVER_ADDR] = md.GetMetaserverAddr()
		row[cobrautil.ROW_TOTAL] = humanize.IBytes(md.GetTotal())
		row[cobrautil.ROW_USED] = humanize.IBytes(md.GetUsed())
		row[cobrautil.ROW_LEFT] = humanize.IBytes(md.GetTotal() - md.GetUsed())
		rows = append(rows, row)
	}
	list := cobrautil.ListMap2ListSortByKeys(rows, mCmd.Header, []string{})
	mCmd.TableNew.AppendBulk(list)
}

func (mCmd *MetadataCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&mCmd.FinalDingoCmd)
}
