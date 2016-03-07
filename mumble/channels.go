package mumble

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/TF2Stadium/fumble/database"
	"github.com/layeh/gumble/gumble"
)

var ErrChanNotFound = errors.New("channel not found")

func printchannels(c gumble.Channels) {
	for _, channel := range c {
		log.Println(channel.Name)
	}
}

func AddLobbyChannel(l *Conn, lobbyID uint, maxplayers int) {
	name := fmt.Sprintf("Lobby #%d", lobbyID)

	l.wait.Add(1)
	l.client.Do(func() { l.client.Channels[0].Add(name, false) })
	l.wait.Wait()

	l.client.Do(func() {
		channel := l.client.Channels[0].Find(name)
		channel.SetDescription("Mumble channel for TF2Stadium " + name)
		channel.SetMaxUsers(uint32(maxplayers))

		l.wait.Add(2)
		channel.Add("RED", false)
		channel.Add("BLU", false)
	})

	l.wait.Wait()
}

//MoveUsersToLobbyRoot moves all users from the RED and BLU channels of the given lobbyID channel
//to the lobby's root channel
func MoveUsersToLobbyRoot(conn *Conn, lobbyID uint) error {
	var err error

	conn.client.Do(func() {
		name := fmt.Sprintf("Lobby #%d", lobbyID)
		root := conn.client.Channels[0].Find(name) // root lobby channel
		if root == nil {
			err = ErrChanNotFound
			return
		}

		totalUsers := 0
		for _, channel := range root.Children {
			totalUsers += len(channel.Users)

			conn.wait.Add(1)
			channel.Remove()
		}

		if totalUsers == 0 { // no users in both channels, remove it entirely
			conn.wait.Add(1)
			root.Remove()
		}
		return
	})

	conn.wait.Wait()
	return err
}

func getLobbyID(channel *gumble.Channel) uint {
	name := channel.Name
	if name[0] != 'L' { // channel name is either "RED" or "BLU"
		name = channel.Parent.Name
	}

	id, _ := strconv.ParseUint(name[strings.Index(name, "#")+1:], 10, 32)
	return uint(id)
}

func isUserAllowed(user *gumble.User, channel *gumble.Channel) bool {
	if channel.IsRoot() {
		return true
	}

	lobbyID := getLobbyID(channel)

	return database.IsAllowed(user.UserID, lobbyID)
}

func (conn Conn) removeEmptyChannels() {
	conn.client.Do(func() {
		for _, c := range conn.client.Channels {
			if len(c.Users) == 0 && !database.IsLobbyClosed(getLobbyID(c)) {
				conn.wait.Add(1)
				c.Remove()
			}
		}
	})

	conn.wait.Wait()
}
