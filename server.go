package main

import (
	"fmt"
	"log"
	"net"
	"strings"
)

type server struct {
	rooms       map[string]*room
	commandChan chan command
}

func newServer() *server {
	return &server{
		rooms:       make(map[string]*room),
		commandChan: make(chan command),
	}
}

func (s *server) newClient(conn net.Conn) *client {
	log.Printf("new client connected : %s", conn.RemoteAddr().String())
	return &client{
		conn:        conn,
		name:        "anonymous",
		commandChan: s.commandChan,
	}
}

func (s *server) run() {
	for cmd := range s.commandChan {
		switch cmd.id {
		case CMD_NICK:
			s.name(cmd.client, cmd.args)
		case CMD_JOIN:
			s.join(cmd.client, cmd.args)
		case CMD_ROOMS:
			s.currentRooms(cmd.client)
		case CMD_MSG:
			s.msg(cmd.client, cmd.args)
		case CMD_QUIT:
			s.quitRoom(cmd.client, cmd.args)
		}
	}

}

func (s *server) name(c *client, args []string) {
	name := args[1]
	c.name = name
	c.msg(fmt.Sprintf("Successfully set name, hey %s", name))
}

func (s *server) join(c *client, args []string) {
	roomName := args[1]

	r, ok := s.rooms[roomName]
	if !ok {
		log.Println("Creating new room", roomName)
		r = &room{
			name:    roomName,
			members: make(map[net.Addr]*client),
		}
		s.rooms[roomName] = r
	}

	r.members[c.conn.RemoteAddr()] = c

	s.quitCurrentRoom(c)

	c.room = r

	r.broadcast(c, fmt.Sprintf("%s has joined the room", c.name))
	c.msg(fmt.Sprintf("Welcome to %s ", r.name))
}

func (s *server) currentRooms(c *client) {
	var rooms []string
	for room, _ := range s.rooms {
		rooms = append(rooms, room)
	}
	res := strings.Join(rooms, " ")
	c.msg(fmt.Sprintf("Available rooms - %s", res))
}

func (s *server) msg(c *client, args []string) {
	msg := strings.Join(args[1:], " ")
	if len(strings.TrimSpace(msg)) == 0 {
		c.msg("Cant send empty message!")
		return
	}
	currRoom := c.room
	currRoom.broadcast(c, fmt.Sprintf("%s: "+strings.Join(args[1:], " "), c.name))
}

func (s *server) quitRoom(c *client, args []string) {
	c.msg("Bye, hope to see you soon.")
	s.quitCurrentRoom(c)

	// Closes the connection completely. Have another option to quit just the room.
	c.conn.Close()
}

func (s *server) quitCurrentRoom(c *client) {
	if c.room != nil {
		delete(c.room.members, c.conn.RemoteAddr())
		c.room.broadcast(c, fmt.Sprintf("%s left the room", c.name))
	}
}
