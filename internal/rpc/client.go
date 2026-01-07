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
	"context"
	"errors"
	"log"
	"time"

	"google.golang.org/grpc"

	"github.com/dingodb/dingofs-tools/internal/errno"
)

var (
	pool *ConnectionPool = NewConnectionPool()
)

type Rpc struct {
	Addrs         []string
	RpcTimeout    time.Duration
	RpcRetryTimes int32
	RpcRetryDelay time.Duration
	RpcFuncName   string
	RpcDataShow   bool
}

func NewRpc(addrs []string, timeout time.Duration, retryTimes int32, retryDelay time.Duration, dataShow bool, funcName string) *Rpc {
	return &Rpc{
		Addrs:         addrs,
		RpcTimeout:    timeout,
		RpcRetryTimes: retryTimes,
		RpcRetryDelay: retryDelay,
		RpcFuncName:   funcName,
		RpcDataShow:   dataShow,
	}
}

type RpcFunc interface {
	NewRpcClient(cc grpc.ClientConnInterface)
	Stub_Func(ctx context.Context) (interface{}, error)
}

type Result struct {
	addr   string
	err    *errno.ErrorCode
	result interface{}
}

func GetRpcResponse(rpc *Rpc, rpcFunc RpcFunc) (interface{}, *errno.ErrorCode) {
	reqAddrs := rpc.Addrs

	// start rpc request
	results := make([]Result, 0)
	for _, address := range reqAddrs {
		conn, err := pool.GetConnection(address, rpc.RpcTimeout)
		if err != nil {
			errRpc := errno.ERR_RPC_FAILED
			errRpc.E(err)
			results = append(results, Result{address, errRpc, nil})

			continue
		}

		rpcFunc.NewRpcClient(conn)
		retryTimes := rpc.RpcRetryTimes

		log.Printf("%s: start to rpc [%s],timeout[%v],retrytimes[%d]", address, rpc.RpcFuncName, rpc.RpcTimeout, retryTimes)
		for {
			ctx, _ := context.WithTimeout(context.Background(), rpc.RpcTimeout)
			res, err := rpcFunc.Stub_Func(ctx)
			if err != nil {
				if retryTimes > 0 { // rpc failed, retrying
					log.Printf("%s: fail to get rpc [%s] response, retrytimes[%d], retrying...", address, rpc.RpcFuncName, retryTimes)
					time.Sleep(rpc.RpcRetryDelay)
					retryTimes = retryTimes - 1
					continue
				} else {
					results = append(results, Result{address, errno.ERR_RPC_FAILED.E(err), nil})
					log.Printf("%s: fail to get rpc [%s] response", address, rpc.RpcFuncName)
					break
				}
			}

			// rpc ok, but return status != ok
			if CheckRpcNeedRetry(res) && retryTimes > 0 {
				log.Printf("%s: rpc [%s] return error, retrytimes[%d], retrying...", address, rpc.RpcFuncName, retryTimes)
				time.Sleep(rpc.RpcRetryDelay)
				retryTimes = retryTimes - 1
				continue
			}
			// rpc success
			results = append(results, Result{address, errno.ERR_OK, res})
			log.Printf("%s: get rpc [%s] response successfully", address, rpc.RpcFuncName)
			break
		}

		// Return Connect to Pool
		pool.PutConnection(address, conn)
		// for mds just choose one available mds
		break
	}

	// get the rpc response result
	var ret interface{}
	var vecErrs []error
	for _, res := range results {
		if res.err.GetCode() != errno.ERR_OK.GetCode() {
			vecErrs = append(vecErrs, res.err)
		} else {
			ret = res.result
			break
		}
	}

	// merge all errors
	if len(vecErrs) >= len(reqAddrs) {
		allErrors := errors.Join(vecErrs...)
		return ret, errno.ERR_RPC_FAILED.E(allErrors)
	}

	return ret, errno.ERR_OK
}
