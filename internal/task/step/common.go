/*
 * Copyright (c) 2026 dingodb.com, Inc. All Rights Reserved
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

package step

import (
	"github.com/dingodb/dingocli/internal/errno"
	"github.com/dingodb/dingocli/internal/task/context"
	"github.com/dingodb/dingocli/internal/utils"
	"github.com/dingodb/dingocli/pkg/module"
)

type (
	LambdaType func(ctx *context.Context) error

	Lambda struct {
		Lambda LambdaType
	}
)

func (s *Lambda) Execute(ctx *context.Context) error {
	return s.Lambda(ctx)
}

func PostHandle(Success *bool, Out *string, out string, err error, ec *errno.ErrorCode) error {
	if Out != nil {
		*Out = utils.TrimSuffixRepeat(out, "\n")
	}

	if Success != nil { // handle error by user
		*Success = (err == nil)
		return nil
	} else if err == nil { // execute success
		return nil
	}

	// execute timed out
	if _, ok := err.(*module.TimeoutError); ok {
		return errno.ERR_EXECUTE_COMMAND_TIMED_OUT.S(ec.GetDescription())
	}

	// execute failed
	if ec == nil {
		ec = errno.ERR_UNKNOWN
	}
	if len(out) > 0 {
		return ec.S(out)
	}
	return ec.E(err)
}
