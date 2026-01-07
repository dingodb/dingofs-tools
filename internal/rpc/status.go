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

package rpc

import (
	mdsError "github.com/dingodb/dingofs-tools/proto/dingofs/proto/error"
)

var (
	mdsRetryErrors = map[mdsError.Errno]bool{
		mdsError.Errno_EREQUEST_FULL:      true,
		mdsError.Errno_EREDIRECT:          true,
		mdsError.Errno_ENOT_SERVE:         true,
		mdsError.Errno_EPARTIAL_SUCCESS:   true,
		mdsError.Errno_ESTORE_MAYBE_RETRY: true,
	}
)

type MdsStatusChecker interface {
	GetError() *mdsError.Error
}

func CheckRpcNeedRetry(result interface{}) bool {
	// check mds retry errors
	if checker, ok := result.(MdsStatusChecker); ok {
		errCode := checker.GetError().GetErrcode()
		if _, exists := mdsRetryErrors[errCode]; exists {
			return true
		}
	}

	return false
}
