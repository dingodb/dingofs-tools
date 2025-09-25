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

package common

import (
	"fmt"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/spf13/cobra"
)

// check fsid and fsname
func GetFsInfoFlagValue(cmd *cobra.Command) (uint32, string, error) {
	var fsId uint32
	var fsName string
	if !cmd.Flag(config.DINGOFS_FSNAME).Changed && !cmd.Flag(config.DINGOFS_FSID).Changed {
		return 0, "", fmt.Errorf("fsname or fsid is required")
	}
	if cmd.Flag(config.DINGOFS_FSID).Changed {
		fsId = config.GetFlagUint32(cmd, config.DINGOFS_FSID)
	} else {
		fsName = config.GetFlagString(cmd, config.DINGOFS_FSNAME)
	}
	if fsId == 0 && len(fsName) == 0 {
		return 0, "", fmt.Errorf("fsname or fsid is invalid")
	}

	return fsId, fsName, nil
}
