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
	mumble.Connection.Create <- lobbyID
	return nil
}

func (Fumble) EndLobby(lobbyID uint, nop *struct{}) error {
	mumble.Connection.Remove <- lobbyID
	return nil
}

func (Fumble) RemovePlayer(playerID uint, _ *struct{}) error {
	mumble.Connection.RemoveUser <- playerID
	return nil
}

func (Fumble) MoveUser(playerID uint, _ *struct{}) error {
	mumble.Connection.MoveUser <- playerID
	return nil
}

func StartRPC(url, queue string) {
	rpc.Register(new(Fumble))
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Fatal(err)
	}

	serverCodec, err := amqprpc.NewServerCodec(conn, queue, amqprpc.JSONCodec{})
	rpc.ServeCodec(serverCodec)
}
