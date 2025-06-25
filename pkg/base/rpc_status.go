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

package base

import (
	mdsv2error "github.com/dingodb/dingofs-tools/proto/dingofs/proto/error"
	"github.com/dingodb/dingofs-tools/proto/dingofs/proto/mds"
	"github.com/dingodb/dingofs-tools/proto/dingofs/proto/metaserver"
)

type MetaServerStatusChecker interface {
	GetStatusCode() metaserver.MetaStatusCode
}

type MdsStatusChecker interface {
	GetStatusCode() mds.FSStatusCode
}

type MdsV2StatusChecker interface {
	GetError() *mdsv2error.Error
}

func CheckRpcNeedRetry(result interface{}) bool {
	if checker, ok := result.(MetaServerStatusChecker); ok {
		return checker.GetStatusCode() != metaserver.MetaStatusCode_OK
	}
	if checker, ok := result.(MdsStatusChecker); ok {
		return checker.GetStatusCode() != mds.FSStatusCode_OK
	}
	if checker, ok := result.(MdsV2StatusChecker); ok {
		return checker.GetError().GetErrcode() != mdsv2error.Errno_OK
	}

	return false
}
