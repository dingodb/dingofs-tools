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
 * Created Date: 2022-06-17
 * Author: chengyi (Cyber-SiKu)
 */

package fs

import (
	"context"
	"fmt"

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	"github.com/dingodb/dingofs-tools/pkg/base"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	"github.com/dingodb/dingofs-tools/proto/dingofs/proto/mds"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

const (
	fsExample = `$ dingo delete fs --fsname test1`
)

type DeleteFsRpc struct {
	Info      *base.Rpc
	Request   *mds.DeleteFsRequest
	mdsClient mds.MdsServiceClient
}

var _ base.RpcFunc = (*DeleteFsRpc)(nil) // check interface

type FsCommand struct {
	basecmd.FinalDingoCmd
	Rpc *DeleteFsRpc
}

var _ basecmd.FinalDingoCmdFunc = (*FsCommand)(nil) // check interface

func (dfRpc *DeleteFsRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	dfRpc.mdsClient = mds.NewMdsServiceClient(cc)
}

func (dfRpc *DeleteFsRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := dfRpc.mdsClient.DeleteFs(ctx, dfRpc.Request)
	output.ShowRpcData(dfRpc.Request, response, dfRpc.Info.RpcDataShow)
	return response, err
}

func NewFsCommand() *cobra.Command {
	fsCmd := &FsCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:     "fs",
			Short:   "delete a fs from dingofs",
			Example: fsExample,
		},
	}
	basecmd.NewFinalDingoCli(&fsCmd.FinalDingoCmd, fsCmd)
	return fsCmd.Cmd
}

func (fCmd *FsCommand) AddFlags() {
	config.AddRpcRetryTimesFlag(fCmd.Cmd)
	config.AddRpcRetryDelayFlag(fCmd.Cmd)
	config.AddRpcTimeoutFlag(fCmd.Cmd)
	config.AddFsMdsAddrFlag(fCmd.Cmd)
	config.AddFsNameRequiredFlag(fCmd.Cmd)
	config.AddNoConfirmOptionFlag(fCmd.Cmd)
}

func (fCmd *FsCommand) Init(cmd *cobra.Command, args []string) error {
	addrs, addrErr := config.GetFsMdsAddrSlice(fCmd.Cmd)
	if addrErr.TypeCode() != cmderror.CODE_SUCCESS {
		fCmd.Error = addrErr
		return fmt.Errorf(addrErr.Message)
	}

	header := []string{cobrautil.ROW_FS_NAME, cobrautil.ROW_RESULT}
	fCmd.SetHeader(header)

	fsName := config.GetFlagString(fCmd.Cmd, config.DINGOFS_FSNAME)

	request := &mds.DeleteFsRequest{
		FsName: &fsName,
	}
	fCmd.Rpc = &DeleteFsRpc{
		Request: request,
	}

	timeout := config.GetRpcTimeout(cmd)
	retrytimes := config.GetRpcRetryTimes(cmd)
	retryDelay := config.GetRpcRetryDelay(cmd)
	verbose := config.GetFlagBool(cmd, config.VERBOSE)
	fCmd.Rpc.Info = base.NewRpc(addrs, timeout, retrytimes, retryDelay, verbose, "DeleteFs")

	return nil
}

func (fCmd *FsCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&fCmd.FinalDingoCmd, fCmd)
}

func (fCmd *FsCommand) RunCommand(cmd *cobra.Command, args []string) error {
	fsName := fCmd.Rpc.Request.GetFsName()
	if !config.GetFlagBool(fCmd.Cmd, config.DINGOFS_NOCONFIRM) && !cobrautil.AskConfirmation(fmt.Sprintf("Are you sure to delete fs %s?", fsName), fsName) {
		return fmt.Errorf("abort delete fs")
	}

	result, err := base.GetRpcResponse(fCmd.Rpc.Info, fCmd.Rpc)
	if err.TypeCode() != cmderror.CODE_SUCCESS {
		return err.ToError()
	}
	response := result.(*mds.DeleteFsResponse)
	errDel := cmderror.ErrDeleteFs(int(response.GetStatusCode()))
	row := map[string]string{
		cobrautil.ROW_FS_NAME: fCmd.Rpc.Request.GetFsName(),
		cobrautil.ROW_RESULT:  errDel.Message,
	}
	fCmd.TableNew.Append(cobrautil.Map2List(row, fCmd.Header))

	res, errTranslate := output.MarshalProtoJson(response)
	if errTranslate != nil {
		errMar := cmderror.ErrMarShalProtoJson()
		errMar.Format(errTranslate.Error())
		return errMar.ToError()
	}

	fCmd.Result = res
	fCmd.Error = cmderror.ErrSuccess()

	return nil
}

func (fCmd *FsCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&fCmd.FinalDingoCmd)
}
