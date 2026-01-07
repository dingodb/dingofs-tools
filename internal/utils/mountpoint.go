/*
 * 	Copyright (c) 2024 dingodb.com Inc.
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

package utils

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cilium/cilium/pkg/mountinfo"
)

const (
	DINGOFS_MOUNTPOINT_FSTYPE  = "fuse.dingofs"
	DINGOFS_MOUNTPOINT_FSTYPE2 = "fuse" //for backward compatibility
)

func GetDingoFSMountPoints() ([]*mountinfo.MountInfo, error) {
	mountpoints, err := mountinfo.GetMountInfo()
	if err != nil {
		return nil, fmt.Errorf("get mountpoint failed.")
	}

	dingofs_mountpoints := make([]*mountinfo.MountInfo, 0)
	for _, m := range mountpoints {
		if m.FilesystemType == DINGOFS_MOUNTPOINT_FSTYPE || m.FilesystemType == DINGOFS_MOUNTPOINT_FSTYPE2 {
			// check if the mountpoint is a dingofs mountpoint
			dingofs_mountpoints = append(dingofs_mountpoints, m)
		}
	}
	return dingofs_mountpoints, nil
}

// make sure path' abs path start with mountpoint.MountPoint
func Path2DingofsPath(path string, mountpoint *mountinfo.MountInfo) string {
	path, _ = filepath.Abs(path)
	mountPoint := mountpoint.MountPoint
	root := mountpoint.Root
	dingofsPath, _ := filepath.Abs(strings.Replace(path, mountPoint, root, 1))
	return dingofsPath
}
