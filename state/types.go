package state

import (
	"sync"
	"time"
)

// State of the game
type State struct {
	Teams  []*Team
	Global globalState
	sync.RWMutex
	Version int // state version, increased by each save
}

type globalState struct {
	DolGuldur []DolGuldurPassword
}

type DolGuldurPassword struct {
	Password string
	From     time.Time
}

// Team holds teams login inforation and game state for it
type Team struct {
	Name   string
	Login  string
	Passwd []byte // hashed by bcrypt
}

type result struct {
	Completed     bool
	CompletedTime time.Time
	Tries         int
}
