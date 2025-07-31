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

package subpath

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	cmdCommon "github.com/dingodb/dingofs-tools/pkg/cli/command/common"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	"github.com/dingodb/dingofs-tools/proto/dingofs/proto/metaserver"
	"github.com/spf13/cobra"
)

type SubPathCommand struct {
	basecmd.FinalDingoCmd
	fsId          uint32 // filesystem id
	parentInodeId uint64 // directory parent inodeId
	dirInodeId    uint64 // directory inodeId
	dirName       string // directory name
	threads       uint32 // threads
}

var _ basecmd.FinalDingoCmdFunc = (*SubPathCommand)(nil) // check interface

func NewSubPathCommand() *cobra.Command {
	subPathCmd := &SubPathCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:   "subpath",
			Short: "delete sub directory in dingofs",
			Example: `$ dingo delete subpath --fsid 1 --path /path1
$ dingo delete subpath --fsname dingofs --path /path1/path2
$ dingo delete subpath --fsid 1 --path /path1 --threads 8`,
		},
	}
	basecmd.NewFinalDingoCli(&subPathCmd.FinalDingoCmd, subPathCmd)
	return subPathCmd.Cmd
}

func (subPathCmd *SubPathCommand) AddFlags() {
	config.AddRpcRetryTimesFlag(subPathCmd.Cmd)
	config.AddRpcTimeoutFlag(subPathCmd.Cmd)
	config.AddFsMdsAddrFlag(subPathCmd.Cmd)
	config.AddFsIdUint32OptionFlag(subPathCmd.Cmd)
	config.AddFsNameStringOptionFlag(subPathCmd.Cmd)
	config.AddFsPathRequiredFlag(subPathCmd.Cmd)
	config.AddThreadsOptionFlag(subPathCmd.Cmd)
}

func (subPathCmd *SubPathCommand) Init(cmd *cobra.Command, args []string) error {
	// TODO: new path instead of DINGOFS_QUOTA_PATH
	path := config.GetFlagString(subPathCmd.Cmd, config.DINGOFS_QUOTA_PATH)
	if len(path) == 0 {
		return fmt.Errorf("path is required")
	}
	fsId, fsErr := cmdCommon.GetFsId(cmd)
	if fsErr != nil {
		return fsErr
	}

	path = filepath.Clean(path)
	parentDirName := filepath.Dir(path)
	dirName := filepath.Base(path)

	parentInodeId, inodeErr := cmdCommon.GetDirPathInodeId(subPathCmd.Cmd, fsId, parentDirName)
	if inodeErr != nil {
		return inodeErr
	}

	subPathCmd.fsId = fsId
	subPathCmd.parentInodeId = parentInodeId
	subPathCmd.dirName = dirName
	subPathCmd.threads = config.GetFlagUint32(subPathCmd.Cmd, config.DINGOFS_THREADS)

	header := []string{cobrautil.ROW_RESULT, cobrautil.ROW_INODES}
	subPathCmd.Header = header
	subPathCmd.SetHeader(header)

	return nil
}

func (subPathCmd *SubPathCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&subPathCmd.FinalDingoCmd, subPathCmd)
}

func (subPathCmd *SubPathCommand) RunCommand(cmd *cobra.Command, args []string) error {
	errDeletePath := cmderror.Success()

	var deleteInodes uint64 = 0

	defer func() { // defer fill result
		rows := make([]map[string]string, 0)
		row := make(map[string]string)
		row[cobrautil.ROW_RESULT] = errDeletePath.Message
		row[cobrautil.ROW_INODES] = fmt.Sprintf("%d", deleteInodes)
		rows = append(rows, row)

		list := cobrautil.ListMap2ListSortByKeys(rows, subPathCmd.Header, []string{cobrautil.ROW_RESULT})
		subPathCmd.TableNew.AppendBulk(list)

		subPathCmd.Result = rows
		subPathCmd.Error = errDeletePath
	}()

	if strings.TrimSpace(subPathCmd.dirName) == "/" {
		errDeletePath = cmderror.ErrDeleteSubPath()
		errDeletePath.Format("root directory can not be deleted")
		return nil
	}
	// check subpath is exist
	ok, inodeId := subPathCmd.CheckPathIsExist(cmd)
	if ok {
		subPathCmd.dirInodeId = inodeId
	} else {
		errDeletePath = cmderror.ErrDeleteSubPath()
		errDeletePath.Format(fmt.Sprintf("directory %s does not exist", subPathCmd.dirName))
		return nil
	}

	summary, err := subPathCmd.DeleteDirectory(cmd, subPathCmd.fsId, subPathCmd.parentInodeId, subPathCmd.dirInodeId, subPathCmd.dirName)
	if err != nil {
		errDeletePath = cmderror.ErrDeleteSubPath()
		errDeletePath.Format(err.Error())
		return nil
	}
	deleteInodes = summary.TotalInodes

	return nil
}

func (subPathCmd *SubPathCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&subPathCmd.FinalDingoCmd)
}

func (subPathCmd *SubPathCommand) CheckPathIsExist(cmd *cobra.Command) (bool, uint64) {
	entries, entErr := cmdCommon.ListDentry(cmd, subPathCmd.fsId, subPathCmd.parentInodeId)
	if entErr != nil {
		return false, 0
	}
	for _, entry := range entries {
		if entry.GetName() == subPathCmd.dirName {
			return true, entry.GetInodeId()
		}
	}

	return false, 0
}

func (subPathCmd *SubPathCommand) DeleteDirectoryAndData(cmd *cobra.Command, fsId uint32, parentInodeId uint64, dirInodeId uint64, name string, summary *cmdCommon.Summary, concurrent chan struct{},
	ctx context.Context, cancel context.CancelFunc) error {
	var err error
	entries, entErr := cmdCommon.ListDentry(cmd, fsId, dirInodeId)
	if entErr != nil {
		return entErr
	}

	var wg sync.WaitGroup
	var errCh = make(chan error, 1)
	for _, entry := range entries {
		if entry.GetType() != metaserver.FsFileType_TYPE_DIRECTORY {
			err1 := cmdCommon.DeleteDentry(cmd, fsId, entry.GetParentInodeId(), entry.GetName(), metaserver.FsFileType_TYPE_DIRECTORY)
			if err1 != nil {
				return err1.ToError()
			}
			err2 := cmdCommon.DeleteInode(cmd, fsId, entry.GetInodeId())
			if err2 != nil {
				return err2.ToError()
			}
			err3 := cmdCommon.UpdateInodeAttr(cmd, subPathCmd.fsId, subPathCmd.parentInodeId)
			if err3 != nil {
				return err3.ToError()
			}
			log.Printf("success delete file:[%d,%s]\n", entry.GetInodeId(), entry.GetName())

			atomic.AddUint64(&summary.TotalInodes, 1)
			continue
		}

		select {
		case err := <-errCh:
			cancel()
			return err
		case <-ctx.Done():
			return fmt.Errorf("cancel delete directory for other goroutine error")
		case concurrent <- struct{}{}:
			wg.Add(1)
			go func(e *metaserver.Dentry) {
				defer wg.Done()
				deleteErr := subPathCmd.DeleteDirectoryAndData(cmd, fsId, e.GetParentInodeId(), e.GetInodeId(), e.GetName(), summary, concurrent, ctx, cancel)
				<-concurrent
				if deleteErr != nil {
					select {
					case errCh <- deleteErr:
					default:
					}
				}
			}(entry)
		default:
			if deleteErr := subPathCmd.DeleteDirectoryAndData(cmd, fsId, entry.GetParentInodeId(), entry.GetInodeId(), entry.GetName(), summary, concurrent, ctx, cancel); deleteErr != nil {
				return deleteErr
			}
		}
	}
	// wait all subdirectory deleted
	wg.Wait()

	select {
	case err = <-errCh:
	default:
		// delete self
		err1 := cmdCommon.DeleteDentry(cmd, fsId, parentInodeId, name, metaserver.FsFileType_TYPE_DIRECTORY)
		if err1 != nil {
			return err1.ToError()
		}
		err2 := cmdCommon.DeleteInode(cmd, fsId, dirInodeId)
		if err2 != nil {
			return err2.ToError()
		}
		err3 := cmdCommon.UpdateInodeAttr(cmd, subPathCmd.fsId, subPathCmd.parentInodeId)
		if err3 != nil {
			return err3.ToError()
		}
		log.Printf("success delete directory:[%d,%s]\n", dirInodeId, name)
		atomic.AddUint64(&summary.TotalInodes, 1)
	}

	return err
}

func (subPathCmd *SubPathCommand) DeleteDirectory(cmd *cobra.Command, fsId uint32, parentInodeId uint64, dirInodeId uint64, name string) (*cmdCommon.Summary, error) {
	log.Printf("start to delete directory[%s], inode[%d]", name, dirInodeId)
	summary := &cmdCommon.Summary{0, 0}
	concurrent := make(chan struct{}, subPathCmd.threads)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deleteErr := subPathCmd.DeleteDirectoryAndData(cmd, fsId, parentInodeId, dirInodeId, name, summary, concurrent, ctx, cancel)
	log.Printf("success delete directory:[%d,%s], TotalInodes[%d]", dirInodeId, name, summary.TotalInodes)
	if deleteErr != nil {
		return nil, deleteErr
	}

	return summary, nil
}
