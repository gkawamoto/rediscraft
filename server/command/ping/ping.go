package ping

import (
	"github.com/gkawamoto/rediscraft/types"
	"github.com/tidwall/redcon"
)

const name = "PING"

func New() *Command {
	return nil
}

type Command struct {
}

func (c *Command) Execute(conn redcon.Conn, args [][]byte) (interface{}, error) {
	return "PONG", nil
}

func (c *Command) Name() string {
	return name
}

func (c *Command) Hint() interface{} {
	return []interface{}{
		types.BulkString(name),
		-1,
		[]interface{}{
			"stale",
			"fast",
		},
		0,
		0,
		0,
		[]interface{}{
			"@stale",
			"@fast",
		},
	}
}
