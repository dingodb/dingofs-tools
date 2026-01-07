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
)

type ConcurrentMetaMap[K comparable, V any] struct {
	metaMap map[K]V
	mux     sync.RWMutex
}

func NewConcurrentMetaMap[K comparable, V any]() *ConcurrentMetaMap[K, V] {
	return &ConcurrentMetaMap[K, V]{
		metaMap: make(map[K]V),
	}
}

func (c *ConcurrentMetaMap[K, V]) Load(key K) (V, bool) {
	c.mux.RLock()
	defer c.mux.RUnlock()
	val, ok := c.metaMap[key]
	return val, ok
}

func (c *ConcurrentMetaMap[K, V]) Store(key K, val V) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.metaMap[key] = val
}

func (c *ConcurrentMetaMap[K, V]) Delete(key K) {
	c.mux.Lock()
	defer c.mux.Unlock()
	delete(c.metaMap, key)
}
