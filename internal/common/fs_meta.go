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

type FsMeta struct {
	fsMap map[uint32]*pbmds.FsInfo // fsid -> fsinfo
	mux   sync.RWMutex
}

func (f *FsMeta) SetFsInfo(fsInfo *pbmds.FsInfo) {
	f.mux.Lock()
	defer f.mux.Unlock()

	f.fsMap[fsInfo.GetFsId()] = fsInfo
}

func (f *FsMeta) GetFsInfo(fsId uint32) (*pbmds.FsInfo, bool) {
	f.mux.RLock()
	defer f.mux.RUnlock()

	fsInfo, ok := f.fsMap[fsId]

	return fsInfo, ok
}

func NewFsMeta() *FsMeta {
	return &FsMeta{
		fsMap: make(map[uint32]*pbmds.FsInfo),
	}
}
