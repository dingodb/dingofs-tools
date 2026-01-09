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

package build

import (
	"encoding/json"
	"fmt"
	"os"
)

const (
	DEBUG_CONFIGURE         = "DEBUG_CONFIGURE"
	DEBUG_HOSTS             = "DEBUG_HOSTS"
	DEBUG_SMART_CONFIGS     = "DEBUG_SMART_CONFIGS"
	DEBUG_TOPOLOGY          = "DEBUG_TOPOLOGY"
	DEBUG_TOOL              = "DEBUG_TOOL"
	DEBUG_CLUSTER           = "DEBUG_CLUSTER"
	DEBUG_CREATE_POOL       = "DEBUG_CREATE_POOL"
	DEBUG_CLIENT_CONFIGURE  = "DEBUG_CLIENT_CONFIGURE"
	DEBUG_MONITOR_CONFIGURE = "DEBUG_MONITOR_CONFIGURE"
)

type Field struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

func DEBUG(name string, a ...interface{}) {
	if !DEBUG_MODLE || os.Getenv(name) != "1" {
		return
	}

	fmt.Printf("%s:\n", name)
	for _, v := range a {
		bytes, _ := json.MarshalIndent(v, "", "    ")
		fmt.Println(string(bytes))
	}
}
