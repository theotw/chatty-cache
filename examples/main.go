/*
 * Copyright (c) The One True Way 2023. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/theotw/chatty-cache/pkg/cache"
	"github.com/theotw/chatty-cache/pkg/chatter"
)

func main() {
	relay, err := chatter.NewNatsMessageChatterRelay()
	if err != nil {
		log.WithError(err).Fatalf("Unable to connect to nats")
	}
	memCache := cache.NewInMemCache(2*1024*1024, relay)
	memCache.Put("cachename", "key1", "value1")

	fmt.Println("Hello")
}
