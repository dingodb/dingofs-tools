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

package status

import (
	"fmt"
	"github.com/dingodb/dingofs-tools/pkg/rpc"

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	"github.com/dingodb/dingofs-tools/pkg/base"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	config "github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	pbmdserror "github.com/dingodb/dingofs-tools/proto/dingofs/proto/error"
	pbmds "github.com/dingodb/dingofs-tools/proto/dingofs/proto/mds"
	"github.com/spf13/cobra"

	"time"
)

const (
	mdsExample = `$ dingo status mds`
)

type MdsCommand struct {
	basecmd.FinalDingoCmd
	Rpc *rpc.GetMdsRpc
}

var _ basecmd.FinalDingoCmdFunc = (*MdsCommand)(nil) // check interface

func NewMdsCommand() *cobra.Command {
	return NewStatusMdsCommand().Cmd
}

func (mdsCmd *MdsCommand) AddFlags() {
	config.AddRpcRetryTimesFlag(mdsCmd.Cmd)
	config.AddRpcRetryDelayFlag(mdsCmd.Cmd)
	config.AddRpcTimeoutFlag(mdsCmd.Cmd)
	config.AddFsMdsAddrFlag(mdsCmd.Cmd)
}

func (mdsCmd *MdsCommand) Init(cmd *cobra.Command, args []string) error {
	// new rpc
	mdsRpc, err := rpc.CreateNewMdsRpc(cmd, "GetMDSList")
	if err != nil {
		return err
	}
	// set request info
	mdsCmd.Rpc = &rpc.GetMdsRpc{Info: mdsRpc, Request: &pbmds.GetMDSListRequest{}}
	// set table header
	header := []string{cobrautil.ROW_ID, cobrautil.ROW_ADDR, cobrautil.ROW_STATE, cobrautil.ROW_LASTONLINETIME, cobrautil.ROW_ONLINE_STATE}
	mdsCmd.SetHeader(header)

	return nil
}

func (mdsCmd *MdsCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&mdsCmd.FinalDingoCmd, mdsCmd)
}

func (mdsCmd *MdsCommand) RunCommand(cmd *cobra.Command, args []string) error {
	// get rpc result
	response, errCmd := base.GetRpcResponse(mdsCmd.Rpc.Info, mdsCmd.Rpc)
	if errCmd.TypeCode() != cmderror.CODE_SUCCESS {
		return fmt.Errorf(errCmd.Message)
	}

	result := response.(*pbmds.GetMDSListResponse)
	if mdsErr := result.GetError(); mdsErr.GetErrcode() != pbmdserror.Errno_OK {
		return cmderror.MDSV2Error(mdsErr).ToError()
	}
	// fill table
	mdsInfos := result.GetMdses()
	rows := make([]map[string]string, 0)
	for _, mdsInfo := range mdsInfos {
		row := make(map[string]string)
		row[cobrautil.ROW_ID] = fmt.Sprintf("%d", mdsInfo.GetId())
		row[cobrautil.ROW_ADDR] = fmt.Sprintf("%s:%d", mdsInfo.GetLocation().GetHost(), mdsInfo.GetLocation().GetPort())
		row[cobrautil.ROW_STATE] = mdsInfo.GetState().String()
		unixTime := int64(mdsInfo.GetLastOnlineTimeMs())
		t := time.Unix(unixTime/1000, (unixTime%1000)*1000000)
		row[cobrautil.ROW_LASTONLINETIME] = t.Format("2006-01-02 15:04:05.000")
		if mdsInfo.GetIsOnline() {
			row[cobrautil.ROW_ONLINE_STATE] = cobrautil.ROW_VALUE_ONLINE
		} else {
			row[cobrautil.ROW_ONLINE_STATE] = cobrautil.ROW_VALUE_OFFLINE
		}
		rows = append(rows, row)
	}
	list := cobrautil.ListMap2ListSortByKeys(rows, mdsCmd.Header, []string{cobrautil.ROW_ID})
	mdsCmd.TableNew.AppendBulk(list)
	// to json
	res, err := output.MarshalProtoJson(result)
	if err != nil {
		return err
	}
	mapRes := res.(map[string]interface{})
	mdsCmd.Result = mapRes
	mdsCmd.Error = cmderror.ErrSuccess()

	return nil
}

func (mdsCmd *MdsCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&mdsCmd.FinalDingoCmd)
}

func NewStatusMdsCommand() *MdsCommand {
	mdsCmd := &MdsCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:     "mds",
			Short:   "get status of mds",
			Example: mdsExample,
		},
	}
	basecmd.NewFinalDingoCli(&mdsCmd.FinalDingoCmd, mdsCmd)
	return mdsCmd
}
