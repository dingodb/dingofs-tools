// Copyright (c) 2025 dingodb.com, Inc. All Rights Reserved
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"fmt"
	"math"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// check quota is consistent
func CheckQuota(capacity int64, usedBytes int64, maxInodes int64, usedInodes int64, realUsedBytes int64, realUsedInodes int64) ([]string, bool) {
	var capacityStr string
	var usedStr string
	var realUsedStr string
	var maxInodesStr string
	var inodeUsedStr string
	var realUsedInodesStr string
	var result []string

	checkResult := true

	if capacity == math.MaxInt64 {
		capacityStr = "unlimited"
	} else { //quota is set
		capacityStr = humanize.Comma(capacity)
	}
	usedStr = humanize.Comma(usedBytes)
	realUsedStr = humanize.Comma(realUsedBytes)
	if usedBytes != realUsedBytes {
		checkResult = false
	}
	result = append(result, capacityStr)
	result = append(result, usedStr)
	result = append(result, realUsedStr)

	if maxInodes == math.MaxInt64 {
		maxInodesStr = "unlimited"
	} else { //inode quota is set
		maxInodesStr = humanize.Comma(int64(maxInodes))
	}
	inodeUsedStr = humanize.Comma(usedInodes)
	realUsedInodesStr = humanize.Comma(int64(realUsedInodes))
	if usedInodes != realUsedInodes {
		checkResult = false
	}
	result = append(result, maxInodesStr)
	result = append(result, inodeUsedStr)
	result = append(result, realUsedInodesStr)

	if checkResult {
		result = append(result, "success")
	} else {
		result = append(result, color.RedString("failed"))
	}
	return result, checkResult
}

// check the quota value from command line
func GetQuotaValue(cmd *cobra.Command) (int64, int64, error) {
	var maxBytes int64 = 0
	var maxInodes int64 = 0

	if !cmd.Flag(DINGOFS_QUOTA_CAPACITY).Changed && !cmd.Flag(DINGOFS_QUOTA_INODES).Changed {
		return 0, 0, fmt.Errorf("capacity or inodes is required")
	}
	if cmd.Flag(DINGOFS_QUOTA_CAPACITY).Changed {
		capacity, err := cmd.Flags().GetUint64(DINGOFS_QUOTA_CAPACITY)
		if err != nil {
			return 0, 0, err
		}
		maxBytes = int64(capacity) * 1024 * 1024 * 1024
	}

	if cmd.Flag(DINGOFS_QUOTA_INODES).Changed {
		inodes, err := cmd.Flags().GetUint64(DINGOFS_QUOTA_INODES)
		if err != nil {
			return 0, 0, err
		}
		maxInodes = int64(inodes)
	}

	if maxBytes == 0 { //  set to 0,unlimited
		maxBytes = math.MaxInt64
	}
	if maxInodes == 0 { // set to 0,unlimited
		maxInodes = math.MaxInt64
	}

	return maxBytes, maxInodes, nil
}

// convert number value to Humanize Value
func ConvertQuotaToHumanizeValue(capacity uint64, usedBytes int64, maxInodes uint64, usedInodes int64) []string {
	var capacityStr string
	var usedPercentStr string
	var maxInodesStr string
	var maxInodesPercentStr string
	var result []string

	if capacity == math.MaxInt64 {
		capacityStr = "unlimited"
		usedPercentStr = ""
	} else {
		capacityStr = humanize.IBytes(capacity)
		usedPercentStr = fmt.Sprintf("%d", int(math.Round((float64(usedBytes) * 100.0 / float64(capacity)))))
	}
	result = append(result, capacityStr)
	result = append(result, humanize.IBytes(uint64(usedBytes))) //TODO usedBytes  may be negative
	result = append(result, usedPercentStr)
	if maxInodes == math.MaxInt64 {
		maxInodesStr = "unlimited"
		maxInodesPercentStr = ""
	} else {
		maxInodesStr = humanize.Comma(int64(maxInodes))
		maxInodesPercentStr = fmt.Sprintf("%d", int(math.Round((float64(usedInodes) * 100.0 / float64(maxInodes)))))
	}
	result = append(result, maxInodesStr)
	result = append(result, humanize.Comma(int64(usedInodes)))
	result = append(result, maxInodesPercentStr)

	return result
}
