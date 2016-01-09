package mumble

import (
	"log"
	"strconv"

	"github.com/layeh/gumble/gumble"
	"github.com/layeh/gumble/gumbleutil"
)

var M = NewMumble()

type Mumble struct {
	Config    *gumble.Config
	Client    *gumble.Client
	Rpc       *Fumble
	KeepAlive chan bool
}

func NewMumble() *Mumble {
	m := new(Mumble)
	m.Rpc = new(Fumble)
	m.KeepAlive = make(chan bool)

	return m
}

func removeTemporaryChannels() {
	for _, c := range Channels {
		if c.Temporary == true && c.IsEmpty() {
			c.Remove()
		}
	}
}

func (m *Mumble) Create() {
	m.Client = gumble.NewClient(m.Config)

	// we'll use this to move users
	// from channels that they're not
	// allowed to get into
	rc := new(gumble.Channel)
	rc.ID = 0 // 0 is the root channel

	e := gumbleutil.Listener{
		// runs when the bot connect
		Connect: onConnect,

		// user events
		UserChange: func(e *gumble.UserChangeEvent) { onUserChange(e, rc) },

		// runs when the bot disconnect
		Disconnect: func(e *gumble.DisconnectEvent) { onDisconnect(e, m) },

		ACL: func(e *gumble.ACLEvent) {
			log.Println("[ACL]: " + strconv.FormatBool(e.ACL.Inherits))
		},

		PermissionDenied: onPermissionDenied,
	}

	// insert events to event listener
	// and attach to gumble client
	el := gumble.EventListener(e)
	m.Client.Attach(el)
}

// connect the bot
func (m *Mumble) Connect() error {
	log.Println("[BOT]: Connecting...")

	err := m.Client.Connect()
	if err != nil {
		return err
	}
	return nil
}

// converts uint32 to string
func UserIdToString(id uint32) string {
	return strconv.FormatUint(uint64(id), 10)
}
