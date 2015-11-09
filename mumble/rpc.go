package mumble

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/layeh/gumble/gumble"
)

type Fumble int
type NoReply struct{}

type Args struct {
	User    *User
	Channel *Channel
}

type LobbyArgs struct {
	User  *User
	Lobby *Lobby
}

type KickArgs struct {
	User   *User
	Reason string
}

// muted or deafened
type MD struct {
	Muted    bool
	Deafened bool
}

///////////////
/// CHANNEL ///
///////////////

func (_ *Fumble) CreateChannel(name string, noreply *NoReply) error {
	c := NewChannel()
	c.Name = name
	c.Create()

	if M.Client.Channels.Find(name) == nil {
		return errors.New("The channel[" + name + "] was not created!")
	} else {
		log.Println("[Fumble]: Created channel: " + name)
	}

	return nil
}

func (_ *Fumble) RemoveChannel(name string, noreply *NoReply) error {
	if M.Client.Channels.Find(name) == nil || Channels[name] == nil {
		return errors.New("The channel doesn't exists!")
	}

	Channels[name].Remove()

	if M.Client.Channels.Find(name) != nil {
		return errors.New("The channel[" + name + "] was not removed!")
	} else {
		log.Println("[Fumble]: Removed channel: " + name)
	}

	return nil
}

////////////
/// USER ///
////////////

func (_ *Fumble) IsUserConnected(user *User, reply *bool) error {
	// "by" is string so we can use for logging
	// the type used to check if the user is connected
	val := ""
	by := ""
	r := false

	// according to http://mumble.sourceforge.net/slice/1.2.7/Murmur/User.html#userid
	// userid of a anonymous user should be -1
	// but for some reason it's 0 in go
	// maybe because the initial value of int is 0
	//
	// only registered users have a id thats not 0
	if user.UserID != 0 {
		val = strconv.Itoa(user.UserID)
		by = "UserID"

		// only users with a certificate have a hash
	} else if user.Hash != "" {
		val = user.Hash
		by = "Hash"

		// registered users have a fixed username
	} else if user.Name != "" {
		val = user.Name
		by = "Name"

		// ip is last because there can be
		// more than 1 user with the same ip
	} else if user.IP != "" {
		val = user.IP
		by = "IP"
	}

	// check if any info was specified
	if by == "" {
		return errors.New("Provide a Name, IP, UserID or Hash!")
	}

	// loop through all connecTed users
	for _, v := range M.Client.Users {
		// we'll check in order from
		// more important to less important
		if (by == "UserID" && v.UserID == uint32(user.UserID)) || // UserID
			(by == "Hash" && v.Hash == user.Hash) || // Hash
			(by == "Name" && v.Name == user.Name) || // Name
			(by == "IP" && v.Stats.IP.String() == user.IP) { // IP
			r = true
			break
		}
	}

	log.Println(fmt.Sprintf("[Fumble]: IsUserConnected? [Type: %s, Value: %s, Result: %s]",
		by, val, strconv.FormatBool(r)))

	*reply = r
	return nil
}

func (_ *Fumble) FindUserByInfo(u *User, reply *User) error {
	r, err := u.Find()

	if err != nil {
		return err
	}

	*reply = *r
	return nil
}

func (_ *Fumble) FindUserByIP(ip string, reply *User) error {
	if ip == "" {
		return errors.New("Provide a IP!")
	}

	r := NewUser()
	for _, v := range M.Client.Users {
		if v.Stats != nil && v.Stats.IP.String() == ip {
			r.CopyInfo(v)
			break
		}
	}

	*reply = *r
	return nil
}

func (_ *Fumble) FindUserByHash(hash string, reply *User) error {
	if hash == "" {
		return errors.New("Provide a Hash!")
	}

	r := NewUser()
	for _, v := range M.Client.Users {
		if v.Hash == hash {
			r.CopyInfo(v)
			break
		}
	}

	*reply = *r
	return nil
}

func (_ *Fumble) FindUserByID(id int, reply *User) error {
	if id == 0 {
		return errors.New("Provide a UserID that isn't 0!")
	}

	uid := uint32(id)
	r := NewUser()

	for _, v := range M.Client.Users {
		if v.UserID == uid {
			r.CopyInfo(v)
			break
		}
	}

	*reply = *r
	return nil
}

func (_ *Fumble) FindUserByName(name string, reply *User) error {
	if name == "" {
		return errors.New("Provide a Name!")
	}

	r := NewUser()
	for _, v := range M.Client.Users {
		if v.Name == name {
			r.CopyInfo(v)
			break
		}
	}

	*reply = *r
	return nil
}

func (f *Fumble) IsUserRegistered(u *User, reply *bool) error {
	var r User
	// get better info about the user
	f.FindUserByInfo(u, &r)

	gUser, err := r.Gumble()
	if err != nil {
		return err
	}

	is := gUser.IsRegistered()
	*reply = is
	return nil
}

func (_ *Fumble) AllowUser(args *Args, noreply *NoReply) error {
	if M.Client.Channels.Find(args.Channel.Name) == nil || Channels[args.Channel.Name] == nil {
		return errors.New("The channel doesn't exists!")
	}

	Channels[args.Channel.Name].AllowUser(args.User)

	ok, err := Channels[args.Channel.Name].IsUserAllowed(args.User)
	if err != nil {
		return err
	}

	if !ok {
		for _, u := range Channels[args.Channel.Name].Allowed {
			log.Println(u)
		}

		return errors.New("Cannot allow user")
	}

	return nil
}

func (_ *Fumble) DisallowUser(args *Args, noreply *NoReply) error {
	if !args.Channel.Exists() {
		return errors.New("The channel doesn't exists!")
	}

	Channels[args.Channel.Name].DisallowUser(args.User)

	ok, err := Channels[args.Channel.Name].IsUserAllowed(args.User)
	if err != nil {
		return err
	}

	if ok {
		return errors.New("Cannot disallow user")
	}

	time.Sleep(1 * time.Second)

	return nil
}

// checks if user is muted or deafened
func (_ *Fumble) IsUserMD(u *User, reply *MD) error {
	muted, deafened, err := u.IsMD()

	if err != nil {
		return err
	}

	md := new(MD)
	md.Muted = muted
	md.Deafened = deafened

	*reply = *md
	return nil
}

// bans a user
func (_ *Fumble) BanUser(k *KickArgs, noreply *NoReply) error {
	err := k.User.Ban(k.Reason)
	if err != nil {
		return err
	}

	return nil
}

// kicks a user
func (_ *Fumble) KickUser(k *KickArgs, noreply *NoReply) error {
	err := k.User.Kick(k.Reason)
	if err != nil {
		return err
	}

	return nil
}

/////////////
/// LOBBY ///
/////////////

func (_ *Fumble) CreateLobby(l *Lobby, reply *Lobby) error {
	err := l.Create()

	if err != nil {
		return err
	}

	*reply = *l
	return nil
}

func (_ *Fumble) AllowPlayer(la *LobbyArgs, reply *Lobby) error {
	la.Lobby.AllowPlayer(la.User)

	*reply = *la.Lobby
	return nil
}

func (_ *Fumble) DisallowPlayer(la *LobbyArgs, reply *Lobby) error {
	la.Lobby.DisallowPlayer(la.User)

	log.Println(la.Lobby.Players)

	*reply = *la.Lobby

	log.Println(la.Lobby.Players)
	return nil
}

func (_ *Fumble) EndLobby(l *Lobby, reply *Lobby) error {
	err := l.End()

	if err != nil {
		return err
	}

	*reply = *l
	return nil
}

type UserInfo struct {
	User *User
	Team string
}

type LobbyArgsTeam struct {
	User  *User
	Lobby *Lobby
	Team  string
}

var WhitelistedUsersLock = new(sync.RWMutex)
var WhitelistedUsers = make(map[int][]UserInfo)

func (_ *Fumble) AddNameToLobbyWhitelist(l LobbyArgs, reply *NoReply) error {
	if l.User.IsConnected() {
		SimplyAllowUser(l.Lobby.ID, UserInfo{l.User, ""})
		return nil
	}

	WhitelistedUsersLock.Lock()
	defer WhitelistedUsersLock.Unlock()

	if _, ok := WhitelistedUsers[l.Lobby.ID]; !ok {
		WhitelistedUsers[l.Lobby.ID] = make([]UserInfo, 0)
	}
	WhitelistedUsers[l.Lobby.ID] = append(WhitelistedUsers[l.Lobby.ID], UserInfo{l.User, ""})
	return nil
}

func (_ *Fumble) AddNameToLobbyWhitelistTeam(l LobbyArgsTeam, reply *NoReply) error {
	if l.User.IsConnected() {
		SimplyAllowUser(l.Lobby.ID, UserInfo{l.User, l.Team})
		return nil
	}

	WhitelistedUsersLock.Lock()
	defer WhitelistedUsersLock.Unlock()

	if _, ok := WhitelistedUsers[l.Lobby.ID]; !ok {
		WhitelistedUsers[l.Lobby.ID] = make([]UserInfo, 0)
	}
	WhitelistedUsers[l.Lobby.ID] = append(WhitelistedUsers[l.Lobby.ID], UserInfo{l.User, l.Team})
	return nil
}

func SimplyAllowUser(lobbyId int, userInfo UserInfo) {
	var channel *Channel
	if userInfo.Team == "" {
		channel = Channels[strconv.Itoa(lobbyId)]
	} else {
		channel = Channels[strconv.Itoa(lobbyId)+"_"+userInfo.Team]
		Channels[channel.Parent].AllowUser(userInfo.User)
	}

	channel.AllowUser(userInfo.User)
	userInfo.User.Move(channel)

	log.Println(fmt.Sprintf("[Fumble]: Allowing User? [Lobby: %d, Name: %s]", lobbyId, userInfo.User.Name))
}

func CheckWhitelist(guser *gumble.User) {
	WhitelistedUsersLock.Lock()
	defer WhitelistedUsersLock.Unlock()

	for lobbyId, users := range WhitelistedUsers {
		toremove := -1
		for i, userInfo := range users {
			if userInfo.User.Name == guser.Name {
				SimplyAllowUser(lobbyId, userInfo)

				toremove = i
				break
			}
		}
		if toremove >= 0 {
			users[toremove] = users[len(users)-1]
			users = users[:len(users)-1]
		} else {
			log.Println(fmt.Sprintf("[Fumble]: User not in whitelist. [Lobby: %d, Name: %s]", lobbyId, guser.Name))
		}
	}
}
