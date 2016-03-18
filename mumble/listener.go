package mumble

import (
	"log"
	"sync/atomic"

	"github.com/TF2Stadium/fumble/database"
	"github.com/layeh/gumble/gumble"
)

var channels = new(uint64)

func (l Conn) OnConnect(e *gumble.ConnectEvent) {
	log.Println("Connected to Mumble!")
	l.wait.Done()
}

func (l Conn) OnDisconnect(e *gumble.DisconnectEvent) {
	log.Fatal("Disconnected from Mumble: ", e.String)
}

func (l Conn) OnUserChange(e *gumble.UserChangeEvent) {
	l.client.Do(func() {
		if e.Type.Has(gumble.UserChangeChannel) {
			if allowed, reason := isUserAllowed(e.User, e.User.Channel); !allowed {
				e.User.Send(reason)
				e.User.Move(e.Client.Channels[0])
			}

			if atomic.LoadUint64(channels) == 30 {
				go l.removeEmptyChannels()
			}
		}
		if e.Type.Has(gumble.UserChangeConnected) {
			steamid := database.GetSteamID(e.User.UserID)
			e.User.SetComment("http://steamcommunity.com/profiles/" + steamid)
			e.User.Send("Welcome to TF2Stadium!")
		}
	})
}

func (l Conn) OnChannelChange(e *gumble.ChannelChangeEvent) {
	if e.Type.Has(gumble.ChannelChangeCreated) {
		l.wait.Done()
		atomic.AddUint64(channels, 1)
	} else if e.Type.Has(gumble.ChannelChangeRemoved) {
		l.wait.Done()
		atomic.AddUint64(channels, ^uint64(0))
	}
}

func (l Conn) OnPermissionDenied(e *gumble.PermissionDeniedEvent)       {}
func (l Conn) OnTextMessage(e *gumble.TextMessageEvent)                 {}
func (l Conn) OnUserList(e *gumble.UserListEvent)                       {}
func (l Conn) OnACL(e *gumble.ACLEvent)                                 {}
func (l Conn) OnBanList(e *gumble.BanListEvent)                         {}
func (l Conn) OnContextActionChange(e *gumble.ContextActionChangeEvent) {}
func (l Conn) OnServerConfig(e *gumble.ServerConfigEvent)               {}
