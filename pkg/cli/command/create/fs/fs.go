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
	fsExample = `
# store in s3
$ dingo create fs --fsname dingofs --storagetype s3 --s3.ak AK --s3.sk SK --s3.endpoint http://localhost:9000 --s3.bucketname dingofs-bucket

# store in rados
$ dingo create fs --fsname dingofs --storagetype rados --rados.username admin --rados.key AQDg3Y2h --rados.mon 10.220.32.1:3300,10.220.32.2:3300,10.220.32.3:3300 --rados.poolname pool1 --rados.clustername ceph
`
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
	config.AddBlockSizeOptionFlag(fCmd.Cmd)
	config.AddChunksizeOptionFlag(fCmd.Cmd)
	config.AddStorageTypeOptionFlag(fCmd.Cmd)
	config.AddCapacityOptionFlag(fCmd.Cmd)
	// s3
	config.AddS3AkOptionFlag(fCmd.Cmd)
	config.AddS3SkOptionFlag(fCmd.Cmd)
	config.AddS3EndpointOptionFlag(fCmd.Cmd)
	config.AddS3BucknameOptionFlag(fCmd.Cmd)
	// rados
	config.AddRadosUsernameOptionFlag(fCmd.Cmd)
	config.AddRadosKeyOptionFlag(fCmd.Cmd)
	config.AddRadosMonOptionFlag(fCmd.Cmd)
	config.AddRadosPoolNameOptionFlag(fCmd.Cmd)
	config.AddRadosClusterNameOptionFlag(fCmd.Cmd)
}

func (fCmd *FsCommand) Init(cmd *cobra.Command, args []string) error {
	addrs, addrErr := config.GetFsMdsAddrSlice(fCmd.Cmd)
	if addrErr.TypeCode() != cmderror.CODE_SUCCESS {
		return fmt.Errorf(addrErr.Message)
	}

	fsName := config.GetFlagString(cmd, config.DINGOFS_FSNAME)
	if !cobrautil.IsValidFsname(fsName) {
		return fmt.Errorf("fsname[%s] is not vaild, it should be match regex: %s", fsName, cobrautil.FS_NAME_REGEX)
	}

	blocksizeStr := config.GetFlagString(cmd, config.DINGOFS_BLOCKSIZE)
	blocksize, err := humanize.ParseBytes(blocksizeStr)
	if err != nil {
		return fmt.Errorf("invalid blocksize: %s", blocksizeStr)
	}
	chunksizeStr := config.GetFlagString(cmd, config.DINGOFS_CHUNKSIZE)
	chunksize, err := humanize.ParseBytes(chunksizeStr)
	if err != nil {
		return fmt.Errorf("invalid chunksize: %s", chunksizeStr)
	}

	storageTypeStr := config.GetFlagString(cmd, config.DINGOFS_STORAGETYPE)
	storageType, errStoragetype := cobrautil.TranslateStorageType(storageTypeStr)
	if errStoragetype.TypeCode() != cmderror.CODE_SUCCESS {
		return errStoragetype.ToError()
	}

	var storageInfo common.StorageInfo
	switch storageType {
	case common.StorageType_TYPE_S3:
		err := SetS3Info(&storageInfo, fCmd.Cmd)
		if err != nil {
			return err
		}
	case common.StorageType_TYPE_RADOS:
		err := SetRadosInfo(&storageInfo, fCmd.Cmd)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid storage type: %s", storageTypeStr)
	}

	owner := config.GetFlagString(cmd, config.DINGOFS_USER)
	capStr := config.GetFlagString(cmd, config.DINGOFS_CAPACITY)
	capability, err := humanize.ParseBytes(capStr)
	if err != nil {
		return fmt.Errorf("invalid capability: %s", capStr)
	}

	request := &mds.CreateFsRequest{
		FsName:      &fsName,
		BlockSize:   &blocksize,
		ChunkSize:   &chunksize,
		StorageInfo: &storageInfo,
		Owner:       &owner,
		Capacity:    &capability,
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

func SetS3Info(storageInfo *common.StorageInfo, cmd *cobra.Command) error {
	ak := config.GetFlagString(cmd, config.DINGOFS_S3_AK)
	sk := config.GetFlagString(cmd, config.DINGOFS_S3_SK)
	endpoint := config.GetFlagString(cmd, config.DINGOFS_S3_ENDPOINT)
	bucketname := config.GetFlagString(cmd, config.DINGOFS_S3_BUCKETNAME)

	if len(ak) == 0 || len(sk) == 0 || len(endpoint) == 0 || len(bucketname) == 0 {
		return fmt.Errorf("s3 info is incomplete, please check s3.ak, s3.sk, s3.endpoint, s3.bucketname")
	}

	storage_s3info := &common.StorageInfo_S3Info{
		S3Info: &common.S3Info{
			Ak:         &ak,
			Sk:         &sk,
			Endpoint:   &endpoint,
			Bucketname: &bucketname,
		},
	}
	storageInfo.Type = new(common.StorageType)
	*storageInfo.Type = common.StorageType_TYPE_S3
	storageInfo.StorageInfo = storage_s3info

	return nil
}

func SetRadosInfo(storageInfo *common.StorageInfo, cmd *cobra.Command) error {
	userName := config.GetFlagString(cmd, config.DINGOFS_RADOS_USERNAME)
	secretKey := config.GetFlagString(cmd, config.DINGOFS_RADOS_KEY)
	monitor := config.GetFlagString(cmd, config.DINGOFS_RADOS_MON)
	poolName := config.GetFlagString(cmd, config.DINGOFS_RADOS_POOLNAME)
	clusterName := config.GetFlagString(cmd, config.DINGOFS_RADOS_CLUSTERNAME)

	if len(userName) == 0 || len(secretKey) == 0 || len(monitor) == 0 || len(poolName) == 0 {
		return fmt.Errorf("rados info is incomplete, please check rados.username, rados.key, rados.mon, rados.poolname")
	}

	storage_radosinfo := &common.StorageInfo_RadosInfo{
		RadosInfo: &common.RadosInfo{
			UserName:    &userName,
			Key:         &secretKey,
			MonHost:     &monitor,
			PoolName:    &poolName,
			ClusterName: &clusterName,
		},
	}
	storageInfo.Type = new(common.StorageType)
	*storageInfo.Type = common.StorageType_TYPE_RADOS
	storageInfo.StorageInfo = storage_radosinfo

	return nil
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
		header := []string{cobrautil.ROW_FS_ID, cobrautil.ROW_FS_NAME, cobrautil.ROW_STATUS, cobrautil.ROW_STORAGE_TYPE, cobrautil.ROW_UUID, cobrautil.ROW_RESULT}
		fCmd.SetHeader(header)
		fsInfo := response.GetFsInfo()
		row[cobrautil.ROW_FS_ID] = fmt.Sprintf("%d", fsInfo.GetFsId())
		row[cobrautil.ROW_STATUS] = fsInfo.GetStatus().String()
		row[cobrautil.ROW_STORAGE_TYPE] = fsInfo.GetStorageInfo().GetType().String()
		row[cobrautil.ROW_UUID] = fsInfo.GetUuid()
	} else {
		header := []string{cobrautil.ROW_FS_NAME, cobrautil.ROW_RESULT}
		fCmd.SetHeader(header)
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
