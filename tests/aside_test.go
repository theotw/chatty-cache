/*
 * Copyright (c) The One True Way 2023. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package tests

import (
	"testing"
)

func TestCacheA(t *testing.T) {
	putCacheName := "cacheA"
	getCacheName := "cacheB"

	runCacheTest(t, putCacheName, getCacheName)
}
