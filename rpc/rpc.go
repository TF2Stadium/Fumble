package rpc

import (
	"log"
	"net/rpc"

	"github.com/TF2Stadium/fumble/mumble"
	"github.com/streadway/amqp"
	"github.com/vibhavp/amqp-rpc"
)

type Fumble struct{}

func (Fumble) CreateLobby(lobbyID uint, nop *struct{}) error {
	mumble.AddLobbyChannel(mumble.Connection, lobbyID, 18)
	return nil
}

func (Fumble) EndLobby(lobbyID uint, nop *struct{}) error {
	return mumble.MoveUsersToRoot(mumble.Connection, lobbyID)
}

func StartRPC(url, event string) {
	rpc.Register(new(Fumble))
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Fatal(err)
	}

	serverCodec, err := amqprpc.NewServerCodec(conn, event, amqprpc.JSONCodec{})
	rpc.ServeCodec(serverCodec)
}
