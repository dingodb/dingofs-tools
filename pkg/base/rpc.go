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
	"context"
	"log"
	"time"

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	config "github.com/dingodb/dingofs-tools/pkg/config"
	"google.golang.org/grpc"
)

var (
	pool *ConnectionPool = NewConnectionPool()
)

type Rpc struct {
	Addrs         []string
	RpcTimeout    time.Duration
	RpcRetryTimes int32
	RpcFuncName   string
	RpcDataShow   bool
}

// TODO field RpcDataShow may be pass by parameter
func NewRpc(addrs []string, timeout time.Duration, retryTimes int32, funcName string) *Rpc {
	return &Rpc{
		Addrs:         addrs,
		RpcTimeout:    timeout,
		RpcRetryTimes: retryTimes,
		RpcFuncName:   funcName,
		RpcDataShow:   false,
	}
}

type RpcFunc interface {
	NewRpcClient(cc grpc.ClientConnInterface)
	Stub_Func(ctx context.Context) (interface{}, error)
}

type Result struct {
	addr   string
	err    *cmderror.CmdError
	result interface{}
}

func GetRpcResponse(rpc *Rpc, rpcFunc RpcFunc) (interface{}, *cmderror.CmdError) {
	var reqAddrs []string
	if config.MDSApiV2 {
		reqAddrs = rpc.Addrs
	} else {
		reqAddrs = GetResuestHosts(rpc.Addrs)
	}
	// start rpc request
	results := make([]Result, 0)
	for _, address := range reqAddrs {
		conn, err := pool.GetConnection(address, rpc.RpcTimeout)
		if err != nil {
			errDial := cmderror.ErrRpcDial()
			errDial.Format(address, err.Error())
			results = append(results, Result{address, errDial, nil})
		} else {
			rpcFunc.NewRpcClient(conn)
			retryTimes := rpc.RpcRetryTimes
			for {
				log.Printf("%s: start to rpc [%s],timeout[%v],retrytimes[%d]", address, rpc.RpcFuncName, rpc.RpcTimeout, retryTimes)
				ctx, _ := context.WithTimeout(context.Background(), rpc.RpcTimeout)
				res, err := rpcFunc.Stub_Func(ctx)
				retryTimes = retryTimes - 1
				if err != nil {
					if retryTimes > 0 {
						log.Printf("%s: fail to get rpc [%s] response,retrying...", address, rpc.RpcFuncName)
						continue
					} else {
						errRpc := cmderror.ErrRpcCall()
						errRpc.Format(rpc.RpcFuncName, err.Error())
						results = append(results, Result{address, errRpc, nil})
						log.Printf("%s: fail to get rpc [%s] response", address, rpc.RpcFuncName)
						break
					}
				} else {
					results = append(results, Result{address, cmderror.ErrSuccess(), res})
					log.Printf("%s: get rpc [%s] response successfully", address, rpc.RpcFuncName)
					break
				}
			}
			pool.PutConnection(address, conn)
			if config.MDSApiV2 { // mdsv2 just choose one available mds
				break
			}
		}
	}
	// get the rpc response result
	var ret interface{}
	var vecErrs []*cmderror.CmdError
	for _, res := range results {
		if res.err.TypeCode() != cmderror.CODE_SUCCESS {
			vecErrs = append(vecErrs, res.err)
		} else {
			ret = res.result
			break
		}
	}
	if len(vecErrs) >= len(reqAddrs) {
		retErr := cmderror.MergeCmdError(vecErrs)
		return ret, retErr
	}

	return ret, cmderror.ErrSuccess()
}

type RpcResult struct {
	position int
	Response interface{}
	Error    *cmderror.CmdError
}

func GetRpcListResponse(rpcList []*Rpc, rpcFunc []RpcFunc) ([]interface{}, []*cmderror.CmdError) {
	results := make([]RpcResult, 0)
	for i := range rpcList {
		res, err := GetRpcResponse(rpcList[i], rpcFunc[i])
		results = append(results, RpcResult{i, res, err})
	}

	retRes := make([]interface{}, 0)
	var vecErrs []*cmderror.CmdError
	for i := 0; i < len(results); i++ {
		res := results[i]
		if res.Error.TypeCode() != cmderror.CODE_SUCCESS {
			// get fail
			vecErrs = append(vecErrs, res.Error)
		} else {
			// success
			retRes = append(retRes, res.Response)
		}
	}

	return retRes, vecErrs
}
