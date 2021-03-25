package subscribe

import (
	"github.com/gkawamoto/rediscraft/types"
	"github.com/tidwall/redcon"
)

const name = "SUBSCRIBE"

func New(pubsub *redcon.PubSub) *Command {
	return &Command{pubsub}
}

type Command struct {
	pubsub *redcon.PubSub
}

func (c *Command) Execute(conn redcon.Conn, args [][]byte) (interface{}, error) {
	c.pubsub.Subscribe(conn, "stdout")
	return nil, nil
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
