/*
 * Copyright (c) 2025 dingodb.com, Inc. All Rights Reserved
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package warmup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dingodb/dingocli/cli/cli"
	"github.com/dingodb/dingocli/internal/output"
	"github.com/dingodb/dingocli/internal/utils"
	"golang.org/x/sys/unix"

	"github.com/spf13/cobra"
)

const (
	WARMUP_ADD_EXAMPLE = `Examples:
   # warmup all files in warmup.lst,file must in dingofs
   $ dingo warmup add --filelist /mnt/warmup.lst

   # warmup one file
   $ dingo warmup add /mnt/bigfile.bin

   # warmup all files in directory dir1
   $ dingo warmup add /mnt/dir1`
)

type addOptions struct {
	filepath string
	daemon   bool
	single   bool
}

func NewWarmupAddCommand(dingocli *cli.DingoCli) *cobra.Command {
	var options addOptions

	cmd := &cobra.Command{
		Use:     "add [OPTIONS]",
		Short:   "Tell client to warmup files(directories) to local cache",
		Args:    utils.ExactArgs(1),
		Example: WARMUP_ADD_EXAMPLE,
		RunE: func(cmd *cobra.Command, args []string) error {
			options.daemon = utils.GetBoolFlag(cmd, utils.DINGOFS_DAEMON)

			filelist := utils.GetStringFlag(cmd, utils.DINGOFS_FILELIST)

			if filelist == "" && len(args) == 0 {
				return fmt.Errorf("no warmup file is specified")
			} else if filelist != "" {
				options.filepath = filelist
				options.single = false

			} else {
				options.filepath = args[0]
				options.single = true
			}

			output.SetShow(utils.GetBoolFlag(cmd, utils.VERBOSE))

			return runAdd(cmd, dingocli, options)
		},
		SilenceUsage:          false,
		DisableFlagsInUseLine: true,
	}

	utils.SetFlagErrorFunc(cmd)

	// add flags
	utils.AddBoolFlag(cmd, utils.DINGOFS_DAEMON, "Run in background")
	utils.AddStringFlag(cmd, utils.DINGOFS_FILELIST, "Full path of file, save the files(dir) to warmup, and should be in dingofs")

	return cmd
}

func runAdd(cmd *cobra.Command, dingocli *cli.DingoCli, options addOptions) error {

	// check has dingofs mountpoint
	mountpoints, err := utils.GetDingoFSMountPoints()
	if err != nil {
		return err
	} else if len(mountpoints) == 0 {
		return fmt.Errorf("no dingofs mountpoint found")
	}

	options.filepath, _ = filepath.Abs(options.filepath)
	options.filepath = filepath.Clean(options.filepath)

	// check file is exist
	info, errStat := os.Stat(options.filepath)
	if errStat != nil {
		if os.IsNotExist(errStat) {
			return fmt.Errorf("[%s]: no such file or directory", options.filepath)
		} else {
			return fmt.Errorf("stat [%s] fail: %v", options.filepath, errStat)
		}
	} else if !options.single && info.IsDir() {
		// --filelist must be a file
		return fmt.Errorf("[%s]: must be a file", options.filepath)
	}

	// check file is in dingofs
	var isInDingofs bool = false
	for _, mountpoint := range mountpoints {
		if strings.HasPrefix(options.filepath, mountpoint.MountPoint) {
			isInDingofs = true
			break
		}
	}
	if !isInDingofs {
		return fmt.Errorf("[%s] is not saved in dingofs", options.filepath)
	}

	// warmup file
	var inodesStr string
	if options.single {
		inodeId, err := utils.GetFileInode(options.filepath)
		if err != nil {
			return err
		}
		inodesStr = fmt.Sprintf("%d", inodeId)
	} else {
		inodes, err := utils.GetInodesAsString(options.filepath)
		if err != nil {
			return err
		}
		inodesStr = inodes
	}

	err = unix.Setxattr(options.filepath, DINGOFS_WARMUP_OP_XATTR, []byte(inodesStr), 0)
	if err == unix.ENOTSUP || err == unix.EOPNOTSUPP {
		return fmt.Errorf("filesystem does not support extended attributes")
	} else if err != nil {
		return fmt.Errorf("%s: %v", DINGOFS_WARMUP_OP_XATTR, err)
	}
	if !options.daemon {
		time.Sleep(1 * time.Second) //wait for 1s
		options := queryOptions{
			path: options.filepath,
		}
		runQuery(cmd, dingocli, options)
	} else {
		fmt.Printf("Successfully run warmup in background, you can run \"dingocli warmup query %s\" to query progress\n", options.filepath)
	}

	return nil
}
