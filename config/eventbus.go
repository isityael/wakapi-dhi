package config

import (
	"strings"
	"sync"
)

type ApplicationEvent struct {
	Type    string
	Payload interface{}
}

const (
	TopicUser                    = "user.*"
	TopicHeartbeat               = "heartbeat.*"
	TopicProjectLabel            = "project_label.*"
	TopicAlias                   = "alias.*"
	EventUserUpdate              = "user.update"
	EventUserDelete              = "user.delete"
	EventHeartbeatCreate         = "heartbeat.create"
	EventProjectLabelCreate      = "project_label.create"
	EventProjectLabelDelete      = "project_label.delete"
	EventAliasCreate             = "alias.create"
	EventAliasDelete             = "alias.delete"
	EventWakatimeFailure         = "wakatime.failure"
	EventLanguageMappingsChanged = "language_mappings.changed"
	EventApiKeyCreate            = "api_key.create"
	EventApiKeyDelete            = "api_key.delete"
	FieldPayload                 = "payload"
	FieldUser                    = "user"
	FieldUserId                  = "user.id"
)

type EventMessage struct {
	Name   string
	Fields map[string]interface{}
}

type EventSubscription struct {
	topics   []string
	Receiver chan EventMessage
}

type EventHub struct {
	lock          sync.RWMutex
	subscriptions []*EventSubscription
}

var eventHub *EventHub

func init() {
	eventHub = NewEventHub()
}

func NewEventHub() *EventHub {
	return &EventHub{}
}

func EventBus() *EventHub {
	return eventHub
}

func SetEventBus(eventBus *EventHub) { // only for testing purposes!
	eventHub = eventBus
}

func (h *EventHub) Subscribe(buffer int, topics ...string) EventSubscription {
	subscription := &EventSubscription{
		topics:   topics,
		Receiver: make(chan EventMessage, buffer),
	}

	h.lock.Lock()
	h.subscriptions = append(h.subscriptions, subscription)
	h.lock.Unlock()

	return *subscription
}

func (h *EventHub) Publish(message EventMessage) {
	h.lock.RLock()
	subscriptions := append([]*EventSubscription(nil), h.subscriptions...)
	h.lock.RUnlock()

	for _, subscription := range subscriptions {
		if !subscription.matches(message.Name) {
			continue
		}

		subscription.deliver(message)
	}
}

func (s *EventSubscription) deliver(message EventMessage) {
	select {
	case s.Receiver <- message:
	default:
		go func() {
			s.Receiver <- message
		}()
	}
}

func (s *EventSubscription) matches(name string) bool {
	for _, topic := range s.topics {
		if topic == name {
			return true
		}
		if strings.HasSuffix(topic, ".*") && strings.HasPrefix(name, strings.TrimSuffix(topic, "*")) {
			return true
		}
	}
	return false
}
