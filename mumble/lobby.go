package mumble

import (
	"log"
	"strconv"
	"time"
)

type Lobby struct {
	ID   uint
	Name string

	Players map[string]*User // string will be the user name
	Channel *Channel
}

func NewLobby() *Lobby {
	l := new(Lobby)
	l.Channel = NewChannel()
	l.Players = make(map[string]*User)

	return l
}

// creates a channel with a child channels for each team
// then allow players to join team's channel
func (l *Lobby) Create() error {
	l.Channel.Name = strconv.FormatUint(uint64(l.ID), 10)
	l.Name = l.Channel.Name

	// red team
	red := NewChannel()
	red.Name = "RED"
	red.Parent = l.Channel.Name
	l.Channel.Children[RED] = red

	// blu team
	blu := NewChannel()
	blu.Name = "BLU"
	blu.Parent = l.Channel.Name
	l.Channel.Children[BLU] = blu

	// allow players to join team's channel
	for _, u := range l.Players {
		if u.Team == BLU || u.Team == RED {
			l.Channel.AllowUser(u)
			l.Channel.Children[u.Team].AllowUser(u)
		}
	}

	l.Channel.Create()
	l.Channel.Children[RED].Create()
	l.Channel.Children[BLU].Create()

	// time to create the channels
	time.Sleep(1 * time.Second)

	for _, u := range l.Players {
		err := u.Move(l.Channel.Children[u.Team])
		if err != nil {
			return err
		}
	}

	return nil
}

func (l *Lobby) AllowPlayer(u *User) {
	// better than loops
	l.Players[u.Name] = u

	l.Channel.AllowUser(u)
	if u.Team != "" {
		l.Channel.Children[u.Team].AllowUser(u)
	}

	// update the Channel variable
	// so the bot allow the player
	// to join the lobby's channel
	l.Channel.update()
	if u.Team != "" {
		u.Move(l.Channel.Children[u.Team])
	}
}

func (l *Lobby) DisallowPlayer(u *User) {
	// remove player from players
	delete(l.Players, u.Name)

	// move player before disallow
	u.Move(GetRootChannel())
	l.Channel.DisallowUser(u)
	l.Channel.Children[u.Team].DisallowUser(u)

	// update the Channel variable
	// so the bot disallow the player
	// from joining the lobby's channel
	l.Channel.update()
	log.Println(l.Players)
}

func (l *Lobby) End() error {
	// move all players to
	// the main lobby's channel
	for _, u := range l.Players {
		if u.IsConnected() {
			err := u.Move(l.Channel)

			if err != nil {
				return err
			}
		}
	}

	// remove red channel
	errR := l.Channel.Children[RED].Remove()
	if errR != nil {
		return errR
	}

	// remove blu channel
	errB := l.Channel.Children[BLU].Remove()
	if errB != nil {
		return errB
	}

	// remove channel when all users leave
	l.Channel.Temporary = true
	l.Channel.update()

	return nil
}
