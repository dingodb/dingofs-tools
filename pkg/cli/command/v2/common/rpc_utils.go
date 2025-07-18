package common

import (
	"fmt"
	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	"github.com/dingodb/dingofs-tools/pkg/base"
	cmdcommon "github.com/dingodb/dingofs-tools/pkg/cli/command/v1/common"
	"github.com/dingodb/dingofs-tools/pkg/config"
	pbmdsv2 "github.com/dingodb/dingofs-tools/proto/dingofs/proto/mdsv2"
	"github.com/spf13/cobra"
)

func ConvertPbPartitionTypeToString(partitionType pbmdsv2.PartitionType) string {
	switch partitionType {
	case pbmdsv2.PartitionType_MONOLITHIC_PARTITION:
		return "monolithic"
	case pbmdsv2.PartitionType_PARENT_ID_HASH_PARTITION:
		return "hash"
	default:
		return "unknown"
	}
}

// retrieve fsid from command-line parameters,if not set, get by GetFsInfo via fsname
func GetFsId(cmd *cobra.Command) (uint32, error) {
	fsId, fsName, fsErr := cmdcommon.CheckAndGetFsIdOrFsNameValue(cmd)
	if fsErr != nil {
		return 0, fsErr
	}
	// fsId is not set,need to get fsId by fsName (fsName -> fsId)
	if fsId == 0 {
		fsInfo, fsErr := GetFsInfo(cmd, 0, fsName)
		if fsErr != nil {
			return 0, fsErr
		}
		fsId = fsInfo.GetFsId()
		if fsId == 0 {
			return 0, fmt.Errorf("fsid is invalid")
		}
	}

	return fsId, nil
}

// retrieve fsid from command-line parameters,if not set, get by GetFsInfo via fsid
func GetFsName(cmd *cobra.Command) (string, error) {
	fsId, fsName, fsErr := cmdcommon.CheckAndGetFsIdOrFsNameValue(cmd)
	if fsErr != nil {
		return "", fsErr
	}
	if len(fsName) == 0 { // fsName is not set,need to get fsName by fsId (fsId->fsName)
		fsInfo, fsErr := GetFsInfo(cmd, fsId, "")
		if fsErr != nil {
			return "", fsErr
		}
		fsName = fsInfo.GetFsName()
		if len(fsName) == 0 {
			return "", fmt.Errorf("fsName is invalid")
		}
	}

	return fsName, nil
}

func CreateNewMdsRpcWithEndPoint(cmd *cobra.Command, endpoint []string, serviceName string) *base.Rpc {
	// new rpc
	timeout := config.GetRpcTimeout(cmd)
	retryTimes := config.GetRpcRetryTimes(cmd)
	retryDelay := config.GetRpcRetryDelay(cmd)
	verbose := config.GetFlagBool(cmd, config.VERBOSE)
	mdsRpc := base.NewRpc(endpoint, timeout, retryTimes, retryDelay, verbose, serviceName)

	return mdsRpc
}

// create new mds rpc
func CreateNewMdsRpc(cmd *cobra.Command, serviceName string) (*base.Rpc, error) {
	// get mds address
	endpoints, addr := config.GetFsMdsAddrSlice(cmd)
	if addr.TypeCode() != cmderror.CODE_SUCCESS {
		return nil, fmt.Errorf(addr.Message)
	}

	mdsRpc := CreateNewMdsRpcWithEndPoint(cmd, endpoints, serviceName)

	return mdsRpc, nil
}
