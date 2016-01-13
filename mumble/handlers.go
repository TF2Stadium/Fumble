package mumble

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/layeh/gumble/gumble"
)

//called when fumble successfully connects
func onConnect(e *gumble.ConnectEvent) {
	log.Println("[BOT]: Connected")

	e.Client.Self.RequestStats()
	// copy channels
	for _, gc := range e.Client.Channels {
		if gc.IsRoot() {
			c := NewChannel()
			c.Name = gc.Name
			c.Temporary = gc.Temporary

			Channels[gc.Name] = c
		}
	}

	// get all users stats on connect
	for _, v := range e.Client.Users {
		log.Printf("Requesting stats for User[Name: %s ID: %s Registered: %s Hash: %s]",
			v.Name, UserIdToString(v.UserID), strconv.FormatBool(v.IsRegistered()), v.Hash)

		// request user stats
		//
		// stats will be available
		// after event: "UserChangeStats"
		v.RequestStats()
	}

	// register bot
	if !e.Client.Self.IsRegistered() {
		e.Client.Self.Register()
	}
}

func onUserChange(e *gumble.UserChangeEvent, rc *gumble.Channel) {
	// user connected
	if e.Type.Has(gumble.UserChangeConnected) {
		log.Println(fmt.Sprintf("[Connected]: User[Name: %s ID: %s]",
			e.User.Name, UserIdToString(e.User.UserID)))

		// "e.User.Stats" is always nil when a user connects
		// so we'll get the user's stats
		// "e.User.Stats" will be available
		// after event: "gumble.UserChangeStats"
		e.User.RequestStats()

		removeTemporaryChannels()
	}

	if e.Type.Has(gumble.UserChangeDisconnected) {
		// log.Printf("[UserChangeDisconnected]: User[Name: %s ID: %s]", e.User.Name,
		// 	UserIdToString(e.User.UserID))

		removeTemporaryChannels()
	}

	if e.Type.Has(gumble.UserChangeKicked) {
		log.Printf("[UserChangeKicked]: User[Name: %s ID: %s Hash: %s]",
			e.User.Name, UserIdToString(e.User.UserID), e.User.Hash)
	}

	if e.Type.Has(gumble.UserChangeBanned) {
		log.Printf("[UserChangeBanned]: User[Name: %s ID: %s Hash: %s]",
			e.User.Name, UserIdToString(e.User.UserID), e.User.Hash)
	}

	// user change registered
	if e.Type.Has(gumble.UserChangeRegistered) {
		log.Printf("[UserChangeRegistered]: User[Name: %s ID: %s Hash: %s]",
			e.User.Name, UserIdToString(e.User.UserID), e.User.Hash)

	}

	if e.Type.Has(gumble.UserChangeUnregistered) {
		log.Printf("[UserChangeUnregistered]: User[Name: %s ID: %s Hash: %s]",
			e.User.Name, UserIdToString(e.User.UserID), e.User.Hash)
	}

	if e.Type.Has(gumble.UserChangeName) {
		log.Println("[UserChangeName]: User changed name to " + e.User.Name)
	}

	// user change channel
	if e.Type.Has(gumble.UserChangeChannel) {
		removeTemporaryChannels()

		// check for nil because when a user connects
		// stats are always nil
		if e.User.Stats != nil {
			log.Println(fmt.Sprintf("[Channel]: User[Name: %s ID: %s Hash: %s IP: %s] is now at: %s",
				e.User.Name, UserIdToString(e.User.UserID), e.User.Hash,
				e.User.Stats.IP.String(), e.User.Channel.Name))

			id := e.User.Channel.Name
			if !e.User.Channel.IsRoot() && !e.User.Channel.Parent.IsRoot() {
				id = e.User.Channel.Parent.Name + "_" + e.User.Channel.Name
			}

			if c, ok := Channels[id]; !e.User.Channel.IsRoot() && ok {
				u := NewUser()
				u.CopyInfo(e.User)

				ok, err := c.IsUserAllowed(u)
				if err != nil {
					log.Println(err)
				}

				if !ok {
					e.User.Move(e.User.Channel.Parent)

					log.Println(fmt.Sprintf("[BOT]: Moved User[%s:%s] to parent channel",
						e.User.Name, UserIdToString(e.User.UserID)))
				}
			}
		} else {
			if !e.User.Channel.IsRoot() {
				e.User.Move(rc)

				log.Println(fmt.Sprintf("[BOT]: Moved User[%s:%s] to root channel",
					e.User.Name, UserIdToString(e.User.UserID)))
			}
		}
	}

	if e.Type.Has(gumble.UserChangeComment) {
		log.Println("User changed comment")
	}

	if e.Type.Has(gumble.UserChangeAudio) {
		log.Println("User changed audio")

		u := NewUser()
		u.CopyInfo(e.User)
		muted, deafened, err := u.IsMD()

		if err != nil {
			log.Println(err)
		}

		log.Println("[MD]: " + strconv.FormatBool(muted) + " - " + strconv.FormatBool(deafened))
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
		log.Println(fmt.Sprintf("[UserChangeStats]: User[%s @ %s, Hash: %s]",
			e.User.Name, e.User.Stats.IP.String(), e.User.Hash))
	}
}

func onDisconnect(e *gumble.DisconnectEvent, m *Mumble) {
	log.Println("[BOT]: Disconnected -> " + e.String)

	// just to make sure the bot reconnects
	ticker := time.NewTicker(5 * time.Second)
	for _ = range ticker.C {
		if m.Client.Conn == nil {
			err := m.Connect()
			if err != nil {
				log.Println(err)
			}

		} else {
			ticker.Stop()
			break
		}
	}

	//m.KeepAlive <- true
}

func onPermissionDenied(e *gumble.PermissionDeniedEvent) {
	var dType string

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
		dType = "I dont know what the denied permission is and at this point im too afraid to ask"
	}

	log.Println("[BOT]: Permission Denied -> " + dType)
	log.Println(e.Permission)
}
