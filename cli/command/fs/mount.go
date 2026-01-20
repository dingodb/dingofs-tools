/*
 * Copyright (c) 2025 dingodb.com, Inc. All Rights Reserved
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package fs

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/dingodb/dingocli/cli/cli"
	"github.com/dingodb/dingocli/internal/utils"
	"github.com/spf13/cobra"
)

const (
	FS_MOUNT_EXAMPLE = `Examples:
   $ dingo fs mount mds://10.220.69.6:7400/myfs /mnt/dingofs
   $ dingo fs mount local://myfs /mnt/dingofs`
)

var (
	DINGOFS_CLIENT_BINARY = fmt.Sprintf("%s/.dingofs/bin/dingo-client", utils.GetHomeDir())
)

type mountOptions struct {
	clientBinary string
	cmdArgs      []string
	mountpoint   string
	daemonize    bool
}

func NewFsMountCommand(dingocli *cli.DingoCli) *cobra.Command {
	var options mountOptions

	cmd := &cobra.Command{
		Use:                "mount METAURL MOUNTPOINT [OPTIONS]",
		Short:              "mount filesystem",
		Args:               utils.RequiresMinArgs(0),
		DisableFlagParsing: true,
		Example:            FS_MOUNT_EXAMPLE,
		RunE: func(cmd *cobra.Command, args []string) error {
			options.clientBinary = DINGOFS_CLIENT_BINARY
			// check flags
			for _, arg := range args {
				if arg == "--help" || arg == "-h" {
					return runCommandHelp(cmd, options.clientBinary)
				}
				if arg == "--daemonize" || arg == "-d" {
					options.daemonize = true
				}
			}

			if len(args) < 2 {
				return fmt.Errorf("\"dingocli fs mount\" requires exactly 2 arguments\n\nUsage: dingocli fs mount METAURL MOUNTPOINT [OPTIONS]")
			}
			options.cmdArgs = args
			options.mountpoint = args[1]

			// check dingo-client is exists
			if !utils.IsFileExists(options.clientBinary) {
				return fmt.Errorf("%s not found", options.clientBinary)
				//TODO(yansp): auto download dingo-client
				//url := "https://github.com/dingodb/dingofs/releases/download/v4.2.0/dingofs-latest.tar.gz"
				// fmt.Printf("[%s] %s not found, download from %s\n", color.RedString("WARNING"), options.client, url)
				// _, err := utils.DownloadFileWithProgress(url, filepath.Dir(options.client), "")
				// if err != nil {
				// 	return fmt.Errorf("failed to download dingo-client: %v", err)
				// }

			}
			// check has execute permission
			if !utils.HasExecutePermission(options.clientBinary) {
				fmt.Printf("no execute permission for %s, now add it\n", options.clientBinary)
				err := utils.AddExecutePermission(options.clientBinary)
				if err != nil {
					return fmt.Errorf("failed to add execute permission for %s,error: %v", options.clientBinary, err)
				}
			}

			return runMount(cmd, dingocli, options)
		},
		SilenceUsage:          false,
		DisableFlagsInUseLine: true,
	}

	utils.SetFlagErrorFunc(cmd)

	return cmd
}

func runMount(cmd *cobra.Command, dingocli *cli.DingoCli, options mountOptions) error {
	var oscmd *exec.Cmd
	var name string

	name = options.clientBinary
	cmdarg := options.cmdArgs

	oscmd = exec.Command(name, cmdarg...)

	oscmd.Stdout = os.Stdout
	oscmd.Stderr = os.Stderr

	if err := oscmd.Start(); err != nil {
		return err
	}

	// forground mode, wait process exit
	if !options.daemonize {
		// wait process complete
		if err := oscmd.Wait(); err != nil {
			return err
		}
		return nil
	}

	// daemonize mode
	isReady := make(chan bool, 1)
	isTimeout := make(chan bool, 1)

	// mount completed
	go func() {
		filename := filepath.Join(options.mountpoint, ".stats")
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if _, err := os.Stat(filename); !os.IsNotExist(err) {
				select {
				case isReady <- true:
				case <-time.After(120 * time.Second):
					isTimeout <- true
				default:
					continue
				}
				return
			}
		}
	}()

	select {
	case <-isReady: // start success
		// continues to read the remaining output
		go func() {
			// wait daemon exit, non block
			go oscmd.Wait()
		}()

		fmt.Printf("Successfully mounted at %s\n", options.mountpoint)
		return nil

	case _ = <-isTimeout: //mount failed
		//umount fs
		tmpOptions := umountOptions{mountpoint: options.mountpoint}
		runUmuont(cmd, dingocli, tmpOptions)
		return fmt.Errorf("Failed mount at %s\n", options.mountpoint)
	}
}

func runCommandHelp(cmd *cobra.Command, command string) error {
	// print dingocli usage
	fmt.Printf("Usage: dingo %s %s\n", cmd.Parent().Use, cmd.Use)
	fmt.Println("")
	fmt.Println(cmd.Short)
	fmt.Println("")

	// print  dingo-client options
	fmt.Println("Options:")

	helpArgs := []string{"--help"}
	oscmd := exec.Command(command, helpArgs...)
	output, err := oscmd.CombinedOutput()
	if err != nil && len(output) == 0 {
		return err
	}

	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "--") {
			fmt.Printf("  %s\n", trimmed)
		}
	}

	// print dingocli example
	fmt.Println("")
	fmt.Println(cmd.Example)

	return nil
}
