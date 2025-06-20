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

package umount

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	"github.com/dingodb/dingofs-tools/pkg/base"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	mds "github.com/dingodb/dingofs-tools/proto/dingofs/proto/mds"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

type UmountFsRpc struct {
	Info      *base.Rpc
	Request   *mds.UmountFsRequest
	mdsClient mds.MdsServiceClient
}

var _ base.RpcFunc = (*UmountFsRpc)(nil) // check interface

type FsCommand struct {
	basecmd.FinalDingoCmd
	Rpc        UmountFsRpc
	fsName     string
	mountpoint string
}

var _ basecmd.FinalDingoCmdFunc = (*FsCommand)(nil) // check interface

func (ufRp *UmountFsRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	ufRp.mdsClient = mds.NewMdsServiceClient(cc)
}

func (ufRp *UmountFsRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := ufRp.mdsClient.UmountFs(ctx, ufRp.Request)
	output.ShowRpcData(ufRp.Request, response, ufRp.Info.RpcDataShow)
	return response, err
}

const (
	fsExample = `$ dingofs umount fs --fsname dingofs --mountpoint dingofs-103:9009:/mnt/dingofs`
)

func NewFsCommand() *cobra.Command {
	fsCmd := &FsCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:     "fs",
			Short:   "umount fs from the dingofs cluster",
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
	config.AddFsNameRequiredFlag(fCmd.Cmd)
	config.AddMountpointFlag(fCmd.Cmd)
}

func (fCmd *FsCommand) Init(cmd *cobra.Command, args []string) error {
	addrs, addrErr := config.GetFsMdsAddrSlice(fCmd.Cmd)
	if addrErr.TypeCode() != cmderror.CODE_SUCCESS {
		fCmd.Error = addrErr
		return fmt.Errorf(addrErr.Message)
	}

	fCmd.Rpc.Request = &mds.UmountFsRequest{}

	fCmd.fsName = config.GetFlagString(fCmd.Cmd, config.DINGOFS_FSNAME)
	fCmd.Rpc.Request.FsName = &fCmd.fsName
	fCmd.mountpoint = config.GetFlagString(fCmd.Cmd, config.DINGOFS_MOUNTPOINT)
	mountpointSlice := strings.Split(fCmd.mountpoint, ":")
	if len(mountpointSlice) != 3 {
		return fmt.Errorf("invalid mountpoint: %s, should be like: hostname:port:path", fCmd.mountpoint)
	}
	port, err := strconv.ParseUint(mountpointSlice[1], 10, 32)
	if err != nil {
		return fmt.Errorf("invalid point: %s", mountpointSlice[1])
	}
	port_ := uint32(port)
	fCmd.Rpc.Request.Mountpoint = &mds.Mountpoint{
		Hostname: &mountpointSlice[0],
		Port:     &port_,
		Path:     &mountpointSlice[2],
	}
	timeout := config.GetRpcTimeout(cmd)
	retrytimes := config.GetRpcRetryTimes(cmd)
	fCmd.Rpc.Info = base.NewRpc(addrs, timeout, retrytimes, "UmountFs")
	fCmd.Rpc.Info.RpcDataShow = config.GetFlagBool(fCmd.Cmd, "verbose")

	header := []string{cobrautil.ROW_FS_NAME, cobrautil.ROW_MOUNTPOINT, cobrautil.ROW_RESULT}
	fCmd.SetHeader(header)
	return nil
}

func (fCmd *FsCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&fCmd.FinalDingoCmd, fCmd)
}

func (fCmd *FsCommand) RunCommand(cmd *cobra.Command, args []string) error {
	response, errCmd := base.GetRpcResponse(fCmd.Rpc.Info, &fCmd.Rpc)
	if errCmd.TypeCode() != cmderror.CODE_SUCCESS {
		return fmt.Errorf(errCmd.Message)
	}
	uf := response.(*mds.UmountFsResponse)
	fCmd.updateTable(uf)
	fCmd.Error = errCmd

	return nil
}

func (fCmd *FsCommand) updateTable(info *mds.UmountFsResponse) *cmderror.CmdError {
	row := make(map[string]string)
	row[cobrautil.ROW_FS_NAME] = fCmd.fsName
	row[cobrautil.ROW_MOUNTPOINT] = fCmd.mountpoint
	err := cmderror.ErrUmountFs(int(info.GetStatusCode()))
	row[cobrautil.ROW_RESULT] = err.Message

	list := cobrautil.Map2List(row, fCmd.Header)
	fCmd.TableNew.Append(list)

	fCmd.Result = row
	return err
}

func (fCmd *FsCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&fCmd.FinalDingoCmd)
}
