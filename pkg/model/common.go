/*
 * Copyright (c) The One True Way 2023. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package model

type CacheRelayMessage struct {
	// CacheName
	CacheName string
	// Cache Key
	CacheKey string
	// Base64 encoded value of the cached jsonifiled bits
	CacheValue string
}
