package mumble

import (
	"log"
	"sync"

	"github.com/layeh/gumble/gumble"
)

type Conn struct {
	wait   *sync.WaitGroup
	client *gumble.Client
}

var Connection = &Conn{new(sync.WaitGroup), nil}

func Connect(config *gumble.Config) {
	Connection.wait.Add(1)
	client := gumble.NewClient(config)
	err := client.Connect()
	if err != nil {
		log.Fatal(err)
	}

	Connection.client = client
	client.Attach(Connection)
	Connection.wait.Wait()

}
