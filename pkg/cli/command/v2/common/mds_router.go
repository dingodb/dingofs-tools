package common

import (
	pbmdsv2 "github.com/dingodb/dingofs-tools/proto/dingofs/proto/mdsv2"
	"sync"
)

type MDSRouter interface {
	Init(mdsSlice []*pbmdsv2.MDS, partitionPolicy *pbmdsv2.PartitionPolicy) bool
	GetMDS(inodeId uint64) (*pbmdsv2.MDS, bool)
}

var _ MDSRouter = (*HashMDSRouter)(nil) // check interface
var _ MDSRouter = (*MonoMDSRouter)(nil) // check interface

type HashMDSRouter struct {
	mux           sync.RWMutex
	hashPartition *pbmdsv2.HashPartition
	mdsIdMap      map[uint32]int64 // bucket_id -> mds_id
	mdsMeta       *MDSMeta
}

func (h *HashMDSRouter) Init(mdsSlice []*pbmdsv2.MDS, partitionPolicy *pbmdsv2.PartitionPolicy) bool {
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

func (h *HashMDSRouter) GetMDS(inodeId uint64) (*pbmdsv2.MDS, bool) {
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
	mds     *pbmdsv2.MDS
	mdsMeta *MDSMeta
}

func (m *MonoMDSRouter) Init(mdsSlice []*pbmdsv2.MDS, partitionPolicy *pbmdsv2.PartitionPolicy) bool {
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

func (m *MonoMDSRouter) GetMDS(inodeId uint64) (*pbmdsv2.MDS, bool) {
	m.mux.RLock()
	defer m.mux.RUnlock()

	if m.mds == nil {
		return nil, false
	}

	return m.mds, true
}

func NewMDSRouter(partitionType pbmdsv2.PartitionType) MDSRouter {
	if partitionType == pbmdsv2.PartitionType_PARENT_ID_HASH_PARTITION {
		return &HashMDSRouter{
			mdsIdMap: make(map[uint32]int64),
			mdsMeta:  NewMDSMeta(),
		}
	}

	return &MonoMDSRouter{
		mdsMeta: NewMDSMeta(),
	}
}
