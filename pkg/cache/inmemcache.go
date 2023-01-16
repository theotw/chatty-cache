/*
 * Copyright (c) The One True Way 2023. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package cache

import (
	"encoding/base64"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"github.com/theotw/chatty-cache/pkg/chatter"
	"github.com/theotw/chatty-cache/pkg/model"
	"sort"
	"sync"
	"time"
)

type cacheEntry struct {
	CacheName string
	CacheKey  string
	CacheData []byte

	// cacheTime time this was cached
	cacheTime   time.Time
	lastTouched time.Time

	//cacheSize is size in bytes of this message when jsonified
	cacheSize uint64
}

func (t *cacheEntry) touch() {
	t.lastTouched = time.Now()
}

type InMemCache struct {
	maxCacheSize       uint64
	caches             map[string]map[string]*cacheEntry
	totalUsedCacheSize uint64
	lock               sync.RWMutex
	chatter            chatter.CacheChatter
}

// NewInMemCache Creates a new in memory cache with maxh size and an optional chatter relay to share messages across processes
func NewInMemCache(maxSize uint64, chatter chatter.CacheChatter) *InMemCache {
	ret := new(InMemCache)
	ret.maxCacheSize = maxSize
	ret.caches = make(map[string]map[string]*cacheEntry, 0)
	ret.chatter = chatter
	if ret.chatter != nil {
		ret.chatter.RegisterListenerForReplicatedObjects(func(message *model.CacheRelayMessage) {
			ret.listenerForMessages(message)
		})
	}
	return ret
}
func (t *InMemCache) listenerForMessages(message *model.CacheRelayMessage) {
	bits, err := base64.StdEncoding.DecodeString(message.CacheValue)
	if err != nil {
		log.WithError(err).Error("Unable to base 64 decode a cache relay message")
		return
	}
	t.putBits(message.CacheName, message.CacheKey, bits)
}

// Put  puts an value into the cache
func (t *InMemCache) Put(cacheName string, cacheKey string, value interface{}) error {

	jsonBits, err := json.Marshal(value)
	if err != nil {
		err := NewCacheError(NotJsonifiable, err)
		return err
	}
	err = t.putBits(cacheName, cacheKey, jsonBits)
	if t.chatter != nil {
		//send a replicate message
		var replicate model.CacheRelayMessage
		replicate.CacheName = cacheName
		replicate.CacheKey = cacheKey
		replicate.CacheValue = base64.StdEncoding.EncodeToString(jsonBits)
		t.chatter.ReplicateCachedObject(&replicate)
	}
	return err
}
func (t *InMemCache) putBits(cacheName, cacheKey string, valueJsonBits []byte) error {
	x := new(cacheEntry)
	x.CacheKey = cacheKey
	x.CacheName = cacheName
	x.cacheTime = time.Now()
	x.touch()

	x.CacheData = valueJsonBits
	x.cacheSize = uint64(len(valueJsonBits))
	if x.cacheSize > t.maxCacheSize {
		return NewCacheError(ExceedsTotalCacheSize, nil)
	}
	var ret *CacheError
	// DO NOT RETURN BETWEEN THESE LOCK/UNLOCK
	//I dont like defers for unlock, I want it unlocked asap, not sitting as waiting on the stack
	t.lock.Lock()
	newTotalSize := t.totalUsedCacheSize + x.cacheSize
	//0 means no size checks
	if t.maxCacheSize > 0 && newTotalSize > t.maxCacheSize {
		ret = t.evict(newTotalSize - t.maxCacheSize)
	}
	if ret == nil {
		t.totalUsedCacheSize = newTotalSize
		m, ok := t.caches[cacheName]
		if !ok {
			m = make(map[string]*cacheEntry)
			t.caches[cacheName] = m
		}
		m[cacheKey] = x
	}

	defer t.lock.Unlock()

	return ret
}

// Get gets a value from the cache, if the item is not found, a CacheError is returned
func (t *InMemCache) Get(cacheName string, cacheKey string, valOut interface{}) error {
	var entry *cacheEntry
	t.lock.RLock()
	cache, ok := t.caches[cacheName]
	if ok {
		entry = cache[cacheKey]
	}
	t.lock.RUnlock()
	if entry == nil {
		return NewCacheError(NoItem, nil)
	}
	entry.touch()
	//if you are wondering how we can get an error on a bit stream we made, it is because it
	//may have been made in another process space and thus mismatched
	err := json.Unmarshal(entry.CacheData, valOut)
	if err != nil {
		NewCacheError(NotJsonifiable, err)
	}
	return err
}

// evict toss out oldest touch entries until evictCount bytes are freed
func (t *InMemCache) evict(evictCount uint64) *CacheError {
	last := t.sortLastTouched()
	var amountFreed uint64
	for _, x := range last {
		entry := t.caches[x.CacheName][x.CacheKey]
		amountFreed = amountFreed + entry.cacheSize
		delete(t.caches[x.CacheName], x.CacheKey)
		if amountFreed >= evictCount {
			break
		}
	}
	if amountFreed < evictCount {
		return NewCacheError(ObjectToLarge, nil)
	}
	return nil
}

func (t *InMemCache) sortLastTouched() []*cacheEntry {
	masterList := make([]*cacheEntry, 0)
	for _, v := range t.caches {
		cacheList := make([]*cacheEntry, len(v))
		i := 0
		for _, vp := range v {
			cacheList[i] = vp
			i++
		}
		masterList = append(masterList, cacheList...)
	}
	sort.Slice(masterList, func(i, j int) bool {
		return masterList[i].lastTouched.Before(masterList[j].lastTouched)
	})
	return masterList
}
