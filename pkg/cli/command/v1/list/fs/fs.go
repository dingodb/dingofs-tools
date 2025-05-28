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
 * Created Date: 2022-06-09
 * Author: chengyi (Cyber-SiKu)
 */

package fs

import (
	"context"
	"fmt"
	"strconv"

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	"github.com/dingodb/dingofs-tools/pkg/base"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	mds "github.com/dingodb/dingofs-tools/proto/dingofs/proto/mds"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc"
)

const (
	fsExample = `$ dingo list fs`
)

type ListFsRpc struct {
	Info      *base.Rpc
	Request   *mds.ListClusterFsInfoRequest
	mdsClient mds.MdsServiceClient
}

var _ base.RpcFunc = (*ListFsRpc)(nil) // check interface

type FsCommand struct {
	basecmd.FinalDingoCmd
	Rpc      *ListFsRpc
	response *mds.ListClusterFsInfoResponse
}

var _ basecmd.FinalDingoCmdFunc = (*FsCommand)(nil) // check interface

func (fRpc *ListFsRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	fRpc.mdsClient = mds.NewMdsServiceClient(cc)
}

func (fRpc *ListFsRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := fRpc.mdsClient.ListClusterFsInfo(ctx, fRpc.Request)
	output.ShowRpcData(fRpc.Request, response, fRpc.Info.RpcDataShow)
	return response, err
}

func NewFsCommand() *cobra.Command {
	return NewListFsCommand().Cmd
}

func NewListFsCommand() *FsCommand {
	fsCmd := &FsCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:     "fs",
			Short:   "list all fs info in the dingofs",
			Example: fsExample,
		},
	}

	basecmd.NewFinalDingoCli(&fsCmd.FinalDingoCmd, fsCmd)
	return fsCmd
}

func (fCmd *FsCommand) AddFlags() {
	config.AddRpcRetryTimesFlag(fCmd.Cmd)
	config.AddRpcTimeoutFlag(fCmd.Cmd)
	config.AddFsMdsAddrFlag(fCmd.Cmd)
}

func (fCmd *FsCommand) Init(cmd *cobra.Command, args []string) error {
	addrs, addrErr := config.GetFsMdsAddrSlice(fCmd.Cmd)
	if addrErr.TypeCode() != cmderror.CODE_SUCCESS {
		fCmd.Error = addrErr
		return fmt.Errorf(addrErr.Message)
	}
	fCmd.Rpc = &ListFsRpc{}
	fCmd.Rpc.Request = &mds.ListClusterFsInfoRequest{}
	timeout := config.GetRpcTimeout(cmd)
	retrytimes := config.GetRpcRetryTimes(cmd)
	fCmd.Rpc.Info = base.NewRpc(addrs, timeout, retrytimes, "ListClusterFsInfo")
	fCmd.Rpc.Info.RpcDataShow = config.GetFlagBool(fCmd.Cmd, "verbose")

	header := []string{cobrautil.ROW_FS_ID, cobrautil.ROW_FS_NAME, cobrautil.ROW_STATUS, cobrautil.ROW_BLOCKSIZE, cobrautil.ROW_STORAGE_TYPE, cobrautil.ROW_OWNER, cobrautil.ROW_MOUNT_NUM, cobrautil.ROW_UUID}
	fCmd.SetHeader(header)
	index_owner := slices.Index(header, cobrautil.ROW_OWNER)
	index_type := slices.Index(header, cobrautil.ROW_STORAGE_TYPE)
	fCmd.TableNew.SetAutoMergeCellsByColumnIndex([]int{index_owner, index_type})

	return nil
}

func (fCmd *FsCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&fCmd.FinalDingoCmd, fCmd)
}

func (fCmd *FsCommand) RunCommand(cmd *cobra.Command, args []string) error {
	response, errCmd := base.GetRpcResponse(fCmd.Rpc.Info, fCmd.Rpc)
	if errCmd.TypeCode() != cmderror.CODE_SUCCESS {
		return fmt.Errorf(errCmd.Message)
	}
	fCmd.response = response.(*mds.ListClusterFsInfoResponse)
	res, err := output.MarshalProtoJson(fCmd.response)
	if err != nil {
		return err
	}
	mapRes := res.(map[string]interface{})
	fCmd.Result = mapRes
	fCmd.updateTable()
	fCmd.Error = cmderror.ErrSuccess()
	return nil
}

func (fCmd *FsCommand) updateTable() {
	fssInfo := fCmd.response.GetFsInfo()
	rows := make([]map[string]string, 0)
	for _, fsInfo := range fssInfo {
		row := make(map[string]string)
		row[cobrautil.ROW_FS_ID] = fmt.Sprintf("%d", fsInfo.GetFsId())
		row[cobrautil.ROW_FS_NAME] = fsInfo.GetFsName()
		row[cobrautil.ROW_STATUS] = fsInfo.GetStatus().String()
		row[cobrautil.ROW_BLOCKSIZE] = fmt.Sprintf("%d", fsInfo.GetBlockSize())
		row[cobrautil.ROW_STORAGE_TYPE] = fsInfo.GetStorageInfo().GetType().String()
		row[cobrautil.ROW_OWNER] = fsInfo.GetOwner()
		row[cobrautil.ROW_MOUNT_NUM] = fmt.Sprintf("%d", fsInfo.GetMountNum())
		row[cobrautil.ROW_UUID] = fsInfo.GetUuid()
		rows = append(rows, row)
	}
	list := cobrautil.ListMap2ListSortByKeys(rows, fCmd.Header, []string{cobrautil.ROW_OWNER, cobrautil.ROW_STORAGE_TYPE, cobrautil.ROW_FS_ID})
	fCmd.TableNew.AppendBulk(list)
}

func (fCmd *FsCommand) ResultPlainOutput() error {
	if fCmd.TableNew.NumLines() == 0 {
		fmt.Println("no fs in cluster")
	}
	return output.FinalCmdOutputPlain(&fCmd.FinalDingoCmd)
}

func GetClusterFsInfo(caller *cobra.Command) (*mds.ListClusterFsInfoResponse, *cmderror.CmdError) {
	listFs := NewListFsCommand()
	listFs.Cmd.SetArgs([]string{"--format", config.FORMAT_NOOUT})
	config.AlignFlagsValue(caller, listFs.Cmd, []string{
		config.RPCRETRYTIMES, config.RPCTIMEOUT, config.DINGOFS_MDSADDR,
	})
	listFs.Cmd.SilenceErrors = true
	err := listFs.Cmd.Execute()
	if err != nil {
		retErr := cmderror.ErrGetClusterFsInfo()
		retErr.Format(err.Error())
		return nil, retErr
	}
	return listFs.response, cmderror.ErrSuccess()
}

func GetFsIds(caller *cobra.Command) ([]string, *cmderror.CmdError) {
	fsInfo, err := GetClusterFsInfo(caller)
	var ids []string
	if err.TypeCode() == cmderror.CODE_SUCCESS {
		for _, fsInfo := range fsInfo.GetFsInfo() {
			id := strconv.FormatUint(uint64(fsInfo.GetFsId()), 10)
			ids = append(ids, id)
		}
	}
	return ids, err
}
