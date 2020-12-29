package pio

import (
	"encoding/json"
	"fmt"
)

type EngineClient struct {
	HOST      string
	AccessKey string
}

func NewEngineClient(HOST string) *EngineClient {
	return &EngineClient{
		HOST: HOST,
	}
}

func (client *EngineClient) Query(query interface{}) ([]byte, error) {
	URL := fmt.Sprintf("%s/query.json", client.HOST)

	JSON, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	resp, err := requestPIO(URL, JSON)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
