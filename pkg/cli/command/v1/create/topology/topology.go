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
 * Created Date: 2022-06-29
 * Author: chengyi (Cyber-SiKu)
 */

package topology

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	cmderror "github.com/dingodb/dingofs-tools/internal/error"
	cobrautil "github.com/dingodb/dingofs-tools/internal/utils"
	basecmd "github.com/dingodb/dingofs-tools/pkg/cli/command"
	"github.com/dingodb/dingofs-tools/pkg/config"
	"github.com/dingodb/dingofs-tools/pkg/output"
	"github.com/dingodb/dingofs-tools/proto/dingofs/proto/topology"
	"github.com/spf13/cobra"
)

const (
	topologyExample = `$ dingo create topology --clustermap /path/to/clustermap.json`
)

type Topology struct {
	Servers []Server `json:"servers"`
	Zones   []Zone   `json:"-"`
	Pools   []Pool   `json:"pools"`
	PoolNum uint64   `json:"npools"`
}

type TopologyCommand struct {
	basecmd.FinalDingoCmd
	topology   Topology
	timeout    time.Duration
	retryTimes int32
	retryDelay time.Duration
	verbose    bool

	addrs []string
	// pool
	clusterPoolsInfo []*topology.PoolInfo
	createPoolRpc    *CreatePoolRpc
	deletePoolRpc    *DeletePoolRpc
	listPoolRpc      *ListPoolRpc
	createPool       []*topology.CreatePoolRequest
	deletePool       []*topology.DeletePoolRequest
	// zone
	clusterZonesInfo []*topology.ZoneInfo
	deleteZoneRpc    *DeleteZoneRpc
	createZoneRpc    *CreateZoneRpc
	listPoolZoneRpc  *ListPoolZoneRpc
	createZone       []*topology.CreateZoneRequest
	deleteZone       []*topology.DeleteZoneRequest
	// server
	clusterServersInfo []*topology.ServerInfo
	deleteServerRpc    *DeleteServerRpc
	createServerRpc    *CreateServerRpc
	listZoneServerRpc  *ListZoneServerRpc
	createServer       []*topology.ServerRegistRequest
	deleteServer       []*topology.DeleteServerRequest

	rows []map[string]string
}

var _ basecmd.FinalDingoCmdFunc = (*TopologyCommand)(nil) // check interface

func NewTopologyCommand() *cobra.Command {
	topologyCmd := &TopologyCommand{
		FinalDingoCmd: basecmd.FinalDingoCmd{
			Use:     "topology",
			Short:   "create dingofs topology",
			Example: topologyExample,
		},
	}
	basecmd.NewFinalDingoCli(&topologyCmd.FinalDingoCmd, topologyCmd)
	return topologyCmd.Cmd
}

func (tCmd *TopologyCommand) AddFlags() {
	config.AddRpcRetryTimesFlag(tCmd.Cmd)
	config.AddRpcRetryDelayFlag(tCmd.Cmd)
	config.AddRpcTimeoutFlag(tCmd.Cmd)
	config.AddFsMdsAddrFlag(tCmd.Cmd)
	config.AddClusterMapRequiredFlag(tCmd.Cmd)
}

func (tCmd *TopologyCommand) Init(cmd *cobra.Command, args []string) error {
	addrs, addrErr := config.GetFsMdsAddrSlice(tCmd.Cmd)
	if addrErr.TypeCode() != cmderror.CODE_SUCCESS {
		tCmd.Error = addrErr
		return fmt.Errorf(addrErr.Message)
	}
	tCmd.addrs = addrs
	tCmd.timeout = config.GetRpcTimeout(cmd)
	tCmd.retryTimes = config.GetRpcRetryTimes(cmd)
	tCmd.retryDelay = config.GetRpcRetryDelay(cmd)
	tCmd.verbose = config.GetFlagBool(cmd, config.VERBOSE)

	filePath := config.GetFlagString(tCmd.Cmd, config.DINGOFS_CLUSTERMAP)
	jsonFile, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer jsonFile.Close()
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		jsonFileErr := cmderror.ErrReadFile()
		jsonFileErr.Format(filePath, err.Error())
		return jsonFileErr.ToError()
	}
	err = json.Unmarshal(byteValue, &tCmd.topology)
	if err != nil {
		jsonFileErr := cmderror.ErrReadFile()
		jsonFileErr.Format(filePath, err.Error())
		return jsonFileErr.ToError()
	}

	updateZoneErr := tCmd.updateZone()
	if updateZoneErr.TypeCode() != cmderror.CODE_SUCCESS {
		return updateZoneErr.ToError()
	}

	header := []string{cobrautil.ROW_NAME, cobrautil.ROW_TYPE, cobrautil.ROW_OPERATION, cobrautil.ROW_PARENT}
	tCmd.SetHeader(header)
	tCmd.TableNew.SetAutoMergeCells(true)

	scanErr := tCmd.scanCluster()
	if scanErr.TypeCode() != cmderror.CODE_SUCCESS {
		return scanErr.ToError()
	}

	return nil
}

func (tCmd *TopologyCommand) updateTopology() *cmderror.CmdError {
	// update cluster topology
	// remove
	// server
	err := tCmd.removeServers()
	if err.TypeCode() != cmderror.CODE_SUCCESS {
		return err
	}
	// zone
	err = tCmd.removeZones()
	if err.TypeCode() != cmderror.CODE_SUCCESS {
		return err
	}
	// pool
	err = tCmd.removePools()
	if err.TypeCode() != cmderror.CODE_SUCCESS {
		return err
	}

	// create
	// pool
	err = tCmd.createPools()
	if err.TypeCode() != cmderror.CODE_SUCCESS {
		return err
	}
	// zone
	err = tCmd.createZones()
	if err.TypeCode() != cmderror.CODE_SUCCESS {
		return err
	}
	// zerver
	err = tCmd.createServers()
	if err.TypeCode() != cmderror.CODE_SUCCESS {
		return err
	}
	return cmderror.ErrSuccess()
}

func (tCmd *TopologyCommand) RunCommand(cmd *cobra.Command, args []string) error {
	err := tCmd.updateTopology()
	if err.TypeCode() != cmderror.CODE_SUCCESS {
		return fmt.Errorf(err.Message)
	}

	tCmd.Result = tCmd.rows
	tCmd.Error = err
	return nil
}

func (tCmd *TopologyCommand) Print(cmd *cobra.Command, args []string) error {
	return output.FinalCmdOutput(&tCmd.FinalDingoCmd, tCmd)
}

func (tCmd *TopologyCommand) ResultPlainOutput() error {
	if len(tCmd.createPool) == 0 && len(tCmd.deletePool) == 0 && len(tCmd.createZone) == 0 && len(tCmd.deleteZone) == 0 && len(tCmd.createServer) == 0 && len(tCmd.deleteServer) == 0 {
		fmt.Println("no change")
		return nil
	}
	return output.FinalCmdOutputPlain(&tCmd.FinalDingoCmd)
}

// Compare the topology in the cluster with json,
// delete in the cluster not in json,
// create in json not in the topology.
func (tCmd *TopologyCommand) scanCluster() *cmderror.CmdError {
	err := tCmd.scanPools()
	if err.TypeCode() != cmderror.CODE_SUCCESS {
		return err
	}
	err = tCmd.scanZones()
	if err.TypeCode() != cmderror.CODE_SUCCESS {
		return err
	}

	err = tCmd.scanServers()
	if err.TypeCode() != cmderror.CODE_SUCCESS {
		return err
	}

	return err
}
