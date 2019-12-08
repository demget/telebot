package tbutil

import (
	"sync"
)

// Users contains all cached users.
var Users = users{
	data: make(map[int]*User),
}

// User represents cached user info.
type User struct {
	ID           int
	CallbackData string
	QueryChan    chan struct{}
}

type users struct {
	sync.RWMutex
	data map[int]*User
}

func (u *users) Get(id int) (*User, bool) {
	u.RLock()
	user, ok := u.data[id]
	u.RUnlock()
	return user, ok
}

func (u *users) Set(id int) *User {
	user := &User{ID: id}
	u.Lock()
	u.data[id] = user
	u.Unlock()
	return user
}
