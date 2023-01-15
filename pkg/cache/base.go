/*
 * Copyright (c) The One True Way 2023. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package cache

// Cache is a simple abstraction of a multi-named space (cacheName) cache that holds key value pairs
type Cache interface {
	// Put  puts an value into the cache, if the type of cache has a size limit, stuff will get tossed out
	Put(cacheName string, cacheKey string, value interface{}) error
	// Get gets a value from the cache
	Get(cacheName string, cacheKey string, valueOut interface{}) error
}
