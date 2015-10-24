package mumble

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/layeh/gumble/gumble"
)

// this is needed because
// gumble.User creates a infinite reference
// so it crashes the RPC/HTTP Server
// when it is returned by some function
//
// this is caused by those fields:
// "Stats" @ gumble.UserStats
// "Channel" @ gumble.User
type User struct {
	IP   string
	Name string

	Hash   string // mumble cert hash
	UserID int    // mumble user id

	Team Team // will be used in a lobby
}

// teams
type Team string

const (
	BLU Team = "BLU"
	RED Team = "RED"
)

func NewUser() *User {
	u := new(User)
	return u
}

// finds a user with minimal info (name, IP, hash or userID) and returns
// the same "User" struct with info from "gumble.User"
func (u *User) Find() (*User, error) {
	if u.IP != "" && u.Name != "" && u.Hash != "" && u.UserID != 0 {
		nu := new(User)
		return nu, errors.New("Provide some info to find the user!")
	}

	// gets the "gumble.User" from the "User"
	gUser, err := u.Gumble()
	if err != nil {
		return new(User), err
	}

	r := NewUser()
	r.CopyInfo(gUser) // copy gumble.User's info

	return r, nil
}

// will be used in all FindUserBy"IP,Name..."
func (u *User) CopyInfo(c *gumble.User) {
	u.Name = c.Name
	u.Hash = c.Hash
	u.UserID = int(c.UserID)

	if c.Stats != nil {
		u.IP = c.Stats.IP.String()
	}
}

// get the user from gumble by using User's info
func (u *User) Gumble() (*gumble.User, error) {
	findName := (u.Name != "")
	findHash := (u.Hash != "")
	findID := (u.UserID != 0)
	findIP := (u.IP != "")

	if findIP && findName && findID && findHash {
		return new(gumble.User), errors.New("Provide some info to find the user!")
	}

	uid := uint32(u.UserID)
	for _, v := range M.Client.Users {
		if findID && v.UserID == uid {
			log.Println("[User@Gumble]: Found user by UserID[" + strconv.Itoa(u.UserID) + "]")
			return v, nil
		}

		if findHash && v.Hash == u.Hash {
			log.Println("[User@Gumble]: Found user by hash[" + v.Hash + "]")
			return v, nil
		}

		if findName && v.Name == u.Name {
			log.Println("[User@Gumble]: Found user by name[" + v.Name + "]")
			return v, nil
		}

		if findIP && !findName &&
			v.Stats != nil && v.Stats.IP.String() == u.IP {
			log.Println("[User@Gumble]: Found user by IP[" + v.Stats.IP.String() + "]")
			return v, nil
		}
	}

	nu := new(gumble.User)
	return nu, nil
}

// moves a user to the given channel
func (u *User) Move(c *Channel) error {
	// does nothing when user is
	// not connecTed
	if !u.IsConnected() {
		return nil
	}

	gu, err := u.Gumble()
	if err != nil {
		return err
	}

	mc := new(gumble.Channel)

	// check if the channel is nil
	// then tries to find it by it's name
	// returns root channel when the channel
	// is not found
	if c.Exists() {
		mc, err = c.Gumble()

		if err != nil {
			return err
		}
	} else {
		return errors.New("Channel not found")
	}

	if mc != nil {
		gu.Move(mc)
	}
	return nil
}

// check if both users contains same data
func (u *User) Equals(a *User) bool {
	return (u.Name == a.Name &&
		u.IP == a.IP &&
		u.Hash == a.Hash &&
		u.UserID == a.UserID)
}

func (u *User) IsConnected() bool {
	// "by" is string so we can use for logging
	// the type used to check if the user is connected
	val := ""
	by := ""
	r := false

	// according to http://mumble.sourceforge.net/slice/1.2.7/Murmur/User.html#userid
	// userid of a anonymous user should be -1 but it's 0 in go
	// maybe because the initial value of int is 0
	//
	// only registered users have a id thats not 0
	if u.UserID != 0 {
		val = strconv.Itoa(u.UserID)
		by = "UserID"

		// only users with a certificate have a hash
	} else if u.Hash != "" {
		val = u.Hash
		by = "Hash"

		// registered users have a fixed username
	} else if u.Name != "" {
		val = u.Name
		by = "Name"

		// ip is last because there can be
		// more than 1 user with the same ip
	} else if u.IP != "" {
		val = u.IP
		by = "IP"
	}

	// check if any info was specified
	if by == "" {
		return false
		//return errors.New("Provide a Name, IP, UserID or Hash!")
	}

	// loop through all connecTed users
	for _, v := range M.Client.Users {
		// we'll check in order from
		// more important to less important
		if (by == "UserID" && v.UserID == uint32(u.UserID)) || // UserID
			(by == "Hash" && v.Hash == u.Hash) || // Hash
			(by == "Name" && v.Name == u.Name) || // Name
			(by == "IP" && v.Stats.IP.String() == u.IP) { // IP
			r = true
		}
	}

	log.Println(fmt.Sprintf("[Fumble]: IsUserConnected? [Type: %s, Value: %s, Result: %s]",
		by, val, strconv.FormatBool(r)))

	return r
}

// check if user is muted and deafened
func (u *User) IsMD() (bool, bool, error) {
	gc, err := u.Gumble()

	if err != nil {
		return false, false, err
	}

	log.Println("[MD]: " + gc.Name + " = M:" + strconv.FormatBool(gc.SelfMuted) + " D:" + strconv.FormatBool(gc.SelfDeafened))

	return gc.SelfMuted, gc.SelfDeafened, nil
}

func (u *User) Kick(reason string) error {
	if u.IsConnected() {
		gb, err := u.Gumble()
		if err != nil {
			return err
		}

		gb.Kick(reason)
	}

	return nil
}

func (u *User) Ban(reason string) error {
	if u.IsConnected() {
		gb, err := u.Gumble()
		if err != nil {
			return err
		}

		gb.Ban(reason)
	} else {
		// TODO: manual ban
		return errors.New("User is not online")
	}

	return nil
}
