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

// cacheEntry contains cache elements meta data
type cacheEntry struct {
	// the target host is used as a unique cache entry identifier
	id string

	// expiry time for the cache entry (calculated on creation time)
	expiryTime int64
}

type cache struct {
	mu sync.Mutex

	// map of cached Prometheus registry for a fast access
	entries map[string]*prometheus.Gatherer

	// a linked ordered list for a faster cache retention
	lru *list.List

	// how long each cache entry should be kept
	retention time.Duration

	// how frequent the cache retention is verified/applied
	pruneDelay time.Duration
}

// add a new cache entry or update it if already exists
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
			if e.Value.(*cacheEntry).id == id {
				c.lru.MoveToBack(e)
				break
			}
			e = e.Next()
		}
	} else {
		c.lru.PushBack(entry)
	}

	c.entries[id] = result
}

// retrieve a cache entry if exists, otherwise return nil
func (c *cache) get(id string) prometheus.Gatherer {
	c.mu.Lock()
	defer c.mu.Unlock()

	result, found := c.entries[id]
	if found {
		return *result
	}

	return nil
}

// prune expired entries from the cache
func (c *cache) prune() {
	c.mu.Lock()
	defer c.mu.Unlock()

	e := c.lru.Front()

	for e != nil {
		entry := e.Value.(*cacheEntry)

		// since the list is ordered, we can stop the iteration once a fresh element is found
		if entry.expiryTime > time.Now().Unix() {
			break
		}

		next := e.Next()
		c.lru.Remove(e)
		delete(c.entries, entry.id)
		e = next
	}
}

// start a time ticker to remove expired cache entries
func (c *cache) start() {
	ticker := time.NewTicker(c.pruneDelay)

	for range ticker.C {
		c.prune()
	}
}

// create a new cache and start the retention worker in the background
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
