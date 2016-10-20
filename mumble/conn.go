package mumble

import (
	"log"
	"sync"

	"github.com/layeh/gumble/gumble"
	"github.com/streadway/amqp"
)

type Conn struct {
	lobbyRootWait *sync.WaitGroup
	client        *gumble.Client

	Create chan uint
	Remove chan uint

	RemoveUser chan uint
	MoveUser   chan uint
}

var Connection = &Conn{
	lobbyRootWait: new(sync.WaitGroup),
	client:        nil,
	Create:        make(chan uint),
	Remove:        make(chan uint),
	RemoveUser:    make(chan uint),
	MoveUser:      make(chan uint),
}

var (
	amqpChannel *amqp.Channel
	queueName   string
)

func Connect(config *gumble.Config, mumbleAddr, amqpURL, eventQueue string) {
	var err error
	amqpConn, err := amqp.Dial(amqpURL)
	if err != nil {
		log.Fatal(err)
	}

	amqpChannel, err = amqpConn.Channel()
	if err != nil {
		log.Fatal(err)
	}

	_, err = amqpChannel.QueueDeclare(eventQueue, false, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	queueName = eventQueue

	client, err := gumble.Dial(mumbleAddr, config)
	if err != nil {
		log.Fatal(err)
	}

	Connection.client = client
	go channelManage(Connection)
	config.Attach(Connection)
}
