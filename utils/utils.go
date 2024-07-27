package utils

import "encoding/json"

func MapToJsonString(v interface{}) string {
	if v == nil {
		return ""
	}
	bytes, err := json.Marshal(&v)
	if err != nil {
		return err.Error()
	}
	return string(bytes)
}
