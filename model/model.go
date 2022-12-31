package model

import "encoding/json"

func ToJSON(thing interface{}) string {
	body, _ := json.Marshal(thing)
	return string(body)
}

func ToJSONPretty(thing interface{}) string {
	body, _ := json.MarshalIndent(thing, "", " ")
	return string(body)
}
