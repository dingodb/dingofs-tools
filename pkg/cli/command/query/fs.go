// Copyright (c) 2025 dingodb.com, Inc. All Rights Reserved
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

package query

import (
	"fmt"
	"github.com/dingodb/dingofs-tools/pkg/common"
	"github.com/dingodb/dingofs-tools/pkg/rpc"
	"slices"
	"strconv"

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	"github.com/dingodb/dingofs-tools/pkg/base"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	pbmdsv2error "github.com/dingodb/dingofs-tools/proto/dingofs/proto/error"
	"github.com/spf13/cobra"

	pbmdsv2 "github.com/dingodb/dingofs-tools/proto/dingofs/proto/mdsv2"
)

const (
	fsExample = `$ dingo query fs --fsid 1,2,3
$ dingo query fs --fsname fs1,fs2,fs3`
)

type QueryFsCommand struct {
	basecmd.FinalDingoCmd
	Rpc []*rpc.GetFsRpc
}

var _ basecmd.FinalDingoCmdFunc = (*QueryFsCommand)(nil) // check interface

func NewQueryFsCommand() *cobra.Command {
	fsCmd := &QueryFsCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:     "fs",
			Short:   "query fs in dingofs by fsname or fsid",
			Long:    "when both fsname and fsid exist, query only by fsid",
			Example: fsExample,
		},
	}
	basecmd.NewFinalDingoCli(&fsCmd.FinalDingoCmd, fsCmd)
	return fsCmd.Cmd
}

func (queryFSCmd *QueryFsCommand) AddFlags() {
	config.AddRpcRetryTimesFlag(queryFSCmd.Cmd)
	config.AddRpcRetryDelayFlag(queryFSCmd.Cmd)
	config.AddRpcTimeoutFlag(queryFSCmd.Cmd)
	config.AddFsMdsAddrFlag(queryFSCmd.Cmd)
	config.AddFsNameSliceOptionFlag(queryFSCmd.Cmd)
	config.AddFsIdSliceOptionFlag(queryFSCmd.Cmd)
}

func (queryFSCmd *QueryFsCommand) Init(cmd *cobra.Command, args []string) error {
	// get fsid or fsname
	// if set simultaneously, fsid has a higher priority
	var fsIds []string
	var fsNames []string
	if !queryFSCmd.Cmd.Flag(config.DINGOFS_FSNAME).Changed && !queryFSCmd.Cmd.Flag(config.DINGOFS_FSID).Changed {
		return fmt.Errorf("fsname or fsid is required")
	}
	if queryFSCmd.Cmd.Flag(config.DINGOFS_FSNAME).Changed && !queryFSCmd.Cmd.Flag(config.DINGOFS_FSID).Changed {
		// fsname is set, but fsid is not set
		fsNames, _ = queryFSCmd.Cmd.Flags().GetStringSlice(config.DINGOFS_FSNAME)
	} else {
		fsIds, _ = queryFSCmd.Cmd.Flags().GetStringSlice(config.DINGOFS_FSID)
	}

	if len(fsIds) == 0 && len(fsNames) == 0 {
		return fmt.Errorf("fsname or fsid is required")
	}

	// query by fsnames
	for _, fsName := range fsNames {
		// set request info
		getFsRpc := &rpc.GetFsRpc{
			Request: &pbmdsv2.GetFsInfoRequest{
				FsName: fsName,
			},
		}
		// new rpc
		mdsRpc, err := rpc.CreateNewMdsRpc(cmd, "GetFsInfo")
		if err != nil {
			return err
		}
		getFsRpc.Info = mdsRpc
		queryFSCmd.Rpc = append(queryFSCmd.Rpc, getFsRpc)
	}
	// query by fsids
	for _, fsId := range fsIds {
		id, err := strconv.ParseUint(fsId, 10, 32)
		if err != nil {
			return fmt.Errorf("invalid fsId: %s", fsId)
		}
		id32 := uint32(id)
		// set request info
		getFsRpc := &rpc.GetFsRpc{
			Request: &pbmdsv2.GetFsInfoRequest{
				FsId: id32,
			},
		}
		// new rpc
		mdsRpc, err := rpc.CreateNewMdsRpc(cmd, "GetFsInfo")
		if err != nil {
			return err
		}
		getFsRpc.Info = mdsRpc
		queryFSCmd.Rpc = append(queryFSCmd.Rpc, getFsRpc)
	}

	// set table header
	header := []string{cobrautil.ROW_FS_ID, cobrautil.ROW_FS_NAME, cobrautil.ROW_STATUS, cobrautil.ROW_BLOCKSIZE, cobrautil.ROW_CHUNK_SIZE, cobrautil.ROW_MDS_NUM, cobrautil.ROW_STORAGE_TYPE, cobrautil.ROW_STORAGE, cobrautil.ROW_MOUNT_NUM, cobrautil.ROW_UUID}
	queryFSCmd.SetHeader(header)
	queryFSCmd.TableNew.SetAutoWrapText(false)

	indexType := slices.Index(header, cobrautil.ROW_STORAGE_TYPE)
	queryFSCmd.TableNew.SetAutoMergeCellsByColumnIndex([]int{indexType})

	return nil
}

func (queryFSCmd *QueryFsCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&queryFSCmd.FinalDingoCmd, queryFSCmd)
}

func (queryFSCmd *QueryFsCommand) RunCommand(cmd *cobra.Command, args []string) error {
	var infos []*base.Rpc
	var funcs []base.RpcFunc
	for _, rpc := range queryFSCmd.Rpc {
		infos = append(infos, rpc.Info)
		funcs = append(funcs, rpc)
	}
	responses, errs := base.GetRpcListResponse(infos, funcs)
	if len(errs) == len(infos) {
		mergeErr := cmderror.MergeCmdErrorExceptSuccess(errs)
		return mergeErr.ToError()
	}
	// traverse query results
	var resList []interface{}
	rows := make([]map[string]string, 0)
	for _, response := range responses {
		result := response.(*pbmdsv2.GetFsInfoResponse)
		if result == nil {
			continue
		}
		if mdsErr := result.GetError(); mdsErr.GetErrcode() != pbmdsv2error.Errno_OK {
			errs = append(errs, cmderror.MDSV2Error(mdsErr))
			continue
		}
		// fill table
		fsInfo := result.GetFsInfo()

		row := make(map[string]string)
		row[cobrautil.ROW_FS_ID] = fmt.Sprintf("%d", fsInfo.GetFsId())
		row[cobrautil.ROW_FS_NAME] = fsInfo.GetFsName()
		row[cobrautil.ROW_STATUS] = fsInfo.GetStatus().String()
		row[cobrautil.ROW_BLOCKSIZE] = fmt.Sprintf("%d", fsInfo.GetBlockSize())
		row[cobrautil.ROW_CHUNK_SIZE] = fmt.Sprintf("%d", fsInfo.GetChunkSize())

		partitionType := fsInfo.GetPartitionPolicy().GetType()
		if partitionType == pbmdsv2.PartitionType_PARENT_ID_HASH_PARTITION {
			row[cobrautil.ROW_STORAGE_TYPE] = fmt.Sprintf("%s(%s %d)", fsInfo.GetFsType().String(),
				common.ConvertPbPartitionTypeToString(partitionType), fsInfo.GetPartitionPolicy().GetParentHash().GetBucketNum())
			row[cobrautil.ROW_MDS_NUM] = fmt.Sprintf("%d", len(fsInfo.GetPartitionPolicy().GetParentHash().GetDistributions()))
		} else {
			row[cobrautil.ROW_STORAGE_TYPE] = fmt.Sprintf("%s(%s)", fsInfo.GetFsType().String(), common.ConvertPbPartitionTypeToString(partitionType))
			row[cobrautil.ROW_MDS_NUM] = "1"
		}

		row[cobrautil.ROW_STORAGE] = common.ConvertFsExtraToString(fsInfo.GetExtra())
		row[cobrautil.ROW_MOUNT_NUM] = fmt.Sprintf("%d", len(fsInfo.GetMountPoints()))
		row[cobrautil.ROW_UUID] = fsInfo.GetUuid()

		rows = append(rows, row)

		// to json
		res, err := output.MarshalProtoJson(result)
		if err != nil {
			errMar := cmderror.ErrMarShalProtoJson()
			errMar.Format(err.Error())
			errs = append(errs, errMar)
		}
		resList = append(resList, res)
	}

	list := cobrautil.ListMap2ListSortByKeys(rows, queryFSCmd.Header, []string{
		cobrautil.ROW_STORAGE_TYPE, cobrautil.ROW_ID,
	})
	queryFSCmd.TableNew.AppendBulk(list)
	queryFSCmd.Result = resList
	queryFSCmd.Error = cmderror.MostImportantCmdError(errs)

	return nil
}

func (queryFSCmd *QueryFsCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&queryFSCmd.FinalDingoCmd)
}
