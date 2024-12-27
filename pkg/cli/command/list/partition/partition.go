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
 * Created Date: 2022-06-14
 * Author: chengyi (Cyber-SiKu)
 */

package partition

import (
	"context"
	"fmt"
	"strconv"

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/list/fs"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	"github.com/dingodb/dingofs-tools/proto/dingofs/proto/common"
	"github.com/dingodb/dingofs-tools/proto/dingofs/proto/topology"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

const (
	partitionExample = `$ dingo list partition
$ dingo list partition --fsid=1,2,3`
)

type ListPartitionRpc struct {
	Info           *basecmd.Rpc
	Request        *topology.ListPartitionRequest
	topologyClient topology.TopologyServiceClient
}

var _ basecmd.RpcFunc = (*ListPartitionRpc)(nil) // check interface

type PartitionCommand struct {
	basecmd.FinalDingoCmd
	Rpc                []*ListPartitionRpc
	fsId2Rows          map[uint32][]map[string]string
	fsId2PartitionList map[uint32][]*common.PartitionInfo
}

var _ basecmd.FinalDingoCmdFunc = (*PartitionCommand)(nil) // check interface

func (lpRp *ListPartitionRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	lpRp.topologyClient = topology.NewTopologyServiceClient(cc)
}

func (lpRp *ListPartitionRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := lpRp.topologyClient.ListPartition(ctx, lpRp.Request)
	output.ShowRpcData(lpRp.Request, response, lpRp.Info.RpcDataShow)
	return response, err
}

func NewPartitionCommand() *cobra.Command {
	pCmd := &PartitionCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:     "partition",
			Short:   "list partition in dingofs by fsid",
			Example: partitionExample,
		},
	}
	basecmd.NewFinalDingoCli(&pCmd.FinalDingoCmd, pCmd)
	return pCmd.Cmd
}

func (pCmd *PartitionCommand) AddFlags() {
	config.AddRpcRetryTimesFlag(pCmd.Cmd)
	config.AddRpcTimeoutFlag(pCmd.Cmd)
	config.AddFsMdsAddrFlag(pCmd.Cmd)
	config.AddFsIdOptionDefaultAllFlag(pCmd.Cmd)
}

func (pCmd *PartitionCommand) Init(cmd *cobra.Command, args []string) error {
	addrs, addrErr := config.GetFsMdsAddrSlice(pCmd.Cmd)
	if addrErr.TypeCode() != cmderror.CODE_SUCCESS {
		pCmd.Error = addrErr
		return fmt.Errorf(addrErr.Message)
	}

	header := []string{cobrautil.ROW_PARTITION_ID, cobrautil.ROW_FS_ID, cobrautil.ROW_POOL_ID, cobrautil.ROW_COPYSET_ID, cobrautil.ROW_START, cobrautil.ROW_END, cobrautil.ROW_STATUS}
	pCmd.SetHeader(header)

	pCmd.TableNew.SetAutoMergeCellsByColumnIndex(cobrautil.GetIndexSlice(
		pCmd.Header, []string{cobrautil.ROW_FS_ID, cobrautil.ROW_POOL_ID,
			cobrautil.ROW_COPYSET_ID},
	))

	fsIds := config.GetFlagStringSliceDefaultAll(pCmd.Cmd, config.DINGOFS_FSID)
	if fsIds[0] == "*" {
		var getFsIdErr *cmderror.CmdError
		fsIds, getFsIdErr = fs.GetFsIds(pCmd.Cmd)
		if getFsIdErr.TypeCode() != cmderror.CODE_SUCCESS {
			return fmt.Errorf(getFsIdErr.Message)
		}
	}

	pCmd.fsId2Rows = make(map[uint32][]map[string]string)
	pCmd.fsId2PartitionList = make(map[uint32][]*common.PartitionInfo)
	timeout := viper.GetDuration(config.VIPER_GLOBALE_RPCTIMEOUT)
	retrytimes := viper.GetInt32(config.VIPER_GLOBALE_RPCRETRYTIMES)
	for _, fsId := range fsIds {
		id, err := strconv.ParseUint(fsId, 10, 32)
		if err != nil {
			return fmt.Errorf("invalid fsId: %s", fsId)
		}
		request := &topology.ListPartitionRequest{}
		id32 := uint32(id)
		request.FsId = &id32
		rpc := &ListPartitionRpc{
			Request: request,
		}
		rpc.Info = basecmd.NewRpc(addrs, timeout, retrytimes, "ListPartition")
		rpc.Info.RpcDataShow = config.GetFlagBool(pCmd.Cmd, "verbose")
		pCmd.Rpc = append(pCmd.Rpc, rpc)
		pCmd.fsId2Rows[id32] = make([]map[string]string, 1)
		pCmd.fsId2Rows[id32][0] = make(map[string]string)
		pCmd.fsId2Rows[id32][0][cobrautil.ROW_FS_ID] = fsId
		pCmd.fsId2Rows[id32][0][cobrautil.ROW_POOL_ID] = cobrautil.ROW_VALUE_DNE
		pCmd.fsId2Rows[id32][0][cobrautil.ROW_COPYSET_ID] = cobrautil.ROW_VALUE_DNE
		pCmd.fsId2Rows[id32][0][cobrautil.ROW_PARTITION_ID] = cobrautil.ROW_VALUE_DNE
		pCmd.fsId2Rows[id32][0][cobrautil.ROW_START] = cobrautil.ROW_VALUE_DNE
		pCmd.fsId2Rows[id32][0][cobrautil.ROW_END] = cobrautil.ROW_VALUE_DNE
		pCmd.fsId2Rows[id32][0][cobrautil.ROW_STATUS] = cobrautil.ROW_VALUE_DNE
		pCmd.fsId2PartitionList[id32] = make([]*common.PartitionInfo, 0)
	}

	return nil
}

func (pCmd *PartitionCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&pCmd.FinalDingoCmd, pCmd)
}

func (pCmd *PartitionCommand) RunCommand(cmd *cobra.Command, args []string) error {
	var infos []*basecmd.Rpc
	var funcs []basecmd.RpcFunc
	if len(pCmd.Rpc) == 0 {
		pCmd.Result = "no partition in cluster"
		pCmd.Error = cmderror.ErrSuccess()
		return nil
	}
	for _, rpc := range pCmd.Rpc {
		infos = append(infos, rpc.Info)
		funcs = append(funcs, rpc)
	}

	results, errs := basecmd.GetRpcListResponse(infos, funcs)
	if len(errs) == len(infos) {
		mergeErr := cmderror.MergeCmdErrorExceptSuccess(errs)
		return mergeErr.ToError()
	}
	var resList []interface{}
	for _, result := range results {
		if result == nil {
			continue
		}
		response := result.(*topology.ListPartitionResponse)
		res, err := output.MarshalProtoJson(response)
		if err != nil {
			errMar := cmderror.ErrMarShalProtoJson()
			errMar.Format(err.Error())
			errs = append(errs, errMar)
		}
		resList = append(resList, res)
		// update fsId2Rows
		partitionList := response.GetPartitionInfoList()
		for _, partition := range partitionList {
			fsId := partition.GetFsId()
			pCmd.fsId2PartitionList[fsId] = append(pCmd.fsId2PartitionList[fsId], partition)
			var row *map[string]string
			if len(pCmd.fsId2Rows[fsId]) == 1 && pCmd.fsId2Rows[fsId][0][cobrautil.ROW_POOL_ID] == cobrautil.ROW_VALUE_DNE {
				row = &pCmd.fsId2Rows[fsId][0]
				pCmd.fsId2Rows[fsId] = make([]map[string]string, 0)
			} else {
				temp := make(map[string]string)
				row = &temp
				(*row)[cobrautil.ROW_FS_ID] = strconv.FormatUint(uint64(fsId), 10)
			}
			(*row)[cobrautil.ROW_POOL_ID] = strconv.FormatUint(uint64(partition.GetPoolId()), 10)
			(*row)[cobrautil.ROW_COPYSET_ID] = strconv.FormatUint(uint64(partition.GetCopysetId()), 10)
			(*row)[cobrautil.ROW_PARTITION_ID] = strconv.FormatUint(uint64(partition.GetPartitionId()), 10)
			(*row)[cobrautil.ROW_START] = strconv.FormatUint(uint64(partition.GetStart()), 10)
			(*row)[cobrautil.ROW_END] = strconv.FormatUint(uint64(partition.GetEnd()), 10)
			(*row)[cobrautil.ROW_STATUS] = partition.GetStatus().String()
			pCmd.fsId2Rows[fsId] = append(pCmd.fsId2Rows[fsId], (*row))
		}
	}

	pCmd.updateTable()
	pCmd.Result = resList
	pCmd.Error = cmderror.MostImportantCmdError(errs)

	return nil
}

func (pCmd *PartitionCommand) ResultPlainOutput() error {
	if pCmd.TableNew.NumLines() == 0 {
		fmt.Println("no partition in cluster")
		return nil
	}
	return output.FinalCmdOutputPlain(&pCmd.FinalDingoCmd)
}

func (pCmd *PartitionCommand) updateTable() {
	var total []map[string]string
	for _, rows := range pCmd.fsId2Rows {
		total = append(total, rows...)
	}
	list := cobrautil.ListMap2ListSortByKeys(total, pCmd.Header, []string{
		cobrautil.ROW_FS_ID, cobrautil.ROW_POOL_ID, cobrautil.ROW_COPYSET_ID,
		cobrautil.ROW_START, cobrautil.ROW_PARTITION_ID,
	})
	pCmd.TableNew.AppendBulk(list)
}

func NewListPartitionCommand() *PartitionCommand {
	pCmd := &PartitionCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:   "partition",
			Short: "list partition in dingofs by fsid",
		},
	}
	basecmd.NewFinalDingoCli(&pCmd.FinalDingoCmd, pCmd)
	return pCmd
}

func GetFsPartition(caller *cobra.Command) (*map[uint32][]*common.PartitionInfo, *cmderror.CmdError) {
	listPartionCmd := NewListPartitionCommand()
	listPartionCmd.Cmd.SetArgs([]string{
		fmt.Sprintf("--%s", config.FORMAT), config.FORMAT_NOOUT,
	})
	config.AlignFlagsValue(caller, listPartionCmd.Cmd, []string{
		config.RPCRETRYTIMES, config.RPCTIMEOUT, config.DINGOFS_MDSADDR,
		config.DINGOFS_FSID,
	})
	listPartionCmd.Cmd.SilenceErrors = true
	err := listPartionCmd.Cmd.Execute()
	if err != nil {
		retErr := cmderror.ErrGetFsPartition()
		retErr.Format(err.Error())
		return nil, retErr
	}
	return &listPartionCmd.fsId2PartitionList, cmderror.ErrSuccess()
}
