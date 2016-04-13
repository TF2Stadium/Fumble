package mumble

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/TF2Stadium/Helen/models/event"
	"github.com/TF2Stadium/fumble/database"
	"github.com/layeh/gumble/gumble"
	"github.com/streadway/amqp"
)

func moveUserToLobbyChannel(client *gumble.Client, user *gumble.User, lobbyID uint, team string) {
	name := fmt.Sprintf("Lobby #%d", lobbyID)
	channel := client.Channels[0].Find(name)
	if channel == nil {
		log.Printf("Couldn't find channel %s", name)
		return
	}

	teamChannel := channel.Find(strings.ToUpper(team))
	if teamChannel == nil {
		log.Printf("Couldn't find channel %s in %s", team, name)
		return
	}

	user.Move(teamChannel)
}

func (l Conn) OnConnect(e *gumble.ConnectEvent) {
	log.Println("Connected to Mumble!")
}

func (l Conn) OnDisconnect(e *gumble.DisconnectEvent) {
	log.Fatal("Disconnected from Mumble: ", e.String)
}

func (l Conn) OnUserChange(e *gumble.UserChangeEvent) {
	l.client.Do(func() {
		switch {
		case e.Type.Has(gumble.UserChangeChannel):
			if !e.User.IsRegistered() {
				// this shouldn't happen, the mumble authenticator
				// is down, so we'll let users join channel by themselves
				e.User.Send("The mumble authentication service is down, please contact admins, or try reconnecting.")
				return
			}
			if database.IsAdmin(e.User.UserID) {
				return
			}
			if e.User.Channel.IsRoot() {
				lobbyID, team := database.GetCurrentLobby(e.User.UserID)

				if lobbyID != 0 {
					moveUserToLobbyChannel(e.Client, e.User, lobbyID, team)
					return
				}
			}

			if allowed, reason := isUserAllowed(e.User, e.User.Channel); !allowed {
				e.User.Send(reason)
				lobbyID, team := database.GetCurrentLobby(e.User.UserID)

				if lobbyID != 0 {
					moveUserToLobbyChannel(e.Client, e.User, lobbyID, team)
				} else {
					e.User.Move(e.Client.Channels[0])
				}

			} else if !e.User.Channel.IsRoot() &&
				!strings.HasPrefix(e.User.Channel.Name, "Lobby") {
				// user joined the correct team channel

				bytes, _ := json.Marshal(event.Event{
					Name:     event.PlayerMumbleJoined,
					PlayerID: e.User.UserID,
				}) // we don't need to know the lobby id, helen can do that

				amqpChannel.Publish(
					"",
					queueName,
					false,
					false,
					amqp.Publishing{
						ContentType: "application/json",
						Body:        bytes,
					})
			} else {
				//Either the lobby ended, and the player joined
				//the root channel, in which case Helen wouldn't
				//do anything, or the player joined the root
				//channel while the lobby was going on,
				//in which case Helen changes the in-mumble
				//status for the player to false
				bytes, _ := json.Marshal(event.Event{
					Name:     event.PlayerMumbleLeft,
					PlayerID: e.User.UserID,
				})
				amqpChannel.Publish(
					"",
					queueName,
					false,
					false,
					amqp.Publishing{
						ContentType: "application/json",
						Body:        bytes,
					})

			}
		case e.Type.Has(gumble.UserChangeConnected):
			if !e.User.IsRegistered() {
				e.User.Send("The mumble authentication service is down, please contact admins, or try reconnecting.")
			}
			e.User.Send("Welcome to TF2Stadium!")

		case e.Type.Has(gumble.UserChangeDisconnected):
			bytes, _ := json.Marshal(event.Event{
				Name:     event.PlayerMumbleLeft,
				PlayerID: e.User.UserID,
			})

			amqpChannel.Publish(
				"",
				queueName,
				false,
				false,
				amqp.Publishing{
					ContentType: "application/json",
					Body:        bytes,
				})

		}
	})
}

func (l Conn) OnChannelChange(e *gumble.ChannelChangeEvent) {
	if e.Type.Has(gumble.ChannelChangeCreated) && e.Channel.Name[0] == 'L' {
		//channel name is "Lobby #..."
		l.lobbyRootWait.Done()
	}
}

func (l Conn) OnPermissionDenied(e *gumble.PermissionDeniedEvent)       {}
func (l Conn) OnTextMessage(e *gumble.TextMessageEvent)                 {}
func (l Conn) OnUserList(e *gumble.UserListEvent)                       {}
func (l Conn) OnACL(e *gumble.ACLEvent)                                 {}
func (l Conn) OnBanList(e *gumble.BanListEvent)                         {}
func (l Conn) OnContextActionChange(e *gumble.ContextActionChangeEvent) {}
func (l Conn) OnServerConfig(e *gumble.ServerConfigEvent)               {}
