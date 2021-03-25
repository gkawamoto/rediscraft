package auth

import (
	"sync"

	"github.com/gkawamoto/rediscraft/types"
	"github.com/tidwall/redcon"
)

const name = "AUTH"

func New(authMap *sync.Map, password string) *Command {
	return &Command{authMap, password}
}

type Command struct {
	authMap  *sync.Map
	password string
}

func (c *Command) Execute(conn redcon.Conn, args [][]byte) (interface{}, error) {
	password := string(args[1])
	if password == c.password {
		c.authMap.Store(conn, true)
		return "OK", nil
	}
	return nil, nil
}

func (c *Command) Name() string {
	return name
}

func (c *Command) Hint() interface{} {
	return []interface{}{
		types.BulkString(name),
		1,
		[]interface{}{
			types.FlagWrite,
		},
		0,
		0,
		0,
		[]interface{}{},
	}
}
