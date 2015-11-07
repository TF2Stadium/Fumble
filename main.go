package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/TF2Stadium/fumble/mumble"
	"github.com/TF2Stadium/fumble/server"
	"github.com/layeh/gumble/gumble"
)

func main() {
	address := os.Getenv("FUMBLE_ADDRESS")
	username := os.Getenv("FUMBLE_USERNAME")
	password := os.Getenv("FUMBLE_PASSWORD")

	// insecure
	ins := os.Getenv("FUMBLE_INSECURE")
	if ins == "" {
		ins = "true"
	}

	insecure, err := strconv.ParseBool(ins)
	if err != nil {
		log.Fatal(err)
	}
	// end: insecure

	// certificate
	certificateFile := os.Getenv("FUMBLE_CERTIFICATE")
	keyFile := os.Getenv("FUMBLE_KEY")
	// end: certificate

	// rpc address
	rpc_port := os.Getenv("FUMBLE_RPC_PORT")
	if rpc_port == "" {
		rpc_port = "7070"
	}
	// end: rpc address

	config := gumble.NewConfig()
	config.Address = address
	config.Username = username
	config.Password = password

	if insecure {
		config.TLSConfig.InsecureSkipVerify = true
	}

	if certificateFile != "" {
		if keyFile == "" {
			keyFile = certificateFile
		}
		if certificate, err := tls.LoadX509KeyPair(certificateFile, keyFile); err != nil {
			fmt.Printf("%s: %s\n", os.Args[0], err)
			os.Exit(1)
		} else {
			config.TLSConfig.Certificates = append(config.TLSConfig.Certificates, certificate)
		}
	}

	m := mumble.M
	m.Config = config
	m.Create()

	err = m.Connect()
	if err != nil {
		log.Fatal(err)
	}

	// rpc server
	server.Start(m, rpc_port)

	<-m.KeepAlive
}
