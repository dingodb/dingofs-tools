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

var (
	metaServerRetryErrors = map[metaserver.MetaStatusCode]bool{
		metaserver.MetaStatusCode_UNKNOWN_ERROR:    true,
		metaserver.MetaStatusCode_RPC_ERROR:        true,
		metaserver.MetaStatusCode_REDIRECTED:       true,
		metaserver.MetaStatusCode_OVERLOAD:         true,
		metaserver.MetaStatusCode_RPC_STREAM_ERROR: true,
	}
	mdsRetryErrors = map[mds.FSStatusCode]bool{
		mds.FSStatusCode_UNKNOWN_ERROR: true,
		mds.FSStatusCode_RPC_ERROR:     true,
		mds.FSStatusCode_FS_BUSY:       true,
		mds.FSStatusCode_LOCK_TIMEOUT:  true,
	}
	mdsV2RetryErrors = map[mdsv2error.Errno]bool{
		mdsv2error.Errno_EREQUEST_FULL:      true,
		mdsv2error.Errno_EGEN_FSID:          true,
		mdsv2error.Errno_EREDIRECT:          true,
		mdsv2error.Errno_ENOT_SERVE:         true,
		mdsv2error.Errno_EPARTIAL_SUCCESS:   true,
		mdsv2error.Errno_ESTORE_MAYBE_RETRY: true,
	}
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
	// check metaServer retry errors
	if checker, ok := result.(MetaServerStatusChecker); ok {
		statusCode := checker.GetStatusCode()
		if _, exists := metaServerRetryErrors[statusCode]; exists {
			return true
		}
	}

	// check mds retry errors
	if checker, ok := result.(MdsStatusChecker); ok {
		statusCode := checker.GetStatusCode()
		if _, exists := mdsRetryErrors[statusCode]; exists {
			return true
		}
	}

	// check mdsV2 retry errors
	if checker, ok := result.(MdsV2StatusChecker); ok {
		errCode := checker.GetError().GetErrcode()
		if _, exists := mdsV2RetryErrors[errCode]; exists {
			return true
		}
	}

	return false
}
