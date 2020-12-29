package pio

import (
	"encoding/json"
	"fmt"
)

type EventClient struct {
	HOST      string
	AccessKey string
}

func NewEventClient(HOST string, accessKey string) *EventClient {
	return &EventClient{
		HOST:      HOST,
		AccessKey: accessKey,
	}
}

type EventResponder struct {
	EventID string `json:"eventId"`
}

func (client *EventClient) SentClient(event *Event) (*EventResponder, error) {
	URL := fmt.Sprintf("%s/events.json?accessKey=%s", client.HOST, client.AccessKey)

	JSON, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	resp, err := requestPIO(URL, JSON)
	if err != nil {
		return nil, err
	}

	var responder EventResponder
	err = json.Unmarshal(resp, &responder)
	if err != nil {
		return nil, err
	}

	return &responder, nil
}
