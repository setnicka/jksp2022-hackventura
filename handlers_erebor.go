package main

import (
	"net/http"
	"time"

	"github.com/coreos/go-log/log"
	"github.com/go-chi/chi"
)

const (
	ereborLogin    = "smacek"
	ereborPassword = "zlatomismackuje"
)

func ereborRouter() *chi.Mux {
	r := newGameRouter("erebor")
	r.Get("/", auth(ereborIndexGet))
	r.Post("/", auth(ereborIndexPost))
	r.Get("/tajne", auth(ereborInternalGet))
	return r
}

func ereborIndexGet(w http.ResponseWriter, r *http.Request) {
	team := server.state.GetTeam(getUser(r))
	if team != nil && team.Erebor.Completed {
		http.Redirect(w, r, "/tajne", http.StatusSeeOther)
		return
	}

	data := getGeneralData("Erebor", w, r)
	executeTemplate(w, "ereborIndex", data)
}

func ereborIndexPost(w http.ResponseWriter, r *http.Request) {
	defer http.Redirect(w, r, "/", http.StatusSeeOther)
	if err := r.ParseForm(); err != nil {
		setFlashMessage(w, r, messageError, "Cannot parse login form")
		return
	}

	server.state.Lock()
	defer server.state.Unlock()

	team := server.state.GetTeam(getUser(r))
	team.Erebor.Tries++
	defer server.state.Save()

	login := r.PostFormValue("login")
	password := r.PostFormValue("password")
	log.Infof("[Erebor - %s] Trying login '%s' and password '%s'", team.Login, login, password)

	if login != ereborLogin || password != ereborPassword {
		setFlashMessage(w, r, messageError, "Nesprávný login, zkuste to znovu nebo kontaktujte podporu")
		return
	}

	log.Infof("[Erebor - %s] Completed", team.Login)
	// Everything completed
	team.Erebor.Completed = true
	team.Erebor.CompletedTime = time.Now()
	http.Redirect(w, r, "/tajne", http.StatusSeeOther)
}

func ereborInternalGet(w http.ResponseWriter, r *http.Request) {
	team := server.state.GetTeam(getUser(r))
	if team == nil || !team.Erebor.Completed {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := getGeneralData("Erebor", w, r)
	data.MessageType = "success"
	data.Message = "Přístup povolen"
	executeTemplate(w, "ereborIndex", data)
}
