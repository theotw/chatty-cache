/*
 * Copyright (c) The One True Way 2023. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package chatter

import "github.com/theotw/chatty-cache/pkg/model"

type ObjectListener func(message *model.CacheRelayMessage)
type CacheChatter interface {
	ReplicateCachedObject(message *model.CacheRelayMessage)
	RegisterListenerForReplicatedObjects(listener ObjectListener)
}
