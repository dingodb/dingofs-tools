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
 * Created Date: 2022-05-11
 * Author: chengyi (Cyber-SiKu)
 */

package cmderror

import (
	"fmt"
	"github.com/dingodb/dingofs-tools/proto/dingofs/proto/cachegroup"
	pbmdsv2error "github.com/dingodb/dingofs-tools/proto/dingofs/proto/error"
)

// It is considered here that the importance of the error is related to the
// code, and the smaller the code, the more important the error is.
// Need to ensure that the smaller the code, the more important the error is
const (
	CODE_BASE_LINE   = 10000
	CODE_SUCCESS     = 0 * CODE_BASE_LINE
	CODE_RPC_RESULT  = 1 * CODE_BASE_LINE
	CODE_HTTP_RESULT = 2 * CODE_BASE_LINE
	CODE_RPC         = 3 * CODE_BASE_LINE
	CODE_HTTP        = 4 * CODE_BASE_LINE
	CODE_INTERNAL    = 9 * CODE_BASE_LINE
	CODE_UNKNOWN     = 10 * CODE_BASE_LINE
)

type CmdError struct {
	Code    int    `json:"code"`    // exit code
	Message string `json:"message"` // exit message
}

var (
	AllError []*CmdError
)

func init() {
	AllError = make([]*CmdError, 0)
}

func (ce *CmdError) ToError() error {
	if ce == nil {
		return nil
	}
	if ce.Code == CODE_SUCCESS {
		return nil
	}
	return fmt.Errorf(ce.Message)
}

func NewSucessCmdError() *CmdError {
	ret := &CmdError{
		Code:    CODE_SUCCESS,
		Message: "success",
	}
	AllError = append(AllError, ret)
	return ret
}

func NewInternalCmdError(code int, message string) *CmdError {
	if code == 0 {
		return NewSucessCmdError()
	}
	ret := &CmdError{
		Code:    CODE_INTERNAL + code,
		Message: message,
	}

	AllError = append(AllError, ret)
	return ret
}

func NewRpcError(code int, message string) *CmdError {
	if code == 0 {
		return NewSucessCmdError()
	}
	ret := &CmdError{
		Code:    CODE_RPC + code,
		Message: message,
	}
	AllError = append(AllError, ret)
	return ret
}

func NewRpcReultCmdError(code int, message string) *CmdError {
	if code == 0 {
		return NewSucessCmdError()
	}
	ret := &CmdError{
		Code:    CODE_RPC_RESULT + code,
		Message: message,
	}
	AllError = append(AllError, ret)
	return ret
}

func NewMdsV2RpcReultCmdError(code int, message string) *CmdError {
	if code == 0 {
		return NewSucessCmdError()
	}
	ret := &CmdError{
		Code:    code,
		Message: message,
	}
	AllError = append(AllError, ret)
	return ret
}

func NewHttpError(code int, message string) *CmdError {
	if code == 0 {
		return NewSucessCmdError()
	}
	ret := &CmdError{
		Code:    CODE_HTTP + code,
		Message: message,
	}
	AllError = append(AllError, ret)
	return ret
}

func NewHttpResultCmdError(code int, message string) *CmdError {
	if code == 0 {
		return NewSucessCmdError()
	}
	ret := &CmdError{
		Code:    CODE_HTTP_RESULT + code,
		Message: message,
	}
	AllError = append(AllError, ret)
	return ret
}

func (cmd CmdError) TypeCode() int {
	return cmd.Code / CODE_BASE_LINE * CODE_BASE_LINE
}

func (cmd CmdError) TypeName() string {
	var ret string
	switch cmd.TypeCode() {
	case CODE_SUCCESS:
		ret = "success"
	case CODE_INTERNAL:
		ret = "internal"
	case CODE_RPC:
		ret = "rpc"
	case CODE_RPC_RESULT:
		ret = "rpcResult"
	case CODE_HTTP:
		ret = "http"
	case CODE_HTTP_RESULT:
		ret = "httpResult"
	default:
		ret = "unknown"
	}
	return ret
}

func (e *CmdError) Format(args ...interface{}) {
	e.Message = fmt.Sprintf(e.Message, args...)
}

// The importance of the error is considered to be related to the code,
// please use it under the condition that the smaller the code,
// the more important the error is.
func MostImportantCmdError(err []*CmdError) *CmdError {
	if len(err) == 0 {
		return NewSucessCmdError()
	}
	ret := err[0]
	for _, e := range err {
		if e.Code < ret.Code {
			ret = e
		}
	}
	return ret
}

// keep the most important wrong id, all wrong message will be kept
// if all success return success
func MergeCmdErrorExceptSuccess(err []*CmdError) *CmdError {
	if len(err) == 0 {
		return NewSucessCmdError()
	}
	var ret CmdError
	ret.Code = CODE_UNKNOWN
	ret.Message = ""
	countSuccess := 0
	for _, e := range err {
		if e != nil {
			if e.Code == CODE_SUCCESS {
				countSuccess++
				continue
			} else if e.Code < ret.Code {
				ret.Code = e.Code
			}
			ret.Message = e.Message + "\n" + ret.Message
		}
	}
	if countSuccess == len(err) {
		return NewSucessCmdError()
	}
	ret.Message = ret.Message[:len(ret.Message)-1]
	return &ret
}

// keep the most important wrong id, all wrong message will be kept
// if have one success return success
func MergeCmdError(err []*CmdError) *CmdError {
	if len(err) == 0 {
		return NewSucessCmdError()
	}
	var ret CmdError
	ret.Code = CODE_UNKNOWN
	ret.Message = ""
	for _, e := range err {
		if e.Code == CODE_SUCCESS {
			return e
		} else if e.Code < ret.Code {
			ret.Code = e.Code
		}
		ret.Message = e.Message + "\n" + ret.Message
	}
	ret.Message = ret.Message[:len(ret.Message)-1]
	return &ret
}

var (
	ErrSuccess = NewSucessCmdError
	Success    = ErrSuccess

	// internal error
	ErrHttpCreateGetRequest = func() *CmdError {
		return NewInternalCmdError(1, "create http get request failed, the error is: %s")
	}
	ErrDataNoExpected = func() *CmdError {
		return NewInternalCmdError(2, "data: %v is not as expected, the error is: %s")
	}
	ErrHttpClient = func() *CmdError {
		return NewInternalCmdError(3, "http client get error: %s")
	}
	ErrRpcDial = func() *CmdError {
		return NewInternalCmdError(4, "dial to rpc server %s failed, the error is: %s")
	}
	ErrUnmarshalJson = func() *CmdError {
		return NewInternalCmdError(5, "unmarshal json error, the json is %s, the error is %s")
	}
	ErrParseMetric = func() *CmdError {
		return NewInternalCmdError(6, "parse metric %s err!")
	}
	ErrGetAddr = func() *CmdError {
		return NewInternalCmdError(7, "invalid %s addr is: %s")
	}
	ErrMarShalProtoJson = func() *CmdError {
		return NewInternalCmdError(8, "marshal proto to json error, the error is: %s")
	}
	ErrSplitMountpoint = func() *CmdError {
		return NewInternalCmdError(9, "invalid mountpoint[%s], should be like: hostname:port:path")
	}
	ErrGetMountpoint = func() *CmdError {
		return NewInternalCmdError(10, "get mountpoint failed! the error is: %s")
	}
	ErrSetxattr = func() *CmdError {
		return NewInternalCmdError(11, "setxattr [%s] failed! the error is: %s")
	}
	ErrGettimeofday = func() *CmdError {
		return NewInternalCmdError(12, "get time of day fail, the error is: %s")
	}
	ErrQueryWarmup = func() *CmdError {
		return NewInternalCmdError(13, "query warmup progress fail, err: %s")
	}
	ErrGetFsUsage = func() *CmdError {
		return NewInternalCmdError(14, "get the usage of the file system fail, err: %s")
	}
	ErrDeleteSubPath = func() *CmdError {
		return NewInternalCmdError(15, "delete sub path fail, err: %s")
	}
	// http error
	ErrHttpUnreadableResult = func() *CmdError {
		return NewHttpResultCmdError(1, "http response is unreadable, the uri is: %s, the error is: %s")
	}
	ErrHttpStatus = func(statusCode int) *CmdError {
		return NewHttpError(statusCode, "the url is: %s, http status code is: %d")
	}

	// rpc error
	ErrRpcCall = func() *CmdError {
		return NewRpcReultCmdError(1, "rpc[%s] is fail, the error is: %s")
	}

	ErrGetFsInfo = func(statusCode int) *CmdError {
		return NewRpcReultCmdError(statusCode, "get fs info failed: status code is %s")
	}

	MDSV2Error = func(mds_error *pbmdsv2error.Error) *CmdError {
		var message string
		code := mds_error.GetErrcode()
		switch code {
		case pbmdsv2error.Errno_OK:
			message = "success"
		default:
			message = fmt.Sprintf("error: %s, errmsg: %s", code.String(), mds_error.Errmsg)
		}
		return NewMdsV2RpcReultCmdError(int(code), message)
	}

	ErrDingoCacheRequest = func(code cachegroup.CacheGroupErrCode) *CmdError {
		var message string
		switch code {
		case cachegroup.CacheGroupErrCode_CacheGroupOk:
			message = "success"
		default:
			message = fmt.Sprintf("dingoCache response error, error is %s", code.String())
		}
		return NewRpcReultCmdError(int(code), message)
	}
)
