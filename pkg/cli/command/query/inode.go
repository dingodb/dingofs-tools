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

package query

import (
	"fmt"
	"github.com/dingodb/dingofs-tools/pkg/rpc"
	"slices"

	"github.com/dingodb/dingofs-tools/proto/dingofs/proto/mdsv2"

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	"github.com/spf13/cobra"
)

const (
	inodeExample = `$ dingo query inode --fsid 1 --inodeid 1024
$ dingo query inode --fsname dingofs --inodeid 1024`
)

type InodeCommand struct {
	basecmd.FinalDingoCmd
	getInodeRpc *rpc.GetInodeRpc
}

var _ basecmd.FinalDingoCmdFunc = (*InodeCommand)(nil) // check interface

func NewGetInodeCommand() *cobra.Command {
	inodeCmd := &InodeCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:     "inode",
			Short:   "query the inode of fs",
			Example: inodeExample,
		},
	}
	basecmd.NewFinalDingoCli(&inodeCmd.FinalDingoCmd, inodeCmd)
	return inodeCmd.Cmd
}

func (inodeCmd *InodeCommand) AddFlags() {
	config.AddRpcRetryTimesFlag(inodeCmd.Cmd)
	config.AddRpcRetryDelayFlag(inodeCmd.Cmd)
	config.AddRpcTimeoutFlag(inodeCmd.Cmd)
	config.AddFsMdsAddrFlag(inodeCmd.Cmd)
	config.AddFsIdUint32OptionFlag(inodeCmd.Cmd)
	config.AddFsNameStringOptionFlag(inodeCmd.Cmd)
	config.AddInodeIdRequiredFlag(inodeCmd.Cmd)
}

func (inodeCmd *InodeCommand) Init(cmd *cobra.Command, args []string) error {
	// set header
	header := []string{
		cobrautil.ROW_FS_ID, cobrautil.ROW_INODE_ID, cobrautil.ROW_LENGTH, cobrautil.ROW_TYPE, cobrautil.ROW_NLINK, cobrautil.ROW_PARENT, cobrautil.ROW_S3CHUNKINFO_CHUNKID, cobrautil.ROW_S3CHUNKINFO_OFFSET, cobrautil.ROW_S3CHUNKINFO_LENGTH, cobrautil.ROW_S3CHUNKINFO_SIZE,
	}
	inodeCmd.Header = header
	inodeCmd.SetHeader(header)
	indexFsId := slices.Index(header, cobrautil.ROW_FS_ID)
	indexInodeId := slices.Index(header, cobrautil.ROW_INODE_ID)
	indexType := slices.Index(header, cobrautil.ROW_TYPE)
	inodeCmd.TableNew.SetAutoMergeCellsByColumnIndex([]int{indexFsId, indexInodeId, indexType})

	return nil
}

func (inodeCmd *InodeCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&inodeCmd.FinalDingoCmd, inodeCmd)
}

func (inodeCmd *InodeCommand) RunCommand(cmd *cobra.Command, args []string) error {
	// set request info
	fsId, getError := rpc.GetFsId(cmd)
	if getError != nil {
		return getError
	}
	// get epoch id
	epoch, epochErr := rpc.GetFsEpochByFsId(cmd, fsId)
	if epochErr != nil {
		return epochErr
	}
	// create router
	routerErr := rpc.InitFsMDSRouter(cmd, fsId)
	if routerErr != nil {
		return routerErr
	}
	inodeId := config.GetFlagUint64(cmd, config.DINGOFS_INODEID)

	inode, err := rpc.GetInode(cmd, fsId, inodeId, 0, epoch)
	if err != nil {
		return err
	}

	tableRows := make([]map[string]string, 0)
	//TODO chunks may be get from readslice
	var chunks []*mdsv2.Chunk
	if len(chunks) == 0 {
		row := make(map[string]string)
		row[cobrautil.ROW_FS_ID] = fmt.Sprintf("%d", inode.GetFsId())
		row[cobrautil.ROW_INODE_ID] = fmt.Sprintf("%d", inode.GetIno())
		row[cobrautil.ROW_LENGTH] = fmt.Sprintf("%d", inode.GetLength())
		row[cobrautil.ROW_TYPE] = inode.GetType().String()
		row[cobrautil.ROW_NLINK] = fmt.Sprintf("%d", inode.GetNlink())
		row[cobrautil.ROW_PARENT] = fmt.Sprintf("%d", inode.GetParents())
		row[cobrautil.ROW_S3CHUNKINFO_CHUNKID] = "-"
		row[cobrautil.ROW_S3CHUNKINFO_OFFSET] = "-"
		row[cobrautil.ROW_S3CHUNKINFO_LENGTH] = "-"
		row[cobrautil.ROW_S3CHUNKINFO_SIZE] = "-"
		tableRows = append(tableRows, row)
	} else {
		rows := make([]map[string]string, 0)
		for _, chunk := range chunks {
			for _, info := range chunk.GetSlices() {
				row := make(map[string]string)
				row[cobrautil.ROW_FS_ID] = fmt.Sprintf("%d", inode.GetFsId())
				row[cobrautil.ROW_INODE_ID] = fmt.Sprintf("%d", inode.GetIno())
				row[cobrautil.ROW_LENGTH] = fmt.Sprintf("%d", inode.GetLength())
				row[cobrautil.ROW_TYPE] = inode.GetType().String()
				row[cobrautil.ROW_NLINK] = fmt.Sprintf("%d", inode.GetNlink())
				row[cobrautil.ROW_PARENT] = fmt.Sprintf("%d", inode.GetParents())
				row[cobrautil.ROW_S3CHUNKINFO_CHUNKID] = fmt.Sprintf("%d", info.GetId())
				row[cobrautil.ROW_S3CHUNKINFO_OFFSET] = fmt.Sprintf("%d", info.GetOffset())
				row[cobrautil.ROW_S3CHUNKINFO_LENGTH] = fmt.Sprintf("%d", info.GetLen())
				row[cobrautil.ROW_S3CHUNKINFO_SIZE] = fmt.Sprintf("%d", info.GetSize())
				rows = append(rows, row)
			}
		}
		tableRows = append(tableRows, rows...)
	}

	list := cobrautil.ListMap2ListSortByKeys(tableRows, inodeCmd.Header, []string{cobrautil.ROW_FS_ID, cobrautil.ROW_INODE_ID, cobrautil.ROW_S3CHUNKINFO_CHUNKID})
	inodeCmd.TableNew.AppendBulk(list)
	inodeCmd.Result = tableRows
	inodeCmd.Error = cmderror.ErrSuccess()

	return nil
}

func (inodeCmd *InodeCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&inodeCmd.FinalDingoCmd)
}
