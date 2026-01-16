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

	pbmds "github.com/dingodb/dingocli/proto/dingofs/proto/mds"
)

type MDSRouter interface {
	Init(mdsSlice []*pbmds.MDS, partitionPolicy *pbmds.PartitionPolicy) bool
	GetMDS(inodeId uint64) (*pbmds.MDS, bool)
}

var _ MDSRouter = (*HashMDSRouter)(nil) // check interface
var _ MDSRouter = (*MonoMDSRouter)(nil) // check interface

type HashMDSRouter struct {
	mux           sync.RWMutex
	hashPartition *pbmds.HashPartition
	mdsIdMap      map[uint32]int64 // bucket_id -> mds_id
	mdsMeta       *MDSMeta
}

func (h *HashMDSRouter) Init(mdsSlice []*pbmds.MDS, partitionPolicy *pbmds.PartitionPolicy) bool {
	h.mux.Lock()
	defer h.mux.Unlock()

	// init mds
	h.mdsMeta.SetMDS(mdsSlice)

	// fill mdsMap
	h.hashPartition = partitionPolicy.GetParentHash()
	for mdsId, bucketSets := range h.hashPartition.GetDistributions() {
		_, ok := h.mdsMeta.GetMDS(int64(mdsId))
		if ok {
			for _, bucketId := range bucketSets.BucketIds {
				h.mdsIdMap[bucketId] = int64(mdsId)
			}
		} else {
			return false
		}
	}

	return true
}

func (h *HashMDSRouter) GetMDS(inodeId uint64) (*pbmds.MDS, bool) {
	h.mux.RLock()
	defer h.mux.RUnlock()

	bucketId := uint32(inodeId % uint64(h.hashPartition.BucketNum))
	mdsId, ok := h.mdsIdMap[bucketId]
	if ok {
		mdsMeta, ok := h.mdsMeta.GetMDS(mdsId)
		if ok {
			return mdsMeta, true
		}
	}

	return nil, false
}

type MonoMDSRouter struct {
	mux     sync.RWMutex
	mds     *pbmds.MDS
	mdsMeta *MDSMeta
}

func (m *MonoMDSRouter) Init(mdsSlice []*pbmds.MDS, partitionPolicy *pbmds.PartitionPolicy) bool {
	m.mux.Lock()
	defer m.mux.Unlock()

	// init mds
	m.mdsMeta.SetMDS(mdsSlice)

	mdsId := partitionPolicy.GetMono().GetMdsId()
	mds, ok := m.mdsMeta.GetMDS(int64(mdsId))
	if !ok {
		return false
	}
	m.mds = mds

	return true
}

func (m *MonoMDSRouter) GetMDS(inodeId uint64) (*pbmds.MDS, bool) {
	m.mux.RLock()
	defer m.mux.RUnlock()

	if m.mds == nil {
		return nil, false
	}

	return m.mds, true
}

func NewMDSRouter(partitionType pbmds.PartitionType) MDSRouter {
	if partitionType == pbmds.PartitionType_PARENT_ID_HASH_PARTITION {
		return &HashMDSRouter{
			mdsIdMap: make(map[uint32]int64),
			mdsMeta:  NewMDSMeta(),
		}
	}

	return &MonoMDSRouter{
		mdsMeta: NewMDSMeta(),
	}
}
