package main

import (
	"log"

	"net/http"
	_ "net/http/pprof"

	"github.com/TF2Stadium/fumble/database"
	"github.com/TF2Stadium/fumble/mumble"
	"github.com/TF2Stadium/fumble/rpc"
	"github.com/kelseyhightower/envconfig"
	"github.com/layeh/gumble/gumble"
)

func main() {
	var config struct {
		MumbleAddress  string `envconfig:"MUMBLE_ADDR" default:"127.0.0.1:64738"`
		MumbleUsername string `envconfig:"MUMBLE_USERNAME" default:"SuperUser"`
		MumblePassword string `envconfig:"MUMBLE_PASSWORD" required:"true"`

		DBAddr     string `envconfig:"DATABASE_ADDR" default:"127.0.0.1:5432"`
		DBName     string `envconfig:"DATABASE_NAME" default:"tf2stadium"`
		DBUsername string `envconfig:"DATABASE_USERNAME" default:"tf2stadium"`
		DBPassword string `envconfig:"DATABASE_PASSWORD" default:"dickbutt"`

		RabbitMQURL        string `envconfig:"RABBITMQ_URL" default:"amqp://guest:guest@localhost:5672/"`
		RabbitMQEventQueue string `envconfig:"RABBITMQ_EVENT_QUEUE" default:"events"`
		RabbitMQRPCQueue   string `envconfig:"RABBITMQ_RPC_QUEUE" default:"fumble"`

		ProfilerAddr string `envconfig:"PROFILER_ADDR"`
	}

	log.SetFlags(log.Lshortfile)
	err := envconfig.Process("FUMBLE", &config)
	if err != nil {
		log.Fatal(err)
	}

	database.Connect(config.DBAddr, config.DBName, config.DBUsername, config.DBPassword)

	mumbleConf := gumble.NewConfig()
	mumbleConf.Username = config.MumbleUsername
	mumbleConf.Password = config.MumblePassword

	if config.ProfilerAddr != "" {
		go func() {
			log.Println(http.ListenAndServe(config.ProfilerAddr, nil))
		}()
		log.Printf("Running profiler at %s", config.ProfilerAddr)
	}

	mumble.Connect(mumbleConf, config.MumbleAddress, config.RabbitMQURL, config.RabbitMQEventQueue)
	rpc.StartRPC(config.RabbitMQURL, config.RabbitMQRPCQueue)
}
