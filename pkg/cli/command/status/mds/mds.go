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
 * Created Date: 2022-06-06
 * Author: chengyi (Cyber-SiKu)
 */

package mds

import (
	"fmt"

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	config "github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/exp/slices"
)

const (
	mdsExample = `$ dingo status mds`
)

type MdsCommand struct {
	basecmd.FinalDingoCmd
	metrics []*basecmd.Metric
	rows    []map[string]string
	health  cobrautil.ClUSTER_HEALTH_STATUS
}

var _ basecmd.FinalDingoCmdFunc = (*MdsCommand)(nil) // check interface

func NewMdsCommand() *cobra.Command {
	return NewStatusMdsCommand().Cmd
}

func (mCmd *MdsCommand) AddFlags() {
	config.AddFsMdsAddrFlag(mCmd.Cmd)
	config.AddHttpTimeoutFlag(mCmd.Cmd)
	config.AddFsMdsDummyAddrFlag(mCmd.Cmd)
}

func (mCmd *MdsCommand) Init(cmd *cobra.Command, args []string) error {
	mCmd.health = cobrautil.HEALTH_ERROR

	header := []string{cobrautil.ROW_ADDR, cobrautil.ROW_DUMMY_ADDR, cobrautil.ROW_VERSION, cobrautil.ROW_STATUS}
	mCmd.SetHeader(header)
	mCmd.TableNew.SetAutoMergeCellsByColumnIndex(cobrautil.GetIndexSlice(
		mCmd.Header, []string{cobrautil.ROW_STATUS, cobrautil.ROW_VERSION},
	))

	// set main addr
	mainAddrs, addrErr := config.GetFsMdsAddrSlice(mCmd.Cmd)
	if addrErr.TypeCode() != cmderror.CODE_SUCCESS {
		mCmd.Error = addrErr
		return fmt.Errorf(addrErr.Message)
	}

	// set dummy addr
	dummyAddrs, addrErr := config.GetFsMdsDummyAddrSlice(mCmd.Cmd)
	if addrErr.TypeCode() != cmderror.CODE_SUCCESS {
		mCmd.Error = addrErr
		return fmt.Errorf(addrErr.Message)
	}
	for _, addr := range dummyAddrs {
		// Use the dummy port to access the metric service
		timeout := viper.GetDuration(config.VIPER_GLOBALE_HTTPTIMEOUT)

		addrs := []string{addr}
		statusMetric := basecmd.NewMetric(addrs, config.STATUS_SUBURI, timeout)
		mCmd.metrics = append(mCmd.metrics, statusMetric)
		versionMetric := basecmd.NewMetric(addrs, config.VERSION_SUBURI, timeout)
		mCmd.metrics = append(mCmd.metrics, versionMetric)
	}

	for i := range mainAddrs {
		row := make(map[string]string)
		row[cobrautil.ROW_ADDR] = mainAddrs[i]
		row[cobrautil.ROW_DUMMY_ADDR] = dummyAddrs[i]
		row[cobrautil.ROW_STATUS] = cobrautil.ROW_VALUE_OFFLINE
		row[cobrautil.ROW_VERSION] = cobrautil.ROW_VALUE_UNKNOWN
		mCmd.rows = append(mCmd.rows, row)
	}

	return nil
}

func (mCmd *MdsCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&mCmd.FinalDingoCmd, mCmd)
}

func (mCmd *MdsCommand) RunCommand(cmd *cobra.Command, args []string) error {
	results := make(chan basecmd.MetricResult, config.MaxChannelSize())
	size := 0
	for _, metric := range mCmd.metrics {
		size++
		go func(m *basecmd.Metric) {
			result, err := basecmd.QueryMetric(m)
			var key string
			if m.SubUri == config.STATUS_SUBURI {
				key = "status"
			} else {
				key = "version"
			}
			var value string
			if err.TypeCode() == cmderror.CODE_SUCCESS {
				value, err = basecmd.GetMetricValue(result)
			}
			results <- basecmd.MetricResult{
				Addr:  m.Addrs[0],
				Key:   key,
				Value: value,
				Err:   err,
			}
		}(metric)
	}

	count := 0
	var errs []*cmderror.CmdError
	var recordAddrs []string
	for res := range results {
		for _, row := range mCmd.rows {
			if res.Err.TypeCode() == cmderror.CODE_SUCCESS && row[cobrautil.ROW_DUMMY_ADDR] == res.Addr {
				row[res.Key] = res.Value
			} else if res.Err.TypeCode() != cmderror.CODE_SUCCESS {
				index := slices.Index(recordAddrs, res.Addr)
				if index == -1 {
					errs = append(errs, res.Err)
					recordAddrs = append(recordAddrs, res.Addr)
				}
			}
		}
		count++
		if count >= size {
			break
		}
	}
	if len(errs) > 0 && len(errs) < len(mCmd.rows) {
		mCmd.health = cobrautil.HEALTH_WARN
	} else if len(errs) == 0 {
		mCmd.health = cobrautil.HEALTH_OK
	}
	mergeErr := cmderror.MergeCmdErrorExceptSuccess(errs)
	mCmd.Error = mergeErr
	list := cobrautil.ListMap2ListSortByKeys(mCmd.rows, mCmd.Header, []string{
		cobrautil.ROW_STATUS, cobrautil.ROW_VERSION,
	})
	mCmd.TableNew.AppendBulk(list)
	mCmd.Result = mCmd.rows
	return nil
}

func (mCmd *MdsCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&mCmd.FinalDingoCmd)
}

func NewStatusMdsCommand() *MdsCommand {
	mdsCmd := &MdsCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:     "mds",
			Short:   "get status of mds",
			Example: mdsExample,
		},
	}
	basecmd.NewFinalDingoCli(&mdsCmd.FinalDingoCmd, mdsCmd)
	return mdsCmd
}

func GetMdsStatus(caller *cobra.Command) (*interface{}, *tablewriter.Table, *cmderror.CmdError, cobrautil.ClUSTER_HEALTH_STATUS) {
	mdsCmd := NewStatusMdsCommand()
	mdsCmd.Cmd.SetArgs([]string{
		fmt.Sprintf("--%s", config.FORMAT), config.FORMAT_NOOUT,
	})
	config.AlignFlagsValue(caller, mdsCmd.Cmd, []string{config.RPCRETRYTIMES, config.RPCTIMEOUT, config.DINGOFS_MDSADDR, config.DINGOFS_MDSDUMMYADDR})
	mdsCmd.Cmd.SilenceErrors = true
	mdsCmd.Cmd.Execute()
	return &mdsCmd.Result, mdsCmd.TableNew, mdsCmd.Error, mdsCmd.health
}
