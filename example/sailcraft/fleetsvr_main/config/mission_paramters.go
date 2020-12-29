package config

import (
	"sailcraft/base"
)

type MissionParameters map[string]interface{}

func (this *MissionParameters) ContainKey(name string) bool {
	if _, ok := (*this)[name]; ok {
		return true
	}

	return false
}

func (this *MissionParameters) Int64(name string, defaultValue int64) int64 {
	if v, ok := (*this)[name]; ok {
		return base.ConvertToInt64(v, defaultValue)
	}

	return defaultValue
}

func (this *MissionParameters) Int(name string, defaultValue int) int {
	if v, ok := (*this)[name]; ok {
		return int(base.ConvertToInt64(v, int64(defaultValue)))
	}

	return defaultValue
}

func (this *MissionParameters) Float64(name string, defaultValue float64) float64 {
	if v, ok := (*this)[name]; ok {
		return base.ConvertToFloat64(v, defaultValue)
	}

	return defaultValue
}

func (this *MissionParameters) String(name string, defaultValue string) string {
	if v, ok := (*this)[name]; ok {
		return base.ConvertToString(v, defaultValue)
	}

	return defaultValue
}
