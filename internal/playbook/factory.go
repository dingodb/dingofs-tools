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

package playbook

import (
	"github.com/dingodb/dingofs-tools/internal/configure"
	"github.com/dingodb/dingofs-tools/internal/configure/topology"
	"github.com/dingodb/dingofs-tools/internal/errno"
	"github.com/dingodb/dingofs-tools/internal/task/task"
	"github.com/dingodb/dingofs-tools/internal/task/task/checker"
	comm "github.com/dingodb/dingofs-tools/internal/task/task/common"
	"github.com/dingodb/dingofs-tools/internal/task/task/fs"
	"github.com/dingodb/dingofs-tools/internal/task/task/monitor"
	"github.com/dingodb/dingofs-tools/internal/tasks"
)

const (
	// checker
	CHECK_TOPOLOGY int = iota
	CHECK_SSH_CONNECT
	CHECK_PERMISSION
	CHECK_KERNEL_VERSION
	CHECK_KERNEL_MODULE
	CHECK_PORT_IN_USE
	CHECK_DESTINATION_REACHABLE
	START_HTTP_SERVER
	CHECK_NETWORK_FIREWALL
	GET_HOST_DATE
	CHECK_HOST_DATE
	CHECK_S3
	CLEAN_PRECHECK_ENVIRONMENT

	// common
	PULL_IMAGE
	CREATE_CONTAINER
	CREATE_MDSV2_CLI_CONTAINER
	SYNC_CONFIG
	START_SERVICE
	START_ETCD
	START_MDS
	START_CHUNKSERVER
	START_SNAPSHOTCLONE
	START_METASERVER
	START_FS_MDS
	START_COORDINATOR
	START_STORE
	START_MDSV2_CLI_CONTAINER
	START_DINGODB_EXECUTOR
	STOP_SERVICE
	RESTART_SERVICE
	CREATE_META_TABLES
	INIT_SERVIE_STATUS
	GET_SERVICE_STATUS
	CLEAN_SERVICE
	BACKUP_ETCD_DATA
	CHECK_MDS_ADDRESS
	CHECK_STORE_HEALTH
	INIT_CLIENT_STATUS
	GET_CLIENT_STATUS

	// dingodb
	START_DINGODB_DOCUMENT
	START_DINGODB_INDEX
	START_DINGODB_DISKANN
	START_DINGODB_PROXY
	START_DINGODB_WEB

	// bs
	FORMAT_CHUNKFILE_POOL
	GET_FORMAT_STATUS
	STOP_FORMAT
	BALANCE_LEADER
	START_NEBD_SERVICE
	CREATE_VOLUME
	MAP_IMAGE
	UNMAP_IMAGE

	// monitor
	PULL_MONITOR_IMAGE
	CREATE_MONITOR_CONTAINER
	SYNC_MONITOR_ORIGIN_CONFIG
	SYNC_MONITOR_ALT_CONFIG
	SYNC_HOSTS_MAPPING
	CLEAN_CONFIG_CONTAINER
	START_MONITOR_SERVICE
	RESTART_MONITOR_SERVICE
	STOP_MONITOR_SERVICE
	INIT_MONITOR_STATUS
	GET_MONITOR_STATUS
	CLEAN_MONITOR_SERVICE
	SYNC_GRAFANA_DASHBOARD

	// fs
	CHECK_CLIENT_S3
	CREATE_DINGOFS
	MOUNT_FILESYSTEM
	UMOUNT_FILESYSTEM

	// playground
	CREATE_PLAYGROUND
	INIT_PLAYGROUND
	START_PLAYGROUND
	REMOVE_PLAYGROUND
	GET_PLAYGROUND_STATUS

	// dingo executor
	SYNC_JAVA_OPTS

	// unknown
	UNKNOWN
)

func (p *Playbook) createTasks(step *PlaybookStep) (*tasks.Tasks, error) {
	// (1) default tasks execute options
	config, err := NewSmartConfig(step.Configs)
	if err != nil {
		return nil, err
	}

	// (2) set key-value pair for options
	for k, v := range step.Options {
		p.dingoadm.MemStorage().Set(k, v)
	}

	// (3) create task one by one and added into tasks
	var t *task.Task
	once := map[string]bool{}
	dingoadm := p.dingoadm
	ts := tasks.NewTasks()
	for i := 0; i < config.Len(); i++ {
		// only need to execute task once per host
		switch step.Type {
		case CHECK_SSH_CONNECT,
			GET_HOST_DATE:
			host := config.GetDC(i).GetHost()
			if once[host] {
				continue
			}
			once[host] = true
		case PULL_IMAGE:
			host := config.GetDC(i).GetHost()
			image := config.GetDC(i).GetContainerImage()
			if once[host+"_"+image] {
				continue
			}
			once[host+"_"+image] = true
		case SYNC_MONITOR_ORIGIN_CONFIG:
			if config.GetMC(i).GetRole() != configure.ROLE_MONITOR_SYNC {
				continue
			}
		case SYNC_MONITOR_ALT_CONFIG:
			if config.GetMC(i).GetRole() == configure.ROLE_MONITOR_SYNC {
				continue
			}
		case SYNC_GRAFANA_DASHBOARD:
			if config.GetMC(i).GetRole() != configure.ROLE_GRAFANA {
				continue
			}
		}

		switch step.Type {
		// checker
		case CHECK_TOPOLOGY:
			t, err = checker.NewCheckTopologyTask(dingoadm, nil)
		case CHECK_SSH_CONNECT:
			t, err = checker.NewCheckSSHConnectTask(dingoadm, config.GetDC(i))
		case CHECK_PERMISSION:
			if config.GetDC(i).GetRole() == topology.ROLE_FS_MDS_CLI {
				continue
			}
			t, err = checker.NewCheckPermissionTask(dingoadm, config.GetDC(i))
		case CHECK_KERNEL_VERSION:
			t, err = checker.NewCheckKernelVersionTask(dingoadm, config.GetDC(i))
		case CHECK_KERNEL_MODULE:
			t, err = checker.NewCheckKernelModuleTask(dingoadm, config.GetCC(i))
		case CHECK_PORT_IN_USE:
			if config.GetDC(i).GetRole() == topology.ROLE_FS_MDS_CLI {
				continue
			}
			t, err = checker.NewCheckPortInUseTask(dingoadm, config.GetDC(i))
		case CHECK_DESTINATION_REACHABLE:
			if config.GetDC(i).GetRole() == topology.ROLE_FS_MDS_CLI {
				continue
			}
			t, err = checker.NewCheckDestinationReachableTask(dingoadm, config.GetDC(i))
		case START_HTTP_SERVER:
			if config.GetDC(i).GetRole() == topology.ROLE_FS_MDS_CLI {
				continue
			}
			t, err = checker.NewStartHTTPServerTask(dingoadm, config.GetDC(i))
		case CHECK_NETWORK_FIREWALL:
			t, err = checker.NewCheckNetworkFirewallTask(dingoadm, config.GetDC(i))
		case GET_HOST_DATE:
			t, err = checker.NewGetHostDate(dingoadm, config.GetDC(i))
		case CHECK_HOST_DATE:
			t, err = checker.NewCheckDate(dingoadm, nil)
		case CHECK_S3:
			t, err = checker.NewCheckS3Task(dingoadm, config.GetDC(i))
		case CHECK_MDS_ADDRESS:
			t, err = checker.NewCheckMdsAddressTask(dingoadm, config.GetCC(i))
		case CHECK_STORE_HEALTH:
			t, err = comm.NewCheckStoreHealthTask(dingoadm, config.GetDC(i))
		case CLEAN_PRECHECK_ENVIRONMENT:
			if config.GetDC(i).GetRole() == topology.ROLE_FS_MDS_CLI {
				continue
			}
			t, err = checker.NewCleanEnvironmentTask(dingoadm, config.GetDC(i))
		// common
		case PULL_IMAGE:
			t, err = comm.NewPullImageTask(dingoadm, config.GetDC(i))
		case CREATE_CONTAINER:
			t, err = comm.NewCreateContainerTask(dingoadm, config.GetDC(i))
		case CREATE_MDSV2_CLI_CONTAINER:
			t, err = comm.NewCreateMdsv2CliContainerTask(dingoadm, config.GetDC(i))
		case SYNC_CONFIG:
			t, err = comm.NewSyncConfigTask(dingoadm, config.GetDC(i))
		case START_SERVICE,
			START_ETCD,
			START_MDS,
			START_CHUNKSERVER,
			START_SNAPSHOTCLONE,
			START_METASERVER,
			START_FS_MDS,
			START_COORDINATOR,
			START_STORE,
			START_DINGODB_DOCUMENT,
			START_DINGODB_INDEX,
			START_DINGODB_DISKANN,
			START_MDSV2_CLI_CONTAINER,
			START_DINGODB_EXECUTOR,
			START_DINGODB_PROXY,
			START_DINGODB_WEB:
			t, err = comm.NewStartServiceTask(dingoadm, config.GetDC(i))
		case STOP_SERVICE:
			t, err = comm.NewStopServiceTask(dingoadm, config.GetDC(i))
		case RESTART_SERVICE:
			t, err = comm.NewRestartServiceTask(dingoadm, config.GetDC(i))
		case CREATE_META_TABLES:
			t, err = comm.NewCreateMetaTablesTask(dingoadm, config.GetDC(i))
		case INIT_SERVIE_STATUS:
			t, err = comm.NewInitServiceStatusTask(dingoadm, config.GetDC(i))
		case GET_SERVICE_STATUS:
			t, err = comm.NewGetServiceStatusTask(dingoadm, config.GetDC(i))
		case CLEAN_SERVICE:
			t, err = comm.NewCleanServiceTask(dingoadm, config.GetDC(i))
		case INIT_CLIENT_STATUS:
			t, err = comm.NewInitClientStatusTask(dingoadm, config.GetAny(i))
		case GET_CLIENT_STATUS:
			t, err = comm.NewGetClientStatusTask(dingoadm, config.GetAny(i))
		// fs
		case CHECK_CLIENT_S3:
			t, err = checker.NewClientS3ConfigureTask(dingoadm, config.GetCC(i))
		case CREATE_DINGOFS:
			t, err = fs.NewCreateDingoFSTask(dingoadm, config.GetCC(i))
		case MOUNT_FILESYSTEM:
			t, err = fs.NewMountFSTask(dingoadm, config.GetCC(i))
		case UMOUNT_FILESYSTEM:
			t, err = fs.NewUmountFSTask(dingoadm, config.GetCC(i))
		// monitor
		case PULL_MONITOR_IMAGE:
			t, err = monitor.NewPullImageTask(dingoadm, config.GetMC(i))
		case CREATE_MONITOR_CONTAINER:
			t, err = monitor.NewCreateContainerTask(dingoadm, config.GetMC(i))
		case SYNC_MONITOR_ORIGIN_CONFIG, SYNC_MONITOR_ALT_CONFIG:
			t, err = monitor.NewSyncConfigTask(dingoadm, config.GetMC(i))
		case SYNC_HOSTS_MAPPING:
			t, err = monitor.NewSyncHostsMappingTask(dingoadm, config.GetMC(i))
		case CLEAN_CONFIG_CONTAINER:
			t, err = monitor.NewCleanConfigContainerTask(dingoadm, config.GetMC(i))
		case START_MONITOR_SERVICE:
			t, err = monitor.NewStartServiceTask(dingoadm, config.GetMC(i))
		case RESTART_MONITOR_SERVICE:
			t, err = monitor.NewRestartServiceTask(dingoadm, config.GetMC(i))
		case STOP_MONITOR_SERVICE:
			t, err = monitor.NewStopServiceTask(dingoadm, config.GetMC(i))
		case INIT_MONITOR_STATUS:
			t, err = monitor.NewInitMonitorStatusTask(dingoadm, config.GetMC(i))
		case GET_MONITOR_STATUS:
			t, err = monitor.NewGetMonitorStatusTask(dingoadm, config.GetMC(i))
		case CLEAN_MONITOR_SERVICE:
			t, err = monitor.NewCleanMonitorTask(dingoadm, config.GetMC(i))
		// dingo executor
		case SYNC_JAVA_OPTS:
			t, err = comm.NewSyncJavaOptsTask(dingoadm, config.GetDC(i))
		case SYNC_GRAFANA_DASHBOARD:
			t, err = monitor.NewSyncGrafanaDashboardTask(dingoadm, config.GetMC(i))

		default:
			return nil, errno.ERR_UNKNOWN_TASK_TYPE.
				F("task type: %d", step.Type)
		}

		if err != nil {
			return nil, err // already is error code
		} else if t == nil {
			continue
		}

		if config.GetType() == TYPE_CONFIG_DEPLOY { // merge task status into one
			t.SetTid(config.GetDC(i).GetId())
			t.SetPtid(config.GetDC(i).GetParentId())
		}
		ts.AddTask(t)
	}

	return ts, nil
}
