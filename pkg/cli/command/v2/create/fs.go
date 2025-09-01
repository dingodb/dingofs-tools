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

package create

import (
	"context"
	"fmt"
	"strings"

	pbmdsv2error "github.com/dingodb/dingofs-tools/proto/dingofs/proto/error"
	pbmdsv2 "github.com/dingodb/dingofs-tools/proto/dingofs/proto/mdsv2"

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	"github.com/dingodb/dingofs-tools/pkg/base"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	"github.com/dingodb/dingofs-tools/pkg/cli/command/v2/common"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

const (
	fsExample = `
# store in s3
$ dingo create fs --fsname dingofs --storagetype s3 --s3.ak AK --s3.sk SK --s3.endpoint http://localhost:9000 --s3.bucketname dingofs-bucket

# store in rados
$ dingo create fs --fsname dingofs --storagetype rados --rados.username admin --rados.key AQDg3Y2h --rados.mon 10.220.32.1:3300,10.220.32.2:3300,10.220.32.3:3300 --rados.poolname pool1 --rados.clustername ceph
`
)

type FsCommand struct {
	basecmd.FinalDingoCmd
	Rpc *common.CreateFsRpc
}

var _ basecmd.FinalDingoCmdFunc = (*FsCommand)(nil) // check interface

func NewCreateFsCommand() *cobra.Command {
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
	config.AddRpcRetryDelayFlag(fCmd.Cmd)
	config.AddRpcTimeoutFlag(fCmd.Cmd)
	config.AddFsMdsAddrFlag(fCmd.Cmd)
	config.AddFsIdUint32OptionFlag(fCmd.Cmd)
	config.AddFsNameRequiredFlag(fCmd.Cmd)
	config.AddUserOptionFlag(fCmd.Cmd)
	config.AddBlockSizeOptionFlag(fCmd.Cmd)
	config.AddChunksizeOptionFlag(fCmd.Cmd)
	config.AddStorageTypeOptionFlag(fCmd.Cmd)
	config.AddCapacityOptionFlag(fCmd.Cmd)
	config.AddPartitionTypeOptionFlag(fCmd.Cmd)
	config.AddMdsNumOptionalFlag(fCmd.Cmd)
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
	// new prc
	mdsRpc, err := common.CreateNewMdsRpc(cmd, "CreateFs")
	if err != nil {
		return err
	}
	// get request parameters
	fsName := config.GetFlagString(cmd, config.DINGOFS_FSNAME)
	if !cobrautil.IsValidFsname(fsName) {
		return fmt.Errorf("fsname[%s] is not vaild, it should be match regex: %s", fsName, cobrautil.FS_NAME_REGEX)
	}
	// block size
	blockSizeStr := config.GetFlagString(cmd, config.DINGOFS_BLOCKSIZE)
	blockSize, err := humanize.ParseBytes(blockSizeStr)
	if err != nil {
		return fmt.Errorf("invalid blocksize: %s", blockSizeStr)
	}
	// chunk size
	chunkSizeStr := config.GetFlagString(cmd, config.DINGOFS_CHUNKSIZE)
	chunkSize, err := humanize.ParseBytes(chunkSizeStr)
	if err != nil {
		return fmt.Errorf("invalid chunksize: %s", chunkSizeStr)
	}
	// owner
	owner := config.GetFlagString(cmd, config.DINGOFS_USER)
	// capability
	capStr := config.GetFlagString(cmd, config.DINGOFS_CAPACITY)
	capability, err := humanize.ParseBytes(capStr)
	if err != nil {
		return fmt.Errorf("invalid capability: %s", capStr)
	}
	// storage type,s3 or rados
	var fsType pbmdsv2.FsType
	var fsExtra pbmdsv2.FsExtra
	storageTypeStr := strings.ToUpper(config.GetFlagString(cmd, config.DINGOFS_STORAGETYPE))
	switch storageTypeStr {
	case "S3":
		fsType = pbmdsv2.FsType_S3
		err := SetS3Info(&fsExtra, fCmd.Cmd)
		if err != nil {
			return err
		}
	case "RADOS":
		fsType = pbmdsv2.FsType_RADOS
		err := SetRadosInfo(&fsExtra, fCmd.Cmd)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid storage type: %s", storageTypeStr)
	}
	// partition type
	var partitionType pbmdsv2.PartitionType
	partitionTypeStr := strings.ToUpper(config.GetFlagString(cmd, config.DINGOFS_PARTITION_TYPE))
	switch partitionTypeStr {
	case "HASH":
		partitionType = pbmdsv2.PartitionType_PARENT_ID_HASH_PARTITION
	case "MONOLITHIC":
		partitionType = pbmdsv2.PartitionType_MONOLITHIC_PARTITION
	default:
		return fmt.Errorf("invalid partition type: %s", partitionTypeStr)
	}

	request := pbmdsv2.CreateFsRequest{
		FsName:        fsName,
		BlockSize:     blockSize,
		ChunkSize:     chunkSize,
		FsType:        fsType,
		Owner:         owner,
		Capacity:      capability,
		FsExtra:       &fsExtra,
		PartitionType: partitionType,
	}

	if cmd.Flag(config.DINGOFS_FSID).Changed {
		fsId := config.GetFlagUint32(cmd, config.DINGOFS_FSID)
		request.FsId = fsId
	}
	if (cmd.Flag(config.DINGOFS_MDS_NUM).Changed) && (partitionType == pbmdsv2.PartitionType_PARENT_ID_HASH_PARTITION) {
		mdsNum := config.GetFlagUint32(cmd, config.DINGOFS_MDS_NUM)
		request.ExpectMdsNum = mdsNum
	}

	// set request info
	fCmd.Rpc = &common.CreateFsRpc{
		Info:    mdsRpc,
		Request: &request,
	}

	return nil
}

func SetS3Info(fsExtra *pbmdsv2.FsExtra, cmd *cobra.Command) error {
	ak := config.GetFlagString(cmd, config.DINGOFS_S3_AK)
	sk := config.GetFlagString(cmd, config.DINGOFS_S3_SK)
	endpoint := config.GetFlagString(cmd, config.DINGOFS_S3_ENDPOINT)
	bucketName := config.GetFlagString(cmd, config.DINGOFS_S3_BUCKETNAME)
	timeout := config.GetRpcTimeout(cmd)
	if len(ak) == 0 || len(sk) == 0 || len(endpoint) == 0 || len(bucketName) == 0 {
		return fmt.Errorf("s3 info is incomplete, please check s3.ak, s3.sk, s3.endpoint, s3.bucketname")
	}

	// check s3 health
	s3Checker, err := cobrautil.NewS3Checker(endpoint, ak, sk, bucketName, timeout)
	if err != nil {
		return err
	}
	ok, checkErr := s3Checker.Check(context.Background())
	if !ok {
		return fmt.Errorf("%s,%w", s3Checker.Name(), checkErr)
	}

	s3Info := &pbmdsv2.S3Info{
		Ak:         ak,
		Sk:         sk,
		Endpoint:   endpoint,
		Bucketname: bucketName,
	}
	fsExtra.S3Info = s3Info

	return nil
}

func SetRadosInfo(fsExtra *pbmdsv2.FsExtra, cmd *cobra.Command) error {
	userName := config.GetFlagString(cmd, config.DINGOFS_RADOS_USERNAME)
	secretKey := config.GetFlagString(cmd, config.DINGOFS_RADOS_KEY)
	monitor := config.GetFlagString(cmd, config.DINGOFS_RADOS_MON)
	poolName := config.GetFlagString(cmd, config.DINGOFS_RADOS_POOLNAME)
	clusterName := config.GetFlagString(cmd, config.DINGOFS_RADOS_CLUSTERNAME)
	timeout := config.GetRpcTimeout(cmd)

	if len(userName) == 0 || len(secretKey) == 0 || len(monitor) == 0 || len(poolName) == 0 {
		return fmt.Errorf("rados info is incomplete, please check rados.username, rados.key, rados.mon, rados.poolname")
	}
	if len(clusterName) == 0 {
		clusterName = "ceph"
	}

	// check rados health
	radosChecker, err := cobrautil.NewRadosChecker(monitor, userName, secretKey, poolName, clusterName, timeout)
	if err != nil {
		return err
	}
	ok, checkErr := radosChecker.Check(context.Background())
	if !ok {
		return fmt.Errorf("%s,%w", radosChecker.Name(), checkErr)
	}

	radosInfo := &pbmdsv2.RadosInfo{
		UserName:    userName,
		Key:         secretKey,
		MonHost:     monitor,
		PoolName:    poolName,
		ClusterName: clusterName,
	}
	fsExtra.RadosInfo = radosInfo

	return nil
}

func (fCmd *FsCommand) RunCommand(cmd *cobra.Command, args []string) error {
	// get rpc result
	response, errCmd := base.GetRpcResponse(fCmd.Rpc.Info, fCmd.Rpc)
	if errCmd.TypeCode() != cmderror.CODE_SUCCESS {
		return fmt.Errorf(errCmd.Message)
	}
	if response == nil {
		return fmt.Errorf("rpc no response")
	}
	result := response.(*pbmdsv2.CreateFsResponse)
	mdsErr := result.GetError()
	row := map[string]string{
		cobrautil.ROW_FS_NAME: fCmd.Rpc.Request.GetFsName(),
		cobrautil.ROW_RESULT:  cmderror.MDSV2Error(mdsErr).Message,
	}
	if mdsErr.GetErrcode() == pbmdsv2error.Errno_OK {
		header := []string{cobrautil.ROW_FS_ID, cobrautil.ROW_FS_NAME, cobrautil.ROW_STATUS, cobrautil.ROW_STORAGE_TYPE, cobrautil.ROW_UUID, cobrautil.ROW_RESULT}
		fCmd.SetHeader(header)
		fsInfo := result.GetFsInfo()
		row[cobrautil.ROW_FS_ID] = fmt.Sprintf("%d", fsInfo.GetFsId())
		row[cobrautil.ROW_STATUS] = fsInfo.GetStatus().String()
		row[cobrautil.ROW_STORAGE_TYPE] = fsInfo.GetFsType().String()
		row[cobrautil.ROW_UUID] = fsInfo.GetUuid()
	} else {
		header := []string{cobrautil.ROW_FS_NAME, cobrautil.ROW_RESULT}
		fCmd.SetHeader(header)
	}
	fCmd.TableNew.Append(cobrautil.Map2List(row, fCmd.Header))
	// to json
	res, errTranslate := output.MarshalProtoJson(result)
	if errTranslate != nil {
		return errTranslate
	}
	fCmd.Result = res
	fCmd.Error = cmderror.ErrSuccess()

	return nil
}

func (fCmd *FsCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&fCmd.FinalDingoCmd, fCmd)
}

func (fCmd *FsCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&fCmd.FinalDingoCmd)
}
