package pio

import (
	"bytes"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
)

func requestPIO(URL string, JSON []byte) ([]byte, error) {
	req, err := http.NewRequest("POST", URL, bytes.NewBuffer(JSON))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 400 {
		return nil, errors.New("Invalid access key")

	} else if resp.StatusCode == 401 {
		return nil, errors.New("Invalid format, content error")

	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
