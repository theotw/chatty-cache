/*
 * Copyright (c) The One True Way 2023. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package chatter

import (
	"encoding/base64"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	"github.com/theotw/chatty-cache/pkg/model"
)

const MessageReplicateChannelEnvVar = "CHATTY_NATS_SUBJECT"
const NatsURLEnvVar = "NATS_SERVER"
const MessageReplicationSubject = "chatty.replicate"
const NatsServerURLDefault = "localhost:30221"

type NatMessagesChatterRelay struct {
	nc               *nats.Conn
	replicateSubject string
	natsURL          string
	objectListener   ObjectListener
	//nodeID random UUID to self reference the node
	nodeID string
}

type protocolVersion int

const noEncryption0 = protocolVersion(0)
const encryption0 = protocolVersion(1)

type replicateCacheMessage struct {
	protocolVersion protocolVersion
	messageData     string
	nodeID          string
}

func NewNatsMessageChatterRelay() (*NatMessagesChatterRelay, error) {
	ret := new(NatMessagesChatterRelay)

	ret.replicateSubject = model.GetEnvVarWithDefault(MessageReplicateChannelEnvVar, MessageReplicationSubject)
	ret.natsURL = model.GetEnvVarWithDefault(NatsURLEnvVar, NatsServerURLDefault)
	u, uuidErr := uuid.NewUUID()
	if uuidErr != nil {
		log.WithError(uuidErr).Errorf("Unable to generate a node UUID.  Defaulting UUID 42")
		ret.nodeID = "42"
	} else {
		ret.nodeID = u.String()
	}
	err := ret.init()
	return ret, err
}

func (t *NatMessagesChatterRelay) ReplicatedCachedObject(message *model.CacheRelayMessage) {
	var syncMsg replicateCacheMessage
	syncMsg.nodeID = t.nodeID
	bits, err := json.Marshal(message)
	if err != nil {
		log.WithError(err).Errorf("Unable to encode a cache releay message")
		return
	}
	syncMsg.messageData = base64.StdEncoding.EncodeToString(bits)
	syncMsg.protocolVersion = noEncryption0
	bits, err = json.Marshal(&syncMsg)
	if err != nil {
		log.WithError(err).Errorf("Unable to encode a replication message")
		return
	}

	t.nc.Publish(t.replicateSubject, bits)
}
func (t *NatMessagesChatterRelay) RegisterListenerForReplicatedObjects(listener ObjectListener) {
	t.objectListener = listener
}
func (t *NatMessagesChatterRelay) init() error {
	// [begin publish_bytes]
	nc, err := nats.Connect(t.natsURL, nats.Name("chatty cache"))
	if err != nil {
		log.Error(err)
		return err
	}
	t.nc = nc
	t.nc.Subscribe(t.replicateSubject, func(msg *nats.Msg) {
		t.handleCacheSync(msg)
	})

	return nil
}

func (t *NatMessagesChatterRelay) handleCacheSync(msg *nats.Msg) {
	var x replicateCacheMessage
	err := json.Unmarshal(msg.Data, &x)
	if err != nil {
		log.WithError(err).Errorf("Error decoding a cache sync message")
		return
	}
	if x.nodeID == t.nodeID {
		// recieved a message for this node, not point in storing it
		return
	}
	switch x.protocolVersion {
	case noEncryption0:
		t.processUnencrypted(&x)
		break
	case encryption0:
		t.processEncrypted0(&x)
		break
	default:
		log.Errorf("Recieved a cache relay message with an unknown protocol version %d", x.protocolVersion)
	}
}
func (t *NatMessagesChatterRelay) processUnencrypted(msg *replicateCacheMessage) {
	var relayMsg model.CacheRelayMessage
	bits, err := base64.StdEncoding.DecodeString(msg.messageData)
	if err != nil {
		log.WithError(err).Errorf("Unable to base 64 decode message data ")
		return
	}
	err = json.Unmarshal(bits, &relayMsg)
	if err != nil {
		log.WithError(err).Errorf("Unable to unmarshal message data ")
		return
	}
	if t.objectListener != nil {
		t.objectListener(&relayMsg)
	}
}
func (t *NatMessagesChatterRelay) processEncrypted0(msg *replicateCacheMessage) {
	panic("not implemented")
}
