package common

import (
	pbmdsv2 "github.com/dingodb/dingofs-tools/proto/dingofs/proto/mdsv2"
	"sync"
)

type FsMeta struct {
	fsMap map[uint32]*pbmdsv2.FsInfo // fsid -> fsinfo
	mux   sync.RWMutex
}

func (f *FsMeta) SetFsInfo(fsInfo *pbmdsv2.FsInfo) {
	f.mux.Lock()
	defer f.mux.Unlock()

	f.fsMap[fsInfo.GetFsId()] = fsInfo
}

func (f *FsMeta) GetFsInfo(fsId uint32) (*pbmdsv2.FsInfo, bool) {
	f.mux.RLock()
	defer f.mux.RUnlock()

	fsInfo, ok := f.fsMap[fsId]

	return fsInfo, ok
}

func NewFsMeta() *FsMeta {
	return &FsMeta{
		fsMap: make(map[uint32]*pbmdsv2.FsInfo),
	}
}
