/*
 * Copyright (c) 2025 dingodb.com, Inc. All Rights Reserved
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mds

import (
	"fmt"
	"time"

	"github.com/dingodb/dingofs-tools/cli/cli"
	"github.com/dingodb/dingofs-tools/internal/common"
	"github.com/dingodb/dingofs-tools/internal/errno"
	"github.com/dingodb/dingofs-tools/internal/output"
	"github.com/dingodb/dingofs-tools/internal/rpc"
	"github.com/dingodb/dingofs-tools/internal/table"
	"github.com/dingodb/dingofs-tools/internal/utils"

	pbmdserror "github.com/dingodb/dingofs-tools/proto/dingofs/proto/error"
	"github.com/dingodb/dingofs-tools/proto/dingofs/proto/mds"
	"github.com/spf13/cobra"
)

const (
	MDS_STATUS_EXAMPLE = `Examples:
   $ dingo mds status`
)

type statusOptions struct {
	format string
}

func NewStatusCommand(dingoadm *cli.DingoAdm) *cobra.Command {
	var options statusOptions

	cmd := &cobra.Command{
		Use:     "status [OPTIONS]",
		Short:   "show mds cluster status",
		Args:    utils.NoArgs,
		Example: MDS_STATUS_EXAMPLE,
		RunE: func(cmd *cobra.Command, args []string) error {
			utils.ReadCommandConfig(cmd)
			output.SetShow(utils.GetBoolFlag(cmd, utils.VERBOSE))

			options.format = utils.GetStringFlag(cmd, utils.FORMAT)

			return runStatus(cmd, dingoadm, options)
		},
		SilenceUsage:          false,
		DisableFlagsInUseLine: true,
	}

	utils.SetFlagErrorFunc(cmd)

	// add flags
	utils.AddBoolFlag(cmd, utils.VERBOSE, "Show more debug info")
	utils.AddConfigFileFlag(cmd)
	utils.AddFormatFlag(cmd)

	utils.AddDurationFlag(cmd, utils.RPCTIMEOUT, "RPC timeout")
	utils.AddDurationFlag(cmd, utils.RPCRETRYDElAY, "RPC retry delay")
	utils.AddUint32Flag(cmd, utils.RPCRETRYTIMES, "RPC retry times")

	utils.AddStringFlag(cmd, utils.DINGOFS_MDSADDR, "Specify mds address")

	return cmd
}

func runStatus(cmd *cobra.Command, dingoadm *cli.DingoAdm, options statusOptions) error {
	outputResult := &common.OutputResult{
		Error: errno.ERR_OK,
	}
	// new rpc
	mdsRpc, err := rpc.CreateNewMdsRpc(cmd, "GetMDSList")
	if err != nil {
		return err
	}

	// set request info
	getMdsRpc := &rpc.GetMdsRpc{
		Info:    mdsRpc,
		Request: &mds.GetMDSListRequest{},
	}

	// get rpc result
	var result *mds.GetMDSListResponse
	response, rpcError := rpc.GetRpcResponse(getMdsRpc.Info, getMdsRpc)
	if rpcError.GetCode() != errno.ERR_OK.GetCode() {
		outputResult.Error = rpcError
	} else {
		result = response.(*mds.GetMDSListResponse)
		if mdsErr := result.GetError(); mdsErr.GetErrcode() != pbmdserror.Errno_OK {
			outputResult.Error = errno.ERR_RPC_FAILED.S(mdsErr.String())
		}
		outputResult.Result = result
	}

	// print result
	if options.format == "json" {
		return output.OutputJson(outputResult)
	}

	if outputResult.Error.GetCode() != errno.ERR_OK.GetCode() {
		return outputResult.Error
	}

	// set table header
	header := []string{common.ROW_ID, common.ROW_ADDR, common.ROW_STATE, common.ROW_LASTONLINETIME, common.ROW_ONLINE_STATE}
	table.SetHeader(header)
	// fill table
	mdsInfos := result.GetMdses()
	rows := make([]map[string]string, 0)
	for _, mdsInfo := range mdsInfos {
		row := make(map[string]string)
		row[common.ROW_ID] = fmt.Sprintf("%d", mdsInfo.GetId())
		row[common.ROW_ADDR] = fmt.Sprintf("%s:%d", mdsInfo.GetLocation().GetHost(), mdsInfo.GetLocation().GetPort())
		row[common.ROW_STATE] = mdsInfo.GetState().String()
		unixTime := int64(mdsInfo.GetLastOnlineTimeMs())
		t := time.Unix(unixTime/1000, (unixTime%1000)*1000000)
		row[common.ROW_LASTONLINETIME] = t.Format("2006-01-02 15:04:05.000")
		if mdsInfo.GetIsOnline() {
			row[common.ROW_ONLINE_STATE] = common.ROW_VALUE_ONLINE
		} else {
			row[common.ROW_ONLINE_STATE] = common.ROW_VALUE_OFFLINE
		}
		rows = append(rows, row)
	}

	list := table.ListMap2ListSortByKeys(rows, header, []string{common.ROW_ID})
	table.AppendBulk(list)
	table.RenderWithNoData("no mds in cluster")

	return nil
}
