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
 * Created Date: 2022-08-10
 * Author: chengyi (Cyber-SiKu)
 */

package cobrautil

import (
	"path/filepath"
	"strings"

	"github.com/cilium/cilium/pkg/mountinfo"
	cmderror "github.com/dingodb/dingofs-tools/internal/error"
)

const (
	DINGOFS_MOUNTPOINT_FSTYPE  = "fuse.dingofs"
	DINGOFS_MOUNTPOINT_FSTYPE2 = "fuse" //for backward compatibility
)

func GetDingoFSMountPoints() ([]*mountinfo.MountInfo, *cmderror.CmdError) {
	mountpoints, err := mountinfo.GetMountInfo()
	if err != nil {
		errMountpoint := cmderror.ErrGetMountpoint()
		errMountpoint.Format(err.Error())
		return nil, errMountpoint
	}
	retMoutpoints := make([]*mountinfo.MountInfo, 0)
	for _, m := range mountpoints {
		if m.FilesystemType == DINGOFS_MOUNTPOINT_FSTYPE || m.FilesystemType == DINGOFS_MOUNTPOINT_FSTYPE2 {
			// check if the mountpoint is a dingofs mountpoint
			retMoutpoints = append(retMoutpoints, m)
		}
	}
	return retMoutpoints, cmderror.ErrSuccess()
}

// make sure path' abs path start with mountpoint.MountPoint
func Path2DingofsPath(path string, mountpoint *mountinfo.MountInfo) string {
	path, _ = filepath.Abs(path)
	mountPoint := mountpoint.MountPoint
	root := mountpoint.Root
	dingofsPath, _ := filepath.Abs(strings.Replace(path, mountPoint, root, 1))
	return dingofsPath
}
