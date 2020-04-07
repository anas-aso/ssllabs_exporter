package main

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func TestCacheOperations(t *testing.T) {
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
	mfs, _ := (*entry.result).Gather()

	// check the content of the cached registry
	if len(mfs) != 1 {
		t.Errorf("Cached registry contains more metrics than expected.\nExpected : %v\nGot : %v\n", 1, len(mfs))
	}
	if mfs[0].GetName() != "metric" {
		t.Errorf("Cached registry contains wrong metric name.\nExpected : %v\nGot : %v\n", "metric", mfs[0].GetName())
	}

	// wait for the prune delay to expire
	time.Sleep(pruneDelay + retention)

	// check the cache expiry
	entry = cache.get("testDomain")
	if entry != nil {
		t.Errorf("Cache contains stale data.\nFound entry : %v\n", entry)
	}

	// check the cache content
	if cache.lru.Len() != 0 || len(cache.entries) != 0 {
		t.Errorf("Cache contains unexpected data.\nFound entry(ies) : %v\n", cache.entries)
	}
}
