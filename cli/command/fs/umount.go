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
	"syscall"

	"github.com/dingodb/dingofs-tools/cli/cli"
	"github.com/dingodb/dingofs-tools/internal/utils"
	"github.com/spf13/cobra"
)

const (
	FS_UMOUNT_EXAMPLE = `Examples:
   $ dingo fs umount /mnt/dingofs`
)

type umountOptions struct {
	mountpoint string
	force      bool
	lazy       bool
}

func NewFsUmountCommand(dingoadm *cli.DingoAdm) *cobra.Command {
	var options umountOptions

	cmd := &cobra.Command{
		Use:     "umount MOUNTPOINT [OPTIONS]",
		Short:   "umount filesystem",
		Args:    utils.ExactArgs(1),
		Example: FS_UMOUNT_EXAMPLE,
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			options.mountpoint = args[0]

			options.force, err = cmd.Flags().GetBool("force")
			if err != nil {
				return err
			}
			options.lazy, err = cmd.Flags().GetBool("lazy")
			if err != nil {
				return err
			}

			return runUmuont(cmd, dingoadm, options)
		},
		SilenceUsage:          false,
		DisableFlagsInUseLine: true,
	}

	utils.SetFlagErrorFunc(cmd)

	// add flags
	cmd.Flags().BoolP("force", "f", false, "Force umount")
	cmd.Flags().BoolP("lazy", "l", false, "Lazy umount")

	return cmd
}

func runUmuont(cmd *cobra.Command, dingoadm *cli.DingoAdm, options umountOptions) error {
	flags := 0
	if options.lazy && options.force {
		return fmt.Errorf("lazy and force options cannot be used simultaneously")
	}

	if options.lazy {
		flags = syscall.MNT_DETACH
	}
	if options.force {
		flags = syscall.MNT_FORCE
	}

	if _, err := os.Stat(options.mountpoint); os.IsNotExist(err) {
		return fmt.Errorf("mountpoint does not exist: %s", options.mountpoint)
	}

	err := syscall.Unmount(options.mountpoint, flags)
	if err != nil {
		switch {
		case err == syscall.EINVAL:
			return fmt.Errorf("invalid mountpoint: %s", options.mountpoint)
		case err == syscall.EPERM:
			// use fusermount3  to umount
			umountErr := runFuseumount(options.mountpoint)
			if umountErr != nil {
				return fmt.Errorf("error unmounting: %v", umountErr)
			}
		case err == syscall.EBUSY:
			return fmt.Errorf("mountpoint %s is busy, try lazy unmount or check processes", options.mountpoint)
		case err == syscall.ENOENT:
			return fmt.Errorf("mountpoint %s does not exist", options.mountpoint)
		default:
			return fmt.Errorf("system error: %v", err)
		}
	}

	fmt.Printf("Successfully unmounted %s\n", options.mountpoint)

	return nil
}

func runFuseumount(mountpoint string) error {

	var oscmd *exec.Cmd

	args := []string{"-u", mountpoint}
	oscmd = exec.Command("fusermount3", args...)
	oscmd.Stderr = os.Stderr
	oscmd.Stdout = os.Stdout

	return oscmd.Run()

}
