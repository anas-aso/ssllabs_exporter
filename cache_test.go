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
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func TestCacheOperations(t *testing.T) {
	// TODO: split the tests per function
	pruneDelay := 1 * time.Second
	retention := 2 * time.Second

	cache := newCache(pruneDelay, retention)

	registry := prometheus.NewRegistry()
	metric := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "metric",
	})
	registry.MustRegister(metric)
	metric.Set(1)
	garether := prometheus.Gatherer(registry)

	cache.add("testDomain", &garether)

	entry := cache.get("testDomain")
	mfs, _ := entry.Gather()

	// check the content of the cached registry
	if len(mfs) != 1 {
		t.Errorf("Cached registry contains more metrics than expected.\nExpected : %v\nGot : %v\n", 1, len(mfs))
	}
	if mfs[0].GetName() != "metric" {
		t.Errorf("Cached registry contains wrong metric name.\nExpected : %v\nGot : %v\n", "metric", mfs[0].GetName())
	}

	// wait for the cache to expire
	time.Sleep(pruneDelay + retention)

	// check the cache staleness
	entry = cache.get("testDomain")
	if entry != nil {
		t.Errorf("Cache contains stale data.\nFound entry : %v\n", entry)
	}

	// check the cache content
	if cache.lru.Len() != 0 || len(cache.entries) != 0 {
		t.Errorf("Cache contains unexpected data.\nFound entry(ies) : %v\n", cache.entries)
	}

	// add the entry element twice
	cache.add("testDomain", &garether)
	cache.add("testDomain", &garether)
	// check the cache content
	if cache.lru.Len() != 1 || len(cache.entries) != 1 {
		var dupEntries []cacheEntry
		for e := cache.lru.Front(); e != nil; e = e.Next() {
			dupEntries = append(dupEntries, *e.Value.(*cacheEntry))
		}
		t.Errorf("Cache contains duplicate entries.\nFound entry(ies) : %v\n", dupEntries)
	}
}
