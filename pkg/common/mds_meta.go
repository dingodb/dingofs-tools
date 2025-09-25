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

package common

import (
	"sync"

	pbmdsv2 "github.com/dingodb/dingofs-tools/proto/dingofs/proto/mdsv2"
)

type MDSMeta struct {
	mdsMap map[int64]*pbmdsv2.MDS // mds_id -> mds_meta
	mux    sync.RWMutex
}

func (m *MDSMeta) SetMDS(mdsSlice []*pbmdsv2.MDS) {
	m.mux.Lock()
	defer m.mux.Unlock()

	// fill map mds_id -> mds_meta
	for _, mds := range mdsSlice {
		m.mdsMap[mds.Id] = mds
	}
}

func (m *MDSMeta) GetMDS(mdsId int64) (*pbmdsv2.MDS, bool) {
	m.mux.RLock()
	defer m.mux.RUnlock()
	mds, ok := m.mdsMap[mdsId]

	return mds, ok
}

func NewMDSMeta() *MDSMeta {
	return &MDSMeta{
		mdsMap: make(map[int64]*pbmdsv2.MDS),
	}
}
