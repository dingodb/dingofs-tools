package common

import (
	pbmdsv2 "github.com/dingodb/dingofs-tools/proto/dingofs/proto/mdsv2"
	"sync"
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
