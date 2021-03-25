package command

import (
	"github.com/gkawamoto/rediscraft/types"
	"github.com/tidwall/redcon"
)

const name = "COMMAND"

func New(serverListFn func() interface{}) *Command {
	return &Command{serverListFn}
}

type Command struct {
	serverListFn func() interface{}
}

func (c *Command) Execute(conn redcon.Conn, args [][]byte) (interface{}, error) {
	return c.serverListFn(), nil
}

func (c *Command) Name() string {
	return name
}

func (c *Command) Hint() interface{} {
	return []interface{}{
		types.BulkString(name),
		-1,
		[]interface{}{},
		0,
		0,
		0,
		[]interface{}{},
	}
}
