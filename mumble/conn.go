package mumble

import (
	"log"
	"sync"

	"github.com/layeh/gumble/gumble"
)

type Conn struct {
	wait   *sync.WaitGroup
	client *gumble.Client

	Create chan uint
	Remove chan uint
}

var Connection = &Conn{
	wait:   new(sync.WaitGroup),
	client: nil,
	Create: make(chan uint),
	Remove: make(chan uint),
}

func Connect(config *gumble.Config) {
	Connection.wait.Add(1)
	client := gumble.NewClient(config)
	err := client.Connect()
	if err != nil {
		log.Fatal(err)
	}

	Connection.client = client
	go channelManage(Connection)
	client.Attach(Connection)
	Connection.wait.Wait()

}
