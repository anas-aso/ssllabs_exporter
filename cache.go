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
	id         string
	expiryTime int64
}

type cache struct {
	mu         sync.Mutex
	entries    map[string]*prometheus.Gatherer
	lru        *list.List
	retention  time.Duration
	pruneDelay time.Duration
}

func (c *cache) add(id string, result *prometheus.Gatherer) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry := &cacheEntry{
		id:         id,
		expiryTime: int64(c.retention.Seconds()) + time.Now().Unix(),
	}

	_, alreadyExists := c.entries[id]
	if alreadyExists {
		e := c.lru.Front()
		for e != nil {
			entry := e.Value.(*cacheEntry)
			if entry.id == id {
				c.lru.Remove(e)
				break
			}
			e = e.Next()
		}
	}

	c.lru.PushBack(entry)
	c.entries[id] = result
}

func (c *cache) get(id string) prometheus.Gatherer {
	c.mu.Lock()
	defer c.mu.Unlock()

	result, found := c.entries[id]
	if found {
		return *result
	}

	return nil
}

func (c *cache) prune() {
	c.mu.Lock()
	defer c.mu.Unlock()

	e := c.lru.Front()

	for e != nil {
		entry := e.Value.(*cacheEntry)

		if entry.expiryTime > time.Now().Unix() {
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
		entries:    make(map[string]*prometheus.Gatherer),
		lru:        list.New(),
		retention:  retention,
		pruneDelay: pruneDelay,
	}

	go c.start()

	return c
}
