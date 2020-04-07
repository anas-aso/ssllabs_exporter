// Copyright 2020 Anas Ait Said Oubrahim

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"container/list"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type cacheEntry struct {
	id           string
	creationTime int64
	result       *prometheus.Gatherer
}

type cache struct {
	mu         sync.Mutex
	entries    map[string]*cacheEntry
	lru        *list.List
	retention  time.Duration
	pruneDelay time.Duration
}

func (c *cache) add(id string, result *prometheus.Gatherer) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry := &cacheEntry{
		id:           id,
		creationTime: time.Now().Unix(),
		result:       result,
	}

	c.entries[id] = entry
	c.lru.PushBack(entry)
}

func (c *cache) get(id string) (result *cacheEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()

	result, found := c.entries[id]
	if found {
		return result
	}

	return nil
}

func (c *cache) prune() {
	c.mu.Lock()
	defer c.mu.Unlock()

	e := c.lru.Front()

	for e != nil {
		entry := e.Value.(*cacheEntry)
		expiryTime := entry.creationTime + int64(c.retention.Seconds())

		if expiryTime > time.Now().Unix() {
			break
		}

		next := e.Next()
		c.lru.Remove(e)
		delete(c.entries, entry.id)
		e = next
	}
}

func (c *cache) start() {
	ticker := time.NewTicker(c.pruneDelay)

	for range ticker.C {
		c.prune()
	}
}

func newCache(pruneDelay, retention time.Duration) *cache {
	c := &cache{
		entries:    make(map[string]*cacheEntry),
		lru:        list.New(),
		retention:  retention,
		pruneDelay: pruneDelay,
	}

	go c.start()

	return c
}
