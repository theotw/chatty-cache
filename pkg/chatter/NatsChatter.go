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
const MasterPassPhraseEnvVar = "CHATTY_PASSPHRASE"
const MessageReplicationSubject = "chatty.replicate"
const NatsServerURLDefault = "localhost:30221"

type encryptedRelayMessage struct {
	//MessageKey base 64 encoded message key
	MessageKey string `json:"messageKey"`
	//CipherData base64 encoded cipher data of the jsonified model.CacheRelayMessag
	CipherData string `json:"cipherData"`
}
type NatMessagesChatterRelay struct {
	nc               *nats.Conn
	replicateSubject string
	natsURL          string
	objectListener   ObjectListener
	//NodeID random UUID to self reference the node
	nodeID           string
	masterPassPhrase string
}

type protocolVersion int

const noEncryption0 = protocolVersion(0)
const encryption0 = protocolVersion(1)

type replicateCacheMessage struct {
	ProtocolVersion protocolVersion `json:"protocolVersion"`
	MessageData     string          `json:"messageData"`
	NodeID          string          `json:"nodeID"`
}

func NewNatsMessageChatterRelay() (*NatMessagesChatterRelay, error) {
	ret := new(NatMessagesChatterRelay)

	ret.replicateSubject = model.GetEnvVarWithDefault(MessageReplicateChannelEnvVar, MessageReplicationSubject)
	ret.natsURL = model.GetEnvVarWithDefault(NatsURLEnvVar, NatsServerURLDefault)
	ret.masterPassPhrase = model.GetEnvVarWithDefault(MasterPassPhraseEnvVar, "")
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

func (t *NatMessagesChatterRelay) ReplicateCachedObject(message *model.CacheRelayMessage) {
	if len(t.masterPassPhrase) == 0 {
		t.sendReplicateMessageNoEncrypt0(message)
	} else {
		t.sendReplicateMessageEncrypt0(message)
	}
}

func (t *NatMessagesChatterRelay) sendReplicateMessageNoEncrypt0(message *model.CacheRelayMessage) {
	var syncMsg replicateCacheMessage
	syncMsg.NodeID = t.nodeID
	bits, err := json.Marshal(message)
	if err != nil {
		log.WithError(err).Errorf("Unable to encode a cache releay message")
		return
	}

	syncMsg.MessageData = base64.StdEncoding.EncodeToString(bits)
	syncMsg.ProtocolVersion = noEncryption0
	bits, err = json.Marshal(&syncMsg)
	if err != nil {
		log.WithError(err).Errorf("Unable to encode a replication message")
		return
	}

	err = t.nc.Publish(t.replicateSubject, bits)
	if err != nil {
		log.WithError(err).Error("Error publishing cache relay message to nats")
	}
	t.nc.Flush()
}
func (t *NatMessagesChatterRelay) sendReplicateMessageEncrypt0(message *model.CacheRelayMessage) {
	var syncMsg replicateCacheMessage
	syncMsg.NodeID = t.nodeID
	bits, err := json.Marshal(message)
	if err != nil {
		log.WithError(err).Errorf("Unable to encode a cache releay message")
		return
	}

	var cipherMessage encryptedRelayMessage
	masterKey := makeAesKey(t.masterPassPhrase)
	messageKeyPlainText := makeRandom256AesKey()
	messageKeyCipherText, _ := DoAesCBCEncrypt(messageKeyPlainText, masterKey)
	cipherMessage.MessageKey = base64.StdEncoding.EncodeToString(messageKeyCipherText)
	messageCipherText, _ := DoAesCBCEncrypt(bits, messageKeyPlainText)
	cipherMessage.CipherData = base64.StdEncoding.EncodeToString(messageCipherText)
	bits, err = json.Marshal(cipherMessage)
	if err != nil {
		log.WithError(err).Errorf("Unable to encode a cipher cache releay message")
		return
	}

	syncMsg.MessageData = base64.StdEncoding.EncodeToString(bits)
	syncMsg.ProtocolVersion = encryption0
	bits, err = json.Marshal(&syncMsg)
	if err != nil {
		log.WithError(err).Errorf("Unable to encode a replication message")
		return
	}

	err = t.nc.Publish(t.replicateSubject, bits)
	if err != nil {
		log.WithError(err).Error("Error publishing cache relay message to nats")
	}
	t.nc.Flush()
}
func (t *NatMessagesChatterRelay) RegisterListenerForReplicatedObjects(listener ObjectListener) {
	t.objectListener = listener
}
func (t *NatMessagesChatterRelay) init() error {

	nc, err := nats.Connect(t.natsURL)
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
	if x.NodeID == t.nodeID {
		log.Tracef("Recieved Message for my node %s, dropping it", t.nodeID)
		// recieved a message for this node, not point in storing it
		return
	}
	switch x.ProtocolVersion {
	case noEncryption0:
		t.processUnencrypted(&x)
		break
	case encryption0:
		t.processEncrypted0(&x)
		break
	default:
		log.Errorf("Recieved a cache relay message with an unknown protocol version %d", x.ProtocolVersion)
	}
}
func (t *NatMessagesChatterRelay) processUnencrypted(msg *replicateCacheMessage) {
	var relayMsg model.CacheRelayMessage
	bits, err := base64.StdEncoding.DecodeString(msg.MessageData)
	if err != nil {
		log.WithError(err).Errorf("Unable to base 64 decode message data ")
		return
	}
	err = json.Unmarshal(bits, &relayMsg)
	if err != nil {
		log.WithError(err).Errorf("Unable to unmarshal message data ")
		return
	}
	log.Tracef("Recieved Cache Sync %s %s", relayMsg.CacheName, relayMsg.CacheKey)
	if t.objectListener != nil {
		t.objectListener(&relayMsg)
	}
}
func (t *NatMessagesChatterRelay) processEncrypted0(msg *replicateCacheMessage) {
	var cipherMessage encryptedRelayMessage
	cipherMessageBits, cbError := base64.StdEncoding.DecodeString(msg.MessageData)
	if cbError != nil {
		log.WithError(cbError).Errorf("Unable to base 64 decode cipher message data ")
		return
	}
	json.Unmarshal(cipherMessageBits, &cipherMessage)
	masterKey := makeAesKey(t.masterPassPhrase)
	messageKeyCipherBits, cbError2 := base64.StdEncoding.DecodeString(cipherMessage.MessageKey)
	if cbError2 != nil {
		log.WithError(cbError).Errorf("Unable to base 64 decode messageKey")
		return
	}
	messageKey, _ := DoAesCBCDecrypt(messageKeyCipherBits, masterKey)
	msgDataCipherBits, mdError := base64.StdEncoding.DecodeString(cipherMessage.CipherData)
	if mdError != nil {
		log.WithError(cbError).Errorf("Unable to base 64 decode message cipher data")
		return
	}
	plainBits, mdDecrError := DoAesCBCDecrypt(msgDataCipherBits, messageKey)
	if mdDecrError != nil {
		log.WithError(cbError).Errorf("Unable to decypt message cipher data")
		return
	}

	var relayMsg model.CacheRelayMessage
	err := json.Unmarshal(plainBits, &relayMsg)
	if err != nil {
		log.WithError(err).Errorf("Unable to unmarshal message data ")
		return
	}
	log.Tracef("Recieved Cache Sync %s %s", relayMsg.CacheName, relayMsg.CacheKey)
	if t.objectListener != nil {
		t.objectListener(&relayMsg)
	}
}
