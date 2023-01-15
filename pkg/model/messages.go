/*
 * Copyright (c) The One True Way 2023. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package model

import "time"

// ExchangeMessage Wrapper for messages to be exchanged between nodes
type ExchangeMessage struct {
	// ExchangeID UUID to identify this message
	ExchangeID string `json:"exchangeID"`

	// CreateTime, time when this exchange message was created
	CreateTime time.Time `json:"createTime"`

	//MessageData base64 encoded message data
	MessageData string `json:"messageData"`
}
