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
	pbmdsv2 "github.com/dingodb/dingofs-tools/proto/dingofs/proto/mdsv2"
)

func ConvertPbPartitionTypeToString(partitionType pbmdsv2.PartitionType) string {
	switch partitionType {
	case pbmdsv2.PartitionType_MONOLITHIC_PARTITION:
		return "MONOLITHIC"
	case pbmdsv2.PartitionType_PARENT_ID_HASH_PARTITION:
		return "HASH"
	default:
		return "unknown"
	}
}

func ConvertFsExtraToString(fsExtra *pbmdsv2.FsExtra) string {
	var result string

	s3Info := fsExtra.GetS3Info()
	if s3Info != nil {
		result = fmt.Sprintf("%s/%s", s3Info.GetEndpoint(), s3Info.GetBucketname())
	}

	radosInfo := fsExtra.GetRadosInfo()
	if radosInfo != nil {
		result = fmt.Sprintf("mon_host: %s\npool_name: %s\ncluster_name: %s", radosInfo.GetMonHost(), radosInfo.GetPoolName(), radosInfo.GetClusterName())
	}

	return result
}
