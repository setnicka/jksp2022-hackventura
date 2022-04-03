package main

import (
	"net/http"
	"time"

	"github.com/coreos/go-log/log"
	"github.com/go-chi/chi"
)

const (
	fbiLogin    = "agentjones"
	fbiPassword = "restart"
)

func fbiRouter() *chi.Mux {
	r := newGameRouter("fbi")
	r.Get("/", auth(fbiIndexGet))
	r.Post("/", auth(fbiIndexPost))
	r.Get("/private", auth(fbiInternalGet))
	return r
}

func fbiIndexGet(w http.ResponseWriter, r *http.Request) {
	team := server.state.GetTeam(getUser(r))
	if team != nil && team.FBI.Completed {
		http.Redirect(w, r, "/private", http.StatusSeeOther)
		return
	}

	data := getGeneralData("FBI", w, r)
	executeTemplate(w, "fbiIndex", data)
}

func fbiIndexPost(w http.ResponseWriter, r *http.Request) {
	defer http.Redirect(w, r, "/", http.StatusSeeOther)
	if err := r.ParseForm(); err != nil {
		setFlashMessage(w, r, messageError, "Cannot parse login form")
		return
	}

	server.state.Lock()
	defer server.state.Unlock()

	team := server.state.GetTeam(getUser(r))
	team.FBI.Tries++
	defer server.state.Save()

	login := r.PostFormValue("login")
	password := r.PostFormValue("pass")
	log.Infof("[FBI - %s] Trying login '%s' and password '%s'", team.Login, login, password)

	if login != fbiLogin {
		setFlashMessage(w, r, messageError, "Incorrect login name, only agents can log in")
		return
	}

	if password != fbiPassword {
		setFlashMessage(w, r, messageError, "Incorrect password")
		return
	}

	log.Infof("[FBI - %s] Completed", team.Login)
	// Everything completed
	team.FBI.Completed = true
	team.FBI.CompletedTime = time.Now()
	http.Redirect(w, r, "/private", http.StatusSeeOther)
}

func fbiInternalGet(w http.ResponseWriter, r *http.Request) {
	team := server.state.GetTeam(getUser(r))
	if team == nil || !team.FBI.Completed {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := getGeneralData("FBI", w, r)
	data.MessageType = "success"
	data.Message = "Access Granted"
	executeTemplate(w, "fbiIndex", data)
}
