package server

import (
	"log"
	"net"
	"net/http"
	"net/rpc"

	"github.com/TF2Stadium/fumble/mumble"
)

func Start(m *mumble.Mumble, port string) {
	rpc.Register(m.Rpc)
	rpc.HandleHTTP()

	l, e := net.Listen("tcp", ":"+port)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	http.Serve(l, nil)

	log.Println("[RPC]: Listening to port: " + port)
}
