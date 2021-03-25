package types

type BulkString string

type Position []int64

type Selectable []string

type Optional struct {
	Value interface{}
}

var FlagWrite = "write"
var FlagReadonly = "readonly"
var FlagDenyoom = "denyoom"
var FlagAdmin = "admin"
var FlagPubSub = "pubsub"
var FlagNoScript = "noscript"
var FlagRandom = "random"
var FlagSortForScript = "sort_for_script"
var FlagLoading = "loading"
var FlagStale = "stale"
var FlagSkipMonitor = "skip_monitor"
var FlagAsking = "asking"
var FlagFast = "fast"
var FlagMovableKeys = "movablekeys"
