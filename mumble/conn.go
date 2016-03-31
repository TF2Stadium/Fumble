package mumble

import (
	"log"
	"sync"

	"github.com/layeh/gumble/gumble"
)

type Conn struct {
	lobbyRootWait *sync.WaitGroup
	client        *gumble.Client

	Create chan uint
	Remove chan uint
}

var Connection = &Conn{
	lobbyRootWait: new(sync.WaitGroup),
	client:        nil,
	Create:        make(chan uint),
	Remove:        make(chan uint),
}

func Connect(config *gumble.Config) {
	client := gumble.NewClient(config)
	err := client.Connect()
	if err != nil {
		log.Fatal(err)
	}

	Connection.client = client
	go channelManage(Connection)
	client.Attach(Connection)
}
