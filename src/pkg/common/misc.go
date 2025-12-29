package common

import "encoding/json"

func ToJSONString(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func ToJSONPretty(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(b)
}
