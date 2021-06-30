package main

import (
	"sync"
	"time"
)

const (
	NoExpiration      time.Duration = -1
	DefaultExpiration time.Duration = 0
)

/* == INTERFACES & STRUCTS == */

type GenericMemoryCache interface {
	Get(key string) (entry interface{}, found bool)
	Set(key string, data interface{}, ttl time.Duration)
	Delete(key string)
}

type cacheLine struct {
	sync.RWMutex //reader/writer mutual exclusion lock
	cleanUP      *cleanUP
	entries      map[string]Entry
	ttl          time.Duration //time to live
}

type Entry struct {
	Expiration int64
	Data       interface{}
}

type cleanUP struct {
	Interval, cacheLifeSpan time.Duration
	notActive               chan bool
}

/* == FUNCTIONS & METHODS == */

// Retruns true if entry's expired
func (entry Entry) Expired() bool {
	if entry.Expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > entry.Expiration
}

// Gets an entriy from the cahce. If key was found --> entry & true.
// Otherwise --> nil & false
func (cache *cacheLine) Get(key string) (interface{}, bool) {
	cache.RLock()         // <-- LOCK
	defer cache.RUnlock() // <-- UNLOCK

	entry, found := cache.entries[key]
	if !found || entry.Expired() {
		return nil, false
	}
	return entry.Data, true
}

// Sets new key and value to cache line within cache, replacing old ones.
// If ttl == DefaultExpiration (0), the default expiration time is used.
// Else if ttl == NoExpiration (-1), the data never expires.
// Else if ttl != (DefaultExpiration || NoExpiration), the expiration time is set to a custom one.
func (cache *cacheLine) Set(key string, data interface{}, ttl time.Duration) {
	var exp int64

	if ttl == DefaultExpiration {
		ttl = cache.ttl
	}
	if ttl > 0 {
		exp = time.Now().Add(ttl).UnixNano()
	}

	cache.Lock()         // <-- LOCK
	defer cache.Unlock() // <-- UNLOCK

	cache.entries[key] = Entry{
		Expiration: exp,
		Data:       data,
	}

}

// Delete's subroutine. Actually deletes the entry from cache line. Returns deleted entry's key and truth if requested key exists. Otherwise returns nil and false
func (cl *cacheLine) delete(key string) (interface{}, bool) {

	//if cl.onEvicted != nil {
	if entry, found := cl.entries[key]; found {
		delete(cl.entries, key)
		return entry.Data, true
	}
	//}
	delete(cl.entries, key)
	return nil, false
}

// Deletes entry from the cache line and puts it on evicted line. Otherwise, if the key doesnt match the request does nothing.
func (cl *cacheLine) Delete(key string) {

	cl.Lock()         // <-- LOCK
	defer cl.Unlock() // <-- UNLOCK

	cl.delete(key)

}

// Deletes all expired entries from cache line.
func (cl *cacheLine) DeleteExpired() {

	cl.Lock()         // <-- LOCK
	defer cl.Unlock() // <-- UNLOCK

	for key, value := range cl.entries {
		if value.Expired() {
			cl.delete(key)
		}
	}
}

//Purges all cacheline entries within Cache
func purgeCache(cl *cacheLine) {
	cl.Lock()
	defer cl.Unlock()

	for key := range cl.entries {
		cl.delete(key)
	}
}

// Runs a cleanup routine
func (cleaner *cleanUP) Run(cl *cacheLine) {
	ticker := time.NewTicker(cleaner.Interval)
	timer := time.NewTimer(cleaner.cacheLifeSpan)
	if cleaner.cacheLifeSpan <= 0 {
		timer.Stop()
	}

	for {
		select {
		case <-ticker.C:
			cl.DeleteExpired()
		case <-timer.C:
			purgeCache(cl)
			ticker.Stop()
			return
		}
	}
}

// Starts cleaner routine
func startCleaner(cl *cacheLine, ttl, cUP time.Duration) {
	cleaner := &cleanUP{
		Interval:      cUP,
		cacheLifeSpan: ttl,
		notActive:     make(chan bool),
	}
	cl.cleanUP = cleaner
	go cleaner.Run(cl)
}

// newACcache()'s subroutine. Creates new default cache line ready to get set to new values.
func newCacheLine(ttl time.Duration, new_entry map[string]Entry) *cacheLine {
	if ttl == DefaultExpiration {
		ttl = NoExpiration
	}

	cl := &cacheLine{
		entries: new_entry,
		ttl:     ttl,
	}
	return cl
}

// New()'s subroutine. Creates a new cache with automatic clean up. If clean up interval set to default (0) or
// no-expiration (-1), the autoclean routine doesn't start
func newACcache(ttl, cUP time.Duration, new_entry map[string]Entry) *cacheLine {
	cl := newCacheLine(ttl, new_entry)

	if cUP > 0 {
		startCleaner(cl, ttl, cUP)
	}
	return cl
}

// Cache constructor
func New(ttl, cUP time.Duration) *cacheLine {
	entries := make(map[string]Entry)
	return newACcache(ttl, cUP, entries)
}

