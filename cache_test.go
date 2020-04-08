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

func TestAddGet(t *testing.T) {
	// initialize cache
	pruneDelay := 1 * time.Minute
	retention := 1 * time.Minute
	cache := newCache(pruneDelay, retention)

	// create test registry
	registry := prometheus.NewRegistry()

	// test adding a cache entry
	entryID := "testDomain"
	metricName := "metric"

	metric := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: metricName,
	})
	registry.MustRegister(metric)

	cache.add(entryID, prometheus.Gatherer(registry))

	// fetch the cached entry and verify contents
	entry := cache.get(entryID)
	mfs, _ := entry.Gather()

	// check the content of the cached registry
	if len(mfs) != 1 {
		t.Errorf("Cached registry contains more metrics than expected.\nExpected : %v\nGot : %v\n", 1, len(mfs))
	}
	if mfs[0].GetName() != metricName {
		t.Errorf("Cached registry contains wrong metric name.\nExpected : %v\nGot : %v\n", metricName, mfs[0].GetName())
	}

	// add 2nd entry
	cache.add(entryID+"_2nd", prometheus.Gatherer(registry))
	if cache.lru.Len() != 2 || len(cache.entries) != 2 {
		var dupEntries []cacheEntry
		for e := cache.lru.Front(); e != nil; e = e.Next() {
			dupEntries = append(dupEntries, *e.Value.(*cacheEntry))
		}
		t.Errorf("Cache doesn't contain expected entries count.\nFound entries : %v\n", dupEntries)
	}

	// add a duplicate entry
	cache.add(entryID+"_2nd", prometheus.Gatherer(registry))
	if cache.lru.Len() != 2 || len(cache.entries) != 2 {
		var dupEntries []cacheEntry
		for e := cache.lru.Front(); e != nil; e = e.Next() {
			dupEntries = append(dupEntries, *e.Value.(*cacheEntry))
		}
		t.Errorf("Cache contains duplicate entries.\nFound entries : %v\n", dupEntries)
	}

	// fetch non-existing content
	nonExistingEntry := cache.get("404")
	if nonExistingEntry != nil {
		t.Errorf("Cache returns unexpected result.\nFound : %v\n", nonExistingEntry)
	}
}

func TestPrune(t *testing.T) {
	// initialize cache
	pruneDelay := 1 * time.Second
	retention := 2 * time.Second
	cache := newCache(pruneDelay, retention)

	// create test registry
	registry := prometheus.NewRegistry()

	// test adding a cache entry
	entryID := "testDomain"
	metricName := "metric"

	metric := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: metricName,
	})
	registry.MustRegister(metric)

	cache.add(entryID, prometheus.Gatherer(registry))

	// wait for the cache to expire
	time.Sleep(retention + pruneDelay)

	// check the cache staleness
	entry := cache.get(entryID)
	if entry != nil {
		t.Errorf("Cache contains stale data")
	}
}
