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
 * Created Date: 2022-06-20
 * Author: chengyi (Cyber-SiKu)
 */

package fs

import (
	"context"
	"fmt"

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	"github.com/dingodb/dingofs-tools/proto/dingofs/proto/common"
	"github.com/dingodb/dingofs-tools/proto/dingofs/proto/mds"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

const (
	fsExample = `$ dingo create fs --fsname dingofs
$ dingo create fs --fsname dingofs --fstype s3 --s3.ak AK --s3.sk SK --s3.endpoint http://localhost:9000 --s3.bucketname dingofs-bucket --s3.blocksize 4MiB --s3.chunksize 4MiB`
)

type CreateFsRpc struct {
	Info      *basecmd.Rpc
	Request   *mds.CreateFsRequest
	mdsClient mds.MdsServiceClient
}

type FsCommand struct {
	basecmd.FinalDingoCmd
	Rpc *CreateFsRpc
}

func (cfRpc *CreateFsRpc) NewRpcClient(cc grpc.ClientConnInterface) {
	cfRpc.mdsClient = mds.NewMdsServiceClient(cc)
}

func (cfRpc *CreateFsRpc) Stub_Func(ctx context.Context) (interface{}, error) {
	response, err := cfRpc.mdsClient.CreateFs(ctx, cfRpc.Request)
	output.ShowRpcData(cfRpc.Request, response, cfRpc.Info.RpcDataShow)
	return response, err
}

var _ basecmd.RpcFunc = (*CreateFsRpc)(nil) // check interface

var _ basecmd.FinalDingoCmdFunc = (*FsCommand)(nil) // check interface

func NewFsCommand() *cobra.Command {
	fsCmd := &FsCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:     "fs",
			Short:   "create a fs in dingofs",
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
	config.AddUserOptionFlag(fCmd.Cmd)
	config.AddCapacityOptionFlag(fCmd.Cmd)
	config.AddBlockSizeOptionFlag(fCmd.Cmd)
	config.AddSumInDIrOptionFlag(fCmd.Cmd)
	config.AddFsTypeOptionFlag(fCmd.Cmd)
	// s3
	config.AddS3AkOptionFlag(fCmd.Cmd)
	config.AddS3SkOptionFlag(fCmd.Cmd)
	config.AddS3EndpointOptionFlag(fCmd.Cmd)
	config.AddS3BucknameOptionFlag(fCmd.Cmd)
	config.AddS3BlocksizeOptionFlag(fCmd.Cmd)
	config.AddS3ChunksizeOptionFlag(fCmd.Cmd)
}

func (fCmd *FsCommand) Init(cmd *cobra.Command, args []string) error {
	addrs, addrErr := config.GetFsMdsAddrSlice(fCmd.Cmd)
	if addrErr.TypeCode() != cmderror.CODE_SUCCESS {
		return fmt.Errorf(addrErr.Message)
	}

	header := []string{cobrautil.ROW_FS_NAME, cobrautil.ROW_RESULT}
	fCmd.SetHeader(header)

	fsName := config.GetFlagString(cmd, config.DINGOFS_FSNAME)
	if !cobrautil.IsValidFsname(fsName) {
		return fmt.Errorf("fsname[%s] is not vaild, it should be match regex: %s", fsName, cobrautil.FS_NAME_REGEX)
	}

	blocksizeStr := config.GetFlagString(cmd, config.DINGOFS_BLOCKSIZE)
	blocksize, err := humanize.ParseBytes(blocksizeStr)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("invalid blocksize: %s", blocksizeStr))
	}

	fsTypeStr := config.GetFlagString(cmd, config.DINGOFS_FSTYPE)
	fsType, errFstype := cobrautil.TranslateFsType(fsTypeStr)
	if errFstype.TypeCode() != cmderror.CODE_SUCCESS {
		return errFstype.ToError()
	}

	var fsDetail mds.FsDetail
	switch fsType {
	case common.FSType_TYPE_S3:
		errS3 := setS3Info(&fsDetail, fCmd.Cmd)
		if errS3.TypeCode() != cmderror.CODE_SUCCESS {
			return fmt.Errorf(errS3.Message)
		}
	default:
		return fmt.Errorf("invalid fs type: %s", fsTypeStr)
	}

	sumInDir := config.GetFlagBool(cmd, config.DINGOFS_SUMINDIR)

	owner := config.GetFlagString(cmd, config.DINGOFS_USER)

	capStr := config.GetFlagString(cmd, config.DINGOFS_CAPACITY)
	capability, err := humanize.ParseBytes(capStr)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("invalid capability: %s", capStr))
	}

	request := &mds.CreateFsRequest{
		FsName:         &fsName,
		BlockSize:      &blocksize,
		FsType:         &fsType,
		FsDetail:       &fsDetail,
		EnableSumInDir: &sumInDir,
		Owner:          &owner,
		Capacity:       &capability,
	}
	fCmd.Rpc = &CreateFsRpc{
		Request: request,
	}

	timeout := config.GetRpcTimeout(cmd)
	retrytimes := config.GetRpcRetryTimes(cmd)
	fCmd.Rpc.Info = basecmd.NewRpc(addrs, timeout, retrytimes, "CreateFs")
	fCmd.Rpc.Info.RpcDataShow = config.GetFlagBool(fCmd.Cmd, "verbose")

	return nil
}

func setS3Info(detail *mds.FsDetail, cmd *cobra.Command) *cmderror.CmdError {
	ak := config.GetFlagString(cmd, config.DINGOFS_S3_AK)
	sk := config.GetFlagString(cmd, config.DINGOFS_S3_SK)
	endpoint := config.GetFlagString(cmd, config.DINGOFS_S3_ENDPOINT)
	bucketname := config.GetFlagString(cmd, config.DINGOFS_S3_BUCKETNAME)
	blocksizeStr := config.GetFlagString(cmd, config.DINGOFS_S3_BLOCKSIZE)
	blocksize, err := humanize.ParseBytes(blocksizeStr)
	if err != nil {
		errParse := cmderror.ErrParse()
		errParse.Format(config.DINGOFS_S3_BLOCKSIZE, blocksizeStr)
		return errParse
	}
	chunksizeStr := config.GetFlagString(cmd, config.DINGOFS_S3_CHUNKSIZE)
	chunksize, err := humanize.ParseBytes(chunksizeStr)
	if err != nil {
		errParse := cmderror.ErrParse()
		errParse.Format(config.DINGOFS_S3_CHUNKSIZE, chunksizeStr)
		return errParse
	}

	info := &common.S3Info{
		Ak:         &ak,
		Sk:         &sk,
		Endpoint:   &endpoint,
		Bucketname: &bucketname,
		BlockSize:  &blocksize,
		ChunkSize:  &chunksize,
	}
	detail.S3Info = info
	return cmderror.ErrSuccess()
}

func (fCmd *FsCommand) RunCommand(cmd *cobra.Command, args []string) error {
	result, errCmd := basecmd.GetRpcResponse(fCmd.Rpc.Info, fCmd.Rpc)
	if errCmd.TypeCode() != cmderror.CODE_SUCCESS {
		return fmt.Errorf(errCmd.Message)
	}

	response := result.(*mds.CreateFsResponse)
	errCreate := cmderror.ErrCreateFs(int(response.GetStatusCode()))
	row := map[string]string{
		cobrautil.ROW_FS_NAME: fCmd.Rpc.Request.GetFsName(),
		cobrautil.ROW_RESULT:  errCreate.Message,
	}
	if response.GetStatusCode() == mds.FSStatusCode_OK {
		fsInfo := response.GetFsInfo()
		row[cobrautil.ROW_ID] = fmt.Sprintf("%d", fsInfo.GetFsId())
		row[cobrautil.ROW_STATUS] = fsInfo.GetStatus().String()
		row[cobrautil.ROW_CAPACITY] = fmt.Sprintf("%d", fsInfo.GetCapacity())
		row[cobrautil.ROW_BLOCKSIZE] = fmt.Sprintf("%d", fsInfo.GetBlockSize())
		row[cobrautil.ROW_FS_TYPE] = fsInfo.GetFsType().String()
		row[cobrautil.ROW_SUM_IN_DIR] = fmt.Sprintf("%t", fsInfo.GetEnableSumInDir())
		row[cobrautil.ROW_OWNER] = fsInfo.GetOwner()
	}

	fCmd.TableNew.Append(cobrautil.Map2List(row, fCmd.Header))

	var errs []*cmderror.CmdError
	res, errTranslate := output.MarshalProtoJson(response)
	if errTranslate != nil {
		errMar := cmderror.ErrMarShalProtoJson()
		errMar.Format(errTranslate.Error())
		errs = append(errs, errMar)
	}

	fCmd.Result = res
	fCmd.Error = cmderror.MostImportantCmdError(errs)
	return nil
}

func (fCmd *FsCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&fCmd.FinalDingoCmd, fCmd)
}

func (fCmd *FsCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&fCmd.FinalDingoCmd)
}
