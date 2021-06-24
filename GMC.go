package main

import (
	"runtime"
	"sync"
	"time"
)

const (
	NoExpiration      time.Duration = -1
	DefaultExpiration time.Duration = 0
)

/* == WHATEVER THIS IS == */

type GenericMemoryCache interface {
	Get(key string) (entry interface{}, found bool)
	Set(key string, data interface{}, ttl time.Duration)
	Delete(key string)
}

type cacheLine struct {
	sync.RWMutex                  //reader/writer mutual exclusion lock
	cleanUP      *cleanUP         //mr.proper (autocleans cache)
	entries      map[string]Entry //entries collection
	ttl          time.Duration    //time to live
}

type Entry struct {
	Expiration int64       // expiration of a particular entry
	Data       interface{} // data copy within requested memory location
}

type cleanUP struct {
	Interval  time.Duration
	notActive chan bool
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
	cache.RLock() // <-- LOCK

	entry, found := cache.entries[key]
	if !found {
		cache.RUnlock() // <-- UNLOCK
		return nil, false
	}
	if entry.Expired() {
		cache.RUnlock() // <-- UNLOCK
		return nil, false
	}
	cache.RUnlock() // <-- UNLOCK
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

	cache.Lock() // <-- LOCK
	cache.entries[key] = Entry{
		Expiration: exp,
		Data:       data,
	}
	cache.Unlock() // <-- UNLOCK
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

	cl.Lock() // <-- LOCK
	cl.delete(key)
	cl.Unlock() // <-- UNLOCK

}

// Deletes all expired entries from cache line. Updates onEvicted using LRU
func (cl *cacheLine) DeleteExpired() {

	cl.Lock() // <-- LOCK
	for key, value := range cl.entries {
		if value.Expired() {
			cl.delete(key)
		}
	}
	cl.Unlock() // <-- UNLOCK
}

// Stops cleaner routine
func stopCleaner(cache *cacheLine) {
	cache.cleanUP.notActive <- true
}

// Runs a cleanup routine
func (cleaner *cleanUP) Run(cL *cacheLine) {
	ticker := time.NewTicker(cleaner.Interval)

	for {
		select {
		case <-ticker.C:
			cL.DeleteExpired()
		case <-cleaner.notActive:
			ticker.Stop()
			return
		}
	}
}

// Starts cleaner routine
func startCleaner(cl *cacheLine, cUP time.Duration) {
	cleaner := &cleanUP{
		Interval:  cUP,
		notActive: make(chan bool),
	}
	cl.cleanUP = cleaner
	go cleaner.Run(cl)
}

// new_Autoclean_Cache()'s subroutine. Creates new default cache line ready to get set to new values.
func new_cacheLine(def_exp time.Duration, new_entry map[string]Entry) *cacheLine {
	if def_exp == DefaultExpiration {
		def_exp = NoExpiration
	}

	cl := &cacheLine{
		entries: new_entry,
		ttl:     def_exp,
	}
	return cl
}

// New()'s subroutine. Creates a new cache with automatic clean up. If clean up interval set to default (0) or
// no-expiration (-1), the autoclean routine doesn't start
func new_Autoclean_Cache(def_exp, cUP time.Duration, new_entry map[string]Entry) *cacheLine {
	cl := new_cacheLine(def_exp, new_entry)
	cache := cl
	if cUP > 0 {
		startCleaner(cl, cUP)
		runtime.SetFinalizer(cache, stopCleaner)
	}
	return cache
}

// Cache constructor
func New(def_exp, cUP time.Duration) *cacheLine {
	entry := make(map[string]Entry)
	return new_Autoclean_Cache(def_exp, cUP, entry)
}
