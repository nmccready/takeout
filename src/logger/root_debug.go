package logger

import debug "github.com/nmccready/go-debug"

var rootDebug = debug.Debug("@znemz/takeout")

var Spawn = rootDebug.Spawn
var Log = rootDebug.Log
