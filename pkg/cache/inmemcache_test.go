/*
 * Copyright (c) The One True Way 2023. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package cache

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type SimpleStruct struct {
	N int
	S string
}

func TestInMemBasic(t *testing.T) {

	cache1 := NewInMemCache(200, nil)
	//these entries have a know size of 8 chars, and this info is used in a test, dont change them
	cache1.Put("space0", "key1", "value0-1")
	cache1.Put("space0", "key2", "value0-2")
	cache1.Put("space0", "key3", "value0-3")

	cache1.Put("space1", "key1", "value1-1")
	cache1.Put("space1", "key2", "value1-2")
	cache1.Put("space1", "key3", "value1-3")

	cache1.Put("space2", "key1", &SimpleStruct{N: 1, S: "one"})
	cache1.Put("space2", "key2", &SimpleStruct{N: 2, S: "two"})
	putError := cache1.Put("space2", "key3", &SimpleStruct{N: 3, S: "three"})
	assert.Nil(t, putError, "Sanity test on put")

	t.Run("Basic Test", func(t *testing.T) {
		BasicTest(t, cache1)
	})

	t.Run("Evict Test", func(t *testing.T) {
		EvictTest(t, cache1)
	})

}

func BasicTest(t *testing.T, cache1 *InMemCache) {
	var val string
	err := cache1.Get("space1", "key2", &val)
	assert.Nil(t, err)
	assert.NotNil(t, val)
	assert.Equal(t, "value1-2", val)

	err = cache1.Get("space1", "notthere", &val)
	assert.NotNil(t, err)
	cacheError := err.(*CacheError)
	assert.Equal(t, NoItem, cacheError.Problem)

	var simpleVal SimpleStruct
	err = cache1.Get("space2", "key2", &simpleVal)
	if assert.Nil(t, err) {
		assert.NotNil(t, val)
		assert.Equal(t, "two", simpleVal.S)
	}
}
func EvictTest(t *testing.T, cache1 *InMemCache) {
	//lets add a value that should eventually evict the 2nd and third entry fo space 0 first two entries

	//first touch the first entry, to make not be least recent
	var val string
	cache1.Get("space0", "key1", &val)
	size := cache1.maxCacheSize - cache1.totalUsedCacheSize + 1
	bits := make([]byte, size)

	cache1.Put("space1", "bigKey", bits)
	err := cache1.Get("space0", "key2", &val)
	assert.NotNil(t, err, "This should be gone")
	err = cache1.Get("space0", "key3", &val)
	assert.NotNil(t, err, "This should be gone")

	err = cache1.Get("space0", "key1", &val)
	assert.Nil(t, err)
	assert.NotNil(t, val)

	bits = make([]byte, 300)
	puterr := cache1.Put("space1", "bigKey", bits)
	if !assert.NotNil(t, puterr, "Bigger than max allowed") {
		cacheError := puterr.(*CacheError)
		assert.Equal(t, ExceedsTotalCacheSize, cacheError.Problem)
	}

}
