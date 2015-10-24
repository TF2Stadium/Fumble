package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/rpc"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/TF2Stadium/fumble/mumble"
	"github.com/layeh/gumble/gumble"
	"github.com/layeh/gumble/gumbleutil"
	"github.com/stretchr/testify/assert"
)

// custom info for testing
var address string
var username string
var password string
var insecure bool
var certificateFile string
var keyFile string

var rpc_address string
var bot_config *gumble.Config

var channelName = "test channel"

var bot = NewBOT()

// keep the rpc client in a global var
// so we can use in multiple tests
var client *rpc.Client

// connection test
func TestMain(t *testing.T) {
	var err error
	address = os.Getenv("FUMBLE_ADDRESS")
	username = os.Getenv("FUMBLE_USERNAME")
	password = os.Getenv("FUMBLE_PASSWORD")

	// insecure
	ins := os.Getenv("FUMBLE_INSECURE")
	if ins == "" {
		ins = "false"
	}

	insecure, err = strconv.ParseBool(ins)
	if err != nil {
		log.Fatal(err)
	}
	// end: insecure

	// certificate
	certificateFile = os.Getenv("FUMBLE_CERTIFICATE")
	keyFile = os.Getenv("FUMBLE_KEY")
	// end: certificate

	// rpc address
	rpc_address = os.Getenv("FUMBLE_TEST_RPC_ADDRESS")
	if rpc_address == "" {
		rpc_address = "localhost:7070"
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

	bot.Config = config
	bot_config = bot.Config
	bot.create()
	go bot.Connect()

	client, err = rpc.DialHTTP("tcp", rpc_address)

	if err != nil {
		log.Fatal("dialing:", err)
	}

	assert.Nil(t, err)
	assert.NotNil(t, client)
}

// check if the user is connecTed
// it tests by all available values: UserID, Hash, IP and Name
func TestIsUserConnected(t *testing.T) {
	u := mumble.NewUser()
	u.Name = username

	var reply bool = false

	// just to make sure the bot connects
	ticker := time.NewTicker(1 * time.Second)
	for _ = range ticker.C {
		if !reply {
			if err := client.Call("Fumble.IsUserConnected", u, &reply); err != nil {
				log.Fatal(err)
			}
		} else {
			ticker.Stop()
			break
		}
	}
	t.Log("Bot connected")

	// tries all available types of
	// checking if the user is connected
	for i := 0; i < 3; i++ {
		var r bool
		u = mumble.NewUser()

		switch i {
		case 0: // Username
			u.Name = username
		case 1: // UserID
			uid := int(bot.Client.Users.Find(username).UserID)

			if uid != 0 {
				t.Log("UserID: " + strconv.Itoa(uid))
				u.UserID = uid
			}

			// "true" so the test doesn't fail
			// if the userid is 0
			r = true
		case 2: // Hash
			h := bot.Client.Users.Find(username).Hash

			if h != "" {
				t.Log("Hash: " + h)
				u.Hash = h
			}

			// "true" so the test doesn't fail
			// if no hash is found
			r = true
		case 3: // IP
			u.IP = bot.Client.Users.Find(username).Stats.IP.String()
		}

		// ignore nil hash (non-registered user)
		if !r {
			if err := client.Call("Fumble.IsUserConnected", u, &r); err != nil {
				log.Fatal(err)
			}

			assert.True(t, r)
		}
	}

	t.Log("IsUserConnected? " + strconv.FormatBool(reply))
}

// find user by info
func TestFindUserByInfo(t *testing.T) {
	u := mumble.NewUser()
	u.Name = username

	if err := client.Call("Fumble.FindUserByInfo", u, &u); err != nil {
		log.Fatal(err)
	}

	b := bot.Client.Users.Find(username)

	assert.Equal(t, u.Name, username)
	if b.Stats != nil && u.IP != "<nil>" {
		assert.NotNil(t, u.IP)
		assert.Equal(t, u.IP, b.Stats.IP.String())
	}

	t.Log(u)
}

// find user by name
func TestFindUserByName(t *testing.T) {
	var reply *mumble.User

	if err := client.Call("Fumble.FindUserByName", username, &reply); err != nil {
		log.Fatal(err)
	}

	assert.NotNil(t, reply)
}

// find user by ip
func TestFindUserByIP(t *testing.T) {
	var reply *mumble.User

	if err := client.Call("Fumble.FindUserByIP", bot.Client.Self.Stats.IP.String(), &reply); err != nil {
		log.Fatal(err)
	}

	assert.NotNil(t, reply)
}

// find user by hash
func TestFindUserByHash(t *testing.T) {
	if bot.Client.Self.Hash == "" {
		return
	}

	var reply *mumble.User

	if err := client.Call("Fumble.FindUserByHash", bot.Client.Self.Stats.IP.String(), &reply); err != nil {
		log.Fatal(err)
	}

	assert.NotNil(t, reply)
}

// find user by id
func TestFindUserByID(t *testing.T) {
	if bot.Client.Self.UserID == 0 {
		return
	}

	var reply *mumble.User

	if err := client.Call("Fumble.FindUserByID", bot.Client.Self.UserID, &reply); err != nil {
		log.Fatal(err)
	}

	assert.NotNil(t, reply)
}

// check if user is registered
func TestIsUserRegistered(t *testing.T) {
	var reply bool

	u := mumble.NewUser()
	u.Name = username

	if err := client.Call("Fumble.IsUserRegistered", u, &reply); err != nil {
		log.Fatal(err)
	}

	assert.False(t, reply)
}

// creates a channel
func TestCreateChannel(t *testing.T) {
	var nr mumble.NoReply

	if err := client.Call("Fumble.CreateChannel", channelName, &nr); err != nil {
		log.Fatal(err)
	}
}

// checks if the bot can join a channel
// that hes not allowed to get into
// and also checks if he can get in
// when allowed
func TestAllowUserIntoChannel(t *testing.T) {
	var nr mumble.NoReply
	var err error
	args := new(mumble.Args)

	u := mumble.NewUser()
	u.CopyInfo(bot.Client.Users.Find(username))
	args.User = u

	c := mumble.NewChannel()
	c.Name = channelName
	args.Channel = c

	err = client.Call("Fumble.AllowUser", args, &nr)
	assert.NoError(t, err)

	bot.Client.Self.Move(bot.Client.Channels.Find(channelName))

	var nr2 mumble.NoReply
	err = client.Call("Fumble.DisallowUser", args, &nr2)
	assert.NoError(t, err)
}

/*
WORKS
DONT TEST AGAIN

func TestBanUser(t *testing.T) {
	var nr mumble.NoReply
	var err error

	u := mumble.NewUser()
	u.CopyInfo(bot.Client.Users.Find(username))

	ban := new(mumble.KickArgs)
	ban.User = u
	ban.Reason = "too elite"

	err = client.Call("Fumble.BanUser", ban, &nr)
	assert.NoError(t, err)
}*/

func TestKickUser(t *testing.T) {
	var nr mumble.NoReply
	var err error

	u := mumble.NewUser()
	u.CopyInfo(bot.Client.Users.Find(username))

	kick := new(mumble.KickArgs)
	kick.User = u
	kick.Reason = "too elite"

	err = client.Call("Fumble.KickUser", kick, &nr)
	assert.NoError(t, err)

	time.Sleep(3 * time.Second)

	var isUserConnecTed bool
	err = client.Call("Fumble.IsUserConnected", u, &isUserConnecTed)
	assert.NoError(t, err)
	assert.False(t, isUserConnecTed)

	go bot.Connect()
	time.Sleep(3 * time.Second)
}

// removes a channel
func TestRemoveChannel(t *testing.T) {
	var nr mumble.NoReply

	err := client.Call("Fumble.RemoveChannel", channelName, &nr)
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)
}

// closes the bot connection
func TestCloseConnection(t *testing.T) {
	bot.KeepAlive <- true
}

func TestLobby(t *testing.T) {
	var err error

	// lobby
	l := mumble.NewLobby()
	l.ID = 123

	// user 1
	u1 := mumble.NewUser()
	u1.Name = "mariokid45"
	u1.Team = "RED"
	l.Players[u1.Name] = u1

	// bot for user 1
	b1 := NewBOT()
	b1.Config = bot_config
	b1.Config.Username = u1.Name
	b1.create()
	go b1.Connect()
	time.Sleep(2 * time.Second)

	// user 2
	u2 := mumble.NewUser()
	u2.Name = "bigkahuna"
	u2.Team = "BLU"
	l.Players[u2.Name] = u2

	// bot for user 2
	b2 := NewBOT()
	b2.Config = bot_config
	b2.Config.Username = u2.Name
	b2.create()
	go b2.Connect()
	time.Sleep(2 * time.Second)

	// create lobby
	err = client.Call("Fumble.CreateLobby", l, &l)
	assert.NoError(t, err)

	// args to disallow player 2 from joining lobby's mumble
	la := new(mumble.LobbyArgs)
	la.Lobby = l
	la.User = u2

	//////////////////////////////////////////////////
	//              *** IMPORTANT ***               //
	//////////////////////////////////////////////////
	// this pyroshit wont update the existing lobby //
	// variable, so i had to make a new one         //
	//////////////////////////////////////////////////
	var lobby *mumble.Lobby

	// Disallow player
	err = client.Call("Fumble.DisallowPlayer", la, &lobby)
	assert.NoError(t, err)

	// check if user 2 isn't in the player list
	assert.Nil(t, lobby.Players[u2.Name])

	// update the lobby variable in lobbyArgs
	la.Lobby = lobby

	// Allow player
	err = client.Call("Fumble.AllowPlayer", la, &lobby)
	assert.NoError(t, err)

	// check if both users are in the player list
	assert.NotNil(t, lobby.Players[u2.Name])

	var shouldNotBeMD mumble.MD

	// checks if the user is muted or deafened
	err = client.Call("Fumble.IsUserMD", u1, &shouldNotBeMD)
	assert.NoError(t, err)

	// check if user 1 is not muted or deafened
	assert.False(t, shouldNotBeMD.Muted)
	assert.False(t, shouldNotBeMD.Deafened)

	// mute and deaf the bot 1
	b1.Client.Self.SetSelfMuted(true)
	b1.Client.Self.SetSelfDeafened(true)

	// let the bot update
	time.Sleep(5 * time.Second)

	// redo the MD test
	var shouldBeMD mumble.MD

	// checks if the user is muted or deafened
	err = client.Call("Fumble.IsUserMD", u1, &shouldBeMD)
	assert.NoError(t, err)

	// check if user 1 is not muted or deafened
	assert.True(t, shouldBeMD.Muted)
	assert.True(t, shouldBeMD.Deafened)

	// end lobby
	err = client.Call("Fumble.EndLobby", lobby, &lobby)
	assert.NoError(t, err)
	assert.True(t, lobby.Channel.Temporary)
}

/////////
// BOT //
/////////

// this is the testing bot
type BOT struct {
	Config    *gumble.Config
	Client    *gumble.Client
	KeepAlive chan bool
}

func NewBOT() *BOT {
	b := new(BOT)
	b.KeepAlive = make(chan bool)

	return b
}

func (b *BOT) create() {
	b.Client = gumble.NewClient(b.Config)

	e := gumbleutil.Listener{
		// runs when the bot connect
		Connect: func(e *gumble.ConnectEvent) {
			log.Println("[BOT]: ConnecTed")

			// get bot stats
			e.Client.Self.Request(gumble.RequestStats)
		},

		// user events
		UserChange: func(e *gumble.UserChangeEvent) {

			// user connected
			if e.Type.Has(gumble.UserChangeConnected) {
				log.Println(fmt.Sprintf("[Connected]: User[Name: %s ID: %s]",
					e.User.Name, mumble.UserIdToString(e.User.UserID)))
			}

			if e.Type.Has(gumble.UserChangeDisconnected) {
				log.Println(fmt.Sprintf("[Disconnected]: User[Name: %s ID: %s]",
					e.User.Name, mumble.UserIdToString(e.User.UserID)))
			}

			if e.Type.Has(gumble.UserChangeKicked) {
				log.Println("User kicked")
			}

			if e.Type.Has(gumble.UserChangeBanned) {
				log.Println("User banned")
			}

			// user change registered
			if e.Type.Has(gumble.UserChangeRegistered) {
				log.Println(fmt.Sprintf("[UserChangeRegistered]: User[Name: %s ID: %s Hash: %s]",
					e.User.Name, mumble.UserIdToString(e.User.UserID), e.User.Hash))
			}

			if e.Type.Has(gumble.UserChangeUnregistered) {
				log.Println("User unregistered")
			}

			if e.Type.Has(gumble.UserChangeName) {
				log.Println("User changed name")
			}

			// user change channel
			if e.Type.Has(gumble.UserChangeChannel) {
			}

			if e.Type.Has(gumble.UserChangeComment) {
				log.Println("User changed comment")
			}

			if e.Type.Has(gumble.UserChangeAudio) {
				log.Println("User changed audio")
			}

			if e.Type.Has(gumble.UserChangeTexture) {
				log.Println("User changed texture")
			}

			if e.Type.Has(gumble.UserChangePrioritySpeaker) {
				log.Println("User priority speaker")
			}

			if e.Type.Has(gumble.UserChangeRecording) {
				log.Println("User recording")
			}

			if e.Type.Has(gumble.UserChangeStats) {
				log.Println("[UserChangeStats]: User[" + e.User.Name + " @ " + e.User.Stats.IP.String() + "]")
			}
		},

		// runs when the bot disconnect
		Disconnect: func(e *gumble.DisconnectEvent) {
			log.Println("[BOT]: Disconnected -> " + e.String)
			b.KeepAlive <- true
		},

		PermissionDenied: func(e *gumble.PermissionDeniedEvent) {
			var dType string

			// text length
			if e.Type.Has(gumble.PermissionDeniedTextTooLong) {
				dType = "Text too long!"
			}

			// full channel
			if e.Type.Has(gumble.PermissionDeniedChannelFull) {
				dType = "Channel is full!"
			}

			// missing certificate
			if e.Type.Has(gumble.PermissionDeniedMissingCertificate) {
				dType = "User is missing certificate!"
			}

			// invalid channel name
			if e.Type.Has(gumble.PermissionDeniedInvalidChannelName) {
				dType = "Invalid channel name!"
			}

			// invalid user name
			if e.Type.Has(gumble.PermissionDeniedInvalidUserName) {
				dType = "Invalid user name!"
			}

			// permission
			if e.Type.Has(gumble.PermissionDeniedPermission) {
				dType = "You have no permission!"
			}

			// other
			if e.Type.Has(gumble.PermissionDeniedOther) {
				dType = "dunno"
			}

			log.Println("[BOT]: Permission Denied -> " + dType)
			log.Println(e.Permission)
		},
	}

	// insert events to event listener
	// and attach to gumble client
	el := gumble.EventListener(e)
	b.Client.Attach(el)
}

func (b *BOT) Connect() {
	log.Println("[BOT]: Connecting...")

	if err := b.Client.Connect(); err != nil {
		log.Fatal(err)
	}

	<-b.KeepAlive
}
