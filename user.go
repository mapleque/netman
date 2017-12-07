package netman

import (
	"errors"
	"net"
	"sync"
)

type UserPool struct {
	Pool map[string]*User `json:"pool"`
	mux  *sync.RWMutex
}

type User struct {
	Name      string `json:"name"`
	Ip        net.IP `json:"ip"`
	GroupName string `json:"group_name"`
}

func newUserPool() *UserPool {
	return &UserPool{
		make(map[string]*User),
		new(sync.RWMutex),
	}
}

func newUser(name string) *User {
	return &User{
		Name: name,
	}
}

func (this *UserPool) get(name string) (*User, bool) {
	user, ok := this.Pool[name]
	return user, ok
}

func (this *UserPool) add(user *User) error {
	if _, exist := this.Pool[user.Name]; exist {
		return errors.New("user already exist")
	}
	this.mux.Lock()
	defer this.mux.Unlock()
	this.Pool[user.Name] = user
	return nil
}

func (this *UserPool) del(name string) error {
	if _, exist := this.Pool[name]; !exist {
		return errors.New("user not exist")
	}
	this.mux.Lock()
	defer this.mux.Unlock()
	delete(this.Pool, name)
	return nil
}
