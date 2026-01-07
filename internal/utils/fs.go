/*
 *  Copyright (c) 2021 NetEase Inc.
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

/*
 * Project: CurveAdm
 * Created Date: 2021-12-19
 * Author: Jingli Chen (Wine93)
 *
 * Project: dingoadm
 * Author: dongwei (jackblack369)
 */

package utils

import (
	"fmt"
	"os"
	"path"
	"strings"
	"syscall"
)

func CheckMountPoint(mountPoint string) error {
	if !PathExist(mountPoint) {
		return fmt.Errorf("%s: path not exist", mountPoint)
	} else if !path.IsAbs(mountPoint) {
		return fmt.Errorf("%s: is not an absolute path", mountPoint)
	}
	return nil
}

// get mountPoint inode
func GetFileInode(path string) (uint64, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	if sst, ok := fi.Sys().(*syscall.Stat_t); ok {
		return sst.Ino, nil
	}
	return 0, nil
}

func GetInodesAsString(listFilePath string) (string, error) {
	content, err := os.ReadFile(listFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file list: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	var inodeStrings []string

	for _, line := range lines {
		filePath := strings.TrimSpace(line)
		if filePath == "" {
			continue
		}

		if !strings.HasPrefix(filePath, "/") {
			return "", fmt.Errorf("filelist[%s] content error, each line requires a full path name", listFilePath)
		}

		inodeId, err2 := GetFileInode(filePath)
		if err2 != nil {
			return "", fmt.Errorf("%s not exist", filePath)
		}
		if inodeId == 0 {
			continue
		}
		inodeStrings = append(inodeStrings, fmt.Sprintf("%d", inodeId))
	}

	inodeStrings = RemoveDuplicates(inodeStrings)

	return strings.Join(inodeStrings, ","), nil
}
