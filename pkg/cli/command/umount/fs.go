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

package umount

import (
	"fmt"
	"github.com/dingodb/dingofs-tools/pkg/rpc"

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	"github.com/dingodb/dingofs-tools/pkg/base"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	pbmdsv2 "github.com/dingodb/dingofs-tools/proto/dingofs/proto/mdsv2"
	"github.com/spf13/cobra"
)

type FsCommand struct {
	basecmd.FinalDingoCmd
	Rpc      *rpc.UmountFsRpc
	fsName   string
	clientId string
}

var _ basecmd.FinalDingoCmdFunc = (*FsCommand)(nil) // check interface

const (
	fsExample = `$ dingo umount fs --fsname dingofs --clientid d708b435-aeba-472b-bbcf-f9d4637aa714`
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
	config.AddRpcRetryDelayFlag(fCmd.Cmd)
	config.AddRpcTimeoutFlag(fCmd.Cmd)
	config.AddFsMdsAddrFlag(fCmd.Cmd)
	config.AddFsNameRequiredFlag(fCmd.Cmd)
	config.AddClientIdRequiredFlag(fCmd.Cmd)
}

func (fCmd *FsCommand) Init(cmd *cobra.Command, args []string) error {
	// new rpc
	mdsRpc, err := rpc.CreateNewMdsRpc(cmd, "UmountFs")
	if err != nil {
		return err
	}
	// set request info
	fCmd.fsName = config.GetFlagString(fCmd.Cmd, config.DINGOFS_FSNAME)
	fCmd.clientId = config.GetFlagString(fCmd.Cmd, config.DINGOFS_CLIENT_ID)

	request := &pbmdsv2.UmountFsRequest{
		FsName:   fCmd.fsName,
		ClientId: fCmd.clientId,
	}
	fCmd.Rpc = &rpc.UmountFsRpc{Info: mdsRpc, Request: request}

	// set header
	header := []string{cobrautil.ROW_FS_NAME, cobrautil.ROW_FS_CLIENTID, cobrautil.ROW_RESULT}
	fCmd.SetHeader(header)

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
	fmt.Println(response)
	result := response.(*pbmdsv2.UmountFsResponse)
	cmdErr := cmderror.MDSV2Error(result.GetError())

	//fill table
	row := make(map[string]string)
	row[cobrautil.ROW_FS_NAME] = fCmd.fsName
	row[cobrautil.ROW_FS_CLIENTID] = fCmd.clientId
	row[cobrautil.ROW_RESULT] = cmdErr.Message
	list := cobrautil.Map2List(row, fCmd.Header)
	fCmd.TableNew.Append(list)

	fCmd.Result = row
	fCmd.Error = cmdErr

	return nil
}

func (fCmd *FsCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&fCmd.FinalDingoCmd)
}
