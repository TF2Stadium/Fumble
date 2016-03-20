package mumble

import (
	"log"

	"github.com/layeh/gumble/gumble"
)

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
			if e.User.Channel.ID == 0 { // is root
				e.User.SetDeafened(true)
				e.User.SetMuted(true)
			}

			if !e.User.IsRegistered() {
				// this shouldn't happen, the mumble authenticator
				// is down, so we'll let users join channel by themselves
				e.User.Send("The mumble authentication service is down, please contact admins, or try reconnecting.")
				e.User.SetDeafened(false)
				e.User.SetMuted(false)
				return
			}

			if allowed, reason := isUserAllowed(e.User, e.User.Channel); !allowed {
				e.User.Send(reason)
				e.User.Move(e.Client.Channels[0])
			} else if !e.User.Channel.IsRoot() {
				e.User.SetDeafened(false)
				e.User.SetMuted(false)
			}
		}
		if e.Type.Has(gumble.UserChangeConnected) {
			if !e.User.IsRegistered() {
				e.User.Send("The mumble authentication service is down, please contact admins, or try reconnecting.")
			}
			e.User.Send("Welcome to TF2Stadium!")
			e.User.SetDeafened(true)
			e.User.SetMuted(true)
		}
	})
}

func (l Conn) OnChannelChange(e *gumble.ChannelChangeEvent) {
	if e.Type.Has(gumble.ChannelChangeCreated) {
		l.wait.Done()
	} else if e.Type.Has(gumble.ChannelChangeRemoved) {
		l.wait.Done()
	}
}

func (l Conn) OnPermissionDenied(e *gumble.PermissionDeniedEvent)       {}
func (l Conn) OnTextMessage(e *gumble.TextMessageEvent)                 {}
func (l Conn) OnUserList(e *gumble.UserListEvent)                       {}
func (l Conn) OnACL(e *gumble.ACLEvent)                                 {}
func (l Conn) OnBanList(e *gumble.BanListEvent)                         {}
func (l Conn) OnContextActionChange(e *gumble.ContextActionChangeEvent) {}
func (l Conn) OnServerConfig(e *gumble.ServerConfigEvent)               {}
