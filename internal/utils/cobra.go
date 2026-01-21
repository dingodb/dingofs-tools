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

package utils

import (
	"errors"
	"fmt"
	"io"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/moby/term"
	"github.com/spf13/cobra"
)

const (
	PREFIX_COBRA_COMMAND_ERROR = "Error:\n"
)

var (
	NoArgs            = cli.NoArgs
	RequiresMinArgs   = cli.RequiresMinArgs
	RequiresMaxArgs   = cli.RequiresMaxArgs
	RequiresRangeArgs = cli.RequiresRangeArgs
	ExactArgs         = cli.ExactArgs

	ShowHelp = command.ShowHelp
)

const (
	usageTemplate = `Usage:
{{- if not .HasSubCommands}}  {{.UseLine}}{{end}}
{{- if .HasSubCommands}}  {{ .CommandPath}} COMMAND {{- if .HasAvailableFlags}} [OPTIONS]{{end}}{{end}}

{{if ne .Long ""}}{{ .Long | trim }}{{ else }}{{ .Short | trim }}{{end}}

{{- if gt .Aliases 0}}

Aliases:
  {{.NameAndAliases}}
{{- end}}

{{- range .Groups}}
{{- $groupId := .ID}}

{{.Title}}
{{- range $.Commands}}
{{- if and (eq .GroupID $groupId) (.IsAvailableCommand) (not .IsAdditionalHelpTopicCommand)}}
  {{rpad .Name .NamePadding }} {{.Short}}
{{- end}}
{{- end}}
{{- end}}

{{- $ungroupedCommands := ungroupedCommands . }}
{{- if $ungroupedCommands}}

Commands:
{{- range $ungroupedCommands }}
  {{rpad .Name .NamePadding }} {{.Short}}
{{- end}}
{{- end}}

{{- if .HasAvailableFlags}}

Options:
{{ wrappedFlagUsages . | trimRightSpace}}

{{- end}}

{{- if .HasExample}}

{{ .Example }}

{{- end}}

{{- if .HasSubCommands }}

Run '{{.CommandPath}} COMMAND --help' for more information on a command.
{{- end}}
`
)

func ungroupedCommands(cmd *cobra.Command) []*cobra.Command {
	cmds := []*cobra.Command{}
	for _, subCmd := range cmd.Commands() {
		if subCmd.IsAvailableCommand() && subCmd.GroupID == "" {
			cmds = append(cmds, subCmd)
		}
	}
	return cmds
}

func managementSubCommands(cmd *cobra.Command) []*cobra.Command {
	cmds := []*cobra.Command{}
	for _, subCmd := range cmd.Commands() {
		if cmd.Name() == "dingocli" && subCmd.Name() == "completion" {
			continue
		} else if subCmd.IsAvailableCommand() && subCmd.HasSubCommands() {
			cmds = append(cmds, subCmd)
		}
	}
	return cmds
}

func operationSubCommands(cmd *cobra.Command) []*cobra.Command {
	cmds := []*cobra.Command{}
	for _, subCmd := range cmd.Commands() {
		if subCmd.IsAvailableCommand() && !subCmd.HasSubCommands() {
			cmds = append(cmds, subCmd)
		}
	}
	return cmds
}

func hasManagementSubCommands(cmd *cobra.Command) bool {
	return len(managementSubCommands(cmd)) > 0
}

func hasOperationSubCommands(cmd *cobra.Command) bool {
	return len(operationSubCommands(cmd)) > 0
}

func wrappedFlagUsages(cmd *cobra.Command) string {
	width := 80
	if ws, err := term.GetWinsize(0); err == nil {
		width = int(ws.Width)
	}
	return cmd.Flags().FlagUsagesWrapped(width - 1)
}

func SetFlagErrorFunc(cmd *cobra.Command) {
	cmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		if err == nil {
			return nil
		}

		return errors.New(fmt.Sprintf("%s\nSee '%s --help'.", err, cmd.CommandPath()))
	})
}

func SetHelpTemplate(cmd *cobra.Command) {
	helpTemplate := `{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}`
	cmd.SetHelpTemplate(helpTemplate)
}

func SetUsageTemplate(cmd *cobra.Command) {
	cobra.AddTemplateFunc("ungroupedCommands", ungroupedCommands)
	cobra.AddTemplateFunc("managementSubCommands", managementSubCommands)
	cobra.AddTemplateFunc("operationSubCommands", operationSubCommands)
	cobra.AddTemplateFunc("hasManagementSubCommands", hasManagementSubCommands)
	cobra.AddTemplateFunc("hasOperationSubCommands", hasOperationSubCommands)
	cobra.AddTemplateFunc("wrappedFlagUsages", wrappedFlagUsages)
	cmd.SetUsageTemplate(usageTemplate)
}

func SetErr(cmd *cobra.Command, writer io.Writer) {
	cmd.SetErr(writer)
}
