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

func channelManage(conn *Conn) {
	for {
		select {
		case lobbyID := <-conn.Create:
			name := fmt.Sprintf("Lobby #%d", lobbyID)

			conn.wait.Add(1)
			conn.client.Do(func() { conn.client.Channels[0].Add(name, false) })
			conn.wait.Wait()

			conn.client.Do(func() {
				channel := conn.client.Channels[0].Find(name)
				channel.SetDescription("Mumble channel for TF2Stadium " + name)

				conn.wait.Add(2)
				log.Printf("#%d: Creating RED and BLU", lobbyID)
				channel.Add("RED", false)
				channel.Add("BLU", false)
			})
			conn.wait.Wait()
			log.Printf("#%d: Done", lobbyID)
		case lobbyID := <-conn.Remove:
			name := fmt.Sprintf("Lobby #%d", lobbyID)

			conn.client.Do(func() {
				root := conn.client.Channels[0].Find(name) // root lobby channel
				if root == nil {
					log.Printf("Couldn't find channel `%s`", name)
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
			log.Printf("#%d: Deleted channels", lobbyID)
		}
	}
}

func getLobbyID(channel *gumble.Channel) uint {
	name := channel.Name
	if name[0] != 'L' { // channel name is either "RED" or "BLU"
		name = channel.Parent.Name
	}

	id, _ := strconv.ParseUint(name[strings.Index(name, "#")+1:], 10, 32)
	return uint(id)
}

func isUserAllowed(user *gumble.User, channel *gumble.Channel) (bool, string) {
	if channel.IsRoot() {
		return true, ""
	}

	lobbyID := getLobbyID(channel)

	return database.IsAllowed(user.UserID, lobbyID, channel.Name)
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
