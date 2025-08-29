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

package stats

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	process "github.com/dingodb/dingofs-tools/internal/utils/process"
	"github.com/dingodb/dingofs-tools/pkg/base"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

// colors
const (
	BLACK = 30 + iota //black  = "\033[0;30m"
	RED               //red    = "\033[0;31m"
	GREEN
	YELLOW
	BLUE
	MAGENTA
	CYAN
	WHITE
	DEFAULT = "00"
)

// \033[0m  unset color attribute \033[1m set highlight
const (
	RESET_SEQ      = "\033[0m"
	COLOR_SEQ      = "\033[1;"
	COLOR_DARK_SEQ = "\033[0;"
	UNDERLINE_SEQ  = "\033[4m"
	CLEAR_SCREEM   = "\033[2J\033[1;1H"
)

// metirc types
const (
	metricByte = 1 << iota
	metricCount
	metricTime
	metricCPU
	metricGauge
	metricCounter
	metricHist
	metricHit
)

const MaxItemSize = 5

type item struct {
	nick string // must be size <= 5 MaxItemSize
	name string
	typ  uint8
}

type section struct {
	name  string
	items []*item
}

type statsWatcher struct {
	colorful   bool
	duration   time.Duration
	interval   int64
	mountPoint string
	header     string
	sections   []*section
	cpuUsage   float64
	count      uint32
}

var _ basecmd.FinalDingoCmdFunc = (*MountpointCommand)(nil) // check interface

type MountpointCommand struct {
	basecmd.FinalDingoCmd
}

// set logout to stdout
func init() {
	process.SetShow(true)
}

func NewMountpointCommand() *cobra.Command {
	mountpointCmd := &MountpointCommand{
		basecmd.FinalDingoCmd{
			Use:   "mountpoint",
			Short: "show real time performance statistics of dingofs mountpoint",
			Example: `dingo stats mountpoint /mnt/dingofs
			
# fuse metrics
dingo stats mountpoint /mnt/dingofs --schema f

# s3 metrics
dingo stats mountpoint /mnt/dingofs --schema o

# More metrics
dingo stats mountpoint /mnt/dingofs --verbose

# Show 3 times
dingo stats mountpoint /mnt/dingofs --count 3`,
		},
	}
	return basecmd.NewFinalDingoCli(&mountpointCmd.FinalDingoCmd, mountpointCmd)
}

// add stats flags
func (mountpointCmd *MountpointCommand) AddFlags() {
	config.AddIntervalOptionFlag(mountpointCmd.Cmd)
	config.AddFsSchemaOptionalFlag(mountpointCmd.Cmd)
	config.AddFsCountOptionalFlag(mountpointCmd.Cmd)
}

func (mountpointCmd *MountpointCommand) Init(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New(`ERROR: This command requires mountPoint
USAGE:
   dingo stats mountpoint <path> [Flags]`)
	}
	return nil
}

// run stats command
func (mountpointCmd *MountpointCommand) RunCommand(cmd *cobra.Command, args []string) error {
	mountPoint := args[0]
	schemaValue := config.GetStatsSchemaFlagOptionFlag(cmd)
	verbose := config.GetFlagBool(cmd, config.VERBOSE)
	duration := config.GetIntervalFlag(cmd)
	count := config.GetStatsCountFlagOptionFlag(cmd)
	if duration < 1*time.Second {
		duration = 1 * time.Second
	}
	realTimeStats(mountPoint, schemaValue, verbose, duration, count)
	return nil
}
func (mountpointCmd *MountpointCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&mountpointCmd.FinalDingoCmd, mountpointCmd)
}

func (mountpointCmd *MountpointCommand) ResultPlainOutput() error {
	return output.FinalCmdOutputPlain(&mountpointCmd.FinalDingoCmd)
}

func (w *statsWatcher) colorize(msg string, color int, dark bool, underline bool) string {
	if !w.colorful || msg == "" || msg == " " {
		return msg
	}
	var cseq, useq string
	if dark {
		cseq = COLOR_DARK_SEQ
	} else {
		cseq = COLOR_SEQ
	}
	if underline {
		useq = UNDERLINE_SEQ
	}
	return fmt.Sprintf("%s%s%dm%s%s", useq, cseq, color, msg, RESET_SEQ)
}

func (w *statsWatcher) buildSchema(schema string, verbose bool) {
	for _, r := range schema {
		var s section
		switch r {
		case 'u':
			s.name = "usage"
			s.items = append(s.items, &item{"cpu", "process_cpu_usage", metricCPU | metricCounter})
			s.items = append(s.items, &item{"mem", "process_memory_resident", metricGauge})
			if verbose {
				s.items = append(s.items, &item{"rbuf", "read_data_cache_byte", metricGauge})
			}
			s.items = append(s.items, &item{"wbuf", "write_data_cache_byte", metricGauge})
		case 'f':
			s.name = "fuse"
			s.items = append(s.items, &item{"ops", "dingofs_fuse_op_all", metricTime | metricHist})
			s.items = append(s.items, &item{"read", "dingofs_vfs_read_bps_total_count", metricByte | metricCounter})
			s.items = append(s.items, &item{"write", "dingofs_vfs_write_bps_total_count", metricByte | metricCounter})
		case 'b':
			s.name = "blockcache"
			s.items = append(s.items, &item{"load", "dingofs_disk_cache_group_load_total_bytes", metricByte | metricCounter})
			s.items = append(s.items, &item{"stage", "dingofs_disk_cache_group_stage_total_bytes", metricByte | metricCounter})
			s.items = append(s.items, &item{"cache", "dingofs_disk_cache_group_cache_total_bytes", metricByte | metricCounter})
		case 'o':
			s.name = "object"
			s.items = append(s.items, &item{"get", "dingofs_block_read_block_bps_total_count", metricByte | metricCounter})
			if verbose {
				s.items = append(s.items, &item{"ops", "dingofs_block_read_block", metricTime | metricHist})
			}
			s.items = append(s.items, &item{"put", "dingofs_block_write_block_bps_total_count", metricByte | metricCounter})
			if verbose {
				s.items = append(s.items, &item{"ops", "dingofs_block_write_block", metricTime | metricHist})
			}
		case 'r':
			s.name = "remotecache"
			s.items = append(s.items, &item{"load", "dingofs_remote_node_group_range_total_bytes", metricByte | metricCounter})
			s.items = append(s.items, &item{"stage", "dingofs_remote_node_group_put_total_bytes", metricByte | metricCounter})
			s.items = append(s.items, &item{"cache", "dingofs_remote_node_group_cache_total_bytes", metricByte | metricCounter})
			s.items = append(s.items, &item{"hit", "dingofs_remote_cache", metricHit})
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

func padding(name string, width int, char byte) string {
	pad := width - len(name)
	if pad < 0 {
		pad = 0
		name = name[0:width]
	}
	prefix := (pad + 1) / 2
	buf := make([]byte, width)
	for i := 0; i < prefix; i++ {
		buf[i] = char
	}

	copy(buf[prefix:], name)
	for i := prefix + len(name); i < width; i++ {
		buf[i] = char
	}
	return string(buf)
}

func (w *statsWatcher) formatHeader() {
	headers := make([]string, len(w.sections))
	subHeaders := make([]string, len(w.sections))
	for i, s := range w.sections {
		subs := make([]string, 0, len(s.items))
		for _, it := range s.items {
			subs = append(subs, w.colorize(padding(it.nick, MaxItemSize, ' '), BLUE, false, true))
			if it.typ&metricHist != 0 {
				if it.typ&metricTime != 0 {
					subs = append(subs, w.colorize(" lat ", BLUE, false, true))
				} else {
					subs = append(subs, w.colorize(" avg ", BLUE, false, true))
				}
			}
		}
		width := 6*len(subs) - 1 // nick(5) + space(1)
		subHeaders[i] = strings.Join(subs, " ")
		headers[i] = w.colorize(padding(s.name, width, '-'), BLUE, true, false)
	}
	w.header = fmt.Sprintf("%s\n%s", strings.Join(headers, " "),
		strings.Join(subHeaders, w.colorize("|", BLUE, true, false)))
}

func (w *statsWatcher) formatU64(v float64, dark, isByte bool) string {
	if v <= 0.0 {
		return w.colorize("   0 ", BLACK, false, false)
	}
	var vi uint64
	var unit string
	var color int
	switch vi = uint64(v); {
	case vi < 10000:
		if isByte {
			unit = "B"
		} else {
			unit = " "
		}
		color = RED
	case vi>>10 < 10000:
		vi, unit, color = vi>>10, "K", YELLOW
	case vi>>20 < 10000:
		vi, unit, color = vi>>20, "M", GREEN
	case vi>>30 < 10000:
		vi, unit, color = vi>>30, "G", BLUE
	case vi>>40 < 10000:
		vi, unit, color = vi>>40, "T", MAGENTA
	default:
		vi, unit, color = vi>>50, "P", CYAN
	}
	return w.colorize(fmt.Sprintf("%4d", vi), color, dark, false) +
		w.colorize(unit, BLACK, false, false)
}

func (w *statsWatcher) formatTime(v float64, dark bool) string {
	var ret string
	var color int
	switch {
	case v <= 0.0:
		ret, color, dark = "   0 ", BLACK, false
	case v < 10.0:
		ret, color = fmt.Sprintf("%4.2f ", v), GREEN
	case v < 100.0:
		ret, color = fmt.Sprintf("%4.1f ", v), YELLOW
	case v < 10000.0:
		ret, color = fmt.Sprintf("%4.f ", v), RED
	default:
		ret, color = fmt.Sprintf("%1.e", v), MAGENTA
	}
	return w.colorize(ret, color, dark, false)
}

func (w *statsWatcher) formatCPU(v float64, dark bool) string {
	var ret string
	var color int
	switch v = v * 100.0; {
	case v <= 0.0:
		ret, color = " 0.0", WHITE
	case v < 30.0:
		ret, color = fmt.Sprintf("%4.1f", v), GREEN
	case v < 100.0:
		ret, color = fmt.Sprintf("%4.1f", v), YELLOW
	default:
		ret, color = fmt.Sprintf("%4.f", v), RED
	}
	return w.colorize(ret, color, dark, false) +
		w.colorize("%", BLACK, false, false)
}

func (w *statsWatcher) formatHits(v float64, dark bool) string {
	var ret string
	var color int = GREEN
	v = v * 100.0
	ret = fmt.Sprintf("%4.1f", v)

	return w.colorize(ret, color, dark, false) +
		w.colorize("%", BLACK, false, false)
}

// read metric data from file
func readStats(mp string) map[string]float64 {

	f, err := os.Open(filepath.Join(mp, ".stats"))
	if err != nil {
		log.Fatalf("open stats file under mount point %s: %s", mp, err)
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		log.Fatalf("read stats file under mount point %s: %s", mp, err)
	}

	if err != nil {
		panic(err)
	}

	outstr := strings.ReplaceAll(string(data), "\r", "")
	outstr = strings.ReplaceAll(outstr, " ", "")

	metricDataMap := make(map[string]float64)
	lines := strings.Split(string(outstr), "\n")

	for _, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) == 2 {
			v, err := strconv.ParseFloat(fields[1], 64)
			if err != nil {
				continue
			}
			metricDataMap[fields[0]] = v
		}
	}
	return metricDataMap
}

func (w *statsWatcher) printDiff(left, right map[string]float64, dark bool) {
	if !w.colorful && dark {
		return
	}
	values := make([]string, len(w.sections))
	for i, s := range w.sections {
		vals := make([]string, 0, len(s.items))
		for _, it := range s.items {
			switch it.typ & 0xF0 {
			case metricGauge: // show current value
				vals = append(vals, w.formatU64(right[it.name], dark, true))
			case metricCounter:
				v := (right[it.name] - left[it.name])
				if !dark {
					v /= float64(w.interval)
				}
				if it.typ&metricByte != 0 {
					vals = append(vals, w.formatU64(v, dark, true))
				} else if it.typ&metricCPU != 0 {
					v = right[it.name] //reset value to current for cpu
					w.cpuUsage += v
					if !dark {
						v = w.cpuUsage
						v /= float64(w.interval)
						w.cpuUsage = 0.0
					}
					vals = append(vals, w.formatCPU(v, dark))
				} else if it.typ&metricTime != 0 {
					vals = append(vals, w.formatTime(v, dark))
				} else { // metricCount
					vals = append(vals, w.formatU64(v, dark, false))
				}
			case metricHist: // metricTime
				count := right[it.name+"_qps_total_count"] - left[it.name+"_qps_total_count"]
				var avg float64
				if count > 0.0 {
					latency := right[it.name+"_lat_total_value"] - left[it.name+"_lat_total_value"]
					if it.typ&metricTime != 0 {
						latency /= 1000 //us -> ms
					}
					avg = latency / count
				}
				if !dark {
					count /= float64(w.interval)
				}
				vals = append(vals, w.formatU64(count, dark, false), w.formatTime(avg, dark))

			case metricHit: // metricHits
				hitCount := right[it.name+"_hit_count"] - left[it.name+"_hit_count"]
				missCount := right[it.name+"_miss_count"] - left[it.name+"_miss_count"]
				totalCount := hitCount + missCount
				var avg float64
				if totalCount > 0.0 {
					avg = hitCount / totalCount
				}
				vals = append(vals, w.formatHits(avg, dark))
			}
		}
		values[i] = strings.Join(vals, " ")
	}
	if w.colorful && dark {
		fmt.Printf("%s\r", strings.Join(values, w.colorize("|", BLUE, true, false)))
	} else {
		fmt.Printf("%s\n", strings.Join(values, w.colorize("|", BLUE, true, false)))
	}
}

// real time read metric data and show in client
func realTimeStats(mountPoint string, schema string, verbose bool, duration time.Duration, count uint32) {
	inode, err := base.GetFileInode(mountPoint)
	if err != nil {
		log.Fatalf("run stats failed, %s", err)
	}
	if inode != 1 {
		log.Fatalf("path %s is not a mount point", mountPoint)
	}
	watcher := &statsWatcher{
		colorful:   isatty.IsTerminal(os.Stdout.Fd()),
		duration:   duration,
		mountPoint: mountPoint,
		interval:   int64(duration) / 1000000000,
		cpuUsage:   0.0,
		count:      count,
	}
	watcher.buildSchema(schema, verbose)
	watcher.formatHeader()

	var tick uint
	var start, last, current map[string]float64
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	current = readStats(watcher.mountPoint)
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
		current = readStats(watcher.mountPoint)
		//for interval > 1s,don't print the middle result for last time
		if uint(math.Ceil(float64(tick)/float64(watcher.interval))) == uint(watcher.count) { //exit
			break
		}
	}

}
