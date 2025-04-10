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
 * Created Date: 2022-06-15
 * Author: chengyi (Cyber-SiKu)
 */

package fs

import (
	"context"
	"fmt"
	"strconv"

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	mds "github.com/dingodb/dingofs-tools/proto/dingofs/proto/mds"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

const (
	fsExample = `$ dingo query fs --fsid 1,2,3
$ dingo query fs --fsname test1,test2,test3`
)

type QueryFsRpc struct {
	Info      *basecmd.Rpc
	Request   *mds.GetFsInfoRequest
	mdsClient mds.MdsServiceClient
}

var _ basecmd.RpcFunc = (*QueryFsRpc)(nil) // check interface

type FsCommand struct {
	basecmd.FinalDingoCmd
	Rpc  []*QueryFsRpc
	Rows []map[string]string
}

var _ basecmd.FinalDingoCmdFunc = (*FsCommand)(nil) // check interface

func (qfRp *QueryFsRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	qfRp.mdsClient = mds.NewMdsServiceClient(cc)
}

func (qfRp *QueryFsRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := qfRp.mdsClient.GetFsInfo(ctx, qfRp.Request)
	output.ShowRpcData(qfRp.Request, response, qfRp.Info.RpcDataShow)
	return response, err
}

func NewFsCommand() *cobra.Command {
	fsCmd := &FsCommand{
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

func (fCmd *FsCommand) AddFlags() {
	config.AddRpcRetryTimesFlag(fCmd.Cmd)
	config.AddRpcTimeoutFlag(fCmd.Cmd)
	config.AddFsMdsAddrFlag(fCmd.Cmd)
	config.AddFsNameSliceOptionFlag(fCmd.Cmd)
	config.AddFsIdSliceOptionFlag(fCmd.Cmd)
}

func (fCmd *FsCommand) Init(cmd *cobra.Command, args []string) error {
	addrs, addrErr := config.GetFsMdsAddrSlice(fCmd.Cmd)
	if addrErr.TypeCode() != cmderror.CODE_SUCCESS {
		fCmd.Error = addrErr
		return fmt.Errorf(addrErr.Message)
	}

	var fsIds []string
	var fsNames []string
	if !fCmd.Cmd.Flag(config.DINGOFS_FSNAME).Changed && !fCmd.Cmd.Flag(config.DINGOFS_FSID).Changed {
		return fmt.Errorf("fsname or fsid is required")
	}
	if fCmd.Cmd.Flag(config.DINGOFS_FSNAME).Changed && !fCmd.Cmd.Flag(config.DINGOFS_FSID).Changed {
		// fsname is set, but fsid is not set
		fsNames, _ = fCmd.Cmd.Flags().GetStringSlice(config.DINGOFS_FSNAME)
	} else {
		fsIds, _ = fCmd.Cmd.Flags().GetStringSlice(config.DINGOFS_FSID)
	}

	if len(fsIds) == 0 && len(fsNames) == 0 {
		return fmt.Errorf("fsname or fsid is required")
	}

	header := []string{cobrautil.ROW_ID, cobrautil.ROW_NAME, cobrautil.ROW_STATUS, cobrautil.ROW_BLOCKSIZE, cobrautil.ROW_FS_TYPE, cobrautil.ROW_SUM_IN_DIR, cobrautil.ROW_OWNER, cobrautil.ROW_MOUNT_NUM}
	fCmd.SetHeader(header)
	fCmd.TableNew.SetAutoMergeCellsByColumnIndex(
		cobrautil.GetIndexSlice(header, []string{cobrautil.ROW_FS_TYPE}),
	)

	fCmd.Rows = make([]map[string]string, 0)
	timeout := config.GetRpcTimeout(cmd)
	retrytimes := config.GetRpcRetryTimes(cmd)
	for i := range fsNames {
		request := &mds.GetFsInfoRequest{
			FsName: &fsNames[i],
		}
		rpc := &QueryFsRpc{
			Request: request,
		}
		rpc.Info = basecmd.NewRpc(addrs, timeout, retrytimes, "GetFsInfo")
		rpc.Info.RpcDataShow = config.GetFlagBool(fCmd.Cmd, "verbose")
		fCmd.Rpc = append(fCmd.Rpc, rpc)
	}

	for i := range fsIds {
		id, err := strconv.ParseUint(fsIds[i], 10, 32)
		if err != nil {
			return fmt.Errorf("invalid fsId: %s", fsIds[i])
		}
		id32 := uint32(id)
		request := &mds.GetFsInfoRequest{
			FsId: &id32,
		}
		rpc := &QueryFsRpc{
			Request: request,
		}
		rpc.Info = basecmd.NewRpc(addrs, timeout, retrytimes, "GetFsInfo")
		rpc.Info.RpcDataShow = config.GetFlagBool(fCmd.Cmd, "verbose")
		fCmd.Rpc = append(fCmd.Rpc, rpc)
	}

	return nil
}

func (fCmd *FsCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&fCmd.FinalDingoCmd, fCmd)
}

func (fCmd *FsCommand) RunCommand(cmd *cobra.Command, args []string) error {
	var infos []*basecmd.Rpc
	var funcs []basecmd.RpcFunc
	for _, rpc := range fCmd.Rpc {
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
		response := result.(*mds.GetFsInfoResponse)
		if response == nil {
			continue
		}
		if response.GetStatusCode() != mds.FSStatusCode_OK {
			code := response.GetStatusCode()
			err := cmderror.ErrGetFsInfo(int(code))
			err.Format(mds.FSStatusCode_name[int32(response.GetStatusCode())])
			errs = append(errs, err)
			continue
		}
		res, err := output.MarshalProtoJson(response)
		if err != nil {
			errMar := cmderror.ErrMarShalProtoJson()
			errMar.Format(err.Error())
			errs = append(errs, errMar)
		}
		resList = append(resList, res)

		fsInfo := response.GetFsInfo()
		row := make(map[string]string)
		row[cobrautil.ROW_ID] = strconv.FormatUint(uint64(fsInfo.GetFsId()), 10)
		row[cobrautil.ROW_NAME] = fsInfo.GetFsName()
		row[cobrautil.ROW_STATUS] = fsInfo.GetStatus().String()
		row[cobrautil.ROW_BLOCKSIZE] = fmt.Sprintf("%d", fsInfo.GetBlockSize())
		row[cobrautil.ROW_FS_TYPE] = fsInfo.GetFsType().String()
		row[cobrautil.ROW_SUM_IN_DIR] = fmt.Sprintf("%t", fsInfo.GetEnableSumInDir())
		row[cobrautil.ROW_OWNER] = fsInfo.GetOwner()
		row[cobrautil.ROW_MOUNT_NUM] = fmt.Sprintf("%d", fsInfo.GetMountNum())
		fCmd.Rows = append(fCmd.Rows, row)
	}

	list := cobrautil.ListMap2ListSortByKeys(fCmd.Rows, fCmd.Header, []string{
		cobrautil.ROW_FS_TYPE, cobrautil.ROW_ID,
	})
	fCmd.TableNew.AppendBulk(list)
	fCmd.Result = resList
	fCmd.Error = cmderror.MostImportantCmdError(errs)

	return nil
}

func (fCmd *FsCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&fCmd.FinalDingoCmd)
}

func NewFsInfoCommand() *FsCommand {
	fsCmd := &FsCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:   "metaserver",
			Short: "get metaserver status of dingofs",
		},
	}
	basecmd.NewFinalDingoCli(&fsCmd.FinalDingoCmd, fsCmd)
	return fsCmd
}

func GetFsInfo(caller *cobra.Command) (map[string]interface{}, error) {
	fsCmd := NewFsInfoCommand()
	fsCmd.Cmd.SetArgs([]string{
		fmt.Sprintf("--%s", config.FORMAT), config.FORMAT_NOOUT,
	})
	config.AlignFlagsValue(caller, fsCmd.Cmd, []string{
		config.RPCRETRYTIMES, config.RPCTIMEOUT, config.DINGOFS_MDSADDR, config.DINGOFS_FSNAME, config.DINGOFS_FSID,
	})
	fsCmd.Cmd.SilenceErrors = true
	fsCmd.Cmd.Execute()
	//check the value
	result := fsCmd.Result.([]interface{})
	if len(result) == 1 {
		tempMap := result[0].(map[string]interface{})
		if statusCode := tempMap["statusCode"].(string); statusCode == "OK" {
			return tempMap["fsInfo"].(map[string]interface{}), nil
		}
	}
	return nil, fmt.Errorf("get fsinfo failed")
}
