package state

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/coreos/go-log/log"
	"github.com/pkg/errors"
)

const (
	logsDir       = "logs"
	stateFilename = "state.json"
)

// Init loads (or creates a new) game state and returns it. It is a caller
// responsibility to correctly use the state.Lock() and state.Rlock()
func Init() *State {
	log.Debug("Initializing game state")
	state := &State{
		Teams: []*Team{},
	}

	// Try to load previously saved state
	jsonFile, err := os.Open(stateFilename)
	if err == nil {
		defer jsonFile.Close()

		log.Debugf("Loading state from file '%s'", stateFilename)
		jsonBytes, _ := ioutil.ReadAll(jsonFile)
		if err = json.Unmarshal(jsonBytes, &state); err != nil {
			log.Errorf("Problem during loading state from file: %v", err)
		} else {
			log.Debug("State loaded")
		}
	}

	return state
}

// Save current state to the file
func (s *State) Save() error {
	log.Debug("Saving actual state into file")
	s.Version++
	// 1. If exists current state move it into folder
	if _, err := os.Stat(stateFilename); err == nil {
		if err := os.MkdirAll(logsDir, os.ModePerm); err != nil {
			return errors.Wrapf(err, "Cannot create directory '%s' for logging old states", logsDir)
		}
		filename := fmt.Sprintf("%s%s", stateFilename, time.Now().Format(".150405.00")) // 2006-01-02_150405
		if err := os.Rename(stateFilename, path.Join(logsDir, filename)); err != nil {
			return errors.Wrapf(err, "Cannot move state to logs directory", logsDir)
		}
	}

	// 2. Marshal state into json
	bytes, err := json.MarshalIndent(s, "", "\t")
	if err != nil {
		return errors.Wrapf(err, "Cannot save actual state into json")
	}

	if err := ioutil.WriteFile(stateFilename, bytes, 0644); err != nil {
		return errors.Wrapf(err, "Cannot save JSON of the actual state into file '%s'", stateFilename)
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// global state handlers

// SetDolGuldurPassword appends a new password for Dol Guldur (and save the state)
func (s *State) SetDolGuldurPassword(password string) error {
	s.Global.DolGuldur = append(s.Global.DolGuldur, DolGuldurPassword{Password: password, From: time.Now()})
	return s.Save()
}

// GetDolGuldurPasswords returns all passwords for Dol Guldur, valid one is the
// last one
func (s *State) GetDolGuldurPasswords() []DolGuldurPassword {
	return s.Global.DolGuldur
}
