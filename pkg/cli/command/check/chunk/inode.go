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

package chunk

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	cmdcommon "github.com/dingodb/dingofs-tools/pkg/cli/command/common"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	"github.com/dingodb/dingofs-tools/proto/dingofs/proto/metaserver"
	"github.com/spf13/cobra"
)

const (
	chunkExample = `
	# check entire file system
	$ dingo check chunk --fsid 1

	# check file chunks for directory inodeid 1024
	$ dingo check chunk --fsid 1 --inodeid 1024

	# check with 16 threads	
	$ dingo check chunk --fsid 1 --threads 16`
)

// check result
type Summary struct {
	Dirs        uint64
	Files       uint64
	Chunks      uint64
	ErrorChunks uint64
	mux         sync.RWMutex
}

type S3ChunkInfo struct {
	ChunkIndex int
	ChunkInfo  *metaserver.S3ChunkInfo
}

type ChunkCommand struct {
	basecmd.FinalDingoCmd
}

var _ basecmd.FinalDingoCmdFunc = (*ChunkCommand)(nil) // check interface

func NewChunkCommand() *cobra.Command {
	inodeCmd := &ChunkCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:     "chunk",
			Short:   "check all the file chunks under directory",
			Example: chunkExample,
		},
	}
	basecmd.NewFinalDingoCli(&inodeCmd.FinalDingoCmd, inodeCmd)
	return inodeCmd.Cmd
}

func (chunkCmd *ChunkCommand) AddFlags() {
	config.AddRpcRetryTimesFlag(chunkCmd.Cmd)
	config.AddRpcTimeoutFlag(chunkCmd.Cmd)
	config.AddFsMdsAddrFlag(chunkCmd.Cmd)
	config.AddFsIdRequiredFlag(chunkCmd.Cmd)
	config.AddInodeIdOptionalFlag(chunkCmd.Cmd)
	config.AddThreadsOptionFlag(chunkCmd.Cmd)
}

func (chunkCmd *ChunkCommand) Init(cmd *cobra.Command, args []string) error {
	return nil
}

func (chunkCmd *ChunkCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&chunkCmd.FinalDingoCmd, chunkCmd)
}

func (chunkCmd *ChunkCommand) RunCommand(cmd *cobra.Command, args []string) error {

	fsId := config.GetFlagUint32(cmd, config.DINGOFS_FSID)
	inodeId := config.GetFlagUint64(cmd, config.DINGOFS_INODEID)
	if inodeId == 0 {
		inodeId = config.ROOTINODEID
	}
	threads := config.GetThreadsOptionFlag(cmd)

	err := chunkCmd.CheckAllChunks(cmd, fsId, inodeId, threads)
	if err != nil {
		return err
	}

	return nil
}

func (chunkCmd *ChunkCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&chunkCmd.FinalDingoCmd)
}

// check all files chunks under directory
func (chunkCmd *ChunkCommand) CheckAllChunks(cmd *cobra.Command, fsId uint32, dirInode uint64, threads uint32) error {
	fmt.Printf("%s: check all file chunks under dirinode[%d]\n", time.Now().Format("2006-01-02 15:04:05.000"), dirInode)

	summary := &Summary{0, 0, 0, 0, sync.RWMutex{}}
	concurrent := make(chan struct{}, int(threads))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := chunkCmd.CheckAllChunksParallel(cmd, fsId, dirInode, summary, concurrent, ctx, cancel)
	fmt.Printf("%s: check over, dirinode[%d], directories[%d], files[%d], chunks[%d], errorChunks[%d]\n", time.Now().Format("2006-01-02 15:04:05.000"), dirInode, summary.Dirs, summary.Files, summary.Chunks, summary.ErrorChunks)
	if err != nil {
		return err
	}

	return nil
}

// parallel check all chunks in a directory
func (chunkCmd *ChunkCommand) CheckAllChunksParallel(cmd *cobra.Command, fsId uint32, inode uint64, summary *Summary, concurrent chan struct{},
	ctx context.Context, cancel context.CancelFunc) error {
	var err error
	entries, entErr := cmdcommon.ListDentry(cmd, fsId, inode)
	if entErr != nil {
		return entErr
	}
	var wg sync.WaitGroup
	var errCh = make(chan error, 1)
	for _, entry := range entries {
		if entry.GetType() == metaserver.FsFileType_TYPE_S3 || entry.GetType() == metaserver.FsFileType_TYPE_FILE {
			atomic.AddUint64(&summary.Files, 1)

			inode, inodeErr := cmdcommon.GetInode(cmd, fsId, entry.GetInodeId())
			if inodeErr != nil {
				return inodeErr
			}

			chunkIdMap := make(map[uint64][]*S3ChunkInfo, 0)
			if len(inode.S3ChunkInfoMap) != 0 {
				infoMap := inode.GetS3ChunkInfoMap()
				for _, infoList := range infoMap {
					for chunk_index, info := range infoList.GetS3Chunks() {
						atomic.AddUint64(&summary.Chunks, 1)
						chunkId := info.GetChunkId()
						chunkIdMap[chunkId] = append(chunkIdMap[chunkId], &S3ChunkInfo{ChunkIndex: chunk_index, ChunkInfo: info})
					}
				}
				//print result
				summary.mux.Lock()
				for chunkId, infoList := range chunkIdMap {
					if len(infoList) > 1 { // duplicate chunkid
						atomic.AddUint64(&summary.ErrorChunks, 1)
						fmt.Printf("- fsid: [%d] inodeId: [%d] name: [%s] duplicate chunkid: [%d] \n", inode.GetFsId(), inode.GetInodeId(), entry.GetName(), chunkId)
						for _, info := range infoList {
							fmt.Printf("	chunkIndex:%v	%v\n", info.ChunkIndex, info.ChunkInfo)
						}
					}
				}
				summary.mux.Unlock()
			}

			continue
		} else { //FsFileType_TYPE_DIRECTORY
			atomic.AddUint64(&summary.Dirs, 1)
		}

		select {
		case err := <-errCh:
			cancel()
			return err
		case <-ctx.Done():
			return fmt.Errorf("cancel scan directory for other goroutine error")
		case concurrent <- struct{}{}:
			wg.Add(1)
			go func(e *metaserver.Dentry) {
				defer wg.Done()
				checkErr := chunkCmd.CheckAllChunksParallel(cmd, fsId, e.GetInodeId(), summary, concurrent, ctx, cancel)
				<-concurrent
				if checkErr != nil {
					select {
					case errCh <- checkErr:
					default:
					}
				}
			}(entry)
		default:
			if checkErr := chunkCmd.CheckAllChunksParallel(cmd, fsId, entry.GetInodeId(), summary, concurrent, ctx, cancel); checkErr != nil {
				return checkErr
			}
		}
	}
	wg.Wait()
	select {
	case err = <-errCh:
	default:
	}
	return err
}
