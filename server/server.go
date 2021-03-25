package server

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/gkawamoto/rediscraft/config"
	"github.com/gkawamoto/rediscraft/server/command/auth"
	"github.com/gkawamoto/rediscraft/server/command/command"
	"github.com/gkawamoto/rediscraft/server/command/minecraft"
	"github.com/gkawamoto/rediscraft/server/command/ping"
	"github.com/gkawamoto/rediscraft/server/command/subscribe"
	"github.com/gkawamoto/rediscraft/types"
	"github.com/tidwall/redcon"
)

type Command interface {
	Execute(owner redcon.Conn, args [][]byte) (interface{}, error)

	Name() string
	Hint() interface{}
}

type commandRef struct {
	ref       Command
	needsAuth bool
}

type process interface {
	AddCommand(string)
}

func New(p process, conf config.Config) *Server {
	s := &Server{
		authMap:  &sync.Map{},
		commands: map[string]*commandRef{},
		pubsub:   &redcon.PubSub{},
	}

	s.Register(auth.New(s.authMap, conf.RedisPassword), false)
	s.Register(ping.New(), false)
	s.Register(command.New(s.ListCommands), false)
	s.Register(subscribe.New(s.pubsub), false)

	for _, cmd := range minecraft.Commands(p) {
		s.Register(cmd, true)
	}

	return s
}

type Server struct {
	authMap  *sync.Map
	commands map[string]*commandRef
	pubsub   *redcon.PubSub
}

func (s *Server) ListCommands() interface{} {
	var result []interface{}
	for _, cmd := range s.commands {
		result = append(result, cmd.ref.Hint())
	}
	return result
}

func (s *Server) Output(channel string, data []byte) {
	s.pubsub.Publish(channel, string(data))
}

func (s *Server) Register(cmd Command, needsAuth bool) {
	s.commands[strings.ToUpper(cmd.Name())] = &commandRef{cmd, needsAuth}
}

func (s *Server) Handle(conn redcon.Conn, cmd redcon.Command) {
	result, err := s.concreteHandle(conn, cmd)

	s.serializeResult(conn, result, err)
}

func (s *Server) serializeResult(conn redcon.Conn, result interface{}, err error) {
	if err != nil {
		conn.WriteError(err.Error())
		return
	}

	switch data := result.(type) {
	case types.BulkString:
		conn.WriteBulkString(string(data))
	case string:
		conn.WriteString(data)
	case nil:
		conn.WriteNull()
	case []interface{}:
		conn.WriteArray(len(data))
		for _, obj := range data {
			s.serializeResult(conn, obj, err)
		}
	case int:
		conn.WriteInt(data)
	case int64:
		conn.WriteInt64(data)
	case uint64:
		conn.WriteUint64(data)
	default:
		conn.WriteError(fmt.Sprintf("invalid data type %+v", result))
	}
}

func (s *Server) concreteHandle(conn redcon.Conn, cmd redcon.Command) (interface{}, error) {
	commandName := strings.ToUpper(string(cmd.Args[0]))
	commandRef, ok := s.commands[commandName]
	if !ok {
		return nil, fmt.Errorf("invalid command '%s'", commandName)
	}
	if commandRef.needsAuth && !s.isAuthenticated(conn) {
		return nil, fmt.Errorf("not authenticated")
	}

	result, err := commandRef.ref.Execute(conn, cmd.Args)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Server) isAuthenticated(conn redcon.Conn) bool {
	result, _ := s.authMap.LoadOrStore(conn, false)
	return result.(bool)
}

func (s *Server) Connect(conn redcon.Conn) bool {
	// Use this function to accept or deny the connection.
	log.Printf("accept: %s", conn.RemoteAddr())
	return true
}

func (s *Server) Disconnect(conn redcon.Conn, err error) {
	s.authMap.Delete(conn)
}
