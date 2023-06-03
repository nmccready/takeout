package json

import (
	"encoding/json"

	"github.com/nmccready/takeout/src/logger"
)

var debug = logger.Spawn("json")

func Stringify(thing interface{}) string {
	body, _ := json.Marshal(thing)
	return string(body)
}

func StringifyPretty(thing interface{}) string {
	body, _ := json.MarshalIndent(thing, "", " ")
	return string(body)
}

func Parse(body string) interface{} {
	var thing interface{}
	err := json.Unmarshal([]byte(body), &thing)
	if err != nil {
		debug.Log("error parsing json", err)
	}
	return thing
}
