// Copyright (c) 2024 dingodb.com, Inc. All Rights Reserved
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

package stats

import (
	"fmt"
	"log"
	"math"
	"os"
	"time"

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	cmdCommon "github.com/dingodb/dingofs-tools/pkg/cli/command/common"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	"github.com/dingodb/dingofs-tools/proto/dingofs/proto/mds"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

type ClusterCommand struct {
	basecmd.FinalDingoCmd
	Rpc      *cmdCommon.GetFsStatsRpc
	Response *mds.GetFsStatsResponse
}

var _ basecmd.FinalDingoCmdFunc = (*ClusterCommand)(nil) // check interface

func NewClusterCommand() *cobra.Command {
	clusterCmd := &ClusterCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:   "cluster",
			Short: "show real time performance statistics of dingofs cluster",
			Example: `$ dingo stats cluster --fsname dingofs
$ dingo stats cluster --fsid 1
$ dingo stats cluster --fsid 1 --count 10
$ dingo stats cluster --fsid 1 --interval 4s`,
		},
	}
	basecmd.NewFinalDingoCli(&clusterCmd.FinalDingoCmd, clusterCmd)
	return clusterCmd.Cmd
}

func (clusterCmd *ClusterCommand) AddFlags() {
	config.AddRpcRetryTimesFlag(clusterCmd.Cmd)
	config.AddRpcTimeoutFlag(clusterCmd.Cmd)
	config.AddFsMdsAddrFlag(clusterCmd.Cmd)
	config.AddFsIdUint32OptionFlag(clusterCmd.Cmd)
	config.AddFsNameStringOptionFlag(clusterCmd.Cmd)
	config.AddIntervalOptionFlag(clusterCmd.Cmd)
	config.AddFsCountOptionalFlag(clusterCmd.Cmd)
}

func (clusterCmd *ClusterCommand) Init(cmd *cobra.Command, args []string) error {
	addrs, getAddrErr := config.GetFsMdsAddrSlice(clusterCmd.Cmd)
	if getAddrErr.TypeCode() != cmderror.CODE_SUCCESS {
		clusterCmd.Error = getAddrErr
		return fmt.Errorf(getAddrErr.Message)
	}
	//check flags values
	fsName, fsErr := cmdCommon.GetFsName(clusterCmd.Cmd)
	if fsErr != nil {
		return fsErr
	}
	//set rpc request
	request := &mds.GetFsStatsRequest{
		Fsname: &fsName,
	}
	clusterCmd.Rpc = &cmdCommon.GetFsStatsRpc{
		Request: request,
	}
	timeout := config.GetRpcTimeout(cmd)
	retrytimes := config.GetRpcRetryTimes(cmd)
	clusterCmd.Rpc.Info = basecmd.NewRpc(addrs, timeout, retrytimes, "GetFsStats")
	clusterCmd.Rpc.Info.RpcDataShow = config.GetFlagBool(clusterCmd.Cmd, config.VERBOSE)

	return nil
}

func (clusterCmd *ClusterCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&clusterCmd.FinalDingoCmd, clusterCmd)
}

func (clusterCmd *ClusterCommand) RunCommand(cmd *cobra.Command, args []string) error {
	duration := config.GetIntervalFlag(cmd)
	count := config.GetStatsCountFlagOptionFlag(cmd)
	if duration < 1*time.Second {
		duration = 1 * time.Second
	}
	clusterCmd.realTimeClusterStats(duration, count)
	return nil
}

func (clusterCmd *ClusterCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&clusterCmd.FinalDingoCmd)
}

func (clusterCmd *ClusterCommand) GetFsStatsData() (map[string]float64, error) {
	result, err := basecmd.GetRpcResponse(clusterCmd.Rpc.Info, clusterCmd.Rpc)
	if err.TypeCode() != cmderror.CODE_SUCCESS {
		return nil, err.ToError()
	}
	response := result.(*mds.GetFsStatsResponse)
	if statusCode := response.GetStatusCode(); statusCode != mds.FSStatusCode_OK {
		return nil, fmt.Errorf("GetFsStats error,errcode[%v]", statusCode)
	}
	fsStatsData := response.GetFsStatsData()
	//convert fsStatsData to map
	metricsData := make(map[string]float64)
	metricsData["readBytes"] = float64(fsStatsData.GetReadBytes())
	metricsData["readQps"] = float64(fsStatsData.GetReadQps())
	metricsData["writeBytes"] = float64(fsStatsData.GetWriteBytes())
	metricsData["writeQps"] = float64(fsStatsData.GetWriteQps())
	metricsData["s3ReadBytes"] = float64(fsStatsData.GetS3ReadBytes())
	metricsData["s3ReadQps"] = float64(fsStatsData.GetS3ReadQps())
	metricsData["s3WriteBytes"] = float64(fsStatsData.GetS3WriteBytes())
	metricsData["s3WriteQps"] = float64(fsStatsData.GetS3WriteQps())

	return metricsData, nil
}

// real time read metric data and show for cluster
func (clusterCmd *ClusterCommand) realTimeClusterStats(duration time.Duration, count uint32) {
	watcher := &statsWatcher{
		colorful: isatty.IsTerminal(os.Stdout.Fd()),
		duration: duration,
		interval: int64(duration) / 1000000000,
		count:    count,
	}
	watcher.buildClusterSchema("fo")
	watcher.formatHeader()

	var tick uint
	var start, last, current map[string]float64
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	current, _ = clusterCmd.GetFsStatsData()
	start = current
	last = current
	for {
		if tick%(uint(watcher.interval)*30) == 0 {
			fmt.Println(watcher.header)
		}
		if tick%uint(watcher.interval) == 0 {
			watcher.printDiff(start, current, false)
			start = current
		} else {
			watcher.printDiff(last, current, true)
		}
		last = current
		tick++
		<-ticker.C
		current, _ = clusterCmd.GetFsStatsData()
		//for interval > 1s,don't print the middle result for last time
		if uint(math.Ceil(float64(tick)/float64(watcher.interval))) == uint(watcher.count) { //exit
			break
		}
	}

}

func (w *statsWatcher) buildClusterSchema(schema string) {
	for _, r := range schema {
		var s section
		switch r {
		case 'f':
			s.name = "fuse"
			s.items = append(s.items, &item{"read", "readBytes", metricByte | metricCounter})
			s.items = append(s.items, &item{"ops", "readQps", metricCounter})
			s.items = append(s.items, &item{"write", "writeBytes", metricByte | metricCounter})
			s.items = append(s.items, &item{"ops", "writeQps", metricCounter})
		case 'o':
			s.name = "object"
			s.items = append(s.items, &item{"get", "s3ReadBytes", metricByte | metricCounter})
			s.items = append(s.items, &item{"ops", "s3ReadQps", metricCounter})
			s.items = append(s.items, &item{"put", "s3WriteBytes", metricByte | metricCounter})
			s.items = append(s.items, &item{"ops", "s3WriteQps", metricCounter})
		default:
			fmt.Printf("Warning: no item defined for %c\n", r)
			continue
		}
		w.sections = append(w.sections, &s)
	}
	if len(w.sections) == 0 {
		log.Fatalln("no section to watch, please check the schema string")
	}
}
