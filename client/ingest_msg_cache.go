package client

import (
	"time"
	"crypto/sha256"
	"sync"
	"github.com/ao-data/albiondata-client/log"
)

const IngestMsgCacheSize = 1024
const IngestMsgCacheLifeTime = 600

type IngestMsgCache struct{
	entries       []IngestMsgCacheEntry
	mutex         sync.RWMutex
	hits          int
	misses        int
}

type IngestMsgCacheEntry struct {
	msghash [32]byte
	added   time.Time
}

func (cache *IngestMsgCache) isDuplicate(msg []byte) bool {
	if !ConfigGlobal.Dedup{
		return false
	}

	duplicate := false
	msghash := sha256.Sum256(msg)

	cache.mutex.RLock()
	for index, entry := range cache.entries {
		if entry.msghash == msghash {
			duplicate = true
			if time.Since(entry.added).Seconds() > IngestMsgCacheLifeTime {
				// A match has been found but it has expired
				duplicate = false
				cache.entries[index].msghash = [32]byte{}
			}
			if duplicate { break }
		}
	}
	cache.mutex.RUnlock()

	if !duplicate {
		if ConfigGlobal.DedupStats{ cache.misses += 1 }
		new_cache_entry := IngestMsgCacheEntry{
			msghash: msghash,
			added: time.Now(),
		}
		// Add the new msg that turned out to be unique to the cache
		cache.mutex.Lock()
		cache.entries = append(cache.entries, new_cache_entry)
		// Chop the Cache to it's max capacity of IngestMsgCacheSize
		// Removing the oldest entries first by removing from the front
		if len(cache.entries) > IngestMsgCacheSize {
			cache.entries = cache.entries[len(cache.entries)-IngestMsgCacheSize:]
		}
		cache.mutex.Unlock()
	} else {
		if ConfigGlobal.DedupStats{ cache.hits += 1 }
	}
	
	return duplicate
}

func (cache *IngestMsgCache) StatsReporter() {
	if !ConfigGlobal.Dedup {
		// Dedup is disabled. No point in collecting stats
		ConfigGlobal.DedupStats = false
	} else {
		for {
			time.Sleep(300 * time.Second)
			if cache.hits > 0 {
				log.Infof("MsgCacheStatistics: %d/%d/%d%% [hits/misses/hitrate]", cache.hits, cache.misses, (cache.hits * 100) / (cache.hits + cache.misses))
			}
		}
	}
}
