package pio

import "time"

const (
	EventTypeSet    = "$set"
	EventTypeUnset  = "$unset"
	EventTypeDelete = "$delete"
)

type Event struct {
	Event            string                 `json:"event"`
	EntityType       string                 `json:"entityType"`
	EntityID         string                 `json:"entityId"`
	TargetEntityType string                 `json:"targetEntityType"`
	TargetEntityID   string                 `json:"targetEntityId"`
	Properties       map[string]interface{} `json:"properties"`
	EventTime        time.Time              `json:"eventTime"`
}

func NewEvent(name string) *Event {
	return &Event{
		Event: name,
	}
}

func (e *Event) SetEvent(name string) *Event {
	e.Event = name
	return e
}

func (e *Event) SetEntityType(entityType string) *Event {
	e.EntityType = entityType
	return e
}

func (e *Event) SetEntityID(entityID string) *Event {
	e.EntityID = entityID
	return e
}

func (e *Event) SetTargetEntityType(targetEntityType string) *Event {
	e.TargetEntityType = targetEntityType
	return e
}

func (e *Event) SetTargetEntityID(targetEntityID string) *Event {
	e.TargetEntityID = targetEntityID
	return e
}

func (e *Event) SetProperties(properties map[string]interface{}) *Event {
	e.Properties = properties
	return e
}

func (e *Event) SetEventTime(eventTime time.Time) *Event {
	e.EventTime = eventTime
	return e
}
