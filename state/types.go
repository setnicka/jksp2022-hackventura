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
	CIA    ResultCIA
	NSA    ResultNSA
	KGB    ResultKGB
	FBI    ResultFBI
	PPL    ResultPPL
	BIS    ResultBIS
	MI5    ResultMI5
}

type result struct {
	Completed     bool
	CompletedTime time.Time
	Tries         int
}

// Results for different hack pages

type ResultCIA struct {
	result
	LastTry time.Time
}

type ResultNSA result
type ResultKGB result
type ResultFBI result
type ResultPPL struct {
	result
	PackagesTracked int
}
type ResultBIS result
type ResultMI5 result
