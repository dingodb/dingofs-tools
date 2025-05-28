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
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"syscall"
	"time"

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	config "github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/spf13/viper"
)

const (
	CURL_VERSION = "curl/7.54.0"
)

type LeaderMetaCache struct {
	mutex      sync.RWMutex
	leaderAddr string
}

type Metric struct {
	Addrs   []string
	SubUri  string
	timeout time.Duration
}

type MetricResult struct {
	Addr  string
	Key   string
	Value string
	Err   *cmderror.CmdError
}

var (
	leaderMetaCache *LeaderMetaCache = &LeaderMetaCache{}
)

func NewMetric(addrs []string, subUri string, timeout time.Duration) *Metric {
	return &Metric{
		Addrs:   addrs,
		SubUri:  subUri,
		timeout: timeout,
	}
}

func QueryMetric(m *Metric) (string, *cmderror.CmdError) {
	response := make(chan string, 1)
	size := len(m.Addrs)
	if size > config.MaxChannelSize() {
		size = config.MaxChannelSize()
	}
	errs := make(chan *cmderror.CmdError, size)
	for _, host := range m.Addrs {
		url := "http://" + host + m.SubUri
		go httpGet(url, m.timeout, response, errs)
	}
	var retStr string
	var vecErrs []*cmderror.CmdError
	count := 0
	for err := range errs {
		if err.Code != cmderror.CODE_SUCCESS {
			vecErrs = append(vecErrs, err)
		} else {
			retStr = <-response
			vecErrs = append(vecErrs, cmderror.ErrSuccess())
			break
		}
		count++
		if count >= len(m.Addrs) {
			// all host failed
			break
		}
	}
	retErr := cmderror.MergeCmdError(vecErrs)
	return retStr, retErr
}

func GetMetricValue(metricRet string) (string, *cmderror.CmdError) {
	kv := cobrautil.RmWitespaceStr(metricRet)
	kvVec := strings.Split(kv, ":")
	if len(kvVec) != 2 {
		err := cmderror.ErrParseMetric()
		err.Format(metricRet)
		return "", err
	}
	kvVec[1] = strings.Replace(kvVec[1], "\"", "", -1)
	return kvVec[1], cmderror.ErrSuccess()
}

func GetKeyValueFromJsonMetric(metricRet string, key string) (string, *cmderror.CmdError) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(metricRet), &data); err != nil {
		err := cmderror.ErrParseMetric()
		err.Format(metricRet)
		return "", err
	}
	return data[key].(string), cmderror.ErrSuccess()
}

// get mds leader server
func GetMdsLeader(mdsAddrs []string) (string, bool) {
	leaderMetaCache.mutex.RLock()
	if leaderMetaCache.leaderAddr != "" {
		return leaderMetaCache.leaderAddr, true
	}
	leaderMetaCache.mutex.RUnlock()
	timeout := viper.GetDuration(config.VIPER_GLOBALE_HTTPTIMEOUT)
	for _, addr := range mdsAddrs {
		addrs := []string{addr}
		statusMetric := NewMetric(addrs, config.STATUS_SUBURI, timeout)
		result, err := QueryMetric(statusMetric)
		if err.TypeCode() == cmderror.CODE_SUCCESS {
			value, err := GetMetricValue(result)
			if err.TypeCode() == cmderror.CODE_SUCCESS && value == "leader" {
				leaderMetaCache.mutex.Lock()
				leaderMetaCache.leaderAddr = addr
				leaderMetaCache.mutex.Unlock()
				return addr, true
			}
		}
	}
	return "", false
}

// get request hosts
func GetResuestHosts(reqAddrs []string) []string {
	var result []string
	if size := len(reqAddrs); size > 1 {
		// mutible host,  get leader
		leaderAddr, ok := GetMdsLeader(reqAddrs)
		if ok {
			result = append(result, leaderAddr)
		} else {
			// fail,remain origin host list
			result = reqAddrs
		}
	} else {
		// only one host
		result = reqAddrs
	}
	return result
}

func httpGet(url string, timeout time.Duration, response chan string, errs chan *cmderror.CmdError) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		interErr := cmderror.ErrHttpCreateGetRequest()
		interErr.Format(err.Error())
		errs <- interErr
	}
	// for get curl url
	req.Header.Set("User-Agent", CURL_VERSION)
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		interErr := cmderror.ErrHttpClient()
		interErr.Format(err.Error())
		errs <- interErr
	} else if resp.StatusCode != http.StatusOK {
		statusErr := cmderror.ErrHttpStatus(resp.StatusCode)
		statusErr.Format(url, resp.StatusCode)
		errs <- statusErr
	} else {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			interErr := cmderror.ErrHttpUnreadableResult()
			interErr.Format(url, err.Error())
			errs <- interErr
		}
		// get response
		response <- string(body)
		errs <- cmderror.ErrSuccess()
	}
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
