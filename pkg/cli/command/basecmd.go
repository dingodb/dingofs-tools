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

package basecmd

import (
	"os"

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	"github.com/dingodb/dingofs-tools/internal/logger"
	"github.com/dingodb/dingofs-tools/internal/utils/process"
	cobratemplate "github.com/dingodb/dingofs-tools/internal/utils/template"
	config "github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// FinalDingoCmd is the final executable command,
// it has no subcommands.
// The execution process is Init->RunCommand->Print.
// Error Use to indicate whether the command is wrong
// and the reason for the execution error
type FinalDingoCmd struct {
	Use      string              `json:"-"`
	Short    string              `json:"-"`
	Long     string              `json:"-"`
	Example  string              `json:"-"`
	Error    *cmderror.CmdError  `json:"error"`
	Result   interface{}         `json:"result"`
	TableNew *tablewriter.Table  `json:"-"`
	Header   []string            `json:"-"`
	Cmd      *cobra.Command      `json:"-"`
	Logger   *logger.DingoLogger `json:"-"`
}

func (fc *FinalDingoCmd) SetHeader(header []string) {
	fc.Header = header
	fc.TableNew.SetHeader(header)
	// width := 80
	// if ws, err := term.GetWinsize(0); err == nil {
	// 	if width < int(ws.Width) {
	// 		width = int(ws.Width)
	// 	}
	// }
	// if len(header) != 0 {
	// 	fc.TableNew.SetColWidth(width/len(header) - 1)
	// }
}

// FinalDingoCmdFunc is the function type for final command
// If there is flag[required] related code should not be placed in init,
// the check for it is placed between PreRun and Run
type FinalDingoCmdFunc interface {
	Init(cmd *cobra.Command, args []string) error
	RunCommand(cmd *cobra.Command, args []string) error
	Print(cmd *cobra.Command, args []string) error
	// result in plain format string
	ResultPlainOutput() error
	AddFlags()
}

// MidDingoCmd is the middle command and has subcommands.
// If you execute this command
// you will be prompted which subcommands are included
type MidDingoCmd struct {
	Use   string
	Short string
	Cmd   *cobra.Command
}

// Add subcommand for MidDingoCmd
type MidDingoCmdFunc interface {
	AddSubCommands()
}

func NewFinalDingoCli(cli *FinalDingoCmd, funcs FinalDingoCmdFunc) *cobra.Command {
	cli.Logger = logger.GetLogger()
	cli.Cmd = &cobra.Command{
		Use:     cli.Use,
		Short:   cli.Short,
		Long:    cli.Long,
		Example: cli.Example,
		RunE: func(cmd *cobra.Command, args []string) error {
			show := config.GetFlagBool(cli.Cmd, config.VERBOSE)
			process.SetShow(show)
			cmd.SilenceUsage = true
			err := funcs.Init(cmd, args)
			if err != nil {
				return err
			}
			err = funcs.RunCommand(cmd, args)
			if err != nil {
				return err
			}
			return funcs.Print(cmd, args)
		},
		SilenceUsage: false,
	}
	
	config.AddFormatFlag(cli.Cmd)
	funcs.AddFlags()
	cobratemplate.SetFlagErrorFunc(cli.Cmd)

	// set table
	cli.TableNew = tablewriter.NewWriter(os.Stdout)
	cli.TableNew.SetRowLine(true)
	cli.TableNew.SetAutoFormatHeaders(true)
	cli.TableNew.SetAutoWrapText(true)
	cli.TableNew.SetAlignment(tablewriter.ALIGN_LEFT)

	return cli.Cmd
}

func NewMidDingoCli(cli *MidDingoCmd, add MidDingoCmdFunc) *cobra.Command {
	cli.Cmd = &cobra.Command{
		Use:   cli.Use,
		Short: cli.Short,
		Args:  cobratemplate.NoArgs,
	}
	add.AddSubCommands()
	return cli.Cmd
}
