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
	Gundabad []GundabadPassword
}

type GundabadPassword struct {
	Password string
	From     time.Time
}

// Team holds teams login inforation and game state for it
type Team struct {
	Name      string
	Login     string
	Passwd    []byte // hashed by bcrypt
	Moria     ResultMoria
	Gundabad  ResultGundabad
	Roklinka  ResultRoklinka
	Erebor    ResultErebor
	Dunharrow ResultDunharrow
	HZP       ResultHZP
	Village		ResultVillage
}

type result struct {
	Completed     bool
	CompletedTime time.Time
	Tries         int
}

// Results for different hack pages
type ResultMoria struct {
	result
	RightAnswers int
}
type ResultGundabad result
type ResultRoklinka result
type ResultErebor result
type ResultHZP result
type ResultVillage result
type ResultDunharrow struct {
	result
	LastTry time.Time
}
