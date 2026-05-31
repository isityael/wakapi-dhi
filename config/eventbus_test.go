package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventHubPublishesToExactAndWildcardSubscribers(t *testing.T) {
	eventBus := NewEventHub()

	exact := eventBus.Subscribe(1, EventAliasCreate)
	wildcard := eventBus.Subscribe(1, TopicAlias)

	eventBus.Publish(EventMessage{
		Name:   EventAliasCreate,
		Fields: map[string]interface{}{FieldUserId: "alice"},
	})

	require.Equal(t, "alice", receiveEvent(t, exact).Fields[FieldUserId])
	require.Equal(t, "alice", receiveEvent(t, wildcard).Fields[FieldUserId])
}

func TestEventHubDoesNotPublishUnmatchedEvents(t *testing.T) {
	eventBus := NewEventHub()

	subscription := eventBus.Subscribe(1, TopicAlias)
	eventBus.Publish(EventMessage{Name: EventHeartbeatCreate})

	select {
	case message := <-subscription.Receiver:
		t.Fatalf("received unexpected event %q", message.Name)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestEventHubDoesNotBlockOnFullSubscriber(t *testing.T) {
	eventBus := NewEventHub()
	subscription := eventBus.Subscribe(1, EventAliasCreate)

	eventBus.Publish(EventMessage{Name: EventAliasCreate})
	eventBus.Publish(EventMessage{Name: EventAliasCreate})

	assert.Equal(t, EventAliasCreate, receiveEvent(t, subscription).Name)
}

func receiveEvent(t *testing.T, subscription EventSubscription) EventMessage {
	t.Helper()

	select {
	case message := <-subscription.Receiver:
		return message
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for event")
	}

	return EventMessage{}
}
