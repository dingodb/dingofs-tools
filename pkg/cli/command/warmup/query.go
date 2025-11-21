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

package warmup

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	"github.com/dustin/go-humanize"
	"github.com/gookit/color"
	"github.com/pkg/xattr"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

const (
	queryExample = `$ dingo warmup query /mnt/warmup `
)

type QueryCommand struct {
	basecmd.FinalDingoCmd
	path     string
	interval time.Duration
}

var _ basecmd.FinalDingoCmdFunc = (*QueryCommand)(nil) // check interface

func NewQueryWarmupCommand() *QueryCommand {
	qCmd := &QueryCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:     "query",
			Short:   "query the warmup progress",
			Example: queryExample,
		},
	}
	basecmd.NewFinalDingoCli(&qCmd.FinalDingoCmd, qCmd)
	return qCmd
}

func NewQueryCommand() *cobra.Command {
	return NewQueryWarmupCommand().Cmd
}

func (qCmd *QueryCommand) AddFlags() {
	config.AddIntervalOptionFlag(qCmd.Cmd)
	config.AddDaemonOptionPFlag(qCmd.Cmd)
}

func (qCmd *QueryCommand) Init(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("need a path")
	}
	qCmd.path = args[0]
	qCmd.interval = config.GetIntervalFlag(qCmd.Cmd)
	return nil
}

func (qCmd *QueryCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&qCmd.FinalDingoCmd, qCmd)
}

func (qCmd *QueryCommand) RunCommand(cmd *cobra.Command, args []string) error {

	isDaemon := config.GetDaemonFlag(qCmd.Cmd)
	qCmd.Logger.Infof("query warmup progress, file: %s", qCmd.path)
	filename := strings.Split(qCmd.path, "/")
	var bar *progressbar.ProgressBar
	if !isDaemon {
		bar = progressbar.NewOptions64(1,
			progressbar.OptionSetDescription("[cyan]Warmup[reset] "+filename[len(filename)-1]+"..."),
			progressbar.OptionShowCount(),
			progressbar.OptionShowIts(),
			progressbar.OptionSpinnerType(14),
			progressbar.OptionFullWidth(),
			progressbar.OptionThrottle(65*time.Millisecond),
			progressbar.OptionSetRenderBlankState(true),
			progressbar.OptionOnCompletion(func() {
				fmt.Fprint(os.Stderr, "\n")
			}),
			progressbar.OptionEnableColorCodes(true),
			progressbar.OptionSetTheme(progressbar.Theme{
				Saucer:        "[green]=[reset]",
				SaucerHead:    "[green]>[reset]",
				SaucerPadding: " ",
				BarStart:      "[",
				BarEnd:        "]",
			}))
	}
	var warmErrors uint64 = 0
	var finished uint64 = 0
	var total uint64 = 0
	var resultStr string
	for {
		// result data format [finished/total/errors]
		qCmd.Logger.Debugf("get warmup xattr")
		result, err := xattr.Get(qCmd.path, DINGOFS_WARMUP_OP_XATTR)
		if err != nil {
			return err
		}
		resultStr = string(result)
		qCmd.Logger.Debugf("warmup xattr: [%s],[finished/total/errors]", resultStr)
		strs := strings.Split(resultStr, "/")
		if len(strs) != 3 {
			return fmt.Errorf("response data format error, should be [finished/total/errors]")
		}
		finished, err = strconv.ParseUint(strs[0], 10, 64)
		if err != nil {
			break
		}
		total, err = strconv.ParseUint(strs[1], 10, 64)
		if err != nil {
			break
		}
		warmErrors, err = strconv.ParseUint(strs[2], 10, 64)
		if err != nil {
			break
		}

		qCmd.Logger.Debugf("warmup result: total[%d], finished[%d], errors[%d]", total, finished, warmErrors)
		if (finished + warmErrors) == total {
			break
		}
		if isDaemon {
			break
		} else {
			bar.ChangeMax64(int64(total))
			bar.Set64(int64(finished))
		}
		time.Sleep(qCmd.interval)
	}
	if warmErrors > 0 { //warmup failed
		fmt.Println(color.Red.Sprintf("\nwarmup finished,%d errors\n", warmErrors))
	}

	if isDaemon { //can't show progress bar
		var progressStr string
		if resultStr == "finished" {
			header := []string{cobrautil.ROW_RESULT}
			qCmd.SetHeader(header)
			qCmd.TableNew.Append([]string{"finished"})

		} else {
			header := []string{"total", "finished", "errors", cobrautil.ROW_RESULT}
			qCmd.SetHeader(header)
			totalStr := humanize.Comma(int64(total))
			finishedStr := humanize.Comma(int64(finished))
			warmErrorStr := humanize.Comma(int64(warmErrors))
			if total > 0 {
				progress := math.Round(float64(warmErrors+finished) * 100 / float64(total))
				progressStr = fmt.Sprintf("%d%%", int(progress))
			}
			qCmd.TableNew.Append([]string{totalStr, finishedStr, warmErrorStr, progressStr})
		}
	} else {
		if total > 0 { //current warmup finished,last time warmup finished total will be 0
			bar.ChangeMax64(int64(total))
			bar.Set64(int64(total))
		}
		bar.Finish()
	}

	return nil
}

func (qCmd *QueryCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&qCmd.FinalDingoCmd)
}

func GetWarmupProgress(caller *cobra.Command, path string) *cmderror.CmdError {
	queryCmd := NewQueryWarmupCommand()
	queryCmd.Cmd.SetArgs([]string{"--format", config.FORMAT_NOOUT, path})
	config.AlignFlagsValue(caller, queryCmd.Cmd, []string{config.DINGOFS_INTERVAL})
	err := queryCmd.Cmd.Execute()
	if err != nil {
		retErr := cmderror.ErrQueryWarmup()
		retErr.Format(err.Error())
		return retErr
	}
	return cmderror.Success()
}
