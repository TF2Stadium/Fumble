package mumble

import (
	"errors"
	//"log"
	"time"

	"github.com/layeh/gumble/gumble"
)

var Channels = make(map[string]*Channel)

type Channel struct {
	Name      string
	Temporary bool

	// i choose string to avoid loops
	Allowed  map[string]*User
	Children map[Team]*Channel
	Parent   string // just the name to access from the child
}

func GetRootGChannel() *gumble.Channel {
	for _, c := range M.Client.Channels {
		if c.IsRoot() {
			return c
		}
	}

	return nil
}

func GetRootChannel() *Channel {
	for _, c := range M.Client.Channels {
		if c.IsRoot() {
			cc := NewChannel()
			cc.Name = c.Name
			return cc
		}
	}

	return nil
}

func NewChannel() *Channel {
	c := new(Channel)

	c.Allowed = make(map[string]*User)
	c.Children = make(map[Team]*Channel)
	c.Temporary = false

	return c
}

// creates a channel
// a name is required
func (c *Channel) Create() error {
	if c.Name == "" {
		return errors.New("[Channel]: a name is required to created a channel!")
	}

	// checks if channels exists
	// returns error when exists
	if c.Exists() {
		return errors.New("[Channel]: Channel[" + c.Name + "] already exists!")
	} else {

		// if the channel is a child (RED or BLU)
		// it's name will be like:
		// Parent123_RED
		if c.IsChildren() {
			// get Parent's id from gumble and insert
			// the child channel into it's Parent
			id := M.Client.Channels.Find(c.Parent).ID
			M.Client.Channels[id].Add(c.Name, c.Temporary)

			Channels[c.Parent+"_"+c.Name] = c
		} else {
			M.Client.Channels[0].Add(c.Name, c.Temporary)
			Channels[c.Name] = c
		}

		time.Sleep(1 * time.Second)
	}

	// checks if channel was created
	// returns error when not
	if !c.Exists() {
		return errors.New("[Channel]: Channel[" + c.Name + "] was not created!")
	}

	c.update()
	return nil
}

func (c *Channel) Exists() bool {
	isChildren := (c.Name == "RED" || c.Name == "BLU") && c.Parent != ""

	if isChildren {
		rc := M.Client.Channels.Find(c.Parent)

		for _, children := range rc.Children {
			if children.Name == c.Name {
				return true
			}
		}
	}

	return (M.Client.Channels.Find(c.Name) != nil)
}

// removes a channel using it's name
func (c *Channel) Remove() error {
	if c.Name == "" {
		return errors.New("[Channel]: a name is required to remove a channel!")
	}

	var cName string

	if c.IsChildren() {
		rc := M.Client.Channels.Find(c.Parent)

		for _, children := range rc.Children {
			if children.Name == c.Name {
				children.Remove()
				delete(Channels, c.Parent+"_"+c.Name)
				break
			}
		}
	} else {
		cName = c.Name
		rc := M.Client.Channels.Find(cName)

		if rc != nil && !rc.IsRoot() {
			rc.Remove()
			delete(Channels, cName)
		}
	}

	time.Sleep(1 * time.Second)

	// returns error when the
	// channel wasn't removed
	if c.Exists() {
		return errors.New("[Channel]: Channel[" + c.Name + "] was not removed!")
	}

	c.update()

	return nil
}

// lets a user join a channel that
// isn't the root channel
func (c *Channel) AllowUser(u *User) {
	c.Allowed[u.Name] = u
	c.update()
}

// remove a user from the allowed list
// returns error when can't move the user
func (c *Channel) DisallowUser(u *User) error {
	// do nothing if the user
	// is already not allowed
	if ok, err := c.IsUserAllowed(u); err == nil && !ok {
		return nil
	}

	// remove user from allowed list
	delete(c.Allowed, u.Name)

	// move user to root channel
	err := u.Move(GetRootChannel())
	if err != nil {
		return err
	}

	c.update()
	return nil
}

func (c *Channel) IsUserAllowed(u *User) (bool, error) {
	if u.Name == "" {
		return false, errors.New("Name is empty")
	}

	if c.Allowed[u.Name] != nil {
		return true, nil
	}

	return false, nil
}

func (c *Channel) IsEmpty() bool {
	gc, _ := c.Gumble()

	if gc == nil {
		return true
	}

	if c.IsChildren() {
		if len(gc.Users) > 0 {
			return false
		}

		// check for users in children channel
	} else {
		total := 0

		for _, children := range gc.Children {
			total += len(children.Users)
		}

		if total > 0 {
			return false
		}
	}

	return true
}

func (c *Channel) IsChildren() bool {
	return (c.Name == "RED" || c.Name == "BLU") && c.Parent != ""
}

func (c *Channel) Gumble() (*gumble.Channel, error) {
	if c.Name == "" {
		return nil, errors.New("Provide a name to find the gumble channel")
	}

	if c.IsChildren() {
		rc := M.Client.Channels.Find(c.Parent)

		for _, children := range rc.Children {
			if children.Name == c.Name {
				return children, nil
			}
		}
	}

	return M.Client.Channels.Find(c.Name), nil
}

func (c *Channel) update() {
	if c.Parent != "" {
		Channels[c.Parent+"_"+c.Name] = c
	} else {
		Channels[c.Name] = c
	}
}
