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

package scripts

import (
	_ "embed"
)

var (
	// Common

	//go:embed shell/wait.sh
	WAIT string

	//go:embed shell/start_nginx.sh
	START_NGINX string

	// DingoFS Mds
	//go:embed shell/create_mds_tables.sh
	CREATE_MDS_TABLES string

	// DingoStore
	//go:embed shell/check_store_health.sh
	CHECK_STORE_HEALTH string

	// DingoFS Executor
	//go:embed shell/sync_java_opts.sh
	SYNC_JAVA_OPTS string

	// DingoFS Monitor
	//go:embed shell/start_monitor_sync.sh
	START_MONITOR_SYNC string

	// Prometheus Node Exporter
	//go:embed shell/sync_prometheus.sh
	SYNC_PROMETHEUS string

	// Grafana dashboard
	//go:embed shell/server_metric_zh.json
	GRAFANA_SERVER_METRIC string

	// Extract /etc/hosts mapping
	//go:embed shell/extract_hosts.sh
	EXTRACT_HOSTS string
)
