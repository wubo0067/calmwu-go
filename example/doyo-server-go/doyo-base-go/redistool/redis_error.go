package redistool

import (
	"fmt"
	"strings"
)

type ErrorType int

const (
	UnknownType ErrorType = iota
	KeyNotExist
	TimeOut
)

func ErrKeyNotExist(key string) error {
	return fmt.Errorf("key[%s] not exists!", key)
}

func ErrTimeOut(key string) error {
	return fmt.Errorf("key[%s] time out!", key)
}

func ErrUnknown(key string) error {
	return fmt.Errorf("key[%s] unknown error", key)
}

func IsKeyNotExist(err error) bool {
	return err != nil && strings.Contains(err.Error(), "not exists")
}

func IsTimeOut(err error) bool {
	return err != nil && strings.Contains(err.Error(), "time out")
}

func GetErrorType(err error) ErrorType {
	errStr := err.Error()
	if strings.Contains(errStr, "not exists") {
		return KeyNotExist
	}

	if strings.Contains(errStr, "time out") {
		return TimeOut
	}

	return UnknownType
}
