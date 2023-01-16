/*
 * Copyright (c) The One True Way 2023. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package tests

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/theotw/chatty-cache/pkg/cache"
	"github.com/theotw/chatty-cache/pkg/chatter"
	"os"
	"testing"
	"time"
)

func runCacheTest(t *testing.T, putCacheName string, getCacheName string) {
	os.Setenv("CHATTY_PASSPHRASE", "bob")
	log.SetLevel(log.TraceLevel)
	relay, err := chatter.NewNatsMessageChatterRelay()
	if err != nil {
		log.WithError(err).Fatalf("Unable to connect to nats")
	}

	memCache := cache.NewInMemCache(2*1024*1024, relay)
	time.Sleep(20 * time.Second)
	limit := 5
	for i := 0; i < limit; i++ {
		key := fmt.Sprintf("key%d", i)
		val := fmt.Sprintf("value-%s-%d", putCacheName, i)
		memCache.Put(putCacheName, key, &val)
	}
	time.Sleep(2 * time.Second)
	start := time.Now()
	for i := 0; i < limit; i++ {
		for {
			key := fmt.Sprintf("key%d", i)
			expected := fmt.Sprintf("value-%s-%d", getCacheName, i)
			var val string
			err := memCache.Get(getCacheName, key, &val)
			if err != nil {
				time.Sleep(2 * time.Second)
			} else {
				if expected != val {
					t.Errorf("Expected %s but got %s", expected, val)
				}
				break
			}
			now := time.Now()
			if now.Sub(start) > 2*time.Minute {
				t.Fatal("timeout waiting for cache")
			}
		}

	}
}
